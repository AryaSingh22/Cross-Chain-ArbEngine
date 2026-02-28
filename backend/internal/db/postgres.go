package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Database wraps a pgxpool connection pool
type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewDatabase creates a new database connection pool
func NewDatabase(ctx context.Context, databaseURL string, maxConns int, logger *zap.Logger) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	config.MaxConns = int32(maxConns)
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("database connected", zap.String("host", config.ConnConfig.Host))

	return &Database{Pool: pool, logger: logger}, nil
}

// RunMigrations applies all pending SQL migrations
func (db *Database) RunMigrations(ctx context.Context, migrationsDir string) error {
	// Create migrations tracking table
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	for _, filename := range upFiles {
		version := strings.TrimSuffix(filename, ".up.sql")

		// Check if already applied
		var count int
		err := db.Pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if count > 0 {
			db.logger.Debug("migration already applied", zap.String("version", version))
			continue
		}

		// Read and execute migration
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		_, err = db.Pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("execute migration %s: %w", filename, err)
		}

		_, err = db.Pool.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		db.logger.Info("migration applied", zap.String("version", version))
	}

	return nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	db.Pool.Close()
}
