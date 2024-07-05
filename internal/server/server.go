package server

import (
	"log"
	"net/http"
	"time"

	"quoridor/internal/router"
)

type Server interface {
}

type ServerImpl struct {
}

func Serve(router router.Router) {
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router.Handler(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Cannot start server: %s", err)
	}
}
