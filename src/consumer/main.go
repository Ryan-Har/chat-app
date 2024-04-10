package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var lavinmqHost string = os.Getenv("lavinmqHost")
var lavinmqPort string = os.Getenv("lavinmqPort")
var apiHost string = os.Getenv("apiHost")
var apiPort string = os.Getenv("apiPort")

var lavinMQURL string = fmt.Sprintf("amqp://guest:guest@%s:%s/", lavinmqHost, lavinmqPort)
var apiBaseUrl string = fmt.Sprintf("http://%s:%s/api", apiHost, apiPort)

const (
	chatQueue     = "ChatUpdateQueue"
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

var brokerSendingChan = make(chan amqp091.Publishing)

func sendToInternalQueue(message *BrokerMessage) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/javascript",
		Body:         b,
	}
	brokerSendingChan <- msg
	return nil
}

func workerManager() {
	workerChan := make(chan *worker, workerCount)

	for i := 0; i < workerCount; i++ {
		i := i
		wk := &worker{id: i}
		go wk.workConsume(workerChan)
	}
	for i := workerCount; i < workerCount*2; i++ {
		i := i
		wk := &worker{id: i}
		go wk.workSend(workerChan, brokerSendingChan)
	}
	for wk := range workerChan {
		log.Printf("amqpWorker %d stopped with err: %s", wk.id, wk.err)
		// reset err
		wk.err = nil
		// a goroutine has ended, restart it
		if wk.id < workerCount {
			go wk.workConsume(workerChan)
		} else {
			go wk.workSend(workerChan, brokerSendingChan)
		}
	}
}

func (wk *worker) workSend(workerChan chan<- *worker, brokerchan chan amqp091.Publishing) (err error) {

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
		return err
	}
	defer conn.Close()
	fmt.Println("connected to lavin instance successfully")

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for msg := range brokerchan {
		err = ch.PublishWithContext(
			ctx,
			"",
			internalQueue,
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
					sendToInternalQueue(&messageBack)
					panic(err)
				}
				panic(err)
			}
		}
	}
	return err
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
		chatQueue,
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
		chatQueue,
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

type ChatUuidTime struct {
	ChatUUID string `json:"chatuuid"`
	Time     string `json:"time"`
}

type ChatMessage struct {
	ChatUUID string `json:"chatuuid"`
	UserID   int64  `json:"userid"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

type JoinLeave struct {
	ChatUUID string `json:"chatuuid"`
	Time     string `json:"time"`
	UserID   int64  `json:"userid"`
}

func processMessage(msg amqp091.Delivery) error {

	bm := BrokerMessage{}
	if err := json.Unmarshal(msg.Body, &bm); err != nil {
		return err
	}

	switch bm.MessageText {
	case "End of chat":
		body := ChatUuidTime{
			ChatUUID: bm.Roomid,
			Time:     bm.Time,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		resp, err := sendPutRequest(apiBaseUrl+"/chat/statusupdate", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		//no content - nothing was updated, likely dealt with out of order, requeue
		if resp.StatusCode == 422 {
			log.Printf("End of chat api request for uuid %s resulted in 422, requeue", body.ChatUUID)
			msg.Nack(false, true)
			return err
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}
		fmt.Println("resp body:", string(respBody))
		sendToInternalQueue(&bm)
		msg.Ack(false)

	case "Start of chat":
		body := ChatUuidTime{
			ChatUUID: bm.Roomid,
			Time:     bm.Time,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Println("error marshalling json:", body)
			return err
		}
		resp, err := sendPostRequest(apiBaseUrl+"/chat/statusupdate", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}
		fmt.Println("resp body:", string(respBody))
		fmt.Println(err)
		sendToInternalQueue(&bm)
		msg.Ack(false)

	case "User joined chat":
		body := JoinLeave{
			ChatUUID: bm.Roomid,
			Time:     bm.Time,
			UserID:   bm.UserID,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Println("error marshalling json:", body)
			return err
		}
		resp, err := sendPostRequest(apiBaseUrl+"/chat/participantupdate", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == 422 {
			log.Printf("Participant join api request for uuid %s resulted in 422, requeue", body.ChatUUID)
			msg.Nack(false, true)
			return err
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}
		fmt.Println("resp body:", string(respBody))
		sendToInternalQueue(&bm)
		msg.Ack(false)

	case "User left chat":
		body := JoinLeave{
			ChatUUID: bm.Roomid,
			Time:     bm.Time,
			UserID:   bm.UserID,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Println("error marshalling json:", body)
			return err
		}
		resp, err := sendPutRequest(apiBaseUrl+"/chat/participantupdate", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == 422 {
			log.Printf("Participant leave api request for uuid %s resulted in 422, requeue", body.ChatUUID)
			msg.Nack(false, true)
			return err
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}
		fmt.Println("resp body:", string(respBody))
		sendToInternalQueue(&bm)
		msg.Ack(false)

	default: //must be a message
		body := ChatMessage{
			ChatUUID: bm.Roomid,
			UserID:   bm.UserID,
			Message:  bm.MessageText,
			Time:     bm.Time,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Println("error marshalling json:", body)
			return err
		}
		resp, err := sendPostRequest(apiBaseUrl+"/chat/addmessage", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 { //requeue
			log.Printf("Addmsg api request for uuid %s resulted in %v, requeue", body.ChatUUID, resp.StatusCode)
			msg.Nack(false, true)
			return err
		}
		sendToInternalQueue(&bm)
		msg.Ack(false)
	}

	return nil
}

func sendPutRequest(url string, content *bytes.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", url, content)
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

func sendGetRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
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

func getTimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05.999999")
}

func main() {
	go workerManager()
	select {}
}
