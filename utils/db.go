package utils

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var UsersCollection *mongo.Collection

func InitDB() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set in environment")
	}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Unable to connect to MongoDB: %v", err)
	}
	MongoClient = client
	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "scratchclone"
	}
	UsersCollection = client.Database(dbName).Collection("users")
}
