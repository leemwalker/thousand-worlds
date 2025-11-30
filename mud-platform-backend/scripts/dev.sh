#!/bin/bash

# Environment variables for local development
export JWT_SECRET="dev-secret-key-change-in-production"
export DATABASE_URL="postgres://admin:password123@127.0.0.1:5432/mud_core?sslmode=disable"
export OLLAMA_HOST="http://localhost:11434"
export PORT="8080"

echo "Running database migrations..."
go run cmd/migrate/main.go

echo "Starting game server in development mode..."
go run cmd/game-server/main.go
