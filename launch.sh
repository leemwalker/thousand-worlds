#!/bin/bash

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_COMPOSE_FILE="mud-platform-backend/deploy/docker-compose.yml"
STOP_DOCKER_ON_EXIT=false

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Function to get local IP address for mobile access
get_local_ip() {
    if command -v ipconfig &> /dev/null; then
        # macOS
        ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo "localhost"
    else
        # Linux
        hostname -I | awk '{print $1}' || echo "localhost"
    fi
}

# Function to kill background processes on exit
cleanup() {
    echo ""
    print_status "Stopping services..."
    
    # Kill all child processes of this script
    pkill -P $$ 2>/dev/null || true
    wait 2>/dev/null || true
    
    if [ "$STOP_DOCKER_ON_EXIT" = true ]; then
        print_status "Stopping Docker containers..."
        (cd mud-platform-backend && docker-compose -f deploy/docker-compose.yml down) || true
        print_status "Docker containers stopped."
    else
        print_info "Docker containers left running for faster restart"
        print_info "To stop: cd mud-platform-backend && docker-compose -f deploy/docker-compose.yml down"
    fi
    
    print_status "Services stopped."
}

# Trap SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

# Banner
echo "================================================"
echo "   Thousand Worlds - Full Stack Launch"
echo "   With Mobile SDK Support"
echo "================================================"
echo ""

# Check Docker prerequisites
print_status "Checking Docker prerequisites..."

if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    print_warning "Please install Docker from: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! docker info > /dev/null 2>&1; then
    print_error "Docker daemon is not running"
    print_warning "Please start Docker Desktop or the Docker daemon"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null 2>&1; then
    print_error "Docker Compose is not installed"
    print_warning "Please install Docker Compose"
    exit 1
fi

print_status "Docker is ready"

# Check Docker Compose file exists
if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
    print_error "Docker Compose file not found: $DOCKER_COMPOSE_FILE"
    exit 1
fi

# Start Docker infrastructure services
echo ""
print_status "Starting Docker infrastructure services..."
print_info "This will start: PostgreSQL, Redis, NATS, MongoDB, Ollama"
echo ""

(cd mud-platform-backend && docker-compose -f deploy/docker-compose.yml up -d) || {
    print_error "Failed to start Docker services"
    exit 1
}

print_status "Docker services starting..."

# Wait for PostgreSQL to be healthy
print_status "Waiting for PostgreSQL to be ready..."
MAX_WAIT=30
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if docker exec mud_postgis pg_isready -U admin -d mud_core > /dev/null 2>&1; then
        print_status "PostgreSQL is ready"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
        print_error "PostgreSQL failed to start within ${MAX_WAIT} seconds"
        cleanup
        exit 1
    fi
done

# Wait for Redis to be ready
print_status "Waiting for Redis to be ready..."
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if docker exec mud_redis redis-cli ping > /dev/null 2>&1; then
        print_status "Redis is ready"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
        print_warning "Redis failed to start within ${MAX_WAIT} seconds (non-critical)"
        break
    fi
done

# Check NATS
if docker ps --filter "name=mud_nats" --filter "status=running" | grep -q mud_nats; then
    print_status "NATS is running"
else
    print_warning "NATS may not be running (non-critical)"
fi

# Check Ollama
if docker ps --filter "name=mud_ollama" --filter "status=running" | grep -q mud_ollama; then
    print_status "Ollama is running"
    print_info "Pull LLM model with: docker exec mud_ollama ollama pull llama3.1:8b"
else
    print_warning "Ollama may not be running (non-critical for basic features)"
fi

# Check for required ports
print_status "Checking required ports..."
PORTS_OK=true

if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    # Check if it's a process other than Docker
    PORT_OWNER=$(lsof -Pi :8080 -sTCP:LISTEN -t)
    if ! docker ps --filter "name=mud" | grep -q .; then
        print_error "Port 8080 is already in use by process $PORT_OWNER (Backend)"
        PORTS_OK=false
    fi
fi

if lsof -Pi :5173 -sTCP:LISTEN -t >/dev/null 2>&1; then
    print_error "Port 5173 is already in use (Frontend/Vite)"
    PORTS_OK=false
fi

if [ "$PORTS_OK" = false ]; then
    print_warning "Please stop services using these ports or change port configuration"
    cleanup
    exit 1
fi
print_status "Required ports are available"

# Check Node dependencies for frontend
print_status "Checking frontend dependencies..."
if [ ! -d "mud-platform-client/node_modules" ]; then
    print_warning "Frontend dependencies not installed. Installing..."
    (cd mud-platform-client && npm install)
    print_status "Frontend dependencies installed"
else
    print_status "Frontend dependencies found"
fi

# Check Go dependencies for backend
print_status "Checking backend dependencies..."
(cd mud-platform-backend && go mod download) > /dev/null 2>&1
print_status "Backend dependencies ready"

# Run mobile SDK tests to ensure everything compiles
print_status "Running Mobile SDK tests..."
if (cd mud-platform-backend && go test ./internal/mobile/... -short > /dev/null 2>&1); then
    print_status "Mobile SDK tests passed"
else
    print_warning "Mobile SDK tests failed (non-critical, continuing...)"
fi

echo ""
echo "================================================"
echo "   Starting Application Services"
echo "================================================"
echo ""

# Start Backend
print_status "Starting Backend (port 8080)..."
(cd mud-platform-backend && ./scripts/dev.sh) &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 3

# Verify backend started
if ! lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    print_error "Backend failed to start"
    cleanup
    exit 1
fi
print_status "Backend is running"

# Start Frontend
print_status "Starting Frontend (port 5173)..."
(cd mud-platform-client && npm run dev) &
FRONTEND_PID=$!

# Wait a moment for frontend to start
sleep 3

# Verify frontend started
if ! lsof -Pi :5173 -sTCP:LISTEN -t >/dev/null 2>&1; then
    print_error "Frontend failed to start"
    cleanup
    exit 1
fi
print_status "Frontend is running"

# Get local IP for mobile access
LOCAL_IP=$(get_local_ip)

echo ""
echo "================================================"
echo "   All Services Running"
echo "================================================"
echo ""
echo "Infrastructure Services (Docker):"
echo "  - PostgreSQL:  localhost:5432"
echo "  - Redis:       localhost:6379"
echo "  - NATS:        localhost:4222"
echo "  - MongoDB:     localhost:27017"
echo "  - Ollama:      localhost:11434"
echo ""
echo "Application Services:"
echo "  - Backend:     http://localhost:8080"
echo "  - Frontend:    http://localhost:5173"
echo ""
echo "Mobile Device Access (same WiFi network):"
echo "  - Frontend:    http://${LOCAL_IP}:5173"
echo "  - Backend API: http://${LOCAL_IP}:8080"
echo "  - WebSocket:   ws://${LOCAL_IP}:8080/api/game/ws"
echo ""
print_info "Ensure your firewall allows connections on ports 8080 and 5173"
echo ""
echo "Quick Commands:"
echo "  - View logs:     docker-compose -f mud-platform-backend/deploy/docker-compose.yml logs -f"
echo "  - Stop Docker:   docker-compose -f mud-platform-backend/deploy/docker-compose.yml down"
echo "  - Pull LLM:      docker exec mud_ollama ollama pull llama3.1:8b"
echo ""
echo "Press Ctrl+C to stop all services"
echo "================================================"
echo ""

# Wait for all background processes
wait
