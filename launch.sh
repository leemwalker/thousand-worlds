#!/bin/bash

# Thousand Worlds - Full Stack Launch Script
# Version: 2.0 (Best Practices Edition)

#============================================================
# Strict Error Handling
#============================================================
set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Optional debug mode
if [ "${DEBUG:-}" = "1" ]; then
    set -x
fi

#============================================================
# Configuration (Override via Environment Variables)
#============================================================
BACKEND_DIR="${BACKEND_DIR:-tw-backend}"
FRONTEND_DIR="${FRONTEND_DIR:-tw-frontend}"
DOCKER_COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-${BACKEND_DIR}/deploy/docker-compose.yml}"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"
POSTGRES_TIMEOUT="${POSTGRES_TIMEOUT:-30}"
REDIS_TIMEOUT="${REDIS_TIMEOUT:-30}"
SERVICE_START_DELAY="${SERVICE_START_DELAY:-3}"
STOP_DOCKER_ON_EXIT="${STOP_DOCKER_ON_EXIT:-false}"
RUN_MOBILE_TESTS="${RUN_MOBILE_TESTS:-true}"
SKIP_MIGRATIONS="${SKIP_MIGRATIONS:-false}"
FORCE_RECREATE="${FORCE_RECREATE:-false}"
VERBOSE="${VERBOSE:-false}"

# Logging configuration
LOG_DIR="${LOG_DIR:-./logs}"
BACKEND_LOG="${LOG_DIR}/backend.log"
FRONTEND_LOG="${LOG_DIR}/frontend.log"

# Global state
BG_PIDS=()
CLEANUP_IN_PROGRESS="false"

#============================================================
# Colors for Output
#============================================================
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

#============================================================
# Logging Functions
#============================================================
log_message() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log to file if LOG_FILE is set
    if [ "${LOG_FILE:-}" != "" ]; then
        echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
    fi
}

print_status() {
    echo -e "${GREEN}✓${NC} $1"
    log_message "INFO" "$1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    log_message "WARNING" "$1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
    log_message "ERROR" "$1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
    log_message "INFO" "$1"
}

print_verbose() {
    if [ "$VERBOSE" = "true" ]; then
        echo -e "${BLUE}→${NC} $1"
        log_message "VERBOSE" "$1"
    fi
}

# === ENVIRONMENT VARIABLE LOADING ===
# Load .env file if it exists (for local development)
if [ -f .env ]; then
    print_verbose "Loading environment from .env file"
    set -a  # automatically export all variables
    source .env
    set +a
fi

# === VALIDATE REQUIRED SECRETS ===
if [ -z "$JWT_SECRET" ] || [ -z "$POSTGRES_PASSWORD" ]; then
    print_error "Required environment variables not set!"
    echo ""
    echo "The following environment variables must be set:"
    echo "  - JWT_SECRET (generate with: openssl rand -hex 32)"
    echo "  - POSTGRES_PASSWORD (generate with: openssl rand -hex 16)"
    echo ""
    echo "Create a .env file or export these variables."
    echo "See .env.example for reference."
    exit 1
fi

# Database configuration
export POSTGRES_PASSWORD  # Ensure this is exported to subshells
export DATABASE_URL="${DATABASE_URL:-postgresql://admin:${POSTGRES_PASSWORD}@localhost:5432/mud_core?sslmode=disable}"

# Normalize Docker Compose command (v1 vs v2)
get_docker_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    elif docker compose version &> /dev/null 2>&1; then
        echo "docker compose"
    else
        print_error "Neither 'docker-compose' nor 'docker compose' found"
        exit 1
    fi
}

DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
declare -a BG_PIDS=()
CLEANUP_IN_PROGRESS=false

#============================================================
# Error Handler
#============================================================
error_handler() {
    local exit_code=$1
    local line_number=$2
    
    if [ "$CLEANUP_IN_PROGRESS" = "true" ]; then
        return
    fi
    
    print_error "Error occurred at line $line_number (exit code: $exit_code)"
    print_info "Enable debug mode with: DEBUG=1 $0"
    cleanup
    exit $exit_code
}

trap 'error_handler $? $LINENO' ERR

#============================================================
# Usage/Help Function
#============================================================
usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

COMMANDS:
    start       Start all services (default)
    stop        Stop all services  
    restart     Restart all services
    status      Show service status

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -d, --debug             Enable debug mode (set -x)
    -f, --force-recreate    Force recreate Docker containers
    -s, --skip-migrations   Skip database migrations
    -c, --stop-docker       Stop Docker containers on exit
    --port PORT             Backend port (default: 8080)
    --frontend-port PORT    Frontend port (default: 5173)
    
ENVIRONMENT VARIABLES:
    DEBUG=1                 Enable debug mode
    VERBOSE=1               Enable verbose output
    FORCE_RECREATE=1        Force recreate containers
    SKIP_MIGRATIONS=1       Skip database migrations
    STOP_DOCKER_ON_EXIT=1   Stop Docker on exit
    LOG_FILE=path           Write logs to file
    BACKEND_PORT=port       Custom backend port
    FRONTEND_PORT=port      Custom frontend port

EXAMPLES:
    # Start with verbose output
    $0 --verbose
    
    # Force recreate containers
    $0 --force-recreate
    
    # Use custom ports
    $0 --port 9000 --frontend-port 3000
    
    # Enable debug mode
    DEBUG=1 $0

For more information, visit: https://github.com/your-username/thousand-worlds
EOF
}

#============================================================
# Argument Parsing
#============================================================
COMMAND="start"

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) 
            usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            ;;
        -d|--debug)
            set -x
            DEBUG=1
            ;;
        -f|--force-recreate)
            FORCE_RECREATE=true
            ;;
        -s|--skip-migrations)
            SKIP_MIGRATIONS=true
            ;;
        -c|--stop-docker)
            STOP_DOCKER_ON_EXIT=true
            ;;
        --port)
            BACKEND_PORT=$2
            shift
            ;;
        --frontend-port)
            FRONTEND_PORT=$2
            shift
            ;;
        start|stop|restart|status|logs)
            COMMAND=$1
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
    shift
done

#============================================================
# Subcommand Handlers
#============================================================

handle_stop() {
    print_status "Stopping all services..."
    
    # Stop application services
    if [ -f "${LOG_DIR}/backend.pid" ]; then
        local backend_pid=$(cat "${LOG_DIR}/backend.pid")
        if kill -0 $backend_pid 2>/dev/null; then
            print_info "Stopping backend (PID: $backend_pid)..."
            kill -TERM $backend_pid 2>/dev/null || true
        fi
    fi
    
    if [ -f "${LOG_DIR}/frontend.pid" ]; then
        local frontend_pid=$(cat "${LOG_DIR}/frontend.pid")
        if kill -0 $frontend_pid 2>/dev/null; then
            print_info "Stopping frontend (PID: $frontend_pid)..."
            kill -TERM $frontend_pid 2>/dev/null || true
        fi
    fi
    
    # Stop Docker services
    print_info "Stopping Docker services..."
    (cd "$BACKEND_DIR" && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml down)
    
    print_status "All services stopped"
}

handle_restart() {
    print_status "Restarting all services..."
    handle_stop
    # After stop, re-run start command
    exec "$0" start
}

handle_status() {
    echo "================================================"
    echo "   Service Status"
    echo "================================================"
    echo ""
    
    # Check Docker services
    echo "Docker Services:"
    if (cd "$BACKEND_DIR" && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml ps); then
        echo ""
    else
        print_warning "Failed to get Docker service status"
    fi
    
    # Check application services
    echo "Application Services:"
    
    if [ -f "${LOG_DIR}/backend.pid" ]; then
        local backend_pid=$(cat "${LOG_DIR}/backend.pid")
        if kill -0 $backend_pid 2>/dev/null; then
            print_status "Backend running (PID: $backend_pid, Port: $BACKEND_PORT)"
            if check_backend_health; then
                print_status "  Health check: OK"
            else
                print_warning "  Health check: FAILED"
            fi
        else
            print_warning "Backend not running (stale PID file)"
        fi
    else
        print_warning "Backend not running (no PID file)"
    fi
    
    if [ -f "${LOG_DIR}/frontend.pid" ]; then
        local frontend_pid=$(cat "${LOG_DIR}/frontend.pid")
        if kill -0 $frontend_pid 2>/dev/null; then
            print_status "Frontend running (PID: $frontend_pid, Port: $FRONTEND_PORT)"
            if check_frontend_health; then
                print_status "  Health check: OK"
            else
                print_warning "  Health check: FAILED"
            fi
        else
            print_warning "Frontend not running (stale PID file)"
        fi
    else
        print_warning "Frontend not running (no PID file)"
    fi
    
    echo ""
    echo "================================================"
    exit 0
}

handle_logs() {
    echo "Tailing logs (Ctrl+C to exit)..."
    echo ""
    
    # Use multitail if available, otherwise fall back to tail
    if command -v multitail &> /dev/null; then
        multitail -s 2 -l "tail -f $BACKEND_LOG" -l "tail -f $FRONTEND_LOG"
    else
        # Simple tail fallback
        print_info "Backend logs: $BACKEND_LOG"
        print_info "Frontend logs: $FRONTEND_LOG"
        print_info "For better log viewing, install multitail: brew install multitail"
        echo ""
        tail -f "$BACKEND_LOG" "$FRONTEND_LOG"
    fi
    
    exit 0
}

# Handle subcommands
case "$COMMAND" in
    stop)
        handle_stop
        exit 0
        ;;
    restart)
        handle_restart
        ;;
    status)
        handle_status
        ;;
    logs)
        handle_logs
        ;;
    start)
        # Continue with normal start flow below
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        usage
        exit 1
        ;;
esac

#============================================================
# Utility Functions
#============================================================

# Get local IP address for mobile access
get_local_ip() {
    if command -v ipconfig &> /dev/null; then
        # macOS
        ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo "localhost"
    else
        # Linux
        hostname -I | awk '{print $1}' || echo "localhost"
    fi
}

# Get container name dynamically
get_container_name() {
    local service=$1
    (cd "$BACKEND_DIR" && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml ps -q "$service" 2>/dev/null | xargs docker inspect -f '{{.Name}}' 2>/dev/null | sed 's/^\///')
}

# Health check functions
check_postgres_health() {
    local container=$(get_container_name postgres)
    if [ -z "$container" ]; then
        container="mud_postgis"  # Fallback to hardcoded name
    fi
    docker exec "$container" pg_isready -U admin -d mud_core 2>/dev/null
}

check_redis_health() {
    local container=$(get_container_name redis)
    if [ -z "$container" ]; then
        container="mud_redis"  # Fallback
    fi
    docker exec "$container" redis-cli ping 2>/dev/null | grep -q PONG
}

check_backend_health() {
    # Check if backend health endpoint responds
    if command -v curl &> /dev/null; then
        curl -sf http://localhost:${BACKEND_PORT}/health > /dev/null 2>&1
    else
        # Fallback to port check
        check_port $BACKEND_PORT
    fi
}

check_frontend_health() {
    # Check if frontend is serving content
    if command -v curl &> /dev/null; then
        curl -sf http://localhost:${FRONTEND_PORT} > /dev/null 2>&1
    else
        check_port $FRONTEND_PORT
    fi
}

# Check if port is available
check_port() {
    local port=$1
    if command -v lsof &> /dev/null; then
        if lsof -Pi :$port -sTCP:LISTEN -t > /dev/null 2>&1; then
            return 0
        fi
    elif command -v netstat &> /dev/null; then
        if netstat -an | grep ":$port.*LISTEN" > /dev/null 2>&1; then
            return 0
        fi
    elif command -v ss &> /dev/null; then
        if ss -ltn | grep ":$port" > /dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# Wait for service with exponential backoff
wait_for_service() {
    local service_name=$1
    local check_command=$2
    local max_wait=${3:-30}
    local retry_delay=1
    local wait_count=0
    
    print_verbose "Waiting for $service_name to be ready (max ${max_wait}s)..."
    
    while [ $wait_count -lt $max_wait ]; do
        if eval "$check_command" > /dev/null 2>&1; then
            print_status "$service_name is ready"
            return 0
        fi
        sleep $retry_delay
        wait_count=$((wait_count + 1))
        
        # Exponential backoff (max 5 seconds)
        if [ $retry_delay -lt 5 ]; then
            retry_delay=$((retry_delay + 1))
        fi
    done
    
    print_warning "$service_name failed to be ready within ${max_wait}s"
    return 1
}

# Run database migrations
run_migrations() {
    if [ "$SKIP_MIGRATIONS" = "true" ]; then
        print_info "Skipping database migrations"
        return 0
    fi
    
    print_status "Running database migrations..."
    
    cd "$BACKEND_DIR"
    
    # Use the Go-based migrator (same as dev.sh) - handles missing .down.sql files gracefully
    if go run cmd/migrate/main.go 2>&1; then
        print_status "Migrations completed successfully"
    else
        print_error "Migration failed"
        cd - > /dev/null
        return 1
    fi
    
    cd - > /dev/null
    return 0
}

#============================================================
# Cleanup Function
#============================================================
cleanup() {
    if [ "$CLEANUP_IN_PROGRESS" = "true" ]; then
        return
    fi
    
    CLEANUP_IN_PROGRESS=true
    echo ""
    print_status "Stopping services gracefully..."
    
    # Send SIGTERM to all tracked processes (check if array is non-empty first)
    if [ ${#BG_PIDS[@]} -gt 0 ]; then
        for pid in "${BG_PIDS[@]}"; do
            if kill -0 $pid 2>/dev/null; then
                print_verbose "Sending SIGTERM to PID $pid"
                kill -TERM $pid 2>/dev/null || true
            fi
        done
        
        # Wait up to 10 seconds for graceful shutdown
        local timeout=10
        while [ $timeout -gt 0 ]; do
            local all_stopped=true
            for pid in "${BG_PIDS[@]}"; do
                if kill -0 "$pid" 2>/dev/null; then
                    all_stopped=false
                    break
                fi
            done
            
            if [ "$all_stopped" = true ]; then
                break
            fi
            
            sleep 1
            timeout=$((timeout - 1))
        done
        
        # Force kill if still running
        for pid in "${BG_PIDS[@]}"; do
            if kill -0 $pid 2>/dev/null; then
                print_warning "Force killing PID $pid"
                kill -9 $pid 2>/dev/null || true
            fi
        done
    fi
    
    # Docker cleanup
    if [ "$STOP_DOCKER_ON_EXIT" = "true" ]; then
        print_status "Stopping Docker containers..."
        (cd "$BACKEND_DIR" && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml down) || true
        print_status "Docker containers stopped."
    else
        print_info "Docker containers left running for faster restart"
        print_info "To stop: cd $BACKEND_DIR && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml down"
    fi
    
    print_status "Services stopped."
}

# Trap signals
trap cleanup SIGINT SIGTERM EXIT

#============================================================
# Main Script
#============================================================

# Create log directory
mkdir -p "$LOG_DIR"

# Banner
echo "================================================"
echo "   Thousand Worlds - Full Stack Launch"
echo "   Version 2.0 - Best Practices Edition"
echo "================================================"
echo ""

# Security check - warn if running as root
if [ "${EUID:-$(id -u)}" -eq 0 ]; then
    print_warning "Running as root is not recommended"
    print_info "Press Ctrl+C to cancel, or wait 3 seconds to continue..."
    sleep 3
fi

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

print_status "Docker is ready"

# Check Docker Compose file exists
if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
    print_error "Docker Compose file not found: $DOCKER_COMPOSE_FILE"
    exit 1
fi

# Check if Docker services are already running
check_docker_services() {
    if (cd "$BACKEND_DIR" && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml ps | grep -q "Up"); then
        print_info "Docker services are already running"
        
        if [ "$FORCE_RECREATE" = "true" ]; then
            print_warning "Force recreate enabled, restarting services..."
            return 1
        else
            print_info "Skipping Docker startup (use -f or FORCE_RECREATE=1 to restart)"
            return 0
        fi
    fi
    return 1
}

# Start Docker infrastructure services
if ! check_docker_services; then
    echo ""
    print_status "Starting Docker infrastructure services..."
    print_info "This will start: PostgreSQL, Redis, NATS, MongoDB, Ollama"
    echo ""
    
    (cd "$BACKEND_DIR" && $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml up -d) || {
        print_error "Failed to start Docker services"
        exit 1
    }
    
    print_status "Docker services starting..."
fi

# Wait for PostgreSQL with robust health check
wait_for_service "PostgreSQL" "check_postgres_health" "$POSTGRES_TIMEOUT" || {
    print_error "PostgreSQL failed to start"
    exit 1
}

# Wait for Redis with robust health check
wait_for_service "Redis" "check_redis_health" "$REDIS_TIMEOUT" || {
    print_warning "Redis failed to start (non-critical)"
}

# Run database migrations
if ! run_migrations; then
    print_error "Database migrations failed"
    print_info "Fix migrations or use --skip-migrations to bypass"
    exit 1
fi

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

if check_port $BACKEND_PORT; then
    # Check if it's us or someone else
    if ! docker ps --filter "name=mud" | grep -q mud; then
        print_error "Port $BACKEND_PORT is already in use (Backend)"
        PORTS_OK=false
    fi
fi

if check_port $FRONTEND_PORT; then
    print_error "Port $FRONTEND_PORT is already in use (Frontend/Vite)"
    PORTS_OK=false
fi

if [ "$PORTS_OK" = false ]; then
    print_warning "Please stop services using these ports or use --port/--frontend-port"
    exit 1
fi
print_status "Required ports are available"

# Check Node dependencies for frontend
print_status "Checking frontend dependencies..."
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
    print_warning "Frontend dependencies not installed. Installing..."
    (cd "$FRONTEND_DIR" && npm install)
    print_status "Frontend dependencies installed"
else
    print_status "Frontend dependencies found"
fi

# Check Go dependencies for backend
print_status "Checking backend dependencies..."
(cd "$BACKEND_DIR" && go mod download) > /dev/null 2>&1
print_status "Backend dependencies ready"

# Run mobile SDK tests
if [ "$RUN_MOBILE_TESTS" = "true" ]; then
    print_status "Running Mobile SDK tests..."
    if (cd "$BACKEND_DIR" && go test ./internal/mobile/... -short > /dev/null 2>&1); then
        print_status "Mobile SDK tests passed"
    else
        print_warning "Mobile SDK tests failed (non-critical, continuing...)"
    fi
fi

echo ""
echo "================================================"
echo "   Starting Application Services"
echo "================================================"
echo ""

# Start Backend
print_status "Starting Backend (port $BACKEND_PORT)..."
print_verbose "Backend logs: $BACKEND_LOG"
(cd "$BACKEND_DIR" && ./scripts/dev.sh) > "$BACKEND_LOG" 2>&1 &
BACKEND_PID=$!
BG_PIDS+=($BACKEND_PID)
echo $BACKEND_PID > "${LOG_DIR}/backend.pid"

# Wait for backend to start
sleep $SERVICE_START_DELAY

# Verify backend started
if ! check_port $BACKEND_PORT; then
    print_error "Backend failed to start"
    print_info "Check logs: tail -f $BACKEND_LOG"
    exit 1
fi
print_status "Backend is running (PID: $BACKEND_PID)"

# Start Frontend
print_status "Starting Frontend (port $FRONTEND_PORT)..."
print_verbose "Frontend logs: $FRONTEND_LOG"
(cd "$FRONTEND_DIR" && npm run dev) > "$FRONTEND_LOG" 2>&1 &
FRONTEND_PID=$!
BG_PIDS+=($FRONTEND_PID)
echo $FRONTEND_PID > "${LOG_DIR}/frontend.pid"

# Wait for frontend to start
sleep $SERVICE_START_DELAY

# Verify frontend started
if ! check_port $FRONTEND_PORT; then
    print_error "Frontend failed to start"
    print_info "Check logs: tail -f $FRONTEND_LOG"
    exit 1
fi
print_status "Frontend is running (PID: $FRONTEND_PID)"

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
echo "  - Backend:     http://localhost:$BACKEND_PORT"
echo "  - Frontend:    http://localhost:$FRONTEND_PORT"
echo ""
echo "Mobile Device Access (same WiFi network):"
echo "  - Frontend:    http://${LOCAL_IP}:$FRONTEND_PORT"
echo "  - Backend API: http://${LOCAL_IP}:$BACKEND_PORT"
echo "  - WebSocket:   ws://${LOCAL_IP}:$BACKEND_PORT/api/game/ws"
echo ""
echo "Logs:"
echo "  - Backend:     tail -f $BACKEND_LOG"
echo "  - Frontend:    tail -f $FRONTEND_LOG"
echo "  - Docker:      $DOCKER_COMPOSE_CMD -f $DOCKER_COMPOSE_FILE logs -f"
echo ""
print_info "Ensure your firewall allows connections on ports $BACKEND_PORT and $FRONTEND_PORT"
echo ""
echo "Quick Commands:"
echo "  - View logs:     $DOCKER_COMPOSE_CMD -f $DOCKER_COMPOSE_FILE logs -f"
echo "  - Stop Docker:   $DOCKER_COMPOSE_CMD -f $DOCKER_COMPOSE_FILE down"
echo "  - Pull LLM:      docker exec mud_ollama ollama pull llama3.1:8b"
echo ""
echo "Press Ctrl+C to stop all services"
echo "================================================"
echo ""

# Wait for all background processes
wait
