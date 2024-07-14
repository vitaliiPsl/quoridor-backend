package sockets

import (
	"encoding/json"
	"log"
	"quoridor/internal/events"
	"quoridor/internal/game"
	"quoridor/internal/matchmaking"
	"sync"
)

type WebsocketService interface {
	RegisterClient(client *Client)
	UnregisterClient(userId string)
	SendMessage(userId string, message *WebsocketMessage)
	HandleMessage(userId string, message *WebsocketMessage)
	HandleMatchFound(event *events.Event)
}

type WebsocketServiceImpl struct {
	mutex       sync.Mutex
	clients     map[string]*Client
	mmService   matchmaking.MatchmakingService
	gameService game.GameService
}

func NewWebsocketService(mmService matchmaking.MatchmakingService, gameService game.GameService) *WebsocketServiceImpl {
	return &WebsocketServiceImpl{
		clients:     map[string]*Client{},
		mmService:   mmService,
		gameService: gameService,
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
	case EventTypeStartGame:
		service.mmService.AddUser(userId)
	default:
		log.Printf("Unknown message type: %v", message.Type)
	}
};

func (service *WebsocketServiceImpl) HandleMatchFound(event *events.Event) {
	data := event.Data.(map[string]string)
	user1Id := data["user1Id"]
	user2Id := data["user2Id"]

	log.Printf("Handling match found: user1Id=%v, user2Id=%v", user1Id, user2Id)

	game, err := service.gameService.CreateGame(user1Id, user2Id)
	if err != nil {
		log.Printf("Error creating game: %v", err)
		return
	}

	payload, err := json.Marshal(game)
	if err != nil {
		log.Printf("Failed to marshal game state: err=%v", err)
		return
	}

	message := WebsocketMessage{Type: EventTypeGameState, Payload: payload}
	service.SendMessage(user1Id, &message)
	service.SendMessage(user2Id, &message)
}
