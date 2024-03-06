package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	lavinMQURL    = "amqp://guest:guest@localhost:32769/"
	workerCount   = 5
	internalQueue = "InternalQueue"
	apiBaseUrl    = "http://localhost:8001/api"
)

type BrokerMessage struct {
	Roomid      string `json:"roomid"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	MessageText string `json:"messagetext"`
	UserID      string `json:"userid"`
}

type worker struct {
	id  int
	err error
}

type timeString string

type ongoingChats map[string]timeString

var oc = make(ongoingChats)

func (oc ongoingChats) add(uuid string) {
	timeLayout := "2006-01-02 15:04:05.999999"
	oc[uuid] = timeString(time.Now().Format(timeLayout))
}

func (oc ongoingChats) remove(uuid string) {
	delete(oc, uuid)
}

type ChatUuidTime struct {
	ChatUUID string `json:"chatuuid"`
	Time     string `json:"time"`
}

func (oc ongoingChats) populateFromDB() error {
	url := apiBaseUrl + "/chat/inprogress"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return nil
	}
	var chatTimeSlice = []ChatUuidTime{}
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &chatTimeSlice); err != nil {
		return err
	}
	timeLayoutdb := "2006-01-02 15:04:05.999999-07"
	timeLayout := "2006-01-02 15:04:05.999999"
	for i := range chatTimeSlice {
		parsedTime, err := time.Parse(timeLayoutdb, chatTimeSlice[i].Time)
		if err != nil {
			return err
		}
		oc[chatTimeSlice[i].ChatUUID] = timeString(parsedTime.Format(timeLayout))
	}
	return nil
}

func workerManager() {
	workerChan := make(chan *worker, workerCount)

	for i := 0; i < workerCount; i++ {
		i := i
		wk := &worker{id: i}
		go wk.workConsume(workerChan)
	}
	for wk := range workerChan {
		log.Printf("amqpWorker %d stopped with err: %s", wk.id, wk.err)
		// reset err
		wk.err = nil
		// a goroutine has ended, restart it
		wk.workConsume(workerChan)
	}
}

func (wk *worker) workConsume(workerChan chan<- *worker) (err error) {
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
		if err = processMessage(msg); err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		time.Sleep(time.Second)
	}
	return err
}

func processMessage(msg amqp091.Delivery) error {

	bm := BrokerMessage{}
	if err := json.Unmarshal(msg.Body, &bm); err != nil {
		return err
	}

	if bm.Name == "" && bm.Address == "" && bm.UserID == "" { //should only be call start and end messages
		switch bm.MessageText {
		case "End of chat":
			oc.remove(bm.Roomid)
			msg.Ack(false)

		case "Start of chat":
			oc.add(bm.Roomid)
			msg.Ack(false)
		}
		return nil
	}
	return nil
}

type ChatContent struct {
	Uuid      string     `json:"uuid"`
	StartTime timeString `json:"starttime"`
}

func streamChats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher := w.(http.Flusher)

	for {
		var message = []ChatContent{}

		for k, v := range oc {
			chatContent := ChatContent{
				Uuid:      k,
				StartTime: v,
			}
			message = append(message, chatContent)
		}

		data, err := json.Marshal(message)
		if err != nil {
			log.Println("Error marshalling message:", err)
			return
		}

		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()

		time.Sleep(5 * time.Second)
	}
}

func main() {
	if err := oc.populateFromDB(); err != nil {
		log.Println(err.Error())
	}
	go workerManager()
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/chats", streamChats)
	http.ListenAndServe(":8005", nil)
}
