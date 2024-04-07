package dbquery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DBQueryHandler interface {
	AddExternalUser(name string, ip string) ([][]any, error)
	GetExternalUser(name string, ip string) ([][]any, error)
	GetExternalUserByID(id int64) ([][]any, error)
	UpdateExternalUserByID(id int64, name string, ip string, email string) ([][]any, error)
	ChatStart(uuid string, startTime string) ([][]any, error)
	ChatEnd(uuid string, endTime string) ([][]any, error)
	AddInternalUser(roleID int64, firstname string, surname string, email string, password string) ([][]any, error)
	GetInternalUserByID(id int64) ([][]any, error)
	UpdateInternalUserByID(id int64, roleID int64, firstname string, surname string, email string, password string) ([][]any, error)
	AddMessageByUUID(uuid string, userid int64, message string, time string) ([][]any, error)
	GetAllMessagesByUUID(uuid string) ([][]any, error)
	GetAllChatsInProgress() ([][]any, error)
	JoinChatParticipant(uuid string, userid int64, time string) ([][]any, error)
	LeaveChatParticipant(uuid string, userid int64, time string) ([][]any, error)
	GetOngoingChatParticipants() ([][]any, error)
	GetOngoingChatMessages() ([][]any, error)
}

type PostgresDBConfig struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
}

type PostgresQueryHandler struct {
	RequestChan chan dbQuery
}

type SqlLiteQueryHandler struct {
	RequestChan chan dbQuery
}

func NewPostgresHandler(dbc PostgresDBConfig) (DBQueryHandler, error) {
	var dbQueryHandler DBQueryHandler

	ch := make(chan error)
	dbRequestChan := make(chan dbQuery, 10) // Buffered for efficient queuing
	go dbc.manageConnections(ch, dbRequestChan)
	err := <-ch
	if err != nil {
		return nil, err
	}
	dbQueryHandler = PostgresQueryHandler{
		RequestChan: dbRequestChan,
	}
	return dbQueryHandler, nil
}

type postgresDBConnector struct {
	err error
	PostgresDBConfig
}

type dbQuery struct {
	Query                   string
	ExpectSingleRow         bool                 //true for single row, false if not
	ReturnChan              chan [][]interface{} // Generic response channel
	NumberOfColumnsExpected int                  //0 for no sql data returned
}

type dbConnectionHandler interface {
	manageConnections()
}

// manages the dbConnectors
func (dbc PostgresDBConfig) manageConnections(ch chan<- error, dbRequestChan chan dbQuery) {
	dbConnectorChan := make(chan *postgresDBConnector, 1)
	wk := &postgresDBConnector{
		err:              nil,
		PostgresDBConfig: dbc,
	}

	//first connection attempt before working to ensure connection is valid
	db, err := connectToPostgres(dbc)
	ch <- err
	close(ch)
	if err != nil {
		return
	} else {
		db.Close()
	}

	go wk.work(dbConnectorChan, dbRequestChan)

	for wk := range dbConnectorChan {
		log.Printf("DBConnector stopped with err: %s", wk.err)
		// reset err
		wk.err = nil
		// a goroutine has ended, restart it
		go wk.work(dbConnectorChan, dbRequestChan)
	}
}

func connectToPostgres(dbc PostgresDBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbc.DBHost, dbc.DBPort, dbc.DBUser, dbc.DBPassword, dbc.DBName)) //consider secure password handling
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// manages the database connection and calls the query workers
func (wk *postgresDBConnector) work(dbConnectorChan chan<- *postgresDBConnector, dbRequestChan chan dbQuery) (err error) {
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

	db, err := connectToPostgres(wk.PostgresDBConfig)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for msg := range dbRequestChan {
		if err := db.Ping(); err != nil {
			dbRequestChan <- msg
			panic(err)
		}

		var sendResults [][]interface{}
		fmt.Println("DB Query Running:", msg)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//returns error if there is an error or bool true if all ok.
		//it doesn't return an error or notify if no rows are changed. TBD
		if msg.NumberOfColumnsExpected == 0 {
			resp, err := db.ExecContext(ctx, msg.Query)
			var singleResult []interface{}
			if err != nil {
				singleResult = append(singleResult, err)
			} else if rows, _ := resp.RowsAffected(); rows == 0 {
				singleResult = append(singleResult, errors.New("no rows changed"))
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

			if err := db.QueryRowContext(ctx, msg.Query).Scan(columnPointers...); err != nil {
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
				queryError := make([]interface{}, 1)
				queryError[0] = err
				sendResults = append(sendResults, queryError)
				msg.ReturnChan <- sendResults
				close(msg.ReturnChan)
				continue // Skipping further processing for this query
			}
			// Process rows and fill the response interface as needed
			defer rows.Close()

			if rows.Next() {
				columns := make([]interface{}, msg.NumberOfColumnsExpected)
				columnPointers := make([]interface{}, len(columns))
				for i := range columns {
					columnPointers[i] = &columns[i]
				}
				if err := rows.Scan(columnPointers...); err != nil {
					fmt.Println("error querying database:", err)
					columns[0] = err
				}
				sendResults = append(sendResults, columns)
				for rows.Next() {
					columns := make([]interface{}, msg.NumberOfColumnsExpected)
					columnPointers := make([]interface{}, len(columns))
					for i := range columns {
						columnPointers[i] = &columns[i]
					}
					if err := rows.Scan(columnPointers...); err != nil {
						fmt.Println("error querying database:", err)
						columns[0] = err
					}
					sendResults = append(sendResults, columns)
				}
			} else {
				noResultsMessage := make([]interface{}, 1)
				noResultsMessage[0] = errors.New("sql: no rows in result set")
				sendResults = append(sendResults, noResultsMessage)
			}
			if err := rows.Err(); err != nil {
				log.Fatal(err)
			}
			msg.ReturnChan <- sendResults
			close(msg.ReturnChan)
		}
	}
	return err
}

func (pqh PostgresQueryHandler) AddExternalUser(name string, ip string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT add_external_user(%s, %s)",
			singleQuote(doubleUpSingleQuotes(name)),
			singleQuote(ip)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 1,
		ExpectSingleRow:         true,
	}

	log.Println("Add external user DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Add external user DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetExternalUser(name string, ip string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT user_id, name, ip_address, email FROM external_users WHERE name = %s AND ip_address = %s",
			singleQuote(doubleUpSingleQuotes(name)),
			singleQuote(ip)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         true,
	}

	log.Println("Get external user DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get external user DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetExternalUserByID(id int64) ([][]any, error) {
	dbq := dbQuery{
		Query:                   fmt.Sprintf("SELECT user_id, name, ip_address, email FROM external_users WHERE user_id = %d", id),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         true,
	}

	log.Println("Get external user by id DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get external user by id DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) UpdateExternalUserByID(id int64, name string, ip string, email string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT given_user_id, updated_name, updated_ip_address, updated_email FROM update_external_user_info(%d, %s, %s, %s)",
			id,
			singleQuote(doubleUpSingleQuotes(name)),
			singleQuote(ip),
			singleQuote(doubleUpSingleQuotes(email))),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         true,
	}

	log.Println("Update external user by id DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Update external user by id DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) ChatStart(uuid string, startTime string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("INSERT INTO chat (uuid, start_time) VALUES (%s, %s)",
			singleQuote(uuid),
			singleQuote(startTime)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 0,
		ExpectSingleRow:         true,
	}

	log.Println("Chat Start DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Chat Start DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) ChatEnd(uuid string, endTime string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("UPDATE chat SET end_time = %s WHERE uuid = %s",
			singleQuote(endTime),
			singleQuote(uuid)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 0,
		ExpectSingleRow:         true,
	}

	log.Println("Chat End DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Chat End DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) AddInternalUser(roleID int64, firstname string, surname string, email string, password string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT add_internal_user(%d, %s, %s, %s, %s)",
			roleID,
			singleQuote(doubleUpSingleQuotes(firstname)),
			singleQuote(doubleUpSingleQuotes(surname)),
			singleQuote(doubleUpSingleQuotes(email)),
			singleQuote(doubleUpSingleQuotes(password))),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 1,
		ExpectSingleRow:         true,
	}

	log.Println("Add internal user DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Add internal user DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetInternalUserByID(id int64) ([][]any, error) {
	dbq := dbQuery{
		Query:                   fmt.Sprintf("SELECT user_id, role_id, firstname, surname, email, password FROM internal_users WHERE user_id = %d", id),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 6,
		ExpectSingleRow:         true,
	}

	log.Println("Get internal user by id DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get internal user by id DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) UpdateInternalUserByID(id int64, roleID int64, firstname string, surname string, email string, password string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("SELECT given_user_id, updated_role_id, updated_firstname, updated_surname, updated_email, updated_password FROM update_internal_user_info(%d, %d, %s, %s, %s, %s)",
			id,
			roleID,
			singleQuote(doubleUpSingleQuotes(firstname)),
			singleQuote(doubleUpSingleQuotes(surname)),
			singleQuote(doubleUpSingleQuotes(email)),
			singleQuote(doubleUpSingleQuotes(password))),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 6,
		ExpectSingleRow:         true,
	}

	log.Println("Update internal user by id DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Update internal user by id DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) AddMessageByUUID(uuid string, userid int64, message string, time string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("INSERT INTO chat_messages (chat_uuid, user_id_from, message, timestamp) VALUES (%s, %d, %s, %s)",
			singleQuote(uuid),
			userid,
			singleQuote(doubleUpSingleQuotes(message)),
			singleQuote(time)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 0,
		ExpectSingleRow:         false,
	}

	log.Println("Add message by uuid DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Add message by uuid DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetAllMessagesByUUID(uuid string) ([][]any, error) {
	dbq := dbQuery{
		Query:                   fmt.Sprintf("SELECT chat_uuid::VARCHAR, user_id_from, message, timestamp::VARCHAR FROM chat_messages WHERE chat_uuid = %s ORDER BY timestamp ASC", singleQuote(uuid)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         false,
	}

	log.Println("Get all messages by uuid DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get all messages by uuid DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetAllChatsInProgress() ([][]any, error) {
	dbq := dbQuery{
		Query:                   "SELECT uuid::VARCHAR, start_time::VARCHAR FROM chat WHERE end_time is null",
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 2,
		ExpectSingleRow:         false,
	}

	log.Println("Get all chats in progress DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get all chats in progress DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) JoinChatParticipant(uuid string, userid int64, time string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("INSERT INTO chat_participant (chat_uuid, user_id, time_joined) VALUES (%s, %d, %s)",
			singleQuote(uuid),
			userid,
			singleQuote(time)),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 0,
		ExpectSingleRow:         false,
	}

	log.Println("Join chat participant DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Join chat participant DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) LeaveChatParticipant(uuid string, userid int64, time string) ([][]any, error) {
	dbq := dbQuery{
		Query: fmt.Sprintf("UPDATE chat_participant SET time_left = %s WHERE chat_uuid = %s AND user_id = %d",
			singleQuote(time),
			singleQuote(uuid),
			userid),
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 0,
		ExpectSingleRow:         false,
	}

	log.Println("Leave chat participant DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Leave chat participant DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetOngoingChatParticipants() ([][]any, error) {
	dbq := dbQuery{
		Query: `SELECT c.uuid::VARCHAR,
				p.user_id,
				CASE WHEN p.time_left IS NULL THEN TRUE ELSE FALSE END AS active,
				u.internal,
				COALESCE(iu.firstname || ' ' || iu.surname, eu.name) AS name
			FROM chat c
			INNER JOIN (
				SELECT UUID
				FROM chat
				WHERE end_time IS null
				) active_chats ON c.uuid = active_chats.uuid
			LEFT JOIN chat_participant p ON c.uuid = p.chat_uuid
			LEFT JOIN users u ON p.user_id = u.id
			LEFT JOIN internal_users iu ON u.id = iu.user_id
			LEFT JOIN external_users eu ON u.id = eu.user_id
			ORDER BY uuid`,
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 5,
		ExpectSingleRow:         false,
	}

	log.Println("Get ongoing chat participants DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get ongoing chat participants DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

func (pqh PostgresQueryHandler) GetOngoingChatMessages() ([][]any, error) {
	dbq := dbQuery{
		Query: `SELECT c.uuid::VARCHAR,
				m.user_id_from,		
				m.message,		
				m.timestamp::VARCHAR as message_time
			FROM chat c
			INNER JOIN (
				SELECT UUID
				FROM chat
				WHERE end_time IS null
				) active_chats ON c.uuid = active_chats.uuid
			LEFT JOIN chat_messages m ON c.uuid = m.chat_uuid
			ORDER BY uuid, message_time asc`,
		ReturnChan:              make(chan [][]interface{}),
		NumberOfColumnsExpected: 4,
		ExpectSingleRow:         false,
	}

	log.Println("Get ongoing chat messages DB Request:", dbq.Query)

	pqh.RequestChan <- dbq
	resp, ok := <-dbq.ReturnChan

	log.Println("Get ongoing chat messages DB Response:", resp)

	if !ok {
		return resp, errors.New("channel closed, no data")
	}
	if err := checkSqlResponseForErrors(resp, dbq.ExpectSingleRow); err != nil {
		return resp, err
	}
	return resp, nil
}

// adds single quotes around any string
func singleQuote(s string) string {
	return "'" + s + "'"
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

func checkSqlResponseForErrors(sqlResponse [][]any, expectSingleRow bool) error {
	if expectSingleRow {
		if len(sqlResponse) != 1 { //only expecting a single response
			return errors.New("db returned a result which is not correct, please review logs")
		}
	}
	if err, ok := sqlResponse[0][0].(error); ok {
		return err
	}
	return nil
}
