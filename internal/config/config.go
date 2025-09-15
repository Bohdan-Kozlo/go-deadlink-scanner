package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl             string
	Port              string
	SessionSecret     string
	SessionMaxAge     time.Duration
	MaxScannerWorkers int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbUrl := GetEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/linkchecker?sslmode=disable")
	sessionSecret := GetEnv("SESSION_SECRET", "supersecretkey")
	serverPort := GetEnv("SERVER_PORT", "8080")
	mwStr := GetEnv("MAX_SCANNER_WORKERS", "100")
	mw, err := strconv.Atoi(mwStr)
	if err != nil {
		log.Printf("invalid MAX_SCANNER_WORKERS '%s', fallback to 10", mwStr)
		mw = 10
	}

	return &Config{
		DBUrl:             dbUrl,
		Port:              serverPort,
		SessionSecret:     sessionSecret,
		SessionMaxAge:     7 * 24 * time.Hour,
		MaxScannerWorkers: mw,
	}
}

func GetEnv(key, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	log.Printf("⚠️ %s not set, using default: %s", key, defaultValue)
	return defaultValue
}
