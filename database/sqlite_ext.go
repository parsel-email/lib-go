//go:build cgo
// +build cgo

// #cgo CFLAGS: -DSQLITE_ENABLE_FTS5 -DSQLITE_ENABLE_JSON1 -DSQLITE_ENABLE_RTREE
// #cgo LDFLAGS: -lm
package database

// This file ensures that when the go-sqlite3 package is built,
// it includes the necessary compile flags to enable extensions.
//
// The build tags here tell Go to:
// - Enable FTS5 full-text search extension
// - Enable JSON1 extension for JSON support
// - Enable RTREE extension for spatial indexing
//
// For this file to take effect, the project must be built with CGO enabled
// by setting CGO_ENABLED=1 in the environment.
