.PHONY: help build run test coverage swagger clean

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the application
	go run cmd/api/main.go

test: ## Run tests
	go test -v ./...

coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

swagger: ## Generate OpenAPI/Swagger documentation
	swag init -g cmd/api/main.go -o docs
	./scripts/generate-openapi.sh

deps: ## Download dependencies
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html

docker-up: ## Start PostgreSQL via docker-compose
	docker-compose up -d

docker-down: ## Stop PostgreSQL
	docker-compose down

migrate: ## Run database migrations (initialize schema)
	go run cmd/api/main.go &
	sleep 2
	pkill -f "cmd/api/main.go"
