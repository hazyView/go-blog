package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for our application
type Config struct {
	Port           string
	DatabaseURL    string
	DatabaseHost   string
	DatabasePort   string
	DatabaseUser   string
	DatabasePass   string
	DatabaseName   string
	LogLevel       string
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int
	MaxConnections int
}

// Load returns a new config struct
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DatabaseHost:   getEnv("DB_HOST", "localhost"),
		DatabasePort:   getEnv("DB_PORT", "5432"),
		DatabaseUser:   getEnv("DB_USER", "postgres"),
		DatabasePass:   getEnv("DB_PASSWORD", ""),
		DatabaseName:   getEnv("DB_NAME", "blog_api"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		ReadTimeout:    getEnvAsInt("READ_TIMEOUT", 10),
		WriteTimeout:   getEnvAsInt("WRITE_TIMEOUT", 10),
		IdleTimeout:    getEnvAsInt("IDLE_TIMEOUT", 120),
		MaxConnections: getEnvAsInt("MAX_DB_CONNECTIONS", 25),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}
