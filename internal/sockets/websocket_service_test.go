package sockets

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"quoridor/internal/errors"
	"quoridor/internal/events"
	"quoridor/internal/game"
)

type MockMatchmakingService struct {
	mock.Mock
}

func (m *MockMatchmakingService) AddUser(userId string) {
	m.Called(userId)
}

func (m *MockMatchmakingService) RemoveUser(userId string) {
	m.Called(userId)
}

func (m *MockMatchmakingService) StartMatchmaking() {
	m.Called()
}

type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) CreateGame(user1Id, user2Id string) (*game.Game, error) {
	args := m.Called(user1Id, user2Id)
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameService) GetGameById(gameId string) (*game.Game, error) {
	args := m.Called(gameId)
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameService) GetPendingGames() ([]*game.Game, error) {
	args := m.Called()
	return args.Get(0).([]*game.Game), args.Error(1)
}

func (m *MockGameService) MakeMove(gameId, userId string, newPos *game.Position) (*game.Game, error) {
	args := m.Called(gameId, userId, newPos)
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameService) PlaceWall(gameId, userId string, wall *game.Wall) (*game.Game, error) {
	args := m.Called(gameId, userId, wall)
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameService) Resign(gameId, userId string) (*game.Game, error) {
	args := m.Called(gameId, userId)
	return args.Get(0).(*game.Game), args.Error(1)
}

func TestRegisterClient(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 8),
	}
	service.RegisterClient(client)

	service.mutex.Lock()
	registeredClient, ok := service.clients["user1"]
	service.mutex.Unlock()

	assert.True(t, ok)
	assert.Equal(t, client, registeredClient)
}

func TestUnregisterClient(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 8),
	}

	mockMMService.On("RemoveUser", "user1").Return()

	service.RegisterClient(client)
	service.UnregisterClient("user1")

	service.mutex.Lock()
	_, ok := service.clients["user1"]
	service.mutex.Unlock()

	assert.False(t, ok)
	mockMMService.AssertCalled(t, "RemoveUser", "user1")
}

func TestSendMessage(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client)

	message := &WebsocketMessage{Type: "test", Payload: []byte("test")}
	service.sendMessage("user1", message)

	receivedMessage := <-client.messages
	assert.Equal(t, message, receivedMessage)
}

func TestHandleMessage_StartGame(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	mockMMService.On("AddUser", "user1").Return()

	message := &WebsocketMessage{Type: EventTypeStartGame}
	service.HandleMessage("user1", message)

	mockMMService.AssertCalled(t, "AddUser", "user1")
}

func TestHandleMessage_UnknownType(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client)

	message := &WebsocketMessage{Type: "unknown"}
	service.HandleMessage("user1", message)

	errorMessage := <-client.messages
	assert.Equal(t, EventTypeError, errorMessage.Type)

	var errorPayload ErrorMessagePayload
	err := json.Unmarshal(errorMessage.Payload, &errorPayload)
	assert.NoError(t, err)
	assert.Equal(t, errors.ErrBadRequest.Error(), errorPayload.ErrorType)
}

func TestHandleMatchFound(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	game := &game.Game{
		GameId: "game1",
		Player1: &game.Player{
			UserId: "user1",
		},
		Player2: &game.Player{
			UserId: "user2",
		},
	}
	mockGameService.On("CreateGame", "user1", "user2").Return(game, nil)

	client1 := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	client2 := &Client{
		userId:   "user2",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client1)
	service.RegisterClient(client2)

	event := &events.Event{
		Type: events.EventTypeMatchFound,
		Data: map[string]string{
			"user1Id": "user1",
			"user2Id": "user2",
		},
	}
	service.HandleMatchFound(event)

	payload, err := json.Marshal(game)
	assert.NoError(t, err)

	expectedMessage := &WebsocketMessage{Type: EventTypeGameState, Payload: payload}

	receivedMessage1 := <-client1.messages
	receivedMessage2 := <-client2.messages

	assert.Equal(t, expectedMessage, receivedMessage1)
	assert.Equal(t, expectedMessage, receivedMessage2)

	mockGameService.AssertCalled(t, "CreateGame", "user1", "user2")
}

func TestHandleMatchFound_GameCreationError(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	mockGameService.On("CreateGame", "user1", "user2").Return((*game.Game)(nil), errors.ErrInternalError)

	client1 := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	client2 := &Client{
		userId:   "user2",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client1)
	service.RegisterClient(client2)

	event := &events.Event{
		Type: events.EventTypeMatchFound,
		Data: map[string]string{
			"user1Id": "user1",
			"user2Id": "user2",
		},
	}
	service.HandleMatchFound(event)

	errorMessage1 := <-client1.messages
	errorMessage2 := <-client2.messages

	assert.Equal(t, EventTypeError, errorMessage1.Type)
	assert.Equal(t, EventTypeError, errorMessage2.Type)

	mockGameService.AssertCalled(t, "CreateGame", "user1", "user2")
}

func TestHandleMove_InvalidPayload(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client)

	message := &WebsocketMessage{Type: EventTypeMakeMove, Payload: []byte("invalid")}
	service.HandleMessage("user1", message)

	errorMessage := <-client.messages
	assert.Equal(t, EventTypeError, errorMessage.Type)

	var errorPayload ErrorMessagePayload
	err := json.Unmarshal(errorMessage.Payload, &errorPayload)
	assert.NoError(t, err)
	assert.Equal(t, errors.ErrBadRequest.Error(), errorPayload.ErrorType)
}

func TestHandleMove_MakeMoveError(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client)

	mockGameService.On("MakeMove", "game1", "user1", mock.Anything).Return((*game.Game)(nil), errors.ErrInvalidMove)

	payload := MakeMovePayload{
		GameId:   "game1",
		Position: game.Position{X: 1, Y: 1},
	}
	payloadBytes, _ := json.Marshal(payload)
	message := &WebsocketMessage{Type: EventTypeMakeMove, Payload: payloadBytes}
	service.HandleMessage("user1", message)

	errorMessage := <-client.messages
	assert.Equal(t, EventTypeError, errorMessage.Type)

	var errorPayload ErrorMessagePayload
	err := json.Unmarshal(errorMessage.Payload, &errorPayload)
	assert.NoError(t, err)
	assert.Equal(t, errors.ErrInvalidMove.Error(), errorPayload.ErrorType)
}

func TestHandlePlaceWall_InvalidPayload(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client)

	message := &WebsocketMessage{Type: EventTypePlaceWall, Payload: []byte("invalid")}
	service.HandleMessage("user1", message)

	errorMessage := <-client.messages
	assert.Equal(t, EventTypeError, errorMessage.Type)

	var errorPayload ErrorMessagePayload
	err := json.Unmarshal(errorMessage.Payload, &errorPayload)
	assert.NoError(t, err)
	assert.Equal(t, errors.ErrBadRequest.Error(), errorPayload.ErrorType)
}

func TestHandlePlaceWall_PlaceWallError(t *testing.T) {
	mockMMService := new(MockMatchmakingService)
	mockGameService := new(MockGameService)
	service := NewWebsocketService(mockMMService, mockGameService)

	client := &Client{
		userId:   "user1",
		messages: make(chan *WebsocketMessage, 1),
	}
	service.RegisterClient(client)

	mockGameService.On("PlaceWall", "game1", "user1", mock.Anything).Return((*game.Game)(nil), errors.ErrInvalidWallPlacement)

	payload := PlaceWallPayload{
		GameId: "game1",
		Wall: game.Wall{
			Direction: game.Horizontal,
			Pos1:      &game.Position{X: 1, Y: 1},
			Pos2:      &game.Position{X: 1, Y: 2},
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	message := &WebsocketMessage{Type: EventTypePlaceWall, Payload: payloadBytes}
	service.HandleMessage("user1", message)

	errorMessage := <-client.messages
	assert.Equal(t, EventTypeError, errorMessage.Type)

	var errorPayload ErrorMessagePayload
	err := json.Unmarshal(errorMessage.Payload, &errorPayload)
	assert.NoError(t, err)
	assert.Equal(t, errors.ErrInvalidWallPlacement.Error(), errorPayload.ErrorType)
}
