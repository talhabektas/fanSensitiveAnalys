package main

import (
	"context"
	"log"
	"time"

	"taraftar-analizi/config"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Konfigürasyon yükle
	config.LoadConfig()
	
	// Veritabanı bağlantısı
	config.ConnectDatabase()
	defer config.DisconnectDatabase()

	collection := config.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Takım isimlerini düzelt
	teamUpdates := map[string]string{
		"Be�ikta�":   "Beşiktaş",
		"Fenerbah�e": "Fenerbahçe",
	}

	for oldName, newName := range teamUpdates {
		filter := bson.M{"name": oldName}
		update := bson.M{
			"$set": bson.M{
				"name":       newName,
				"updated_at": time.Now(),
			},
		}

		result, err := collection.UpdateMany(ctx, filter, update)
		if err != nil {
			log.Printf("Error updating team %s: %v", oldName, err)
			continue
		}

		log.Printf("Updated %d teams: %s -> %s", result.ModifiedCount, oldName, newName)
	}

	log.Println("Team name fix completed!")
}