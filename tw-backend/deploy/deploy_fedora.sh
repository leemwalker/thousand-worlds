#!/bin/bash

# Deployment script for Fedora 42 Server

echo "Starting deployment for Fedora 42 Server..."

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed."
    echo "Install with: sudo dnf install docker-ce docker-ce-cli containerd.io docker-compose-plugin"
    exit 1
fi

# Check Docker is running
if ! docker info &> /dev/null; then
    echo "Error: Docker daemon is not running or you don't have permission."
    echo "Try: sudo systemctl start docker"
    echo "Or add yourself to docker group: sudo usermod -aG docker $USER"
    exit 1
fi

# Check for Docker Compose
if ! docker compose version &> /dev/null; then
    echo "Error: Docker Compose plugin is not installed."
    echo "Install with: sudo dnf install docker-compose-plugin"
    exit 1
fi

echo "✓ Docker and Docker Compose detected"

# Change to tw-backend directory (script may be run from deploy/)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/.."

# Ensure .env exists
if [ ! -f .env ]; then
    echo "Creating .env from .env.template..."
    if [ -f .env.template ]; then
        cp .env.template .env
        echo ""
        echo "⚠️  Please edit .env with your production secrets!"
        echo "Required variables: JWT_SECRET, POSTGRES_PASSWORD, MONGO_PASSWORD"
        echo ""
        read -p "Press enter to continue after editing .env (or Ctrl+C to abort)"
    else
        echo "Error: .env.template not found."
        exit 1
    fi
fi

echo "✓ Environment file ready"

# Build and start services
echo "Building and starting services..."
docker compose -f docker-compose.prod.yml up -d --build

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Services:"
echo "  - Frontend:   http://localhost:3000"
echo "  - Backend:    http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis:      localhost:6379"
echo "  - NATS:       localhost:4222"
echo "  - Ollama:     http://localhost:11434"
echo ""
echo "Commands:"
echo "  Check status: docker compose -f docker-compose.prod.yml ps"
echo "  View logs:    docker compose -f docker-compose.prod.yml logs -f"
echo "  Stop:         docker compose -f docker-compose.prod.yml down"
