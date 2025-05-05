/*
Copyright © 2025 Parsel Email <contact@parsel.email>
*/
package database

import (
	"database/sql"
	"fmt"
)

// DB is an interface that defines the methods a database implementation should provide
type DB interface {
	// Close closes the database connection
	Close() error

	// RunMigrations runs all migrations for the database
	RunMigrations() error

	// GetDB returns the underlying sql.DB instance
	GetDB() *sql.DB

	// VerifyExtensions checks that required database extensions are loaded
	VerifyExtensions() error
}

// GetDB returns the underlying sql.DB instance for SQLite
func (s *SQLite) GetDB() *sql.DB {
	return s.DB
}

// New creates a new database connection based on the specified type
func New(dbType, dbPath string) (DB, error) {
	switch dbType {
	case "sqlite":
		return NewSQLite(dbPath)
	case "libsql":
		return NewLibSQL(dbPath)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
