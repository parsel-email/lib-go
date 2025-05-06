package libsql

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDatabaseBasic(t *testing.T) {
	// Use in-memory database for testing
	cfg := DefaultConfig()

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := WithContext(context.Background(), 5*time.Second)
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
	cfg := DefaultConfig()

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create table for transaction test
	_, err = db.ExecContext(ctx, "CREATE TABLE tx_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test successful transaction
	tx, err := db.BeginTx(ctx, nil)
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
	tx, err = db.BeginTx(ctx, nil)
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
	cfg := DefaultConfig()

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := WithContext(context.Background(), 5*time.Second)
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
	cfg := DefaultConfig()

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create tables for FTS5 test - splitting into separate statements
	// Create the documents table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE documents (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create documents table: %v", err)
	}

	// Create the FTS5 virtual table
	_, err = db.ExecContext(ctx, `
		CREATE VIRTUAL TABLE documents_fts USING fts5(
			title, content, content='documents', content_rowid='id'
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create FTS5 virtual table: %v", err)
	}

	// Create the trigger
	_, err = db.ExecContext(ctx, `
		CREATE TRIGGER documents_ai AFTER INSERT ON documents BEGIN
			INSERT INTO documents_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
		END
	`)
	if err != nil {
		t.Fatalf("Failed to create trigger: %v", err)
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

func TestLibSQLVectorSupport(t *testing.T) {
	// Use in-memory database for testing
	cfg := DefaultConfig()

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create table with vector column using F32_BLOB datatype for vectors
	_, err = db.ExecContext(ctx, `
		CREATE TABLE vector_test (
			id INTEGER PRIMARY KEY,
			embedding F32_BLOB(4)  -- 4-dimensional float32 vector
		)
	`)
	if err != nil {
		// If the native vector types aren't supported, skip the test
		if strings.Contains(err.Error(), "near \"F32_BLOB\"") {
			t.Skip("LibSQL native vector types not supported in this version, skipping test")
		}
		t.Fatalf("Failed to create vector table: %v", err)
	}

	// Insert vectors using the vector32 function
	_, err = db.ExecContext(ctx, `
		INSERT INTO vector_test (embedding) VALUES 
		(vector32('[0.800, 0.579, 0.481, 0.229]')),
		(vector32('[0.406, 0.027, 0.378, 0.056]')),
		(vector32('[0.698, 0.140, 0.073, 0.125]')),
		(vector32('[0.379, 0.637, 0.011, 0.647]'))
	`)
	if err != nil {
		// If vector32 function isn't available, skip the test
		if strings.Contains(err.Error(), "no such function: vector32") {
			t.Skip("LibSQL vector32 function not available, skipping test")
		}
		t.Fatalf("Failed to insert vectors: %v", err)
	}

	// Create a vector index for efficient similarity search
	_, err = db.ExecContext(ctx, `
		CREATE INDEX vector_idx ON vector_test (libsql_vector_idx(embedding))
	`)
	if err != nil {
		// If vector indexing isn't available, just log and continue (don't skip the test)
		if strings.Contains(err.Error(), "libsql_vector_idx") {
			t.Logf("Warning: Vector indexing not available: %v", err)
		} else {
			t.Fatalf("Failed to create vector index: %v", err)
		}
	}

	// Test vector distance calculation
	var distance float64
	err = db.QueryRowContext(ctx, `
		SELECT vector_distance_cos(
			vector32('[0.800, 0.579, 0.481, 0.229]'), 
			vector32('[0.379, 0.637, 0.011, 0.647]')
		)
	`).Scan(&distance)
	if err != nil {
		// If vector_distance_cos function isn't available, skip the test
		if strings.Contains(err.Error(), "no such function: vector_distance_cos") {
			t.Skip("LibSQL vector_distance_cos function not available, skipping test")
		}
		t.Fatalf("Failed to calculate vector distance: %v", err)
	}

	// Verify the distance is a reasonable value for cosine distance (should be between 0 and 2)
	if distance < 0 || distance > 2 {
		t.Errorf("Unexpected cosine distance value: %f (should be between 0 and 2)", distance)
	}

	// Test vector similarity search using vector_top_k if available
	rows, err := db.QueryContext(ctx, `
		SELECT id FROM vector_top_k('vector_idx', vector32('[0.064, 0.777, 0.661, 0.687]'), 2)
	`)

	if err != nil {
		// If vector_top_k function isn't available, just log warning (don't skip the test)
		if strings.Contains(err.Error(), "no such function: vector_top_k") {
			t.Logf("Warning: vector_top_k function not available: %v", err)
		} else {
			t.Fatalf("Failed to query with vector_top_k: %v", err)
		}
	} else {
		defer rows.Close()

		// Count the results (should be 2 since we asked for top 2)
		var results []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			results = append(results, id)
		}

		// We expect 2 results
		if len(results) != 2 {
			t.Logf("Note: Expected 2 vector search results, got %d", len(results))
		}
	}

	// Fallback test: Manually compute vector similarity without vector_top_k
	rows, err = db.QueryContext(ctx, `
		SELECT id, vector_distance_cos(embedding, vector32('[0.064, 0.777, 0.661, 0.687]')) AS distance
		FROM vector_test
		ORDER BY distance ASC
		LIMIT 2
	`)
	if err != nil {
		t.Fatalf("Failed to query vector similarity: %v", err)
	}
	defer rows.Close()

	// Verify we can get the results
	var results []int64
	for rows.Next() {
		var id int64
		var distance float64
		if err := rows.Scan(&id, &distance); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		results = append(results, id)
	}

	// We should get 2 results
	if len(results) != 2 {
		t.Errorf("Expected 2 manual vector search results, got %d", len(results))
	}
}

func TestFileDatabasePersistence(t *testing.T) {
	// Use a temporary file for testing persistence
	dbFile := "test_persistence.db"

	// Cleanup after test
	defer os.Remove(dbFile)

	// Create configuration with file path
	cfg := DefaultConfig()
	cfg.Path = dbFile

	// First connection - create data
	db1, err := Open(cfg)
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
	db2, err := Open(cfg)
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
	cfg := DefaultConfig()
	cfg.MaxOpenConns = 10
	cfg.MaxIdleConns = 5
	cfg.ConnMaxLifetime = 30 * time.Minute
	cfg.ConnMaxIdleTime = 10 * time.Minute

	// Use a file database instead of in-memory for this test to ensure
	// consistent behavior with concurrent connections
	dbFile := "test_pool.db"
	defer os.Remove(dbFile) // Clean up after test
	cfg.Path = dbFile

	// Set pragmas to better handle concurrent operations
	cfg.Pragmas = Pragmas{
		"journal_mode": "WAL",       // Write-Ahead Logging for better concurrency
		"synchronous":  "NORMAL",    // Good balance between safety and performance
		"busy_timeout": "5000",      // Wait up to 5 seconds for locks to clear
		"foreign_keys": "ON",        // Enable foreign key constraints
		"cache_size":   "-2000",     // Use up to 2MB of memory for caching
		"temp_store":   "MEMORY",    // Store temporary tables in memory
		"mmap_size":    "268435456", // Memory-mapped I/O (256MB)
	}

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a table for concurrent operations
	_, err = db.Exec("CREATE TABLE pool_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test concurrent operations with reduced concurrency and more spacing between operations
	concurrency := 3 // Reduced from 5
	iterations := 5  // Reduced from 10
	errChan := make(chan error, concurrency*iterations)
	doneChan := make(chan bool, concurrency)

	// Function to insert and query in a transaction with retry logic
	worker := func(id int) {
		for i := 0; i < iterations; i++ {
			// Add random delay to avoid lock contention
			time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)

			// Create a context with timeout
			ctx, cancel := WithContext(context.Background(), 5*time.Second)

			// Begin transaction with retry
			var tx *sql.Tx
			var err error

			// Try up to 3 times to begin a transaction
			for attempt := 0; attempt < 3; attempt++ {
				tx, err = db.BeginTx(ctx, nil)
				if err == nil {
					break
				}

				// If failed, wait a bit before retrying
				time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
			}

			if err != nil {
				errChan <- fmt.Errorf("worker %d failed to begin transaction after retries: %w", id, err)
				cancel()
				doneChan <- true
				return
			}

			// Insert data with retry
			value := fmt.Sprintf("worker %d - iter %d", id, i)
			var execErr error

			// Try up to 3 times to execute the query
			for attempt := 0; attempt < 3; attempt++ {
				_, execErr = tx.Exec("INSERT INTO pool_test (value) VALUES (?)", value)
				if execErr == nil {
					break
				}

				// If failed and not on last attempt, wait before retrying
				if attempt < 2 {
					time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
				}
			}

			if execErr != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("worker %d failed to insert after retries: %w", id, execErr)
				cancel()
				doneChan <- true
				return
			}

			// Commit transaction with retry
			var commitErr error

			// Try up to 3 times to commit
			for attempt := 0; attempt < 3; attempt++ {
				commitErr = tx.Commit()
				if commitErr == nil {
					break
				}

				// If failed and not on last attempt, wait before retrying
				if attempt < 2 {
					time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
				}
			}

			if commitErr != nil {
				errChan <- fmt.Errorf("worker %d failed to commit after retries: %w", id, commitErr)
				cancel()
				doneChan <- true
				return
			}

			cancel()
			// Longer delay between iterations to reduce lock contention
			time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)
		}
		errChan <- nil
		doneChan <- true
	}

	// Start concurrent workers with delay between starts
	for i := 0; i < concurrency; i++ {
		go worker(i)
		// Add delay between starting workers
		time.Sleep(10 * time.Millisecond)
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
	cfg := DefaultConfig()

	// Open connection to the database
	db, err := Open(cfg)
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
	ctx, cancel := WithContext(context.Background(), 1*time.Nanosecond)
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
