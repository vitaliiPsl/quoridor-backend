package sockets

import (
	"encoding/json"
	"log"
	"quoridor/internal/errors"
	"quoridor/internal/events"
	"quoridor/internal/game"
	"quoridor/internal/matchmaking"
	"sync"
)

type WebsocketService interface {
	RegisterClient(client *Client)
	UnregisterClient(userId string)
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
	service.mmService.RemoveUser(userId)
}

func (service *WebsocketServiceImpl) sendMessage(userId string, message *WebsocketMessage) {
	log.Printf("Sending websocket message: userId=%v, type=%v", userId, message.Type)

	if client, ok := service.clients[userId]; ok {
		client.messages <- message
	}
}

func (service *WebsocketServiceImpl) sendErrorMessage(userId string, err error) {
	log.Printf("Error: userId=%v, err=%v", userId, err)

	errorPayload := ErrorMessagePayload{ErrorType: err.Error()}
	payload, err := json.Marshal(errorPayload)
	if err != nil {
		log.Printf("Failed to marshal error payload: err=%v", err)
		return
	}

	message := WebsocketMessage{Type: EventTypeError, Payload: payload}
	service.sendMessage(userId, &message)
}

func (service *WebsocketServiceImpl) broadcastGameState(game *game.Game) {
	payload, err := json.Marshal(game)
	if err != nil {
		log.Printf("Failed to marshal game state: err=%v", err)
		return
	}

	message := WebsocketMessage{Type: EventTypeGameState, Payload: payload}
	service.sendMessage(game.Player1.UserId, &message)
	service.sendMessage(game.Player2.UserId, &message)
}

func (service *WebsocketServiceImpl) HandleMessage(userId string, message *WebsocketMessage) {
	log.Printf("Handling websocket message: userId=%v, type=%v", userId, message.Type)

	switch message.Type {
	case EventTypeStartGame:
		service.handStartGame(userId, message)
	case EventTypeMakeMove:
		service.handleMove(userId, message)
	case EventTypePlaceWall:
		service.handlePlaceWall(userId, message)
	case EventTypeResign:
		service.handleResign(userId, message)
	case EventTypeReconnect:
		service.handleReconnect(userId, message)
	default:
		log.Printf("Unknown message type: %v", message.Type)
		service.sendErrorMessage(userId, errors.ErrBadRequest)
	}
}

func (service *WebsocketServiceImpl) HandleMatchFound(event *events.Event) {
	data := event.Data.(map[string]string)
	user1Id := data["user1Id"]
	user2Id := data["user2Id"]

	log.Printf("Handling match found: user1Id=%v, user2Id=%v", user1Id, user2Id)

	game, err := service.gameService.CreateGame(user1Id, user2Id)
	if err != nil {
		service.sendErrorMessage(user1Id, err)
		service.sendErrorMessage(user2Id, err)
		return
	}

	service.broadcastGameState(game)
}

func (service *WebsocketServiceImpl) handStartGame(userId string, _ *WebsocketMessage) {
	log.Printf("Handling start game: userId=%v", userId)

	activeGame, err := service.gameService.GetActiveGameByUserId(userId)
	if err != nil {
		service.sendErrorMessage(userId, errors.ErrInternalError)
		return
	}

	if activeGame == nil {
		log.Printf("Adding user to matchmaking queue: userId=%v.", userId)
		service.mmService.AddUser(userId)
		return
	}

	payload, err := json.Marshal(activeGame)
	if err != nil {
		log.Printf("Failed to marshal game state: err=%v", err)
		return
	}

	message := WebsocketMessage{Type: EventTypeGameState, Payload: payload}
	service.sendMessage(userId, &message)
}

func (service *WebsocketServiceImpl) handleMove(userId string, message *WebsocketMessage) {
	log.Printf("Handling move: userId=%v", userId)

	payload := MakeMovePayload{}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal move request: userId=%v, err=%v", userId, err)
		service.sendErrorMessage(userId, errors.ErrBadRequest)
		return
	}

	game, err := service.gameService.MakeMove(payload.GameId, userId, &payload.Position)
	if err != nil {
		service.sendErrorMessage(userId, err)
		return
	}

	service.broadcastGameState(game)
}

func (service *WebsocketServiceImpl) handlePlaceWall(userId string, message *WebsocketMessage) {
	log.Printf("Handling wall placement: userId=%v", userId)

	payload := PlaceWallPayload{}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal wall placement request: userId=%v, err=%v", userId, err)
		service.sendErrorMessage(userId, errors.ErrBadRequest)
		return
	}

	game, err := service.gameService.PlaceWall(payload.GameId, userId, &payload.Wall)
	if err != nil {
		service.sendErrorMessage(userId, err)
		return
	}

	service.broadcastGameState(game)
}

func (service *WebsocketServiceImpl) handleResign(userId string, message *WebsocketMessage) {
	log.Printf("Handling resign: userId=%v", userId)

	payload := ResignPayload{}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal resign request: userId=%v, err=%v", userId, err)
		service.sendErrorMessage(userId, errors.ErrBadRequest)
		return
	}

	game, err := service.gameService.Resign(payload.GameId, userId)
	if err != nil {
		service.sendErrorMessage(userId, err)
		return
	}

	service.broadcastGameState(game)
}

func (service *WebsocketServiceImpl) handleReconnect(userId string, message *WebsocketMessage) {
	log.Printf("Handling reconnect: userId=%v", userId)

	payload := ReconnectPayload{}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal reconnect request: userId=%v, err=%v", userId, err)
		service.sendErrorMessage(userId, errors.ErrBadRequest)
		return
	}

	game, err := service.gameService.Reconnect(payload.GameId, userId)
	if err != nil {
		service.sendErrorMessage(userId, err)
		return
	}

	service.broadcastGameState(game)
}
