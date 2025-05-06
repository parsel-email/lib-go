package helpers

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestCreatePostgresContainer(t *testing.T) {
	ctx := context.Background()

	// Create container
	container, err := CreatePostgresContainer(ctx)
	if err != nil {
		t.Fatalf("could not create postgres container: %v", err)
	}

	// Ensure cleanup
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	})

	// Verify connection string
	if container.ConnectionString == "" {
		t.Error("expected connection string to not be empty")
	}

	// Verify connection string contains expected parts
	expectedParts := []string{"postgres://", "postgres:postgres@"}
	for _, part := range expectedParts {
		if !strings.Contains(container.ConnectionString, part) {
			t.Errorf("connection string missing expected part: %s", part)
		}
	}

	// Test actual database connection
	db, err := sql.Open("postgres", container.ConnectionString)
	if err != nil {
		t.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	// Add retry logic for connection
	var connected bool
	for i := 0; i < 5; i++ {
		err = db.Ping()
		if err == nil {
			connected = true
			break
		}
		time.Sleep(time.Second)
	}

	if !connected {
		t.Fatalf("failed to connect to database after retries: %v", err)
	}

	// Verify we can execute a simple query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("failed to execute test query: %v", err)
	}
	if result != 1 {
		t.Errorf("expected query result to be 1, got %d", result)
	}
}
