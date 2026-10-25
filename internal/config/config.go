package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
}

func Load() (*Config, error) {
	// Attempt to load .env file if it exists, without failing if missing
	_ = godotenv.Load()

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		if port := os.Getenv("PORT"); port != "" {
			httpAddr = ":" + port
		} else {
			httpAddr = ":8080"
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "agro-supply-erp-secret-key-change-in-production-2026"
	}

	cfg := &Config{
		AppEnv:      appEnv,
		HTTPAddr:    httpAddr,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
	}

	return cfg, nil
}

func (c *Config) ValidateDB() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required but not set")
	}
	return nil
}
