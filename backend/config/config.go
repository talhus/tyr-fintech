package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	RabbitmqURL string
	JWTSecret   string

	// API
	APIHost string
	APIPort string
}

func New() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://admin:secretpassword@localhost:5432/fintech?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis:6379"),
		RabbitmqURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:   getEnv("JWT_SECRET", "my_secret_key"),
		APIHost:     getEnv("API_HOST", "localhost"),
		APIPort:     getEnv("API_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
