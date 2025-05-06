REPO=github.com/parsel-email/lib-go

# Go build flags for SQLite extensions
GO_BUILD_FLAGS=-tags "fts5 json1 sqlite_vec"

# CGO settings
export CGO_ENABLED=1

# Build the application with SQLite extensions
build:
	@echo "Building with SQLite extensions..."
	@go build $(GO_BUILD_FLAGS) ./...

# Test the application with SQLite extensions
test:
	@echo "Testing with SQLite extensions..."
	@go test $(GO_BUILD_FLAGS) ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f *.db
	@rm -f *.db-journal
	@find . -type f -name "*.o" -delete
	@find . -type f -name "*.a" -delete
	@find . -type f -name "*.so" -delete

db-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Use 'make db-new name=migration_name'"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	@go run $(GO_BUILD_FLAGS) cmd/migrate/main.go new $(name)
	
# Run migrations up
db-migrate-up:
	@echo "Running migrations up..."
	@go run $(GO_BUILD_FLAGS) cmd/migrate/main.go up

# Run migrations down
db-migrate-down:
	@echo "Running migrations down..."
	@go run $(GO_BUILD_FLAGS) cmd/migrate/main.go down

# Get current migration version
db-migrate-version:
	@echo "Current migration version:"
	@go run $(GO_BUILD_FLAGS) cmd/migrate/main.go version

# Generate SQLC code from SQL queries
sqlc-generate:
	@echo "Generating code from SQL queries..."
	@sqlc generate

# Check for proper SQLite installation and extensions
check-sqlite:
	@echo "Checking SQLite installation and extensions..."
	@echo "SQLite version:"
	@sqlite3 --version
	@echo "\nAvailable SQLite extensions:"
	@echo "PRAGMA compile_options;" | sqlite3 | grep -E 'ENABLE_FTS5|ENABLE_JSON1'
	
.PHONY: build test clean db-new db-migrate-up db-migrate-down db-migrate-version sqlc-generate check-sqlite