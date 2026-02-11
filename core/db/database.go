package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/lhommenul/brique/migrations"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// Database wraps the SQL database connection
type Database struct {
	DB     *sql.DB
	logger *slog.Logger
}

// NewDatabase creates a new database connection and runs migrations
func NewDatabase(dbPath string, logger *slog.Logger) (*Database, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	database := &Database{
		DB:     db,
		logger: logger,
	}

	// Run migrations
	if err := database.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database initialized successfully", "path", dbPath)

	return database, nil
}

// runMigrations runs all pending database migrations
func (d *Database) runMigrations() error {
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(d.DB, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	d.logger.Info("Migrations completed successfully")

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if err := d.DB.Close(); err != nil {
		d.logger.Error("Failed to close database", "error", err)
		return err
	}
	d.logger.Info("Database connection closed")
	return nil
}
