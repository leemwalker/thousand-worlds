# Service Architecture

The Thousand Worlds backend uses a microservices architecture with 7 services communicating via NATS.

## Service Overview

| Service | Status | Port/Protocol | Purpose |
|---------|--------|---------------|---------|
| `game-server` | ✅ Production | HTTP :8080, WebSocket | Primary API, game loop, WebSocket connections |
| `world-service` | ⚠️ Partial | NATS | World tickers, weather simulation |
| `ai-gateway` | ✅ Ready | NATS | LLM request routing to Ollama |
| `auth-service` | ⚠️ Partial | NATS | Token validation, rate limiting |
| `player-service` | 🔲 Scaffold | NATS | Player state management (TODO) |
| `migrate` | ✅ Production | CLI | Database migrations |
| `admin` | ✅ Production | CLI | Data management utilities |

---

## Service Details

### game-server (Primary)
**Location**: `cmd/game-server/`  
**Status**: ✅ Fully Integrated  
**Files**: 21 files including handlers, websocket, middleware

The primary entry point for the platform. Handles:
- REST API endpoints (`/api/*`)
- WebSocket connections for real-time gameplay
- User authentication (JWT + cookies)
- Character management
- World interview (LLM-based world creation)
- Game command processing

**Dependencies**: PostgreSQL, Redis, NATS (optional), Ollama

```bash
go run cmd/game-server/main.go
```

---

### world-service
**Location**: `cmd/world-service/`  
**Status**: ⚠️ Partially Integrated  
**Files**: 1 file (159 lines)

Manages world simulation tickers and weather. Currently:
- ✅ Connects to PostgreSQL and NATS
- ✅ Initializes event store, weather service
- ✅ Creates world registry and ticker manager
- ⏳ TODO: NATS command subscriptions (world.create, world.pause, etc.)

**Dependencies**: PostgreSQL, NATS

```bash
go run cmd/world-service/main.go
```

---

### ai-gateway
**Location**: `cmd/ai-gateway/`  
**Status**: ✅ Ready  
**Files**: 1 file (50 lines)

NATS-based service that routes LLM requests to Ollama. Runs as a separate worker to offload AI processing from game-server.

**Features**:
- Subscribes to AI request topics via NATS
- Routes to Ollama client
- Supports parallel processing via gateway worker

**Dependencies**: NATS, Ollama

```bash
go run cmd/ai-gateway/main.go
```

---

### auth-service
**Location**: `cmd/auth-service/`  
**Status**: ⚠️ Partially Integrated  
**Files**: 4 files (handlers + tests)

NATS-based authentication service with:
- ✅ Token management (JWT signing/encryption)
- ✅ Password hashing (Argon2id)
- ✅ Session management (Redis)
- ✅ Rate limiting
- ✅ Login subscription (`auth.login`)

**Note**: game-server has its own auth handling. This service is for distributed deployments.

**Dependencies**: NATS, Redis

```bash
go run cmd/auth-service/main.go
```

---

### player-service
**Location**: `cmd/player-service/`  
**Status**: 🔲 Scaffold Only  
**Files**: 1 file (37 lines)

Placeholder for player state management. Contains:
- NATS connection setup
- TODO comments for event store, repository, service initialization

**Not ready for deployment.**

---

### migrate
**Location**: `cmd/migrate/`  
**Status**: ✅ Production Ready  
**Files**: 1 file (70 lines)

Database migration tool that:
- Enables PostGIS extension
- Runs all `*.up.sql` files in `migrations/postgres/`
- Handles "already exists" errors gracefully

```bash
go run cmd/migrate/main.go
# or
make migrate
```

---

### admin
**Location**: `cmd/admin/`  
**Status**: ✅ Production Ready  
**Files**: 1 file (wipe_data.go)

Administrative utilities:
- `wipe_data.go`: Deletes all users (cascades) and non-Lobby worlds

```bash
go run cmd/admin/wipe_data.go
```

---

## Communication Patterns

```
┌─────────────┐     HTTP/WS      ┌──────────────┐
│   Frontend  │ ───────────────► │ game-server  │
└─────────────┘                  └──────┬───────┘
                                        │ NATS
                    ┌───────────────────┼───────────────────┐
                    │                   │                   │
                    ▼                   ▼                   ▼
            ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
            │  ai-gateway   │   │ world-service │   │ auth-service  │
            └───────────────┘   └───────────────┘   └───────────────┘
                    │                   │                   │
                    ▼                   ▼                   ▼
                 Ollama             PostgreSQL            Redis
```

## Deployment Modes

### Monolith (Development)
Run only `game-server` - it contains all functionality inline.

### Microservices (Production)
Run separate services for:
- Scalability (multiple ai-gateways for LLM requests)
- Isolation (auth-service can be rate-limited independently)
- Resilience (world-service can restart without affecting API)
