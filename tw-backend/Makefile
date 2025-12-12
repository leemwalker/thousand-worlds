.PHONY: up down test test-coverage build deploy clean lint run migrate help

# Default target
help:
	@echo "Thousand Worlds Backend - Available Commands:"
	@echo ""
	@echo "  make up             - Start all Docker services"
	@echo "  make down           - Stop all Docker services"
	@echo "  make run            - Run the game server locally"
	@echo "  make build          - Build the game server binary"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter"
	@echo "  make migrate        - Run database migrations"
	@echo "  make deploy         - Deploy to production"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

# Docker commands
up:
	docker-compose -f docker-compose.prod.yml up -d

down:
	docker-compose -f docker-compose.prod.yml down

# Run the server locally
run:
	./scripts/dev.sh

# Build commands
build:
	@mkdir -p bin
	go build -o bin/game-server cmd/game-server/main.go
	@echo "Built: bin/game-server"

build-all:
	@mkdir -p bin
	go build -o bin/game-server cmd/game-server/main.go
	go build -o bin/ai-gateway cmd/ai-gateway/main.go
	go build -o bin/world-service cmd/world-service/main.go
	go build -o bin/auth-service cmd/auth-service/main.go
	@echo "Built all services in bin/"

# Test commands
test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	@mkdir -p .coverage
	go test -coverprofile=.coverage/coverage.out ./...
	go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
	@echo "Coverage report: .coverage/coverage.html"

test-coverage-text:
	@mkdir -p .coverage
	go test -coverprofile=.coverage/coverage.out ./...
	go tool cover -func=.coverage/coverage.out

# Linting
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	elif [ -x "$(HOME)/go/bin/golangci-lint" ]; then \
		$(HOME)/go/bin/golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Database migrations
migrate:
	go run cmd/migrate/main.go

migrate-up:
	migrate -path migrations/postgres -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path migrations/postgres -database "$${DATABASE_URL}" down 1

# Deployment
deploy:
	./scripts/deploy.sh

# Cleanup
clean:
	rm -rf bin/
	rm -rf .coverage/
	rm -f game-server ai-gateway main
	rm -f *.out
	rm -f coverage*.out
	rm -f *_coverage*.out
	@echo "Cleaned build artifacts and coverage files"

# Development helpers
dev-deps:
	docker-compose -f docker-compose.prod.yml up -d postgis mongo nats redis
	@echo "Development dependencies started"

pull-model:
	docker exec mud_ollama ollama pull llama3.2:3b
	@echo "LLM model pulled"
