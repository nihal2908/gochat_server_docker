package db

import (
	"gochat_server/config"
	"log"

	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB holds the database connection
var DB *mongo.Database

// ConnectDB initializes the database connection
func ConnectDB() {
	clientOptions := options.Client().ApplyURI(config.Cfg.DbURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	DB = client.Database("whatsapp_clone")
}

// Example of a simple query
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}

func GetDB() *mongo.Database {
	return DB
}