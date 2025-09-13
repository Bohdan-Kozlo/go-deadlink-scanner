package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl         string
	Port          string
	SessionSecret string
	SessionMaxAge time.Duration
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbUrl := GetEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/linkchecker?sslmode=disable")
	sessionSecret := GetEnv("SESSION_SECRET", "supersecretkey")
	serverPort := GetEnv("SERVER_PORT", "8080")

	return &Config{
		DBUrl:         dbUrl,
		Port:          serverPort,
		SessionSecret: sessionSecret,
		SessionMaxAge: 7 * 24 * time.Hour,
	}
}

func GetEnv(key, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	log.Printf("⚠️ %s not set, using default: %s", key, defaultValue)
	return defaultValue
}
