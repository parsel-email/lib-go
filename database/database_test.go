package database_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/parsel-email/lib-go/database"
)

func TestDatabaseBasic(t *testing.T) {
	// Use in-memory database for testing
	cfg := database.DefaultConfig()

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := database.WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Test basic query
	_, err = db.ExecContext(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert data
	_, err = db.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "test value")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Query data
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM test WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}

	if name != "test value" {
		t.Errorf("Expected 'test value', got '%s'", name)
	}
}

func TestTransaction(t *testing.T) {
	// Use in-memory database for testing
	cfg := database.DefaultConfig()

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := database.WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create table for transaction test
	_, err = db.ExecContext(ctx, "CREATE TABLE tx_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test successful transaction
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tx_test (value) VALUES (?)", "commit value")
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Test rollback transaction
	tx, err = db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tx_test (value) VALUES (?)", "rollback value")
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify only committed value exists
	rows, err := db.QueryContext(ctx, "SELECT value FROM tx_test")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		values = append(values, value)
	}

	if len(values) != 1 || values[0] != "commit value" {
		t.Errorf("Expected only committed value, got: %v", values)
	}
}

func TestJSONExtension(t *testing.T) {
	// Use in-memory database for testing
	cfg := database.DefaultConfig()

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := database.WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create table with JSON column
	_, err = db.ExecContext(ctx, `
		CREATE TABLE json_test (
			id INTEGER PRIMARY KEY,
			data JSON
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert JSON data
	jsonData := `{"name": "John", "age": 30, "tags": ["developer", "golang"]}`
	_, err = db.ExecContext(ctx, "INSERT INTO json_test (data) VALUES (json(?))", jsonData)
	if err != nil {
		t.Fatalf("Failed to insert JSON data: %v", err)
	}

	// Test JSON extraction
	var name string
	err = db.QueryRowContext(ctx, "SELECT json_extract(data, '$.name') FROM json_test WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("Failed to extract JSON: %v", err)
	}

	if name != "John" {
		t.Errorf("Expected 'John', got '%s'", name)
	}

	// Test JSON array extraction
	var firstTag string
	err = db.QueryRowContext(ctx, "SELECT json_extract(data, '$.tags[0]') FROM json_test WHERE id = 1").Scan(&firstTag)
	if err != nil {
		t.Fatalf("Failed to extract JSON array: %v", err)
	}

	if firstTag != "developer" {
		t.Errorf("Expected 'developer', got '%s'", firstTag)
	}
}

func TestFTS5Extension(t *testing.T) {
	// Use in-memory database for testing
	cfg := database.DefaultConfig()

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := database.WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create tables for FTS5 test
	_, err = db.ExecContext(ctx, `
		CREATE TABLE documents (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT
		);

		CREATE VIRTUAL TABLE documents_fts USING fts5(
			title, content, content='documents', content_rowid='id'
		);

		CREATE TRIGGER documents_ai AFTER INSERT ON documents BEGIN
			INSERT INTO documents_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
		END;
	`)
	if err != nil {
		t.Fatalf("Failed to create FTS5 tables: %v", err)
	}

	// Insert test documents
	docs := []struct {
		title   string
		content string
	}{
		{"Golang Database", "How to use database/sql package in Go"},
		{"SQLite in Go", "Using SQLite with Go is simple and efficient"},
		{"JSON in SQLite", "SQLite supports JSON data format for flexible storage"},
	}

	for _, doc := range docs {
		_, err = db.ExecContext(ctx, "INSERT INTO documents (title, content) VALUES (?, ?)",
			doc.title, doc.content)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}
	}

	// Test FTS5 search
	rows, err := db.QueryContext(ctx,
		"SELECT rowid, title FROM documents_fts WHERE documents_fts MATCH ? ORDER BY rank",
		"sqlite")
	if err != nil {
		t.Fatalf("Failed to search with FTS5: %v", err)
	}
	defer rows.Close()

	results := []string{}
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		results = append(results, title)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(results))
	}

	// First result should be "SQLite in Go" as it's more relevant
	if len(results) > 0 && results[0] != "SQLite in Go" {
		t.Errorf("Expected first result to be 'SQLite in Go', got '%s'", results[0])
	}
}

func TestSQLiteVecExtension(t *testing.T) {
	// Use in-memory database for testing
	cfg := database.DefaultConfig()

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if sqlite-vec extension is available
	// If not available, skip this test
	var hasVec int
	err = db.QueryRow(`
		SELECT count(*) FROM pragma_function_list 
		WHERE name='vector_to_json' OR name='json_to_vector'
	`).Scan(&hasVec)

	if err != nil || hasVec == 0 {
		t.Skip("sqlite-vec extension not available, skipping test")
	}

	// Create a context with timeout
	ctx, cancel := database.WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create table for vector test
	_, err = db.ExecContext(ctx, `
		CREATE TABLE vector_test (
			id INTEGER PRIMARY KEY,
			embedding BLOB
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create vector table: %v", err)
	}

	// Test vector operations
	// This is a simplified test as the actual embedding operations would be more complex
	_, err = db.ExecContext(ctx, `
		WITH vector AS (
			SELECT json_to_vector('[1.0, 2.0, 3.0]') as vec
		)
		INSERT INTO vector_test (embedding) SELECT vec FROM vector
	`)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	// Verify vector can be retrieved
	var vecJSON string
	err = db.QueryRowContext(ctx, `
		SELECT vector_to_json(embedding) FROM vector_test WHERE id = 1
	`).Scan(&vecJSON)
	if err != nil {
		t.Fatalf("Failed to retrieve vector: %v", err)
	}

	if vecJSON != "[1.0,2.0,3.0]" && vecJSON != "[1,2,3]" {
		t.Errorf("Unexpected vector value: %s", vecJSON)
	}
}

func TestFileDatabasePersistence(t *testing.T) {
	// Use a temporary file for testing persistence
	dbFile := "test_persistence.db"

	// Cleanup after test
	defer os.Remove(dbFile)

	// Create configuration with file path
	cfg := database.DefaultConfig()
	cfg.Path = dbFile

	// First connection - create data
	db1, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open first database connection: %v", err)
	}

	// Create table and insert data
	_, err = db1.Exec("CREATE TABLE persist_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db1.Exec("INSERT INTO persist_test (value) VALUES (?)", "persistent data")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Close first connection
	if err := db1.Close(); err != nil {
		t.Fatalf("Failed to close first connection: %v", err)
	}

	// Second connection - verify data persists
	db2, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open second database connection: %v", err)
	}
	defer db2.Close()

	var value string
	err = db2.QueryRow("SELECT value FROM persist_test WHERE id = 1").Scan(&value)
	if err != nil {
		t.Fatalf("Failed to query persisted data: %v", err)
	}

	if value != "persistent data" {
		t.Errorf("Expected 'persistent data', got '%s'", value)
	}
}

func TestConnectionPool(t *testing.T) {
	// Configure connection pool settings
	cfg := database.DefaultConfig()
	cfg.MaxOpenConns = 10
	cfg.MaxIdleConns = 5
	cfg.ConnMaxLifetime = 30 * time.Minute
	cfg.ConnMaxIdleTime = 10 * time.Minute

	// Use a file database instead of in-memory for this test to ensure
	// consistent behavior with concurrent connections
	dbFile := "test_pool.db"
	defer os.Remove(dbFile) // Clean up after test
	cfg.Path = dbFile

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a table for concurrent operations
	_, err = db.Exec("CREATE TABLE pool_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test concurrent operations
	concurrency := 5
	iterations := 10
	errChan := make(chan error, concurrency*iterations)
	doneChan := make(chan bool, concurrency)

	// Function to insert and query in a transaction
	worker := func(id int) {
		for i := 0; i < iterations; i++ {
			// Create a context with timeout
			ctx, cancel := database.WithContext(context.Background(), 5*time.Second)

			// Begin transaction
			tx, err := db.BeginTx(ctx)
			if err != nil {
				errChan <- fmt.Errorf("worker %d failed to begin transaction: %w", id, err)
				cancel()
				doneChan <- true
				return
			}

			// Insert data
			value := fmt.Sprintf("worker %d - iter %d", id, i)
			_, err = tx.Exec("INSERT INTO pool_test (value) VALUES (?)", value)
			if err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("worker %d failed to insert: %w", id, err)
				cancel()
				doneChan <- true
				return
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				errChan <- fmt.Errorf("worker %d failed to commit: %w", id, err)
				cancel()
				doneChan <- true
				return
			}

			cancel()
			// Small delay to avoid overwhelming the SQLite lock
			time.Sleep(1 * time.Millisecond)
		}
		errChan <- nil
		doneChan <- true
	}

	// Start concurrent workers
	for i := 0; i < concurrency; i++ {
		go worker(i)
	}

	// Wait for all workers to finish
	for i := 0; i < concurrency; i++ {
		<-doneChan
	}

	// Check for any errors
	var workerErrors []error
	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			workerErrors = append(workerErrors, err)
		}
	}

	if len(workerErrors) > 0 {
		for _, err := range workerErrors {
			t.Errorf("Worker error: %v", err)
		}
	}

	// Verify total number of rows
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pool_test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	expectedCount := concurrency*iterations - len(workerErrors)
	if count != expectedCount {
		t.Errorf("Expected %d rows, got %d", expectedCount, count)
	}
}

func TestContextTimeout(t *testing.T) {
	// Use in-memory database for testing
	cfg := database.DefaultConfig()

	// Open connection to the database
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a table for the long-running query test
	_, err = db.Exec("CREATE TABLE timeout_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert some data
	for i := 0; i < 100; i++ {
		_, err = db.Exec("INSERT INTO timeout_test (value) VALUES (?)", fmt.Sprintf("value %d", i))
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	// Test with a very short timeout
	ctx, cancel := database.WithContext(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This query should timeout
	_, err = db.QueryContext(ctx, "SELECT value FROM timeout_test WHERE id IN (SELECT id FROM timeout_test)")

	// We expect a context deadline exceeded error
	if err == nil {
		t.Error("Expected context deadline exceeded error, got nil")
	} else if err.Error() != "context deadline exceeded" && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}
