# Parsel Go Library

This repository contains shared Go library code for the Parsel project.

## Components

- **database**: libSQL database interface with support for both local and remote databases
- **logger**: Structured logging utilities
- **metrics**: Prometheus metrics collection
- **tracing**: Distributed tracing support

## Database Package

The `database` package provides a simple, idiomatic libSQL interface. libSQL is a SQLite fork that's optimized for serverless and edge computing environments, maintaining API compatibility with SQLite.

### Features

- Lightweight wrapper around the standard `database/sql` package
- Support for both local file databases and remote libSQL databases (Turso)
- Transaction support with easy-to-use Begin/Commit/Rollback methods
- Connection pool configuration with sensible defaults
- Context-based operations for proper timeout handling

### Usage

```go
import (
    "context"
    "log"
    "time"

    "github.com/parsel-email/lib-go/database"
)

func main() {
    // Create a database with default configuration (local in-memory database)
    cfg := database.DefaultConfig()
    cfg.Path = "my-database.db" // Use a local file
    
    // For remote libSQL databases:
    // cfg.Path = "libsql://my-db.turso.io"
    // cfg.AuthToken = "your-auth-token"

    // Open connection to the database
    db, err := database.Open(cfg)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    // Create a context with timeout
    ctx, cancel := database.WithContext(context.Background(), 5*time.Second)
    defer cancel()

    // Using transactions
    tx, err := db.BeginTx(ctx)
    if err != nil {
        log.Fatalf("Failed to begin transaction: %v", err)
    }

    // Execute SQL within transaction
    if _, err := tx.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
        tx.Rollback()
        log.Fatalf("Failed to create table: %v", err)
    }

    if err := tx.Commit(); err != nil {
        log.Fatalf("Failed to commit transaction: %v", err)
    }

    // Using JSON capabilities
    jsonQuery := `
    INSERT INTO users (name, metadata) 
    VALUES (?, json(?))
    `
    metadata := `{"roles": ["admin", "user"], "settings": {"theme": "dark"}}`
    _, err = db.ExecContext(ctx, jsonQuery, "John Doe", metadata)
    if err != nil {
        log.Fatalf("Failed to insert with JSON: %v", err)
    }

    // Using FTS5 (Full-Text Search)
    // Assuming you've set up FTS5 tables as shown in the example.go file
    rows, err := db.QueryContext(ctx, 
        "SELECT id, title FROM documents_fts WHERE documents_fts MATCH ?", 
        "search term")
    if err != nil {
        log.Fatalf("Failed to query: %v", err)
    }
    defer rows.Close()
    
    // Process search results
    for rows.Next() {
        var id int64
        var title string
        if err := rows.Scan(&id, &title); err != nil {
            log.Fatalf("Failed to scan row: %v", err)
        }
        log.Printf("Result: %d - %s", id, title)
    }
}
```

## License

[Add your license information here]