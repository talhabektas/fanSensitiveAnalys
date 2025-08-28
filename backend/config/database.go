package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB     *mongo.Database
	Client *mongo.Client
)

func ConnectDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(AppConfig.MongoURI)
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMinPoolSize(5)
	clientOptions.SetMaxConnIdleTime(30 * time.Second)
	clientOptions.SetServerSelectionTimeout(5 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second)
	
	// UTF-8 encoding için
	clientOptions.SetReadPreference(nil)
	clientOptions.SetWriteConcern(nil)

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	DB = Client.Database(AppConfig.MongoDatabase)
	log.Printf("Successfully connected to MongoDB database: %s", AppConfig.MongoDatabase)

	createIndexes()
}

func DisconnectDatabase() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("Successfully disconnected from MongoDB")
		}
	}
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}

func createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	commentsCollection := GetCollection("comments")
	
	_, err := commentsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"source_id": 1,
				"source":    1,
			},
		},
		{
			Keys: map[string]interface{}{
				"created_at": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"team_id": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"sentiment.label": 1,
			},
		},
	})
	
	if err != nil {
		log.Printf("Warning: Failed to create indexes for comments: %v", err)
	}

	sentimentsCollection := GetCollection("sentiments")
	
	_, err = sentimentsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"comment_id": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"team_id": 1,
				"created_at": -1,
			},
		},
	})
	
	if err != nil {
		log.Printf("Warning: Failed to create indexes for sentiments: %v", err)
	}

	teamsCollection := GetCollection("teams")
	
	_, err = teamsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{
			"slug": 1,
		},
	})
	
	if err != nil {
		log.Printf("Warning: Failed to create index for teams: %v", err)
	}

	log.Println("Database indexes created successfully")
}

func HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return Client.Ping(ctx, nil)
}