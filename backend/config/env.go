package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	GinMode            string
	APISecret          string
	MongoURI           string
	MongoDatabase      string
	RedditClientID     string
	RedditClientSecret string
	RedditUsername     string
	RedditPassword     string
	HuggingFaceToken   string
	YouTubeAPIKey      string
	FrontendURL        string
	N8NWebhookURL      string
	GroqAPIKey         string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	AppConfig = &Config{
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "release"),
		APISecret:          getEnv("API_SECRET", "default_secret_change_this"),
		MongoURI:           getEnv("MONGODB_URI", ""),
		MongoDatabase:      getEnv("MONGODB_DATABASE", "taraftar_analizi"),
		RedditClientID:     getEnv("REDDIT_CLIENT_ID", ""),
		RedditClientSecret: getEnv("REDDIT_CLIENT_SECRET", ""),
		RedditUsername:     getEnv("REDDIT_USERNAME", ""),
		RedditPassword:     getEnv("REDDIT_PASSWORD", ""),
		HuggingFaceToken:   getEnv("HUGGINGFACE_TOKEN", ""),
		YouTubeAPIKey:      getEnv("YOUTUBE_API_KEY", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		N8NWebhookURL:      getEnv("N8N_WEBHOOK_URL", "http://localhost:5678/webhook"),
		GroqAPIKey:         getEnv("GROQ_API_KEY", ""),
	}

	validateConfig()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func validateConfig() {
	if AppConfig.MongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is required")
	}
	
	if AppConfig.HuggingFaceToken == "" {
		log.Println("Warning: HUGGINGFACE_TOKEN not set, sentiment analysis will not work")
	}
	
	if AppConfig.GroqAPIKey == "" {
		log.Println("Warning: GROQ_API_KEY not set, Grok AI features will not work")
	}
	
	if AppConfig.RedditClientID == "" || AppConfig.RedditClientSecret == "" {
		log.Println("Warning: Reddit credentials not set, Reddit data collection will not work")
	}
}