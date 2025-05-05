/*
Copyright © 2025 Parsel Email <contact@parsel.email>
*/
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Import for side effects to register the driver
)

// LibSQL represents a libsql database connection
type LibSQL struct {
	DB     *sql.DB
	DBPath string
}

// NewLibSQL creates a new libsql database connection
func NewLibSQL(dbPath string) (*LibSQL, error) {
	// Create the directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Format the URL based on the path (libsql:// for remote or file:// for local)
	var url string
	if isRemoteURL(dbPath) {
		url = dbPath
	} else {
		url = fmt.Sprintf("file:%s", dbPath)
	}

	// Open database connection with libsql
	// Following the guide for local-only connections
	db, err := sql.Open("libsql", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open libsql database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping libsql database: %w", err)
	}

	// Set pragmas for better performance and reliability
	pragmas := []string{
		"PRAGMA journal_mode = WAL",   // Use Write-Ahead Logging for better concurrency
		"PRAGMA synchronous = NORMAL", // Good balance between safety and speed
		"PRAGMA temp_store = MEMORY",  // Store temp tables and indices in memory
		"PRAGMA cache_size = -2000",   // 2MB page cache
		"PRAGMA foreign_keys = ON",    // Enforce foreign key constraints
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Printf("Warning: Failed to set pragma '%s': %v", pragma, err)
		}
	}

	libsqlDB := &LibSQL{
		DB:     db,
		DBPath: dbPath,
	}

	// Verify extensions are available
	if err := libsqlDB.VerifyExtensions(); err != nil {
		db.Close()
		return nil, fmt.Errorf("extension verification failed: %w", err)
	}

	return libsqlDB, nil
}

// Helper function to check if a path is a remote URL
func isRemoteURL(path string) bool {
	return len(path) > 8 && (path[:8] == "libsql://" || path[:9] == "https://")
}

// Close closes the database connection
func (s *LibSQL) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

// GetDB returns the underlying sql.DB instance
func (s *LibSQL) GetDB() *sql.DB {
	return s.DB
}

// RunMigrations runs all migrations for the libsql database
func (s *LibSQL) RunMigrations() error {
	log.Printf("Running migrations on %s", s.DBPath)

	// Create a new driver instance
	driver, err := sqlite3.WithInstance(s.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite driver for migrations: %w", err)
	}

	// Set the path to migration files
	migrationsPath := "file://db/migrations/sqlite"

	// Create a new migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Printf("Migrations complete")
	return nil
}

// VerifyExtensions checks that required database extensions are loaded
func (s *LibSQL) VerifyExtensions() error {
	// Try to create a simple FTS5 table
	_, err := s.DB.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS temp_fts5_test USING fts5(content); DROP TABLE IF EXISTS temp_fts5_test;")
	if err != nil {
		return fmt.Errorf("FTS5 extension is not available: %w", err)
	}
	log.Printf("FTS5 is available")

	// Test a JSON function
	_, err = s.DB.Exec("SELECT json_extract('{\"key\": \"value\"}', '$.key');")
	if err != nil {
		return fmt.Errorf("JSON1 extension is not available: %w", err)
	}
	log.Printf("JSON1 is available")

	log.Printf("libsql extensions verified: FTS5 and JSON are available")
	return nil
}
