package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
)

type UserInfo struct {
	Conn   *websocket.Conn
	Name   string
	UserID int64 //corresponding id of user in database, if it exists
	IPAddr string
}

var lavinmqHost string = os.Getenv("lavinmqHost")
var lavinmqPort string = os.Getenv("lavinmqPort")
var apiHost string = os.Getenv("apiHost")
var apiPort string = os.Getenv("apiPort")

var lavinMQURL string = fmt.Sprintf("amqp://guest:guest@%s:%s/", lavinmqHost, lavinmqPort)
var apiBaseUrl string = fmt.Sprintf("http://%s:%s/api", apiHost, apiPort)

const (
	queueName   = "ChatUpdateQueue"
	workerCount = 5
)

type worker struct {
	id  int
	err error
}

type BrokerMessage struct {
	Roomid      string `json:"roomid"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	MessageText string `json:"messagetext"`
	UserID      int64  `json:"userid"`
	Time        string `json:"time"`
}

var brokerSendingChan = make(chan amqp091.Publishing)

func sendToBroker(message *BrokerMessage) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}
	var newMessage BrokerMessage
	json.Unmarshal(b, &newMessage)
	fmt.Println(newMessage)

	msg := amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/javascript",
		Body:         b,
	}
	brokerSendingChan <- msg
	return nil
}

func amqpManager() {
	workerChan := make(chan *worker, workerCount)

	for i := 0; i < workerCount; i++ {
		i := i
		wk := &worker{id: i}
		go wk.work(workerChan, brokerSendingChan)
	}
	for wk := range workerChan {
		log.Printf("amqpWorker %d stopped with err: %s", wk.id, wk.err)
		// reset err
		wk.err = nil
		// a goroutine has ended, restart it
		go wk.work(workerChan, brokerSendingChan)
	}
}

func (wk *worker) work(workerChan chan<- *worker, brokerchan chan amqp091.Publishing) (err error) {
	// make goroutine signal its death, whether it's a panic or a return
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
		workerChan <- wk
	}()

	conn, err := amqp091.Dial(lavinMQURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("connected to lavin instance successfully")

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange": "message.deadletter",
			"x-max-priority":         10,
		},
	)
	if err != nil {
		wk.err = err
		workerChan <- wk
	}

	for msg := range brokerchan {
		err = ch.PublishWithContext(
			ctx,
			"",
			queueName,
			false,
			false,
			msg,
		)
		if err != nil {
			//if it has this exception, something happened to make the connection close
			//we must resend the message if we don't want to lose it.
			if strings.HasPrefix(err.Error(), "Exception (504)") {
				var messageBack BrokerMessage
				err = json.Unmarshal(msg.Body, &messageBack)
				if err == nil {
					sendToBroker(&messageBack)
					panic(err)
				}
				panic(err)
			}
		}
	}
	return err
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Add your origin validation logic here if needed
		return true
	},
}

// Room to store connected clients
var room = make(map[string]map[*websocket.Conn]*UserInfo)

type ExternalUserInfo struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	IPAddr    string `json:"ipaddr"`
	EmailAddr string `json:"email,omitempty"`
}

type InternalUserInfo struct {
	ID             int64  `json:"id,omitempty"`
	RoleID         int64  `json:"roleid"`
	FirstName      string `json:"firstname"`
	Surname        string `json:"surname"`
	EmailAddr      string `json:"email,omitempty"`
	HashedPassword string `json:"password"`
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Read the GUID from the request URL
	guid := r.URL.Query().Get("guid")
	if guid == "" {
		log.Println("GUID is required.")
		return
	}
	name := r.URL.Query().Get("name")

	urluserid := r.URL.Query().Get("userid")
	var userid int64
	// connect to api and check if user exists already by comparing the
	// the ip and name provided to records.
	// if it doesn't exist then create an external user for them and retrieve the
	// new id for use here
	if urluserid == "" && name != "" { //external users will provide name and no id
		log.Println("external user joining")
		body := ExternalUserInfo{
			Name:   name,
			IPAddr: r.RemoteAddr,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Println("error marshalling json:", body)
			return
		}
		resp, err := sendGetRequest(apiBaseUrl+"/users/getexternal", bytes.NewReader(jsonBody))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		//no content - nothing was updated, likely dealt with out of order, requeue
		if resp.StatusCode == 204 {
			resp2, err := sendPostRequest(apiBaseUrl+"/users/addexternal", bytes.NewReader(jsonBody))
			if err != nil {
				return
			}
			defer resp2.Body.Close()
			data, _ := io.ReadAll(resp2.Body)
			if err := json.Unmarshal(data, &body); err != nil {
				log.Println("error unmarshalling json with status code 204", err)
			}
			userid = body.ID
		} else {
			data, _ := io.ReadAll(resp.Body)
			if err := json.Unmarshal(data, &body); err != nil {
				log.Printf("error unmarshalling json with status code %d: %v \n", resp.StatusCode, err.Error())
			}
			userid = body.ID
		}
	} else if name == "" && urluserid != "" { //internal users will provide id but no name
		log.Println("internal user joining")
		var iui *InternalUserInfo

		resp, err := sendGetRequest(apiBaseUrl+"/users/getinternalbyid/"+urluserid, nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(data, &iui); err != nil {
			log.Println(string(data))
			log.Println("error unmarshalling json", err)
			return
		}
		name = fmt.Sprintf("%s %s", iui.FirstName, iui.Surname)
		userid = iui.ID
	}

	//extract just the ip address from the remote connection
	var ip string
	//this might not work correctly, need to test externally
	tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		fmt.Println("Not a TCP connection.")
		ip = ""
	} else {
		ip = tcpAddr.String()
	}
	fmt.Println(ip)
	userinfo := UserInfo{
		Conn:   conn,
		Name:   name,
		UserID: userid,
		IPAddr: ip,
	}

	// Create a room for the GUID if it doesn't exist
	if _, err := room[guid]; !err {
		room[guid] = make(map[*websocket.Conn]*UserInfo)
		log.Println("Start of chat:", guid)
		brokerMessage := BrokerMessage{
			Roomid:      guid,
			MessageText: "Start of chat",
			Time:        getTimeNow(),
		}

		if err := sendToBroker(&brokerMessage); err != nil {
			log.Println(err)
			return
		}
	}

	// Add the client to the room
	room[guid][conn] = &userinfo

	//send user joined message to broker
	brokerMessage := BrokerMessage{
		Roomid:      guid,
		MessageText: "User joined chat",
		UserID:      userinfo.UserID,
		Time:        getTimeNow(),
	}

	if err := sendToBroker(&brokerMessage); err != nil {
		log.Println(err)
		return
	}

	// Listen for messages from the client
	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		//only handle text messagetype currently. Need to add Binary types upload
		brokerMessage := BrokerMessage{
			Roomid:      guid,
			Name:        userinfo.Name,
			UserID:      userinfo.UserID,
			Address:     userinfo.IPAddr,
			MessageText: string(payload),
			Time:        getTimeNow(),
		}

		if err := sendToBroker(&brokerMessage); err != nil {
			log.Println(err)
			break
		}
		// Broadcast the message to all clients in the room
		for client := range room[guid] {
			if err := client.WriteMessage(messageType, []byte(userinfo.Name+": "+string(payload))); err != nil {
				log.Println(err)
				return
			}
		}
	}

	// Remove the client from the room when the connection is closed
	delete(room[guid], conn)

	//send user left message to broker
	brokerMessage = BrokerMessage{
		Roomid:      guid,
		MessageText: "User left chat",
		UserID:      userinfo.UserID,
		Time:        getTimeNow(),
	}

	if err := sendToBroker(&brokerMessage); err != nil {
		log.Println(err)
		return
	}

	if len(room[guid]) == 0 {
		brokerMessage := BrokerMessage{
			Roomid:      guid,
			MessageText: "End of chat",
			Time:        getTimeNow(),
		}
		log.Println("End of chat:", guid)
		if err := sendToBroker(&brokerMessage); err != nil {
			log.Println(err)
			return
		}
	}
}

func sendPostRequest(url string, content *bytes.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, content)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func sendGetRequest(url string, content *bytes.Reader) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if content != nil {
		req.Body = io.NopCloser(content)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getTimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05.999999")
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	go amqpManager()
	fmt.Printf("Starting server  at port 8002\n")
	log.Fatal(http.ListenAndServe(":8002", nil))
}
