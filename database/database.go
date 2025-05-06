// Package database provides a simple libSQL interface with extensions
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

// DB represents a database connection
type DB struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	Path            string
	AuthToken       string // libSQL auth token for remote connections
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
		AuthToken:       "",         // Default to no auth token
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
		Pragmas:         DefaultPragmas(),
	}
}

// Open creates a new database connection with libSQL
func Open(cfg Config) (*DB, error) {
	var db *sql.DB

	// Check if the connection string is for a remote database or local file
	if strings.HasPrefix(cfg.Path, "libsql://") {
		// Configure for remote libSQL database
		connOpts := []libsql.Option{}

		// Add auth token if provided
		if cfg.AuthToken != "" {
			connOpts = append(connOpts, libsql.WithAuthToken(cfg.AuthToken))
		}

		// Create a connector for the remote database
		connector, err := libsql.NewConnector(cfg.Path, connOpts...)
		if err != nil {
			return nil, fmt.Errorf("creating libSQL connector for remote database: %w", err)
		}

		// Open the database with the connector
		db = sql.OpenDB(connector)
	} else {
		// For local file or in-memory database
		dsn := formatDSN(cfg.Path, cfg.Pragmas)

		// Open the database with the libSQL connector
		connector, err := libsql.NewConnector(dsn, nil)
		if err != nil {
			return nil, fmt.Errorf("creating libSQL connector for local database: %w", err)
		}
		db = sql.OpenDB(connector)
	}

	if db == nil {
		return nil, fmt.Errorf("failed to create a database connection")
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close() // Close the failed connection
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
