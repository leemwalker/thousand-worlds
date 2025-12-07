#!/bin/bash

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set defaults if not provided
export JWT_SECRET="${JWT_SECRET:-dev-secret-key-CHANGE-IN-PRODUCTION}"
export DATABASE_URL="${DATABASE_URL:-postgres://admin:${POSTGRES_PASSWORD:-password123}@127.0.0.1:5432/mud_core?sslmode=disable}"
export OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
export PORT="${PORT:-8080}"

echo "Running database migrations..."
go run cmd/migrate/main.go

echo "Building game server..."
go build -o bin/game-server ./cmd/game-server
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Starting game server in development mode..."
./bin/game-server
