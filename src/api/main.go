package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Ryan-Har/chat-app/src/api/dbquery"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type ExternalUserInfo struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	IPAddr    string `json:"ipaddr"`
	EmailAddr string `json:"email,omitempty"`
}

type idResponse struct {
	ID int64
}

// responds with user information of added user
func addExternalUser(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
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
	log.Println("Add external user api request:", eui)

	resp, err := dbqh.AddExternalUser(eui.Name, eui.IPAddr)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
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

// responds with user information of added user, if exists
func getExternalUser(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
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
	log.Println("Get external user api request:", eui)

	resp, err := dbqh.GetExternalUser(eui.Name, eui.IPAddr)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
		return
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

func getExternalUserByID(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
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

	log.Println("Get external user by ID api request:", eui.ID)

	resp, err := dbqh.GetExternalUserByID(eui.ID)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
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

func updateExternalUserByID(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
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

	log.Println("Update external user by ID api request:", eui.ID)

	resp, err := dbqh.UpdateExternalUserByID(eui.ID, eui.Name, eui.IPAddr, eui.EmailAddr)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
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
		fmt.Fprintf(w, "error verifying db results. Error: %v", err)
		return
	}
	err = json.NewEncoder(w).Encode(structToVerify)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}

}

func updateChatStatus(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {

	var cut *ChatUuidTime
	err := json.NewDecoder(r.Body).Decode(&cut)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	verifiedTime, err := verifyTimeFormat(cut.Time)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		log.Println("Chat start api request:", cut.ChatUUID)
		if _, err := dbqh.ChatStart(cut.ChatUUID, verifiedTime); err != nil {
			verifyDBErrorsAndReturn(w, err)
		}
	} else {
		log.Println("Chat end api request:", cut.ChatUUID)
		if _, err := dbqh.ChatEnd(cut.ChatUUID, verifiedTime); err != nil {
			verifyDBErrorsAndReturn(w, err)
		}
	}
}

func addInternalUser(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
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

	log.Println("Add internal user api request:", iui)

	resp, err := dbqh.AddInternalUser(iui.RoleID, iui.FirstName, iui.Surname, iui.EmailAddr, iui.HashedPassword)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
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

func getInternalUserByID(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	idString := mux.Vars(r)["id"]
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iui := &InternalUserInfo{}
	iui.ID = int64(idInt)

	log.Println("Get internal user by ID api request:", iui.ID)

	resp, err := dbqh.GetInternalUserByID(iui.ID)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
		return
	}

	structToVerify := InternalUserInfo{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(structToVerify)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

func updateInternalUserByID(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	idString := mux.Vars(r)["id"]
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var iui *InternalUserInfo
	err = json.NewDecoder(r.Body).Decode(&iui)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//overwrite id with int provided in the url, incase it's different to that applied in the body
	iui.ID = int64(idInt)

	log.Println("Update internal user by ID api request:", iui.ID)

	resp, err := dbqh.UpdateInternalUserByID(iui.ID, iui.RoleID, iui.FirstName, iui.Surname, iui.EmailAddr, iui.HashedPassword)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
		return
	}

	structToVerify := InternalUserInfo{}
	intToStruct := interface{}(&structToVerify)

	if err := convertSliceToStruct(resp[0], intToStruct); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
		return
	}
	if iui.ID != structToVerify.ID {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error verifying db results")
		return
	}
	err = json.NewEncoder(w).Encode(structToVerify)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}

}

func addMessage(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	var cm *ChatMessage
	err := json.NewDecoder(r.Body).Decode(&cm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Add chat message api request:", cm)

	_, err = dbqh.AddMessageByUUID(cm.ChatUUID, cm.UserID, cm.Message, cm.Time)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
		return
	}
}

func getAllMessages(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	uuid := mux.Vars(r)["uuid"]

	log.Println("get all message request for uuid api request:", uuid)

	resp, err := dbqh.GetAllMessagesByUUID(uuid)
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
		return
	}

	respSlice := []ChatMessage{}

	for i := range resp {
		structToVerify := ChatMessage{}
		intToStruct := interface{}(&structToVerify)

		if err := convertSliceToStruct(resp[i], intToStruct); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
			return
		}
		respSlice = append(respSlice, structToVerify)
	}

	err = json.NewEncoder(w).Encode(respSlice)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

func getChatsInProgress(w http.ResponseWriter, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	log.Println("get chats in progress api request")

	resp, err := dbqh.GetAllChatsInProgress()
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
	}

	respSlice := []ChatUuidTime{}

	for i := range resp {
		structToVerify := ChatUuidTime{}
		intToStruct := interface{}(&structToVerify)

		if err := convertSliceToStruct(resp[i], intToStruct); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
			return
		}
		respSlice = append(respSlice, structToVerify)
	}

	err = json.NewEncoder(w).Encode(respSlice)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

func chatParticipantUpdate(w http.ResponseWriter, r *http.Request, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	var jl *JoinLeave
	err := json.NewDecoder(r.Body).Decode(&jl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	verifiedTime, err := verifyTimeFormat(jl.Time)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idint64, err := strconv.ParseInt(jl.UserID, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		log.Println("Join chat participant api request:", jl)
		if _, err := dbqh.JoinChatParticipant(jl.ChatUUID, idint64, verifiedTime); err != nil {
			verifyDBErrorsAndReturn(w, err)
		}
	} else if r.Method == "PUT" {
		log.Println("Leave chat participant api request:", jl)
		if _, err := dbqh.LeaveChatParticipant(jl.ChatUUID, idint64, verifiedTime); err != nil {
			verifyDBErrorsAndReturn(w, err)
		}
	} else {
		http.Error(w, errors.New("only POST or PUT requests accepted").Error(), http.StatusBadRequest)
		return
	}
}

func GetAllOngoingChatParticipants(w http.ResponseWriter, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	log.Println("Get all ongoing chat participants api request:")

	resp, err := dbqh.GetOngoingChatParticipants()
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
	}

	respSlice := []chatParticipant{}

	for i := range resp {
		structToVerify := chatParticipant{}
		intToStruct := interface{}(&structToVerify)

		if err := convertSliceToStruct(resp[i], intToStruct); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
			return
		}
		respSlice = append(respSlice, structToVerify)
	}

	err = json.NewEncoder(w).Encode(respSlice)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error converting result to JSON")
		return
	}
}

func GetAllOngoingChatMessages(w http.ResponseWriter, dbqh dbquery.DBQueryHandler) {
	enableCors(&w)
	respondJson(&w)

	log.Println("Get all ongoing chat messages api request:")

	resp, err := dbqh.GetOngoingChatMessages()
	if err != nil {
		verifyDBErrorsAndReturn(w, err)
	}

	respSlice := []ChatMessage{}

	for i := range resp {
		structToVerify := ChatMessage{}
		intToStruct := interface{}(&structToVerify)

		if err := convertSliceToStruct(resp[i], intToStruct); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error converting db results to struct. Error: %v", err.Error())
			return
		}
		respSlice = append(respSlice, structToVerify)
	}

	err = json.NewEncoder(w).Encode(respSlice)
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

type chatParticipant struct {
	ChatUUID  string `json:"chatuuid"`
	StartTime string `json:"starttime"`
	UserID    int64  `json:"userid"`
	Active    bool   `json:"active"`
	Internal  bool   `json:"internal"`
	Name      string `json:"name"`
}

type JoinLeave struct {
	ChatUUID string `json:"chatuuid"`
	Time     string `json:"time"`
	UserID   string `json:"userid"`
}

type ChatUuidTime struct {
	ChatUUID string `json:"chatuuid"`
	Time     string `json:"time"`
}

type InternalUserInfo struct {
	ID             int64  `json:"id,omitempty"`
	RoleID         int64  `json:"roleid"`
	FirstName      string `json:"firstname"`
	Surname        string `json:"surname"`
	EmailAddr      string `json:"email,omitempty"`
	HashedPassword string `json:"password"`
}

type ChatMessage struct {
	ChatUUID string `json:"chatuuid"`
	UserID   int64  `json:"userid"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

func verifyDBErrorsAndReturn(w http.ResponseWriter, err error) {
	switch {
	case err.Error() == "sql: no rows in result set":
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "no results")
	case err.Error() == "pq: record not found":
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "record not found")
	case strings.HasPrefix(err.Error(), "pq: invalid input syntax"):
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
	case err.Error() == "no rows changed": //this is used for updates, when we're expecting a row to be updated
		w.WriteHeader(http.StatusUnprocessableEntity)
	case strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint"):
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
	case strings.Contains(err.Error(), "violates foreign key constraint"):
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, err.Error())
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error when running database query: %v", err.Error())
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
	var dbuser string = os.Getenv("POSTGRES_USER")
	var dbpassword string = os.Getenv("POSTGRES_PASSWORD")
	var dbname string = os.Getenv("POSTGRES_DB")
	var dbhost string = os.Getenv("dbhost")
	var dbport string = os.Getenv("dbport")

	dbQueryHandler, err := dbquery.NewPostgresHandler(dbquery.PostgresDBConfig{
		DBUser:     dbuser,
		DBPassword: dbpassword,
		DBName:     dbname,
		DBHost:     dbhost,
		DBPort:     dbport,
	})
	if err != nil {
		log.Panicln("error connecting to database", err.Error())
	}

	// Where ORIGIN_ALLOWED is like `scheme://dns[:port]`, or `*` (insecure)
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Accept"})
	//originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	//methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST"})

	r := mux.NewRouter()
	r.HandleFunc("/api/users/addexternal", func(w http.ResponseWriter, r *http.Request) {
		addExternalUser(w, r, dbQueryHandler)
	}).Methods("POST")
	r.HandleFunc("/api/users/getexternal", func(w http.ResponseWriter, r *http.Request) {
		getExternalUser(w, r, dbQueryHandler)
	}).Methods("GET")
	r.HandleFunc("/api/users/getexternalbyid/{id}", func(w http.ResponseWriter, r *http.Request) {
		getExternalUserByID(w, r, dbQueryHandler)
	}).Methods("GET")
	r.HandleFunc("/api/users/updateexternalbyid/{id}", func(w http.ResponseWriter, r *http.Request) {
		updateExternalUserByID(w, r, dbQueryHandler)
	}).Methods("PUT")
	r.HandleFunc("/api/chat/statusupdate", func(w http.ResponseWriter, r *http.Request) {
		updateChatStatus(w, r, dbQueryHandler)
	}).Methods("POST", "PUT")
	r.HandleFunc("/api/users/addinternal", func(w http.ResponseWriter, r *http.Request) {
		addInternalUser(w, r, dbQueryHandler)
	}).Methods("POST")
	r.HandleFunc("/api/users/getinternalbyid/{id}", func(w http.ResponseWriter, r *http.Request) {
		getInternalUserByID(w, r, dbQueryHandler)
	}).Methods("GET")
	r.HandleFunc("/api/users/updateinternalbyid/{id}", func(w http.ResponseWriter, r *http.Request) {
		updateInternalUserByID(w, r, dbQueryHandler)
	}).Methods("PUT")
	r.HandleFunc("/api/chat/addmessage", func(w http.ResponseWriter, r *http.Request) {
		addMessage(w, r, dbQueryHandler)
	}).Methods("POST")
	r.HandleFunc("/api/chat/getallmessages/{uuid}", func(w http.ResponseWriter, r *http.Request) {
		getAllMessages(w, r, dbQueryHandler)
	}).Methods("GET")
	r.HandleFunc("/api/chat/inprogress", func(w http.ResponseWriter, r *http.Request) {
		getChatsInProgress(w, dbQueryHandler)
	}).Methods("GET")
	r.HandleFunc("/api/chat/participantupdate", func(w http.ResponseWriter, r *http.Request) {
		chatParticipantUpdate(w, r, dbQueryHandler)
	}).Methods("POST", "PUT")
	r.HandleFunc("/api/chat/inprogress/messages", func(w http.ResponseWriter, r *http.Request) {
		GetAllOngoingChatMessages(w, dbQueryHandler)
	}).Methods("GET")
	r.HandleFunc("/api/chat/inprogress/participants", func(w http.ResponseWriter, r *http.Request) {
		GetAllOngoingChatParticipants(w, dbQueryHandler)
	}).Methods("GET")

	fmt.Printf("Starting server  at port 8001\n")
	log.Fatal(http.ListenAndServe(":8001", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}
