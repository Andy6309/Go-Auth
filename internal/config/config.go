package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	JWTSecret  string
	JWTExpiry  time.Duration
	DBPath     string
}

func LoadConfig() *Config {
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Set default values
	config := &Config{
		ServerPort: getEnv("PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "default_jwt_secret"),
		DBPath:     getEnv("DB_PATH", "./db-data/auth.db"),
	}

	// Parse JWT expiration
	expiry, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRATION format. Use format like '24h', '1h30m', etc.")
	}
	config.JWTExpiry = expiry

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}