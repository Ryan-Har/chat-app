package chatstate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

var apiHost string = os.Getenv("apiHost")
var apiPort string = os.Getenv("apiPort")

var apiBaseUrl string = fmt.Sprintf("http://%s:%s/api", apiHost, apiPort)

type ChatStateHandler interface {
	populateFromDB() error
	GetChats() []ChatInformation
	RemoveChat(chatUUID string)
	AddChat(chatUUID string, chatStartTime string)
	AddMessage(chatUUID string, userID int64, message string, time string)
	RemoveParticipant(chatUUID string, userID int64)
	AddParticipant(chatUUID string, userID int64)
	GetApiBaseUrl() string
}

type StateUpdateHandler struct {
	Chats      []ChatInformation `json:"chats"`
	ApiBaseUrl string
	mutex      sync.Mutex
}

type ChatInformation struct {
	ChatUUID      string            `json:"chatuuid"`      // UUID of the chat
	Participants  []ChatParticipant `json:"participants"`  // List of participants in the chat
	Messages      []ChatMessage     `json:"messages"`      // List of messages in the chat
	ChatStartTime string            `json:"chatStartTime"` // Time the chat started (datetime format)
}

type ChatParticipant struct {
	UserID   int64  `json:"userid"`
	Active   bool   `json:"active"`
	Internal bool   `json:"internal"`
	Name     string `json:"name"`
}

type ChatMessage struct {
	UserID  int64  `json:"userid"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

func NewChatStateHandler() (ChatStateHandler, error) {

	var handler = &StateUpdateHandler{
		ApiBaseUrl: fmt.Sprintf("http://%s:%s/api", apiHost, apiPort),
		mutex:      sync.Mutex{},
	}
	err := handler.populateFromDB()
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func (suh *StateUpdateHandler) populateFromDB() error {
	url := apiBaseUrl + "/chat/inprogress/info"
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

	var chatInfoSlice = []ChatInformation{}
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &chatInfoSlice); err != nil {
		return err
	}
	suh.Chats = append(suh.Chats, chatInfoSlice...)

	return nil
}

func (suh *StateUpdateHandler) GetChats() []ChatInformation {
	return suh.Chats
}

func (suh *StateUpdateHandler) RemoveChat(chatUUID string) {
	suh.mutex.Lock()
	defer suh.mutex.Unlock()
	for i, chat := range suh.Chats {
		if chat.ChatUUID == chatUUID {
			suh.Chats = append(suh.Chats[:i], suh.Chats[i+1:]...)
			break
		}
	}
}

func (suh *StateUpdateHandler) AddChat(chatUUID string, chatStartTime string) {
	suh.mutex.Lock()
	defer suh.mutex.Unlock()
	suh.Chats = append(suh.Chats, ChatInformation{
		ChatUUID:      chatUUID,
		Participants:  []ChatParticipant{},
		Messages:      []ChatMessage{},
		ChatStartTime: chatStartTime,
	})
}

func (suh *StateUpdateHandler) AddMessage(chatUUID string, userID int64, message string, time string) {
	suh.mutex.Lock()
	defer suh.mutex.Unlock()
	for i, chat := range suh.Chats {
		if chat.ChatUUID == chatUUID {
			suh.Chats[i].Messages = append(suh.Chats[i].Messages, ChatMessage{
				UserID:  userID,
				Message: message,
				Time:    time,
			})
			break
		}
	}
}

func (suh *StateUpdateHandler) RemoveParticipant(chatUUID string, userID int64) {
	suh.mutex.Lock()
	defer suh.mutex.Unlock()
	for i, chat := range suh.Chats {
		if chat.ChatUUID == chatUUID {
			for j, participant := range suh.Chats[i].Participants {
				if participant.UserID == userID {
					suh.Chats[i].Participants[j].Active = false
					break
				}
			}
			break
		}
	}
}

type BasicUserInfo struct {
	ID          int64  `json:"id"`
	TimeCreated string `json:"timecreated"`
	Internal    bool   `json:"internal"`
	Name        string `json:"name"`
}

func (suh *StateUpdateHandler) AddParticipant(chatUUID string, userID int64) {
	bui := BasicUserInfo{}
	url := fmt.Sprintf("%s/users/getbasicbyid/%d", suh.ApiBaseUrl, userID)
	resp, err := sendGetRequest(url)
	if err != nil {
		log.Println("error with addparticipant api request: ", err.Error())
		return
	}

	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &bui); err != nil {
		log.Println("error with addparticipant json unmarshal: ", err.Error())
		return
	}

	suh.mutex.Lock()
	defer suh.mutex.Unlock()
	for i, chat := range suh.Chats {
		if chat.ChatUUID == chatUUID {
			//just set active if already in chat
			for j, participant := range suh.Chats[i].Participants {
				if participant.UserID == userID {
					suh.Chats[i].Participants[j].Active = true
					return
				}
			}
			suh.Chats[i].Participants = append(suh.Chats[i].Participants, ChatParticipant{
				UserID:   userID,
				Active:   true,
				Internal: bui.Internal,
				Name:     bui.Name,
			})
			break
		}
	}
}

func (suh *StateUpdateHandler) GetApiBaseUrl() string {
	return suh.ApiBaseUrl
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
