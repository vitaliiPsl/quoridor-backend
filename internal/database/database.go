package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	host = os.Getenv("DB_HOST")
	port = os.Getenv("DB_PORT")
)

func SetupDatabase() *mongo.Database {
	log.Println("Connecting to the database...")

	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port))

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

	db := client.Database("quoridory_game")
	
	log.Println("Conneced to the database.")
	return db
}
