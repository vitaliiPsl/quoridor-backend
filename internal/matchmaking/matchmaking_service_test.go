package matchmaking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"quoridor/internal/events"
)

type MockMatchmakingQueue struct {
	mock.Mock
}

func (m *MockMatchmakingQueue) AddUserToQueue(userId string) {
	m.Called(userId)
}

func (m *MockMatchmakingQueue) RemoveUserFromQueue(userId string) {
	m.Called(userId)
}

func (m *MockMatchmakingQueue) FindMatches() []*Match {
	args := m.Called()
	return args.Get(0).([]*Match)
}

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) Publish(event *events.Event) {
	m.Called(event)
}

func (m *MockEventService) RegisterHandler(eventType events.EventType, handler events.EventHandler) {
	m.Called(eventType, handler)
}

func TestMatchmakingService_AddUser(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	mockEventService := new(MockEventService)
	service := NewMatchmakingService(mockQueue, mockEventService)

	mockQueue.On("AddUserToQueue", "user1").Return()

	service.AddUser("user1")

	mockQueue.AssertCalled(t, "AddUserToQueue", "user1")
}

func TestMatchmakingService_RemoveUser(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	mockEventService := new(MockEventService)
	service := NewMatchmakingService(mockQueue, mockEventService)

	mockQueue.On("RemoveUserFromQueue", "user1").Return()

	service.RemoveUser("user1")

	mockQueue.AssertCalled(t, "RemoveUserFromQueue", "user1")
}

func TestMatchmakingService_StartMatchmaking(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	mockEventService := new(MockEventService)
	service := NewMatchmakingService(mockQueue, mockEventService)

	mockQueue.On("FindMatches").Return([]*Match{
		{User1Id: "user1", User2Id: "user2"},
	}).Once()

	mockQueue.On("AddUserToQueue", "user1").Return()
	mockQueue.On("AddUserToQueue", "user2").Return()

	mockEventService.On("Publish", mock.AnythingOfType("*events.Event")).Return()

	service.StartMatchmaking()
	service.AddUser("user1")
	service.AddUser("user2")

	time.Sleep(1500 * time.Millisecond)

	mockQueue.AssertCalled(t, "FindMatches")
	mockEventService.AssertCalled(t, "Publish", mock.MatchedBy(func(event *events.Event) bool {
		return event.Type == events.EventTypeMatchFound &&
			event.Data.(map[string]string)["user1Id"] == "user1" &&
			event.Data.(map[string]string)["user2Id"] == "user2"
	}))
}

func TestMatchmakingService_notifyAboutMatch(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	mockEventService := new(MockEventService)
	service := NewMatchmakingService(mockQueue, mockEventService)

	match := &Match{User1Id: "user1", User2Id: "user2"}

	mockEventService.On("Publish", mock.AnythingOfType("*events.Event")).Return()

	service.notifyAboutMatch(match)

	mockEventService.AssertCalled(t, "Publish", mock.MatchedBy(func(event *events.Event) bool {
		return event.Type == events.EventTypeMatchFound &&
			event.Data.(map[string]string)["user1Id"] == "user1" &&
			event.Data.(map[string]string)["user2Id"] == "user2"
	}))
}
