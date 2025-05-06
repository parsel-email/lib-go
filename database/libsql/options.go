package libsql

import (
	"strings"
)

// Pragmas represents SQLite/libSQL connection pragmas
type Pragmas map[string]string

// DefaultPragmas returns the default pragmas for optimized performance
func DefaultPragmas() Pragmas {
	return Pragmas{
		"journal_mode": "WAL",       // Write-Ahead Logging for better concurrency
		"synchronous":  "NORMAL",    // Good balance between safety and performance
		"foreign_keys": "ON",        // Enable foreign key constraints
		"cache_size":   "-2000",     // Use up to 2MB of memory for caching
		"temp_store":   "MEMORY",    // Store temporary tables in memory
		"mmap_size":    "268435456", // Memory-mapped I/O (256MB)
	}
}

// formatDSN formats a DSN (Data Source Name) string with required pragmas
func formatDSN(path string, pragmas Pragmas) string {
	// Start with the base path
	dsn := path

	// Build query parameters
	var params []string

	// Add pragmas
	for key, value := range pragmas {
		params = append(params, key+"="+value)
	}

	// Add query string if parameters exist
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}
