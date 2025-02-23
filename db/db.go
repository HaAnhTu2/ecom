package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGOURI")))
	if err != nil {
		log.Fatal("Error: " + err.Error())
	}
	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error: " + err.Error())
	}
	log.Println("Connected to MongoDB")
	return client
}
