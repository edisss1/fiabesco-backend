package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
)

var (
	Client       *mongo.Client
	Once         sync.Once
	Database     *mongo.Database
	DatabaseName = "fiabesco"
)

func Connect() {
	Once.Do(func() {
		mongodbURI := os.Getenv("MONGODB_URI")
		clientOptions := options.Client().ApplyURI(mongodbURI)

		var err error
		Client, err = mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		if err = Client.Ping(context.Background(), nil); err != nil {
			log.Fatalf("MongoDB not reachable: %v", err)
		}

		Database = Client.Database(DatabaseName)
		log.Println("✅ Connected to MongoDB")
	})
}
