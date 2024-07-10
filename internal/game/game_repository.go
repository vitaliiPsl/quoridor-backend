package game

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GameRepository interface {
	SaveGame(state *Game) error
	GetGameById(gameID string) (*Game, error)
	GetGamesByStatus(status GameStatus) ([]*Game, error)
}

type MongoGameRepository struct {
	database   *mongo.Database
	collection *mongo.Collection
}

func NewGameRepository() GameRepository {
	return &MongoGameRepository{}
}

func NewMongoGameRepository(database *mongo.Database, collectionName string) *MongoGameRepository {
	collection := database.Collection(collectionName)
	return &MongoGameRepository{
		database:   database,
		collection: collection,
	}
}

func (r *MongoGameRepository) SaveGame(state *Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": state.GameId},
		state,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		log.Printf("Error saving game state: %v", err)
		return err
	}
	return nil
}

func (r *MongoGameRepository) GetGameById(gameID string) (*Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var state Game
	err := r.collection.FindOne(ctx, bson.M{"_id": gameID}).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Printf("Error loading game state: %v", err)
		return nil, err
	}
	return &state, nil
}

func (r *MongoGameRepository) GetGamesByStatus(status GameStatus) ([]*Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		log.Printf("Error loading games by status: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	games := []*Game{}
	for cursor.Next(ctx) {
		var game Game
		err := cursor.Decode(&game)
		if err != nil {
			log.Printf("Error decoding game: %v", err)
			continue
		}
		games = append(games, &game)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		return nil, err
	}

	return games, nil
}
