// Package database provides a simple SQLite interface with extensions
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB represents a database connection
type DB struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	Pragmas         Pragmas
}

// DefaultConfig returns a default database configuration
func DefaultConfig() Config {
	return Config{
		Path:            ":memory:", // Default to in-memory database
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
		Pragmas:         DefaultPragmas(),
	}
}

// Open creates a new database connection with JSON1, FTS5, and sqlite-vec extensions
func Open(cfg Config) (*DB, error) {
	// Format the DSN with required extensions and pragmas
	dsn := formatDSN(cfg.Path, cfg.Pragmas)

	// Open the database connection
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &DB{DB: db}, nil
}

// WithContext returns a context with timeout for database operations
func WithContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

// Transaction represents a database transaction
type Transaction struct {
	*sql.Tx
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	return &Transaction{Tx: tx}, nil
}

// Commit commits the transaction
func (tx *Transaction) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// Rollback rolls back the transaction
func (tx *Transaction) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		return fmt.Errorf("rolling back transaction: %w", err)
	}
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if err := db.DB.Close(); err != nil {
		return fmt.Errorf("closing database: %w", err)
	}
	return nil
}
