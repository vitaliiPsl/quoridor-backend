package matchmaking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddUserToQueue(t *testing.T) {
	mq := NewInMemoryMatchmakingQueue()

	mq.AddUserToQueue("user1")

	mq.mu.Lock()
	defer mq.mu.Unlock()

	assert.Len(t, mq.queue, 1)
	assert.Equal(t, "user1", mq.queue[0].UserId)
}

func TestRemoveUserFromQueue(t *testing.T) {
	mq := NewInMemoryMatchmakingQueue()

	mq.AddUserToQueue("user1")
	mq.AddUserToQueue("user2")

	mq.RemoveUserFromQueue("user1")

	mq.mu.Lock()
	defer mq.mu.Unlock()

	assert.Len(t, mq.queue, 1)
	assert.Equal(t, "user2", mq.queue[0].UserId)
}

func TestFindMatches(t *testing.T) {
	mq := NewInMemoryMatchmakingQueue()

	mq.AddUserToQueue("user1")
	time.Sleep(200 * time.Millisecond)
	mq.AddUserToQueue("user2")
	mq.AddUserToQueue("user3")

	matches := mq.FindMatches()

	assert.Len(t, matches, 1)
	assert.Equal(t, "user1", matches[0].User1Id)
	assert.Equal(t, "user2", matches[0].User2Id)

	mq.mu.Lock()
	defer mq.mu.Unlock()

	assert.Len(t, mq.queue, 1)
	assert.Equal(t, "user3", mq.queue[0].UserId)
}

func TestFindMatches_notEnoughUsers(t *testing.T) {
	mq := NewInMemoryMatchmakingQueue()

	mq.AddUserToQueue("user1")

	matches := mq.FindMatches()

	assert.Len(t, matches, 0)

	mq.mu.Lock()
	defer mq.mu.Unlock()

	assert.Len(t, mq.queue, 1)
	assert.Equal(t, "user1", mq.queue[0].UserId)
}

func TestFindMatches_multipleMatches(t *testing.T) {
	mq := NewInMemoryMatchmakingQueue()

	mq.AddUserToQueue("user1")
	mq.AddUserToQueue("user2")
	mq.AddUserToQueue("user3")
	mq.AddUserToQueue("user4")

	matches := mq.FindMatches()

	assert.Len(t, matches, 2)
	assert.Equal(t, "user1", matches[0].User1Id)
	assert.Equal(t, "user2", matches[0].User2Id)
	assert.Equal(t, "user3", matches[1].User1Id)
	assert.Equal(t, "user4", matches[1].User2Id)

	mq.mu.Lock()
	defer mq.mu.Unlock()

	assert.Len(t, mq.queue, 0)
}
