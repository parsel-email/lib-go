// Package main provides a command-line utility for managing database migrations
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

const migrationsDir = "./db/migrations"

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Command is required: new, up, down, version")
	}

	cmd := args[0]

	switch cmd {
	case "new":
		if len(args) != 2 {
			log.Fatal("Migration name is required")
		}
		createMigration(args[1])
	case "up":
		runMigration(func(m *migrate.Migrate) error {
			return m.Up()
		})
	case "down":
		runMigration(func(m *migrate.Migrate) error {
			return m.Down()
		})
	case "version":
		getMigrationVersion()
	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}

func createMigration(name string) {
	timestamp := time.Now().Unix()
	upMigration := filepath.Join(migrationsDir, fmt.Sprintf("%d_%s.up.sql", timestamp, name))
	downMigration := filepath.Join(migrationsDir, fmt.Sprintf("%d_%s.down.sql", timestamp, name))

	// Ensure migrations directory exists
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Create up migration file
	if err := os.WriteFile(upMigration, []byte("-- Migration Up\n"), 0644); err != nil {
		log.Fatalf("Failed to create up migration file: %v", err)
	}

	// Create down migration file
	if err := os.WriteFile(downMigration, []byte("-- Migration Down\n"), 0644); err != nil {
		log.Fatalf("Failed to create down migration file: %v", err)
	}

	fmt.Printf("Created migration files:\n%s\n%s\n", upMigration, downMigration)
}

func getDBPath() string {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "parsel.db" // Default DB path
	}
	return dbPath
}

func runMigration(migrateFn func(*migrate.Migrate) error) {
	dbPath := getDBPath()

	// Connect to database
	db, err := sqlite.WithInstance(openDB(dbPath), &sqlite.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"sqlite",
		db,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	// Run migration function
	if err := migrateFn(m); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migration needed")
			return
		}
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration successful")
}

func getMigrationVersion() {
	dbPath := getDBPath()

	// Connect to database
	db, err := sqlite.WithInstance(openDB(dbPath), &sqlite.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"sqlite",
		db,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("No migrations applied yet")
			return
		}
		log.Fatalf("Failed to get migration version: %v", err)
	}

	fmt.Printf("Current migration version: %d (dirty: %v)\n", version, dirty)
}

func openDB(dbPath string) *sql.DB {
	// Check if this is a libSQL URL or a local file
	var db *sql.DB
	var err error

	if strings.HasPrefix(dbPath, "libsql://") {
		// For remote libSQL URLs

		connOpts := []libsql.Option{}

		connector, err := libsql.NewConnector(dbPath, connOpts...)
		if err != nil {
			log.Fatalf("Failed to create libSQL connector: %v", err)
		}
		db = sql.OpenDB(connector)
	} else {
		// For local files
		connector, err := libsql.NewConnector(dbPath, nil)
		if err != nil {
			log.Fatalf("Failed to create libSQL connector: %v", err)
		}
		db = sql.OpenDB(connector)
	}

	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return db
}
