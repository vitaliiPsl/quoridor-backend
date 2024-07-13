package matchmaking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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

func TestMatchmakingService_AddUser(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	service := &MatchmakingServiceImpl{queue: mockQueue}

	mockQueue.On("AddUserToQueue", "user1").Return()

	service.AddUser("user1")

	mockQueue.AssertCalled(t, "AddUserToQueue", "user1")
}

func TestMatchmakingService_RemoveUser(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	service := &MatchmakingServiceImpl{queue: mockQueue}

	mockQueue.On("RemoveUserFromQueue", "user1").Return()

	service.RemoveUser("user1")

	mockQueue.AssertCalled(t, "RemoveUserFromQueue", "user1")
}

func TestMatchmakingService_StartMatchmaking(t *testing.T) {
	mockQueue := new(MockMatchmakingQueue)
	service := &MatchmakingServiceImpl{queue: mockQueue}

	mockQueue.On("FindMatches").Return([]*Match{
		{User1Id: "user1", User2Id: "user2"},
	}).Once()

	mockQueue.On("AddUserToQueue", "user1").Return()
	mockQueue.On("AddUserToQueue", "user2").Return()

	service.StartMatchmaking()
	service.AddUser("user1")
	service.AddUser("user2")

	time.Sleep(1500 * time.Millisecond)

	mockQueue.AssertCalled(t, "FindMatches")
}