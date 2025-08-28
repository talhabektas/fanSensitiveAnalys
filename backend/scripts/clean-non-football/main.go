package main

import (
	"context"
	"log"
	"strings"
	"time"

	"taraftar-analizi/config"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	config.LoadConfig()
	config.ConnectDatabase()
	defer config.DisconnectDatabase()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("Cleaning non-football comments from database...")

	collection := config.GetCollection("comments")
	
	// Futbol dışı anahtar kelimeler
	nonFootballKeywords := []string{
		"saç kesimi", "saç şekli", "tıraş", "berber", "keko",
		"hükümet", "politika", "akp", "chp", "mhp", 
		"ekonomi", "enflasyon", "dolar", "euro",
		"askerlik", "bedelli", "muvazzaf",
		"eğitim", "okul", "üniversite",
		"covid", "aşı", "maske", "salgın",
		"ukrayna", "rusya", "savaş", "putin",
		"iran", "suriye", "irak", "afganistan",
		"terör", "pkk", "ypg", "isis",
		"kültür", "sanat", "müzik", "film",
		"teknoloji", "telefon", "bilgisayar",
		"yasaklama", "yasaklandı", "yasak",
		"çocuk", "anne", "baba", "aile",
	}

	// Pozitif futbol anahtar kelimeler
	footballKeywords := []string{
		"galatasaray", "fenerbahçe", "beşiktaş", "trabzonspor",
		"futbol", "maç", "gol", "lig", "takım", "oyuncu",
		"derbi", "şampiyonluk", "kupa", "uefa", "champions",
		"transfer", "teknik", "direktör", "antrenör",
		"stadyum", "taraftar", "tezahürat", "tribün",
		"puan", "skor", "hakem", "ofsayt", "penaltı",
		"milli takım", "euro", "dünya kupası",
		"süper lig", "tff", "beinsports",
	}

	// Silinecek yorumları bul
	var deletedCount int64

	// Non-football keywords içeren yorumları bul
	for _, keyword := range nonFootballKeywords {
		filter := bson.M{
			"text": bson.M{"$regex": keyword, "$options": "i"},
		}
		
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			log.Printf("Error deleting comments with keyword '%s': %v", keyword, err)
			continue
		}
		
		if result.DeletedCount > 0 {
			log.Printf("Deleted %d comments containing '%s'", result.DeletedCount, keyword)
			deletedCount += result.DeletedCount
		}
	}

	// Çok kısa yorumları sil (< 10 karakter)
	shortFilter := bson.M{
		"$expr": bson.M{
			"$lt": bson.A{bson.M{"$strLenCP": "$text"}, 10},
		},
	}
	
	result, err := collection.DeleteMany(ctx, shortFilter)
	if err != nil {
		log.Printf("Error deleting short comments: %v", err)
	} else {
		log.Printf("Deleted %d short comments", result.DeletedCount)
		deletedCount += result.DeletedCount
	}

	// Futbol anahtar kelimesi içermeyen yorumları sil
	footballRegex := strings.Join(footballKeywords, "|")
	noFootballFilter := bson.M{
		"text": bson.M{
			"$not": bson.M{
				"$regex": footballRegex, 
				"$options": "i",
			},
		},
	}
	
	result, err = collection.DeleteMany(ctx, noFootballFilter)
	if err != nil {
		log.Printf("Error deleting non-football comments: %v", err)
	} else {
		log.Printf("Deleted %d comments without football keywords", result.DeletedCount)
		deletedCount += result.DeletedCount
	}

	log.Printf("Database cleanup completed! Total deleted: %d comments", deletedCount)
}