package main

import (
	"log"
	"quoridor/internal/config"
	"quoridor/internal/database"
	"quoridor/internal/events"
	"quoridor/internal/game"
	"quoridor/internal/matchmaking"
	"quoridor/internal/router"
	"quoridor/internal/server"
	"quoridor/internal/sockets"
)

func main() {
	log.Println("Starting Quoridory server...")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.ReadConfig()
	database := database.SetupDatabase(cfg)

	eventService := events.NewEventService()

	gameEngine := game.NewGameEngine()
	gameRepository := game.NewMongoGameRepository(database, "games")
	gameService := game.NewGameService(gameEngine, gameRepository)

	mmQueue := matchmaking.NewInMemoryMatchmakingQueue()
	mmService := matchmaking.NewMatchmakingService(mmQueue, eventService)
	mmService.StartMatchmaking()

	websocketService := sockets.NewWebsocketService(mmService, gameService)
	websocketHandler := sockets.NewWebsocketHandler(websocketService)

	eventService.RegisterHandler(events.EventTypeMatchFound, websocketService.HandleMatchFound)

	router := router.NewRouter(websocketHandler)
	server.Serve(router)
}
