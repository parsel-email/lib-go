// Package sqlite3 provides a simple sqlite interface with extensions
package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

// Config holds database configuration
type Config struct {
	Path            string
	AuthToken       string // sqlite3 auth token for remote connections
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

// Open creates a new database connection with sqlite3
func Open(cfg Config) (*sql.DB, error) {
	var db *sql.DB

	// Check if the connection string is for a remote database or local file
	// For local file or in-memory database
	dsn := formatDSN(cfg.Path, cfg.Pragmas)

	// For local SQLite databases, use the sqlite3 connector with file: prefix
	if dsn != ":memory:" && !strings.HasPrefix(dsn, "file:") {
		dsn = "file:" + dsn
	}

	sqlite_vec.Auto()
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
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

	return db, nil
}

// WithContext returns a context with timeout for database operations
func WithContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}
