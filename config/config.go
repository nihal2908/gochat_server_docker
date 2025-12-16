package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds configuration values
type Config struct {
	ServerPort string
	DbURI      string
	// Add other configurations like Firebase, JWT secret, etc.
}

var Cfg Config

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	// Load configurations
	Cfg.ServerPort = getEnv("PORT", ":8080")
	Cfg.DbURI = getEnv("DB_URL", "mongodb://localhost:27017/whatsapp_clone")
	// Load other configuration variables as needed
}

// getEnv gets the value of an environment variable, or a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
