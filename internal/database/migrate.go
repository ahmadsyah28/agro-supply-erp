package database

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type MigrationRunner struct {
	db *DB
	fs fs.FS
}

func NewMigrationRunner(db *DB, migrationFS fs.FS) *MigrationRunner {
	return &MigrationRunner{
		db: db,
		fs: migrationFS,
	}
}

func (m *MigrationRunner) RunMigrations(ctx context.Context) error {
	// Create schema_migrations table if not exists
	initQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := m.db.Pool.Exec(ctx, initQuery); err != nil {
		return fmt.Errorf("failed to initialize schema_migrations table: %w", err)
	}

	// Read migration directory
	entries, err := fs.ReadDir(m.fs, ".")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	var upMigrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}

	sort.Strings(upMigrations)

	for _, filename := range upMigrations {
		version := strings.TrimSuffix(filename, ".up.sql")

		var exists bool
		err := m.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration version %s: %w", version, err)
		}

		if exists {
			continue
		}

		log.Printf("[MIGRATION] Applying %s...", filename)
		content, err := fs.ReadFile(m.fs, filepath.ToSlash(filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := m.db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		recordQuery := `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`
		if _, err := tx.Exec(ctx, recordQuery, version, time.Now().UTC()); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("[MIGRATION] Successfully applied %s", filename)
	}

	return nil
}
