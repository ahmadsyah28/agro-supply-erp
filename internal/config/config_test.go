package config

import (
	"os"
	"testing"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clear relevant env vars
	os.Unsetenv("APP_ENV")
	os.Unsetenv("HTTP_ADDR")
	os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("expected AppEnv 'development', got '%s'", cfg.AppEnv)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("expected HTTPAddr ':8080', got '%s'", cfg.HTTPAddr)
	}

	if err := cfg.ValidateDB(); err == nil {
		t.Errorf("expected error when DATABASE_URL is empty, got nil")
	}
}

func TestConfigCustomEnv(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("HTTP_ADDR", ":9000")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db")
	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("HTTP_ADDR")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Errorf("expected AppEnv 'production', got '%s'", cfg.AppEnv)
	}

	if cfg.HTTPAddr != ":9000" {
		t.Errorf("expected HTTPAddr ':9000', got '%s'", cfg.HTTPAddr)
	}

	if err := cfg.ValidateDB(); err != nil {
		t.Errorf("expected ValidateDB to pass, got: %v", err)
	}
}
