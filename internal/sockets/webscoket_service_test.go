package sockets

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	service.RegisterClient(client)
	service.UnregisterClient("user1")

	service.mutex.Lock()
	_, ok := service.clients["user1"]
	service.mutex.Unlock()

	assert.False(t, ok)
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
	service.SendMessage("user1", message)

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

	mockGameService.On("CreateGame", "user1", "user2").Return((*game.Game)(nil), errors.New("creation error"))

	event := &events.Event{
		Type: events.EventTypeMatchFound,
		Data: map[string]string{
			"user1Id": "user1",
			"user2Id": "user2",
		},
	}
	service.HandleMatchFound(event)

	mockGameService.AssertCalled(t, "CreateGame", "user1", "user2")
}
