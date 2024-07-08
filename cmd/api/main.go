package main

import (
	"log"
	"quoridor/internal/config"
	"quoridor/internal/database"
	"quoridor/internal/router"
	"quoridor/internal/server"
	"quoridor/internal/sockets"
)

func main() {
	log.Println("Starting Quoridory server...")

	cfg := config.ReadConfig()
	_ = database.SetupDatabase(cfg)

	websocketHandler := sockets.NewWebsocketHandler()

	router := router.NewRouter(websocketHandler)
	server.Serve(router)
}
