# Thousand Worlds - Backend

A Go 1.24 microservices backend for the Thousand Worlds MUD Platform, featuring event sourcing, spatial queries with PostGIS, real-time WebSocket communication, and LLM-powered world generation.

## 🛠️ Technology Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.24 |
| **Web Framework** | Chi Router |
| **Databases** | PostgreSQL 14+ (PostGIS), MongoDB 7+, Redis 7+ |
| **Messaging** | NATS JetStream |
| **AI** | Ollama (Llama 3.1) |
| **Monitoring** | Prometheus, zerolog |
| **Testing** | testify, testcontainers |

---

## 📁 Project Structure

```
tw-backend/
├── cmd/                          # Service entry points
│   ├── game-server/              # Main API + WebSocket server
│   │   ├── api/                  # HTTP handlers (auth, session, health)
│   │   └── websocket/            # WebSocket hub and client handlers
│   ├── auth-service/             # Authentication microservice
│   ├── ai-gateway/               # LLM integration gateway
│   ├── world-service/            # World ticker service
│   ├── player-service/           # Player management
│   ├── admin/                    # Admin utilities
│   └── migrate/                  # Database migration tool
│
├── internal/                     # Business logic packages
│   ├── ai/                       # LLM client, prompts, caching
│   ├── auth/                     # JWT, passwords, sessions, rate limiting
│   ├── cache/                    # Multi-level caching (L1/L2)
│   ├── character/                # Character creation, attributes
│   ├── combat/                   # Action queue, damage calculation, status effects
│   ├── economy/                  # Crafting, trading, resources, tech trees
│   ├── ecosystem/                # Geological processes, environmental simulation
│   ├── errors/                   # Custom error types
│   ├── eventstore/               # Event sourcing, CQRS, replay engine
│   ├── formatter/                # Output formatting utilities
│   ├── game/                     # Command processing, lobby, game services
│   ├── health/                   # Health check endpoints
│   ├── item/                     # Item definitions, properties
│   ├── logging/                  # Structured logging with zerolog
│   ├── memory/                   # In-memory data structures
│   ├── metrics/                  # Prometheus metrics collection
│   ├── mobile/                   # Mobile-optimized endpoints
│   ├── nats/                     # NATS event listener
│   ├── npc/                      # NPC systems (genetics, memory, behavior, dialogue)
│   ├── player/                   # Stamina, movement, inventory
│   ├── pubsub/                   # Pub/sub abstractions
│   ├── repository/               # Database repositories
│   ├── service/                  # Service layer abstractions
│   ├── skills/                   # Skill system, progression
│   ├── spatial/                  # PostGIS queries, coordinate systems
│   ├── testutil/                 # Test utilities and mocks
│   ├── validation/               # Input validation
│   ├── world/                    # World management, interview
│   ├── worldentity/              # World entity definitions
│   └── worldgen/                 # Procedural generation (geography, weather, flora/fauna)
│
├── migrations/postgres/          # Database migrations (60 files)
├── scripts/                      # Development and deployment scripts
├── deploy/                       # Docker and deployment configs
├── tests/                        # Integration and E2E tests
└── data/                         # Static data (recipes, tech trees)
```

---

## 🔧 Internal Packages

### Core Systems

| Package | Description |
|---------|-------------|
| `auth` | JWT token generation/validation with AES-256, Argon2id password hashing, Redis session management, rate limiting |
| `game` | Command processing, lobby system, game services, entity management |
| `player` | Player stamina system, coordinate-based movement, inventory management |
| `character` | Character creation (inhabit NPC or generate new), 15 attributes, point-buy system |

### World Systems

| Package | Description |
|---------|-------------|
| `world` | World CRUD operations, world state management, interview system |
| `worldgen` | Procedural generation: tectonic plates, heightmaps, biomes, weather, flora/fauna |
| `spatial` | PostGIS integration, 3D coordinate system (X, Y, Z), spatial indexing, radius queries |
| `ecosystem` | Geological processes, environmental simulation, terrain evolution |

### NPC Systems

| Package | Description |
|---------|-------------|
| `npc/genetics` | Mendelian inheritance, trait mutations, appearance generation |
| `npc/memory` | MongoDB-backed memory with decay/rehearsal, emotional weighting |
| `npc/behavior` | Desire engine, personality-driven actions |
| `npc/dialogue` | LLM-enhanced conversations |
| `npc/relationships` | Affection, trust, fear tracking |
| `combat` | Real-time action queue, damage calculation, status effects |
| `skills` | Use-based progression, skill categories, checks and synergies |

### Economy & Items

| Package | Description |
|---------|-------------|
| `economy` | Crafting system, trading, resource distribution, tech trees |
| `item` | Item definitions, properties, durability |

### Infrastructure

| Package | Description |
|---------|-------------|
| `ai` | Ollama LLM client, prompt templates, aggressive caching (15-min TTL) |
| `cache` | Multi-level caching (memory L1, Redis L2) |
| `eventstore` | Append-only event log, CQRS read models, event replay engine |
| `pubsub` | NATS pub/sub abstractions |
| `metrics` | Prometheus metrics collection |
| `health` | Health check endpoints for all services |
| `logging` | Structured JSON logging with correlation IDs |
| `testutil` | Mock implementations, test helpers |
| `validation` | Input validation utilities |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 14+ with PostGIS extension
- Redis 7+
- MongoDB 7+
- (Optional) Ollama with Llama 3.1 8B for AI features

### 1. Start Infrastructure Services

```bash
docker-compose -f docker-compose.prod.yml up -d postgis mongo nats redis
```

### 2. Run Database Migrations

```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations/postgres -database "postgresql://admin:password123@localhost:5432/mud_core?sslmode=disable" up
```

### 3. Set Environment Variables

```bash
export DATABASE_URL="postgresql://admin:password123@localhost:5432/mud_core?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export NATS_URL="nats://localhost:4222"
export OLLAMA_HOST="http://localhost:11434"
export JWT_SECRET="your-secret-key"
export PORT="8080"
```

Or copy and configure from template:
```bash
cp .env.template .env
# Edit .env with your values
```

### 4. Run the Server

```bash
# Option A: Direct run
go run cmd/game-server/main.go

# Option B: Use dev script
./scripts/dev.sh
```

Server available at: **http://localhost:8080**

---

## 📡 API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/register` | Create new account |
| `POST` | `/api/auth/login` | Authenticate and receive JWT |
| `GET` | `/api/auth/me` | Get current user (requires auth) |

### Session Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/session/state` | Get current session state |
| `POST` | `/api/session/world/select` | Select a world to enter |
| `POST` | `/api/session/character/create` | Create new character |
| `POST` | `/api/session/character/select` | Select existing character |

### World Interview

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/interview/start` | Start world creation interview |
| `POST` | `/api/interview/respond` | Send interview response |
| `POST` | `/api/interview/finalize` | Complete interview and generate world |

### WebSocket

| Protocol | Endpoint | Description |
|----------|----------|-------------|
| `WS` | `/api/game/ws` | Game WebSocket connection (requires auth) |

### Health & Metrics

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Service health check |
| `GET` | `/metrics` | Prometheus metrics |

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output and coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/character/...
go test ./internal/combat/...

# Run E2E tests
go test -v ./tests/e2e/...

# Run benchmarks
go test -bench=. ./...
```

### Coverage Target
All packages should maintain **80%+ code coverage**.

---

## 🗄️ Database Schema

### Event Store
```sql
events (
  id UUID PRIMARY KEY,
  event_type VARCHAR,
  aggregate_id UUID,
  version INT,
  timestamp TIMESTAMPTZ,
  payload JSONB
)
```

### Worlds
```sql
worlds (
  id UUID PRIMARY KEY,
  name VARCHAR,
  shape VARCHAR,  -- 'spherical', 'bounded_cube', 'infinite'
  radius NUMERIC,
  bounds JSONB,
  owner_id UUID
)
```

### Spatial Entities
```sql
entities (
  id UUID PRIMARY KEY,
  world_id UUID,
  position GEOMETRY(POINTZ, 4326),  -- X, Y, Z coordinates
  entity_type VARCHAR
)
```

---

## 📜 Scripts

| Script | Description |
|--------|-------------|
| `scripts/dev.sh` | Start development server with hot reload |
| `scripts/deploy.sh` | Deploy to production |
| `scripts/validate-env.sh` | Validate environment variables |
| `scripts/verify_security.sh` | Run security verification tests |
| `scripts/verify_session.sh` | Verify session management |
| `scripts/verify_interview.sh` | Test interview endpoints |
| `scripts/verify_websocket.js` | WebSocket connection test |

---

## 🐳 Docker

### Build Game Server Image
```bash
docker build -f Dockerfile.game-server -t thousand-worlds/game-server:latest .
```

### Run Full Stack
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Pull LLM Model
```bash
docker exec mud_ollama ollama pull llama3.1:8b
```

---

## 📚 Related Documentation

- [API Specification (OpenAPI)](api/openapi.yaml) - Full Swagger/OpenAPI documentation
- [DEPLOYMENT.md](DEPLOYMENT.md) - Production deployment guide
- [../features.md](../features.md) - Detailed feature specifications
- [../roadmap.md](../roadmap.md) - Development roadmap
- [../SECURITY.md](../SECURITY.md) - Security documentation
