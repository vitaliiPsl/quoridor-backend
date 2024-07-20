package matchmaking

import (
	"math"
	"quoridor/internal/utils"
	"slices"
	"sync"
	"time"
)

type MatchmakingQueue interface {
	FindMatches() []*Match
	AddUserToQueue(userId string)
	RemoveUserFromQueue(userId string)
}

type InMemoryMatchmakingQueue struct {
	mu    sync.Mutex
	queue map[string]*MatchRequest
}

func NewInMemoryMatchmakingQueue() *InMemoryMatchmakingQueue {
	return &InMemoryMatchmakingQueue{
		queue: make(map[string]*MatchRequest),
	}
}

func (mq *InMemoryMatchmakingQueue) AddUserToQueue(userId string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if _, exists := mq.queue[userId]; !exists {
		mq.queue[userId] = &MatchRequest{
			UserId:   userId,
			JoinTime: time.Now(),
		}
	}
}

func (mq *InMemoryMatchmakingQueue) RemoveUserFromQueue(userId string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	delete(mq.queue, userId)
}

func (mq *InMemoryMatchmakingQueue) FindMatches() []*Match {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	matches := []*Match{}
	if len(mq.queue) < 2 {
		return matches
	}

	queue := utils.MapToValuesSlice(mq.queue)
	slices.SortFunc(queue, func(req1, req2 *MatchRequest) int {
		if req1.JoinTime.Before(req2.JoinTime) {
			return -1
		}
		return 1
	})

	for i := 0; i < len(queue); i++ {
		req1 := queue[i]
		bestMatch := -1
		bestScore := math.MaxFloat64

		for j := i + 1; j < len(queue); j++ {
			req2 := queue[j]
			score := mq.calculateScore(req1, req2)
			if score < bestScore {
				bestMatch = j
				bestScore = score
			}
		}

		if bestMatch != -1 {
			match := &Match{
				User1Id: req1.UserId,
				User2Id: queue[bestMatch].UserId,
			}
			matches = append(matches, match)
			queue = append(queue[:i], queue[i+1:]...)
			if bestMatch > i {
				bestMatch--
			}
			queue = append(queue[:bestMatch], queue[bestMatch+1:]...)
			i--
			
			delete(mq.queue, match.User1Id)
			delete(mq.queue, match.User2Id)
		}
	}

	return matches
}

func (mq *InMemoryMatchmakingQueue) calculateScore(req1, req2 *MatchRequest) float64 {
	waitTimeDiff := math.Abs(time.Since(req1.JoinTime).Seconds() - time.Since(req2.JoinTime).Seconds())
	return waitTimeDiff
}
