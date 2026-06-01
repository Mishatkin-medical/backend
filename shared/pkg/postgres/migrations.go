package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RunMigrations runs all SQL files in migrations directory
func RunMigrations(ctx context.Context, pool *Pool, migrationsDir string) error {
	// Read migrations directory
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	// Filter SQL files and sort
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Execute each migration
	for _, file := range files {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}

// CreateMigrationsTable creates the schema_migrations table if not exists
const CreateMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);
'

// RecordMigration records a migration as applied
func RecordMigration(ctx context.Context, pool *Pool, version string) error {
	_, err := pool.Exec(ctx, 
		"INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT (version) DO NOTHING",
		version,
	)
	return err
}

// GetAppliedMigrations returns list of applied migrations
func GetAppliedMigrations(ctx context.Context, pool *Pool) ([]string, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
