package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	dbuser     = "guest"
	dbpassword = "guest"
	dbname     = "chat-app"
	dbhost     = "127.0.0.1"
	dbport     = "5432"
)

type dbQuery struct {
	Query                   string
	ExpectSingleRow         bool                 //true for single row, false if not
	ReturnChan              chan [][]interface{} // Generic response channel
	NumberOfColumnsExpected int                  //0 for no sql data returned
}

type dbConnector struct {
	err error
}

var dbRequestChan = make(chan dbQuery, 10) // Buffered for efficient queuing

// manages the dbConnectors
func dbManager() {
	dbConnectorChan := make(chan *dbConnector, 1)
	wk := &dbConnector{err: nil}
	go wk.work(dbConnectorChan, dbRequestChan)

	for wk := range dbConnectorChan {
		log.Printf("DBConnector stopped with err: %s", wk.err)
		// reset err
		wk.err = nil
		// a goroutine has ended, restart it
		go wk.work(dbConnectorChan, dbRequestChan)
	}
}

// manages the database connection and calls the query workers
func (wk *dbConnector) work(dbConnectorChan chan<- *dbConnector, dbchan <-chan dbQuery) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				wk.err = err
			} else {
				wk.err = fmt.Errorf("panic happened with %v", r)
			}
		} else {
			wk.err = err
		}
		dbConnectorChan <- wk
	}()

	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbhost, dbport, dbuser, dbpassword, dbname)) //consider secure password handling
	if err != nil {
		return err
	}
	defer db.Close()

	for msg := range dbchan {
		var sendResults [][]interface{}
		fmt.Println("DB Query Running:", msg)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//returns error if there is an error or bool true if all ok.
		//it doesn't return an error or notify if no rows are changed. TBD
		if msg.NumberOfColumnsExpected == 0 {
			_, err := db.ExecContext(ctx, msg.Query)
			var singleResult []interface{}
			if err != nil {
				singleResult = append(singleResult, err)
			} else {
				singleResult = append(singleResult, true)
			}
			sendResults = append(sendResults, singleResult)
			fmt.Println("DB Query Response:", sendResults)
			msg.ReturnChan <- sendResults
			close(msg.ReturnChan)
			continue
		}

		// Choose the appropriate method based on whether you expect a single value:
		switch msg.ExpectSingleRow {
		case true:
			columns := make([]interface{}, msg.NumberOfColumnsExpected)
			// Create a slice of pointers to the elements in the columns slice
			columnPointers := make([]interface{}, len(columns))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			err := db.QueryRowContext(ctx, msg.Query).Scan(columnPointers...)
			if err != nil {
				fmt.Println("error querying database:", err)
				columns[0] = err
			}

			sendResults = append(sendResults, columns)
			msg.ReturnChan <- sendResults
			close(msg.ReturnChan)
		case false:

			rows, err := db.QueryContext(ctx, msg.Query)
			if err != nil {
				fmt.Println("error querying database:", err)
				continue // Skipping further processing for this query
			}
			// Process rows and fill the response interface as needed
			defer rows.Close()
			msg.ReturnChan <- sendResults
			close(msg.ReturnChan)
		}
	}
	return err
}

type ExternalUserInfo struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	IPAddr    string `json:"ipaddr"`
	EmailAddr string `json:"email,omitempty"`
}

func singleQuote(s string) string {
	return "'" + s + "'"
}

type idResponse struct {
	ID int64
}

// responds with user information of added user
func addExternalUser(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	respondJson(&w)

	var eui *ExternalUserInfo
	err := json.NewDecoder(r.Body).Decode(&eui)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if eui.Name == "" || eui.IPAddr == "" {
		http.Error(w, errors.New("name and IP address must be provided").Error(), http.StatusBadRequest)
		return
	}
	log.Println("Add external User Request:", eui)

	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT add_external_user(%s, %s)",
			singleQuote(doubleUpSingleQuotes(eui.Name)),
			singleQuote(eui.IPAddr)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 1,
		ExpectSingleRow:         true,
	}

	resp, err := dbq.processRequest(w, "addExternalUser")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	structToVerify := idResponse{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}
	eui.ID = structToVerify.ID
	err = json.NewEncoder(w).Encode(eui)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

// takes a slice of fields provided by a db query and a struct as an interface and converts to the requested struct as an interface.
func convertSliceToStruct(sl []interface{}, str interface{}) error {
	//TODO: add checking for this function to ensure that the length of each interface is correct.

	strValue := reflect.ValueOf(str)
	if strValue.Kind() == reflect.Ptr && strValue.Elem().Kind() == reflect.Struct {
		// Iterate over the fields of the struct and set values
		for i := 0; i < strValue.Elem().NumField(); i++ {

			//skip if the slice index interface is nil
			if checkNil(sl[i]) {
				continue
			}

			field := strValue.Elem().Field(i)
			if field.CanSet() {
				// Set a value based on the field type
				switch field.Kind() {
				case reflect.String:
					if i < len(sl) && reflect.TypeOf(sl[i]).Kind() == reflect.String {
						field.SetString(sl[i].(string))
					}
				case reflect.Int:
					if i < len(sl) && reflect.TypeOf(sl[i]).Kind() == reflect.Int64 { //all structs should only use int64
						field.SetInt(sl[i].(int64))
					}
				case reflect.Int64:
					if i < len(sl) && reflect.TypeOf(sl[i]).Kind() == reflect.Int64 {
						field.SetInt(sl[i].(int64))
					}
				case reflect.Bool:
					if i < len(sl) && reflect.TypeOf(sl[i]).Kind() == reflect.Bool {
						field.SetBool(sl[i].(bool))
					}
				}
			}
		}

		return nil
	} else {
		return fmt.Errorf("%v is not a pointer to a struct", str)
	}
}

// responds with user information of added user, if exists
func getExternalUser(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	respondJson(&w)

	var eui *ExternalUserInfo
	err := json.NewDecoder(r.Body).Decode(&eui)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if eui.Name == "" || eui.IPAddr == "" {
		http.Error(w, errors.New("name and IP address must be provided").Error(), http.StatusBadRequest)
		return
	}
	log.Println("Get external User Request:", eui)

	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT user_id, name, ip_address, email FROM external_users WHERE name = %s AND ip_address = %s",
			singleQuote(doubleUpSingleQuotes(eui.Name)),
			singleQuote(eui.IPAddr)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         true,
	}

	resp, err := dbq.processRequest(w, "getExternalUser")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	structToVerify := ExternalUserInfo{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}

	eui.ID, eui.EmailAddr = structToVerify.ID, structToVerify.EmailAddr

	err = json.NewEncoder(w).Encode(eui)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}

}

func getExternalUserByID(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	respondJson(&w)

	idString := mux.Vars(r)["id"]
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eui := &ExternalUserInfo{}
	eui.ID = int64(idInt)

	log.Println("Get external User By ID Request:", eui.ID)

	dbq := dbQuery{
		Query:                   fmt.Sprintf("SELECT user_id, name, ip_address, email FROM external_users WHERE user_id = %d", eui.ID),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         true,
	}

	resp, err := dbq.processRequest(w, "getExternalUserByID")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	structToVerify := ExternalUserInfo{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}

	eui.Name, eui.IPAddr, eui.EmailAddr = structToVerify.Name, structToVerify.IPAddr, structToVerify.EmailAddr

	err = json.NewEncoder(w).Encode(eui)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

func updateExternalByID(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	respondJson(&w)

	idString := mux.Vars(r)["id"]
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var eui *ExternalUserInfo
	//get any fields to update from body
	err = json.NewDecoder(r.Body).Decode(&eui)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//overwrite id with int provided in the url, incase it's different to that applied in the body
	eui.ID = int64(idInt)

	log.Println("Update external User By ID Request:", eui.ID)

	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT given_user_id, updated_name, updated_ip_address, updated_email FROM update_external_user_info(%d, %s, %s, %s)",
			eui.ID,
			singleQuote(doubleUpSingleQuotes(eui.Name)),
			singleQuote(eui.IPAddr),
			singleQuote(doubleUpSingleQuotes(eui.EmailAddr))),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         true,
	}

	resp, err := dbq.processRequest(w, "updateExternalByID")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	structToVerify := ExternalUserInfo{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}
	if eui.ID != structToVerify.ID {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error verifying db results. Error: %v", err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(structToVerify)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}

}

type UpdateChatTime struct {
	ChatUUID     string `json:"chatuuid"`
	TimeToUpdate string `json:"time"`
}

func updateChatStatus(w http.ResponseWriter, r *http.Request) {

	var uct *UpdateChatTime
	err := json.NewDecoder(r.Body).Decode(&uct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	verifiedTime, err := verifyTimeFormat(uct.TimeToUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Chat Status Request:", uct.ChatUUID)

	var queryStr string
	if r.Method == "POST" {
		queryStr = fmt.Sprintf("INSERT INTO chat (uuid, start_time) VALUES (%s, %s)",
			singleQuote(uct.ChatUUID), singleQuote(verifiedTime))
	} else {
		queryStr = fmt.Sprintf("UPDATE chat SET end_time = %s WHERE uuid = %s",
			singleQuote(verifiedTime), singleQuote(uct.ChatUUID))
	}

	dbq := dbQuery{
		Query:                   queryStr,
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 0,
		ExpectSingleRow:         false,
	}

	_, err = dbq.processRequest(w, "updateChatStatus")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
}

type InternalUserInfo struct {
	ID             int64  `json:"id,omitempty"`
	RoleID         int64  `json:"roleid"`
	FirstName      string `json:"firstname"`
	Surname        string `json:"surname"`
	EmailAddr      string `json:"email,omitempty"`
	HashedPassword string `json:"password"`
}

func addInternalUser(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	respondJson(&w)

	var iui *InternalUserInfo
	err := json.NewDecoder(r.Body).Decode(&iui)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if iui.RoleID == 0 || iui.FirstName == "" || iui.Surname == "" || iui.EmailAddr == "" || iui.HashedPassword == "" {
		http.Error(w, errors.New("the following information needs to be provided: roleid, firstname, surname, email, password").Error(), http.StatusBadRequest)
		return
	}

	log.Println("Add internal user request:", iui)

	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT add_internal_user(%d, %s, %s, %s, %s)",
			iui.RoleID,
			singleQuote(doubleUpSingleQuotes(iui.FirstName)),
			singleQuote(doubleUpSingleQuotes(iui.Surname)),
			singleQuote(doubleUpSingleQuotes(iui.EmailAddr)),
			singleQuote(doubleUpSingleQuotes(iui.HashedPassword))),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 1,
		ExpectSingleRow:         true,
	}

	resp, err := dbq.processRequest(w, "addInternalUser")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	structToVerify := idResponse{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}
	iui.ID = structToVerify.ID
	err = json.NewEncoder(w).Encode(iui)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

func (dbq *dbQuery) processRequest(w http.ResponseWriter, funcCaller string) ([][]interface{}, error) {
	dbRequestChan <- *dbq

	for {
		resp, ok := <-dbq.ReturnChan // read from the channel
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return resp, errors.New("channel closed, no data")
		}

		log.Printf("%s Response: %v", funcCaller, resp)

		if dbq.ExpectSingleRow {
			if len(resp) != 1 { //only expecting a single response
				w.WriteHeader(http.StatusInternalServerError)
				return resp, errors.New("db returned a result which is not correct, please review logs")
			}
		}

		if err, ok := resp[0][0].(error); ok {
			if err.Error() == "sql: no rows in result set" {
				w.WriteHeader(http.StatusNoContent)
				return resp, errors.New("no results")
			}

			w.WriteHeader(http.StatusInternalServerError)
			return resp, fmt.Errorf("error when running database query: %v", err.Error())
		}

		return resp, nil
	}
}

// returns true if interface is nil
func checkNil(myInterface interface{}) bool {
	switch myInterface.(type) {
	case nil:
		return true
	default:
		return false
	}
}

// returns doubled up single quotes. e.g Lay's -> Lay”s. Need to escape for storing in database
func doubleUpSingleQuotes(str string) string {
	var modified string
	for _, char := range str {
		switch char {
		case '\'':
			modified += "''"
		default:
			modified += string(char)
		}
	}
	return modified
}

func verifyTimeFormat(st string) (rs string, err error) {
	layout := "2006-01-02 15:04:05.999999"

	parsedTime, err := time.Parse(layout, st)
	if err != nil {
		return st, fmt.Errorf("time not valid for this application, format needed: %s", layout)
	}
	return parsedTime.Format(layout), nil
}

func respondJson(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func main() {
	go dbManager()
	// Where ORIGIN_ALLOWED is like `scheme://dns[:port]`, or `*` (insecure)
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Accept"})
	//originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	//methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST"})

	r := mux.NewRouter()
	r.HandleFunc("/api/users/addexternal", addExternalUser).Methods("POST")
	r.HandleFunc("/api/users/getexternal", getExternalUser).Methods("GET")
	r.HandleFunc("/api/users/getexternalbyid/{id}", getExternalUserByID).Methods("GET")
	r.HandleFunc("/api/users/updateexternalbyid/{id}", updateExternalByID).Methods("PUT")
	r.HandleFunc("/api/chat/statusupdate", updateChatStatus).Methods("POST", "PUT")
	r.HandleFunc("/api/users/addinternal", addInternalUser).Methods("POST")

	fmt.Printf("Starting server  at port 8001\n")
	log.Fatal(http.ListenAndServe(":8001", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}
