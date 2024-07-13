package matchmaking

import (
	"log"
	"time"
)

type MatchmakingService interface {
	AddUser(userId string)
	RemoveUser(userId string)
	StartMatchmaking()
}

type MatchmakingServiceImpl struct {
	queue MatchmakingQueue
}

func NewMatchmakingService() *MatchmakingServiceImpl {
	return &MatchmakingServiceImpl{
		queue: NewInMemoryMatchmakingQueue(),
	}
}

func (service *MatchmakingServiceImpl) AddUser(userId string) {
	log.Printf("Adding user to the matchmaking queue: userId=%v", userId)
	service.queue.AddUserToQueue(userId)
}

func (service *MatchmakingServiceImpl) RemoveUser(userId string) {
	log.Printf("Removing user from the matchmaking queue: userId=%v", userId)
	service.queue.RemoveUserFromQueue(userId)
}

func (service *MatchmakingServiceImpl) StartMatchmaking() {
	log.Println("Starting matchmaking goroutine...")
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			service.matchUsers()
		}
	}()
}

func (service *MatchmakingServiceImpl) matchUsers() {
	matches := service.queue.FindMatches()

	for _, match := range matches {
		log.Printf("Found match: match=%v", match)
	}
}
