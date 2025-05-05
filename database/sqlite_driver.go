//go:build cgo
// +build cgo

package database

// Import the go-sqlite3 driver with support for FTS5 and JSON1
import (
	_ "github.com/mattn/go-sqlite3"
)

// This file imports the go-sqlite3 driver with the necessary compile-time
// flags to enable the FTS5 and JSON1 extensions.
//
// The SQLite extensions are enabled at compile time by using the appropriate
// build tags in go.mod or by using a custom build with specific CGO_CFLAGS.
