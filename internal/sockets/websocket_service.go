package sockets

import (
	"encoding/json"
	"log"
	"sync"
)

type WebsocketService interface {
	RegisterClient(client *Client)
	UnregisterClient(userId string)
	SendMessage(userId string, message *WebsocketMessage)
	HandleMessage(userId string, message *WebsocketMessage)
}

type WebsocketServiceImpl struct {
	mutex   sync.Mutex
	clients map[string]*Client
}

func NewWebsocketService() *WebsocketServiceImpl {
	return &WebsocketServiceImpl{
		clients: map[string]*Client{},
	}
}

func (service *WebsocketServiceImpl) RegisterClient(client *Client) {
	log.Printf("Registering client: userId=%v", client.userId)

	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.clients[client.userId] = client
}

func (service *WebsocketServiceImpl) UnregisterClient(userId string) {
	log.Printf("Unregistering client: userId=%v", userId)

	service.mutex.Lock()
	defer service.mutex.Unlock()

	delete(service.clients, userId)
}

func (service *WebsocketServiceImpl) SendMessage(userId string, message *WebsocketMessage) {
	log.Printf("Sending websocket message: userId=%v, type=%v", userId, message.Type)

	if client, ok := service.clients[userId]; ok {
		client.messages <- message
	}
}

func (service *WebsocketServiceImpl) HandleMessage(userId string, message *WebsocketMessage) {
	log.Printf("Handling websocket message: userId=%v, type=%v", userId, message.Type)

	switch message.Type {
	case EventTypeSendMessage:
		service.handleSendMessageEvent(userId, message)
	default:
		log.Printf("Unknown message type: %v", message.Type)
	}
}

func (service *WebsocketServiceImpl) handleSendMessageEvent(userId string, message *WebsocketMessage) {
	log.Printf("Handling send message event: userId=%v", userId)

	sendMessagePayload := SendMessagePayload{}
	if err := json.Unmarshal(message.Payload, &sendMessagePayload); err != nil {
		log.Printf("Failed to unmarshal websocket message payload: userId=%v", userId)
		return
	}

	receiveMessagePayload := ReceiveMessagePayload{
		SenderId: userId,
		Message: sendMessagePayload.Message,
	}
	
	payload, err := json.Marshal(receiveMessagePayload)
	if err != nil {
		log.Printf("Failed to marshal websocket message payload: userId=%v", userId)
		return	
	}

	message = &WebsocketMessage{
		Type: EventTypeReceiveMessage,
		Payload: payload,
	}
	service.SendMessage(sendMessagePayload.ReceiverId, message)
}
