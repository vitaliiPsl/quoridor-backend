package matchmaking

import (
	"math"
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
	queue []*MatchRequest
}

func NewInMemoryMatchmakingQueue() *InMemoryMatchmakingQueue {
	return &InMemoryMatchmakingQueue{
		queue: []*MatchRequest{},
	}
}

func (mq *InMemoryMatchmakingQueue) AddUserToQueue(userId string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	req := MatchRequest{
		UserId:   userId,
		JoinTime: time.Now(),
	}
	mq.queue = append(mq.queue, &req)
}

func (mq *InMemoryMatchmakingQueue) RemoveUserFromQueue(userId string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for idx, req := range mq.queue {
		if req.UserId == userId {
			mq.queue = append(mq.queue[:idx], mq.queue[idx+1:]...)
			return
		}
	}
}

func (mq *InMemoryMatchmakingQueue) FindMatches() []*Match {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	matches := []*Match{}
	if len(mq.queue) < 2 {
		return matches
	}

	slices.SortFunc(mq.queue, func(req1, req2 *MatchRequest) int {
		if req1.JoinTime.Before(req2.JoinTime) {
			return -1
		}
		return 1
	})

	for i := 0; i < len(mq.queue); i++ {
		req1 := mq.queue[i]
		bestMatch := -1
		bestScore := math.MaxFloat64

		for j := i + 1; j < len(mq.queue); j++ {
			req2 := mq.queue[j]
			score := mq.calculateScore(req1, req2)
			if score < bestScore {
				bestMatch = j
				bestScore = score
			}
		}

		if bestMatch != -1 {
			match := &Match{
				User1Id: req1.UserId,
				User2Id: mq.queue[bestMatch].UserId,
			}
			matches = append(matches, match)
			mq.queue = append(mq.queue[:i], mq.queue[i+1:]...)
			if bestMatch > i {
				bestMatch--
			}
			mq.queue = append(mq.queue[:bestMatch], mq.queue[bestMatch+1:]...)
			i--
		}
	}

	return matches
}

func (mq *InMemoryMatchmakingQueue) calculateScore(req1, req2 *MatchRequest) float64 {
	waitTimeDiff := math.Abs(time.Since(req1.JoinTime).Seconds() - time.Since(req2.JoinTime).Seconds())
	return waitTimeDiff
}
