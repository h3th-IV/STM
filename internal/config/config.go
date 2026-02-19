package config

import (
	"os"
	"strconv"
)

// Config holds application configuration loaded from environment.
type Config struct {
	Port         int
	JWTSecret    string
	JWTIssuer    string
	JWTExpiry    int    // minutes
	RefreshExpiry int    // days
	DBPath       string
	Env          string
}

// Load reads configuration from environment variables.
func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "15"))
	refreshExpiry, _ := strconv.Atoi(getEnv("REFRESH_EXPIRY_DAYS", "7"))

	return &Config{
		Port:          port,
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTIssuer:     getEnv("JWT_ISSUER", "stm-api"),
		JWTExpiry:     jwtExpiry,
		RefreshExpiry: refreshExpiry,
		DBPath:        getEnv("DB_PATH", "./data/tasks.db"),
		Env:           getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
