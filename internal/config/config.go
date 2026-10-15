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
		httpAddr = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")

	cfg := &Config{
		AppEnv:      appEnv,
		HTTPAddr:    httpAddr,
		DatabaseURL: databaseURL,
	}

	return cfg, nil
}

func (c *Config) ValidateDB() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required but not set")
	}
	return nil
}
