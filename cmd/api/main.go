package main

import (
	"log"

	"quoridor/internal/database"
	"quoridor/internal/router"
	"quoridor/internal/server"
)

func main() {
	log.Println("Starting Quoridory server...")

	_ = database.SetupDatabase()

	router := router.NewRouter()
	server.Serve(router)
}
