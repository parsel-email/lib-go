package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// CreatePostgresContainer creates a new Postgres container for testing
func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	files, _ := os.ReadDir(filepath.Join(basepath, "../", "../", "db"))
	initScripts := make([]string, 0)
	for _, file := range files {
		fmt.Println(filepath.Join(basepath, "../", "../", "db", file.Name()))
		initScripts = append(initScripts, filepath.Join(basepath, "../", "../", "db", file.Name()))
	}

	pgContainer, err := postgres.Run(ctx,
		"pgvector/pgvector:pg17",
		postgres.WithInitScripts(initScripts...),
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	container := &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}

	return container, nil
}

// CreatePostgresContainerWithMigrations creates a Postgres container and applies migrations
func CreatePostgresContainerWithMigrations(ctx context.Context) (*PostgresContainer, error) {
	container, err := CreatePostgresContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres container: %w", err)
	}

	// Get the path to the migrations directory
	migrationsPath := filepath.Join(basepath, "../", "../", "db", "migrations")

	// Apply migrations
	err = RunMigrations(migrationsPath, container.ConnectionString)
	if err != nil {
		// Clean up the container if migrations fail
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return container, nil
}

// RunMigrations applies database migrations from the specified path
func RunMigrations(migrationsPath, connectionString string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		connectionString,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
