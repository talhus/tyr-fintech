package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	RabbitmqURL       string
	JWTSecret         string
	CardEncryptionKey string

	// Environment
	Env string

	// API
	APIHost string
	APIPort string

	// Frontend / Checkout URL
	FrontendURL string

	// CORS Whitelist
	AllowedOrigins []string
}

func New() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}
	return &Config{
		Env:               getEnv("ENV", "development"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://admin:secretpassword@localhost:5432/fintech?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", getEnv("REDIS_ADDR", "redis:6379")),
		RabbitmqURL:       getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:         getEnv("JWT_SECRET", "my_secret_key"),
		APIHost:           getEnv("API_HOST", "localhost"),
		APIPort:           getEnv("PORT", getEnv("API_PORT", "8081")),
		CardEncryptionKey: getEnv("CARD_ENCRYPTION_KEY", "12345678901234567890123456789012"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3005"),
		AllowedOrigins:    parseCommaSeparated(getEnv("ALLOWED_ORIGINS", "http://localhost:3005,http://localhost:3000")),
	}
}

func parseCommaSeparated(val string) []string {
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
