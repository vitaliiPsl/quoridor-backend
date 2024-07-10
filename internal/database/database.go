package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"quoridor/internal/config"
)

func SetupDatabase(config *config.Config) *mongo.Database {
	log.Println("Connecting to the database...")

	clientOptions := options.Client().ApplyURI(config.DatabaseURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database("quoridor_game")

	log.Println("Connected to the database.")
	return db
}
