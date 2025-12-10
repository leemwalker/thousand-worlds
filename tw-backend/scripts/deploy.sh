#!/bin/bash

# Production deployment script
set -e

echo "=== Thousand Worlds Deployment ==="
echo ""

# Get script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Change to project root
cd "$PROJECT_ROOT"

# Check prerequisites
echo "Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting." >&2; exit 1; }

# Check environment variables
echo "Validating required environment variables..."

# Load .env if it exists
if [ -f .env ]; then
    echo "Loading environment from .env file..."
    export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
fi

# Validate required secrets
MISSING_VARS=()
[ -z "$JWT_SECRET" ] && MISSING_VARS+=("JWT_SECRET")
[ -z "$POSTGRES_PASSWORD" ] && MISSING_VARS+=("POSTGRES_PASSWORD")
[ -z "$MONGO_PASSWORD" ] && MISSING_VARS+=("MONGO_PASSWORD")

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo "ERROR: Required environment variables not set: ${MISSING_VARS[*]}"
    echo "Please create a .env file or set these variables."
    echo "See .env.template for reference."
    exit 1
fi

echo "✓ All required secrets are set"

# Stop any existing containers that might conflict
echo "Stopping existing containers..."
docker-compose -f deploy/docker-compose.yml down 2>/dev/null || true
docker-compose -f docker-compose.prod.yml down 2>/dev/null || true

# Build images
echo "Building Docker images..."
docker build -f Dockerfile.game-server -t thousand-worlds/game-server:latest .

# Run migrations
echo "Running database migrations..."
docker-compose -f docker-compose.prod.yml up -d postgis
sleep 5  # Wait for PostgreSQL to be ready

# Run migration container
docker run --rm \
  --network mud_net \
  -e DATABASE_URL="postgres://${POSTGRES_USER:-admin}:${POSTGRES_PASSWORD}@postgis:5432/${POSTGRES_DB:-mud_core}?sslmode=disable" \
  thousand-worlds/game-server:latest \
  ./game-server migrate || echo "Migrations may have already run"

# Pull LLM model
echo "Ensuring LLM model is available..."
docker-compose -f docker-compose.prod.yml up -d ollama
sleep 10
docker exec mud_ollama ollama pull llama3.1:8b || echo "Model may already exist"

# Start all services
echo "Starting all services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for health checks
echo "Waiting for services to be healthy..."
sleep 10

# Check health
echo "Checking service health..."
curl -f http://localhost:8080/health || echo "Game server not yet ready"

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Services:"
echo "  - Game Server: http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo "  - NATS: localhost:4222"
echo "  - Ollama: http://localhost:11434"
echo ""
echo "View logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "Stop: docker-compose -f docker-compose.prod.yml down"
