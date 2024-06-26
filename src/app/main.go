package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ryan-Har/chat-app/src/app/chatstate"
	"github.com/rabbitmq/amqp091-go"
)

var lavinmqHost string = os.Getenv("lavinmqHost")
var lavinmqPort string = os.Getenv("lavinmqPort")

var lavinMQURL string = fmt.Sprintf("amqp://guest:guest@%s:%s/", lavinmqHost, lavinmqPort)

const (
	workerCount   = 5
	internalQueue = "AppQueue"
)

type BrokerMessage struct {
	Roomid      string `json:"roomid"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	MessageText string `json:"messagetext"`
	UserID      int64  `json:"userid"`
	Time        string `json:"time"`
}

type worker struct {
	id  int
	err error
}

func workerManager(stateHandler chatstate.ChatStateHandler) {
	workerChan := make(chan *worker, workerCount)

	for i := 0; i < workerCount; i++ {
		i := i
		wk := &worker{id: i}
		go wk.workConsume(workerChan, stateHandler)
	}
	for wk := range workerChan {
		log.Printf("amqpWorker %d stopped with err: %s", wk.id, wk.err)
		// reset err
		wk.err = nil
		// a goroutine has ended, restart it
		wk.workConsume(workerChan, stateHandler)
	}
}

func (wk *worker) workConsume(workerChan chan<- *worker, stateHandler chatstate.ChatStateHandler) (err error) {
	// make my goroutine signal its death, whether it's a panic or a return
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

	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	_, err = ch.QueueDeclare(
		internalQueue,
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
		panic(err)
	}

	msgs, err := ch.Consume(
		internalQueue,
		fmt.Sprintf("consumer %d", wk.id),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	for msg := range msgs {
		if err = processMessage(msg, stateHandler); err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		time.Sleep(time.Second)
	}
	return err
}

func processMessage(msg amqp091.Delivery, stateHandler chatstate.ChatStateHandler) error {

	bm := BrokerMessage{}
	if err := json.Unmarshal(msg.Body, &bm); err != nil {
		return err
	}
	log.Println("Received message:", bm)
	switch bm.MessageText {
	case "End of chat":
		stateHandler.RemoveChat(bm.Roomid)
		msg.Ack(false)
	case "Start of chat":
		stateHandler.AddChat(bm.Roomid, bm.Time)
		msg.Ack(false)
	case "User joined chat":
		stateHandler.AddParticipant(bm.Roomid, bm.UserID)
		msg.Ack(false)
	case "User left chat":
		stateHandler.RemoveParticipant(bm.Roomid, bm.UserID)
		msg.Ack(false)
	default: //must be a message
		stateHandler.AddMessage(bm.Roomid, bm.UserID, bm.MessageText, bm.Time)
		msg.Ack(false)

	}
	return nil
}

func streamChats(w http.ResponseWriter, r *http.Request, stateHandler chatstate.ChatStateHandler) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher := w.(http.Flusher)

	for {
		data, err := json.Marshal(stateHandler.GetChats())
		if err != nil {
			log.Println("Error marshalling message:", err)
			return
		}

		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()

		time.Sleep(2 * time.Second)
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	err := tmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/navbar.html"))
	err := tmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Title":    "Chat App",
		"ChatHost": os.Getenv("chatHost"),
		"ChatPort": os.Getenv("chatPort"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func chatPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("templates/chats.html"))
	err := tmpl.ExecuteTemplate(w, "chats.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type loginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginWithUsernameAndPassword(w http.ResponseWriter, r *http.Request, apiBaseUrl string) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Println(r.Body)
	fmt.Printf("apiBaseUrl: %s\n", apiBaseUrl)

	var li *loginInfo
	err := json.NewDecoder(r.Body).Decode(&li)
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}

	fmt.Println(li)
	payloadJson, err := json.Marshal(&li)
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}

	fmt.Println("sending post")
	resp, err := sendPostRequest(fmt.Sprintf("%s/users/login", apiBaseUrl), bytes.NewReader(payloadJson))
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}
	fmt.Println("received response")
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(w, "%s", string(data))
		fmt.Printf("response: %v", data)
		return
	} else {
		fmt.Fprintf(w, "0")
		return
	}
}

func sendPostRequest(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func main() {
	chatHandler, err := chatstate.NewChatStateHandler()
	if err != nil {
		log.Println("error creating chat state handler", err.Error())
	}

	go workerManager(chatHandler)
	// Handle the root URL

	//serve js and css files
	cssfs := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", cssfs))

	jsfs := http.FileServer(http.Dir("js"))
	http.Handle("/js/", http.StripPrefix("/js/", jsfs))

	//http.Handle("/login", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	http.HandleFunc("/chatstream", func(w http.ResponseWriter, r *http.Request) {
		streamChats(w, r, chatHandler)
	})
	http.HandleFunc("/handlelogin", func(w http.ResponseWriter, r *http.Request) {
		loginWithUsernameAndPassword(w, r, chatHandler.GetApiBaseUrl())
	})
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/chats", chatPage)
	http.ListenAndServe(":8005", nil)
}
