// Package db provides PostgreSQL connection and migration utilities.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Register postgres driver
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/config"
)

// Connect establishes a PostgreSQL connection pool using sqlx.
func Connect(cfg *config.DBConfig, log zerolog.Logger) (*sqlx.DB, error) {
	// Log connection attempt with DSN (if enabled)
	connLog := log.Info().
		Str("host", cfg.Host).
		Str("name", cfg.Name)

	if cfg.LogDSN {
		connLog.Str("dsn", cfg.DSN())
	} else {
		connLog.Str("dsn", cfg.MaskedDSN())
	}
	connLog.Msg("attempting to connect to postgres")

	db, err := sqlx.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connectivity with retry.
	if err := pingWithRetry(db.DB, 5, 2*time.Second, log); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}

	log.Info().Msg("successfully connected to postgres")

	return db, nil
}

// RunMigrations applies pending goose migrations from the given directory.
func RunMigrations(db *sqlx.DB, dir string, log zerolog.Logger) error {
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	if err := goose.Up(db.DB, dir); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Info().Str("dir", dir).Msg("database migrations applied")
	return nil
}

func pingWithRetry(db *sql.DB, attempts int, delay time.Duration, log zerolog.Logger) error {
	ctx := context.Background()
	for i := 1; i <= attempts; i++ {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		log.Warn().Int("attempt", i).Int("max", attempts).Msg("waiting for database...")
		time.Sleep(delay)
	}
	return fmt.Errorf("database not reachable after %d attempts", attempts)
}
