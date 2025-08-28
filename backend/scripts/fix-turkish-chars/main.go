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
	// Konfigürasyon yükle
	config.LoadConfig()
	
	// Veritabanı bağlantısı
	config.ConnectDatabase()
	defer config.DisconnectDatabase()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("Starting Turkish character fix for all collections...")

	// 1. Teams koleksiyonunu düzelt
	fixTeamsCollection(ctx)
	
	// 2. Comments koleksiyonunu düzelt
	fixCommentsCollection(ctx)

	log.Println("Turkish character fix completed!")
}

func fixTeamsCollection(ctx context.Context) {
	collection := config.GetCollection("teams")
	
	// Tüm takımları çek
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error finding teams: %v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var team bson.M
		if err := cursor.Decode(&team); err != nil {
			continue
		}

		update := bson.M{}
		needsUpdate := false

		// Name alanını kontrol et
		if name, ok := team["name"].(string); ok {
			fixedName := fixTurkishChars(name)
			if fixedName != name {
				update["name"] = fixedName
				needsUpdate = true
			}
		}

		// Keywords alanını kontrol et
		if keywords, ok := team["keywords"].([]interface{}); ok {
			fixedKeywords := make([]string, len(keywords))
			keywordsChanged := false
			
			for i, keyword := range keywords {
				if keywordStr, ok := keyword.(string); ok {
					fixedKeyword := fixTurkishChars(keywordStr)
					fixedKeywords[i] = fixedKeyword
					if fixedKeyword != keywordStr {
						keywordsChanged = true
					}
				}
			}
			
			if keywordsChanged {
				update["keywords"] = fixedKeywords
				needsUpdate = true
			}
		}

		if needsUpdate {
			update["updated_at"] = time.Now()
			filter := bson.M{"_id": team["_id"]}
			
			_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": update})
			if err != nil {
				log.Printf("Error updating team %v: %v", team["_id"], err)
			} else {
				log.Printf("Fixed team: %s", update["name"])
			}
		}
	}
}

func fixCommentsCollection(ctx context.Context) {
	collection := config.GetCollection("comments")
	
	// Tüm yorumları çek
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error finding comments: %v", err)
		return
	}
	defer cursor.Close(ctx)

	updateCount := 0
	for cursor.Next(ctx) {
		var comment bson.M
		if err := cursor.Decode(&comment); err != nil {
			continue
		}

		update := bson.M{}
		needsUpdate := false

		// Text alanını kontrol et
		if text, ok := comment["text"].(string); ok {
			fixedText := fixTurkishChars(text)
			if fixedText != text {
				update["text"] = fixedText
				needsUpdate = true
			}
		}

		// Author alanını kontrol et
		if author, ok := comment["author"].(string); ok {
			fixedAuthor := fixTurkishChars(author)
			if fixedAuthor != author {
				update["author"] = fixedAuthor
				needsUpdate = true
			}
		}

		if needsUpdate {
			update["updated_at"] = time.Now()
			filter := bson.M{"_id": comment["_id"]}
			
			_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": update})
			if err != nil {
				log.Printf("Error updating comment %v: %v", comment["_id"], err)
			} else {
				updateCount++
				if updateCount%10 == 0 {
					log.Printf("Fixed %d comments...", updateCount)
				}
			}
		}
	}
	
	log.Printf("Fixed total %d comments", updateCount)
}

// Türkçe karakterleri düzelten fonksiyon
func fixTurkishChars(text string) string {
	// Bozuk karakterleri düzelt
	replacements := map[string]string{
		// Ç karakteri
		"Ã§": "ç",
		"Ã‡": "Ç",
		
		// Ğ karakteri
		"Ä": "ğ",
		"ÄŸ": "ğ",
		
		// I, İ, ı karakterleri
		"Ä±": "ı",
		"Ä°": "İ",
		
		// Ö karakteri
		"Ã¶": "ö",
		"Ã–": "Ö",
		
		// Ş karakteri
		"Å": "ş",
		"Åž": "ş",
		"ÅŸ": "ş",
		
		// Ü karakteri
		"Ã¼": "ü",
		"Ãœ": "Ü",
		
		// Diğer bozuk karakterler
		"â€™": "'",
		"â€œ": "\"",
		"â€": "\"",
		"â": "",
		"€": "",
		"™": "",
		"Â": "",
		"³": "",
		"¼": "ü",
		"¶": "ö",
		"±": "ı",
		"°": "İ",
		"�": "",
		
		// Belirli kelimeler için özel düzeltmeler
		"Beikta": "Beşiktaş",
		"Fenerbahe": "Fenerbahçe",
		"krkl": "kırıklığı",
		"yaratyor": "yaratıyor",
		"dnemi": "dönemi",
		"gerekten": "gerçekten",
		"baarl": "başarılı",
		"umutlarm": "umutlarım",
		"artt": "arttı",
		"mthi": "müthiş",
		"sper": "süper",
		"kampanyas": "kampanyası",
	}
	
	result := text
	for broken, fixed := range replacements {
		result = strings.ReplaceAll(result, broken, fixed)
	}
	
	return result
}