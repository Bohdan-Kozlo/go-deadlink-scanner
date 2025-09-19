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
	SessionMaxAge     time.Duration
	MaxScannerWorkers int
	EnableTLS         bool
	TLSCertFile       string
	TLSKeyFile        string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbUrl := GetEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/linkchecker?sslmode=disable")
	serverPort := GetEnv("SERVER_PORT", "8080")
	enableTLSStr := GetEnv("ENABLE_TLS", "false")
	tlsCertFile := GetEnv("TLS_CERT_FILE", "cert/dev-cert.pem")
	tlsKeyFile := GetEnv("TLS_KEY_FILE", "cert/dev-key.pem")
	mwStr := GetEnv("MAX_SCANNER_WORKERS", "100")
	mw, err := strconv.Atoi(mwStr)
	if err != nil {
		log.Printf("invalid MAX_SCANNER_WORKERS '%s', fallback to 10", mwStr)
		mw = 10
	}

	return &Config{
		DBUrl:             dbUrl,
		Port:              serverPort,
		SessionMaxAge:     7 * 24 * time.Hour,
		MaxScannerWorkers: mw,
		EnableTLS:         enableTLSStr == "true" || enableTLSStr == "1" || enableTLSStr == "on",
		TLSCertFile:       tlsCertFile,
		TLSKeyFile:        tlsKeyFile,
	}
}

func GetEnv(key, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	log.Printf("⚠️ %s not set, using default: %s", key, defaultValue)
	return defaultValue
}
