package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	BcryptCost  int
}

// Load reads configuration from environment variables and returns a Config.
// It returns an error if any required variable is missing.
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	bcryptCost := 12
	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		cost, err := strconv.Atoi(costStr)
		if err != nil {
			return nil, fmt.Errorf("BCRYPT_COST must be an integer: %w", err)
		}
		if cost < 10 || cost > 20 {
			return nil, fmt.Errorf("BCRYPT_COST must be between 10 and 20")
		}
		bcryptCost = cost
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		BcryptCost:  bcryptCost,
	}, nil
}
