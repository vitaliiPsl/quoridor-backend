package matchmaking

import (
	"log"
	"quoridor/internal/events"
	"time"
)

type MatchmakingService interface {
	AddUser(userId string)
	RemoveUser(userId string)
	StartMatchmaking()
}

type MatchmakingServiceImpl struct {
	queue        MatchmakingQueue
	eventService events.EventService
}

func NewMatchmakingService(queue MatchmakingQueue, eventService events.EventService) *MatchmakingServiceImpl {
	return &MatchmakingServiceImpl{
		queue:        queue,
		eventService: eventService,
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
		service.notifyAboutMatch(match)
	}
}

func (service *MatchmakingServiceImpl) notifyAboutMatch(match *Match) {
	log.Printf("Found match: match=%v", match)

	matchEvent := &events.Event{
		Type: events.EventTypeMatchFound,
		Data: map[string]string{
			"user1Id": match.User1Id,
			"user2Id": match.User2Id,
		},
	}
	service.eventService.Publish(matchEvent)
}
