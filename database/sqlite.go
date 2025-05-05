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
	_ "github.com/mattn/go-sqlite3"
)

// SQLite represents a SQLite database connection
type SQLite struct {
	DB     *sql.DB
	DBPath string
}

// NewSQLite creates a new SQLite database connection
func NewSQLite(dbPath string) (*SQLite, error) {
	// Create the directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with extension support using a DSN (Data Source Name)
	// that includes parameters to enable the extensions
	dsn := fmt.Sprintf("%s?_fts5=1&_json1=1&_foreign_keys=1&mode=rwc", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
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

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	sqlite := &SQLite{
		DB:     db,
		DBPath: dbPath,
	}

	// Verify extensions are available
	if err := sqlite.VerifyExtensions(); err != nil {
		db.Close()
		return nil, fmt.Errorf("extension verification failed: %w", err)
	}

	return sqlite, nil
}

// Close closes the database connection
func (s *SQLite) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

// RunMigrations runs all migrations for the SQLite database
func (s *SQLite) RunMigrations() error {
	log.Printf("Running migrations on %s", s.DBPath)

	// Create a new driver instance
	driver, err := sqlite3.WithInstance(s.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite driver: %w", err)
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

// VerifyExtensions checks that required SQLite extensions are loaded
func (s *SQLite) VerifyExtensions() error {
	// Check FTS5
	var result string
	err := s.DB.QueryRow("SELECT sqlite_compileoption_used('ENABLE_FTS5')").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to check FTS5 compile option: %w", err)
	}
	if result != "1" {
		// Try an alternate approach by attempting to create a simple FTS5 table
		_, err := s.DB.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS temp_fts5_test USING fts5(content); DROP TABLE IF EXISTS temp_fts5_test;")
		if err != nil {
			return fmt.Errorf("FTS5 extension is not available: %w", err)
		}
		log.Printf("FTS5 is available but not reported in compile options")
	} else {
		log.Printf("FTS5 extension is compiled in")
	}

	// Check JSON functions
	err = s.DB.QueryRow("SELECT sqlite_compileoption_used('ENABLE_JSON1')").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to check JSON1 compile option: %w", err)
	}
	if result != "1" {
		// Try an alternate approach by testing a JSON function
		_, err := s.DB.Exec("SELECT json_extract('{\"key\": \"value\"}', '$.key');")
		if err != nil {
			return fmt.Errorf("JSON1 extension is not available: %w", err)
		}
		log.Printf("JSON1 is available but not reported in compile options")
	} else {
		log.Printf("JSON1 extension is compiled in")
	}

	log.Printf("SQLite extensions verified: FTS5 and JSON are available")
	return nil
}
