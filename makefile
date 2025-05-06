REPO=github.com/parsel-email/lib-go


# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

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