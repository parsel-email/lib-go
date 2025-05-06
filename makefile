REPO=github.com/parsel-email/lib-go

# Go build flags for SQLite extensions

# CGO settings
export CGO_ENABLED=1

# Build the application with SQLite extensions
build:
	@echo "Building with SQLite extensions..."
	@go build ./...

# Test the application with SQLite extensions
test:
	@echo "Testing with SQLite extensions..."
	@go test ./... -v

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
	@go run cmd/migrate/main.go new $(name)
	
# Run migrations up
db-migrate-up:
	@echo "Running migrations up..."
	@go run cmd/migrate/main.go up

# Run migrations down
db-migrate-down:
	@echo "Running migrations down..."
	@go run cmd/migrate/main.go down

# Get current migration version
db-migrate-version:
	@echo "Current migration version:"
	@go run cmd/migrate/main.go version

# Generate SQLC code from SQL queries
sqlc-generate:
	@echo "Generating code from SQL queries..."
	@sqlc generate

# Check for proper SQLite installation and extensions
check-sqlite:
	@echo "Checking SQLite installation and extensions..."
	@echo "SQLite version:"
	@sqlite3 --version
	@echo "\nAvailable SQLite compile options (for system SQLite):"
	@echo "PRAGMA compile_options;" | sqlite3 | grep -E 'ENABLE_FTS5|ENABLE_JSON1'
	@echo "\nNote: For the Go application, extensions like fts5 and json1 are enabled via build tags."
	@echo "Vector operations are supported natively in libSQL with F32_BLOB type and vector functions."
	
.PHONY: build test clean db-new db-migrate-up db-migrate-down db-migrate-version sqlc-generate check-sqlite