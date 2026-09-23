.PHONY: help build run test test-integration coverage swagger clean docker-up docker-down migrate

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the application from source with the dev-auth build tag (AUTH_MODE=bypass allowed)
	go run -tags devauth cmd/api/main.go

test: ## Run tests
	go test -v ./...

coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

swagger: ## Generate OpenAPI/Swagger documentation
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	./scripts/generate-openapi.sh

deps: ## Download dependencies
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html

test-integration: ## Run integration tests against the migrated database (see docs/backend/testing.md)
	./test_integration.sh

docker-up: ## Start the local stack from ../docker-compose.yml (host only; the dev container starts it itself)
	docker compose -f ../docker-compose.yml up -d --build

docker-down: ## Stop the local stack
	docker compose -f ../docker-compose.yml down

migrate: ## Apply pending migrations with the Compose defaults (works inside the dev container)
	DB_ADMIN_CONN_STRING="$${DB_ADMIN_CONN_STRING:-host=$${INTEGRATION_TEST_DB_HOST:-localhost} port=5432 user=easi password=easi dbname=easi sslmode=disable}" \
	EASI_APP_PASSWORD="$${EASI_APP_PASSWORD:-localdev}" EASI_ADMIN_PASSWORD="$${EASI_ADMIN_PASSWORD:-localdev}" \
	go run cmd/migrate/main.go
