# libsql Database Library

The `libsql` package provides a simple, high-performance interface to SQLite and libSQL databases using the `go-libsql` driver. It includes default connection pragmas, connection pooling, context-based timeouts, and transaction support.

## Installation

Ensure your module uses Go 1.23+ and run:

    go get github.com/parsel-email/lib-go/database/libsql

## Import

```go
import "github.com/parsel-email/lib-go/database/libsql"
```

## Configuration

The package exposes a `Config` struct to customize your connection:

```go
cfg := libsql.DefaultConfig()
cfg.Path = "file:my.db"             // Local file or in-memory
cfg.AuthToken = ""                  // For remote libSQL URLs (e.g., Turso)
cfg.MaxOpenConns = 10
cfg.MaxIdleConns = 5
cfg.ConnMaxLifetime = time.Hour
cfg.ConnMaxIdleTime = 30 * time.Minute
// Override default pragmas if needed:
// cfg.Pragmas["busy_timeout"] = "5000"
```

## Opening a Connection

Use `Open` to establish a connection:

```go
ctx := context.Background()
db, err := libsql.Open(cfg)
if err != nil {
    log.Fatalf("failed to open db: %v", err)
}
defer db.Close()
```

The `Open` function applies default pragmas (WAL, synchronous=NORMAL, foreign_keys=ON, etc.) for optimal performance. It also supports remote URLs (prefix `libsql://...`).

## Context-Based Operations

Use `WithContext` to create contexts with timeouts:

```go
ctx, cancel := libsql.WithContext(context.Background(), 5*time.Second)
defer cancel()

// Execute statements
_, err = db.ExecContext(ctx, "CREATE TABLE foo (id INTEGER PRIMARY KEY)")
```

Queries, Exec, and transactions all respect the context deadline.

## Transactions

Begin a transaction with:

```go
tx, err := db.BeginTx(ctx)
if err != nil {
    // handle error
}

defer tx.Rollback()

// Use tx.Exec or tx.QueryContext
err = tx.Commit()
```

## Example

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/parsel-email/lib-go/database/libsql"
)

func main() {
    cfg := libsql.DefaultConfig()
    cfg.Path = "my.db"
    db, err := libsql.Open(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx, cancel := libsql.WithContext(context.Background(), 5*time.Second)
    defer cancel()

    // Create table
    _, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)")
    if err != nil {
        log.Fatal(err)
    }

    // Insert
    _, err = db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Alice")
    if err != nil {
        log.Fatal(err)
    }

    // Query
    var name string
    row := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = 1")
    if err := row.Scan(&name); err != nil {
        log.Fatal(err)
    }
    log.Printf("User 1: %s", name)
}
```

## Testing

A comprehensive test suite is available in `database/libsql/database_test.go`. Run:

    go test github.com/parsel-email/lib-go/database/libsql

to verify functionality, including JSON, FTS5, vector, and transaction tests.

## License

Licensed under MIT. See the repository LICENSE for details.
