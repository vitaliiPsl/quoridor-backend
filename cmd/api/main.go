package main

import (
	"log"

	"quoridor/internal/config"
	"quoridor/internal/database"
	"quoridor/internal/router"
	"quoridor/internal/server"
)

func main() {
	log.Println("Starting Quoridory server...")

	cfg := config.ReadConfig()
	_ = database.SetupDatabase(cfg)

	router := router.NewRouter()
	server.Serve(router)
}
