package core

import (
	"log"
	"os"
	"strconv"

	"pharmacyclaims/internal/database"
)

type Config struct {
	Database      database.Connection
	Port          int
	DataDir       string
	LogDir        string
	MigrationsDir string
}

func LoadConfig() Config {
	config := Config{
		Database: database.Connection{
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvIntWithDefault("DB_PORT", 5432),
			User:     getEnvWithDefault("DB_USER", "pharmacy_user"),
			Password: getEnvWithDefault("DB_PASSWORD", "pharmacy_password"),
			DBName:   getEnvWithDefault("DB_NAME", "pharmacy_claims"),
			SSLMode:  getEnvWithDefault("DB_SSLMODE", "disable"),
		},
		Port:          getEnvIntWithDefault("PORT", 8080),
		DataDir:       getEnvWithDefault("DATA_DIR", "./data"),
		LogDir:        getEnvWithDefault("LOG_DIR", "./logs"),
		MigrationsDir: getEnvWithDefault("MIGRATIONS_DIR", "./migrations"),
	}

	return config
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default %d", key, value, defaultValue)
	}
	return defaultValue
}
