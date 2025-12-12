# Thousand Worlds MUD Platform

A next-generation text-based multiplayer game combining modern infrastructure with innovative gameplay. Experience living, breathing worlds where NPCs have memories, genetics matter, and every item tells a story.

## 🎮 What is Thousand Worlds?

Thousand Worlds is an ambitious MUD (Multi-User Dungeon) platform featuring:

- **🤖 AI-Driven World Creation**: LLM-guided interviews create unique, procedurally-generated worlds
- **🧬 Advanced NPC System**: NPCs with genetic inheritance, memory, relationships, and emergent behavior
- **🌍 Multi-World Coordinate System**: Support for spherical planets, bounded spaces, and infinite worlds
- **⚔️ Real-Time Combat**: Action queue system with stamina, status effects, and skill-based progression
- **🎭 Dual Character Creation**: Inhabit existing NPCs or generate custom characters with point-buy system
- **📱 Mobile-First PWA**: Progressive Web App optimized for mobile gameplay
- **🔄 Event Sourcing**: Complete audit trail with time-travel debugging capabilities

For a detailed feature list, see [`features.md`](features.md).

---

## 📁 Project Structure

```
thousand-worlds/
├── tw-backend/               # Go microservices backend
│   ├── cmd/                  # Service entry points
│   │   ├── game-server/     # Main game server (API + WebSocket)
│   │   ├── auth-service/    # Authentication service
│   │   ├── ai-gateway/      # LLM integration gateway
│   │   └── world-service/   # World ticker service
│   └── internal/            # Business logic
│       ├── auth/           # JWT, passwords, sessions, rate limiting
│       ├── character/      # Character creation, attributes
│       ├── combat/         # Action queue, damage calculation
│       ├── economy/        # Crafting, trading, resources
│       ├── eventstore/     # Event sourcing, CQRS
│       ├── game/           # Command processing, lobby
│       ├── npc/            # Genetics, memory, behavior, dialogue
│       ├── player/         # Stamina, movement, inventory
│       ├── skills/         # Skill system, progression
│       ├── spatial/        # PostGIS queries, coordinate systems
│       ├── world/          # World generation, interview
│       └── worldgen/       # Procedural generation algorithms
│
├── tw-frontend/              # SvelteKit PWA frontend
│   ├── src/
│   │   ├── routes/          # SvelteKit pages
│   │   │   ├── +page.svelte    # Landing/login page
│   │   │   └── game/           # Game interface
│   │   └── lib/
│   │       ├── components/     # UI components
│   │       ├── services/       # API clients
│   │       ├── stores/         # State management
│   │       └── types/          # TypeScript types
│   └── static/              # PWA assets, icons
│
└── tw-backend/deploy/        # Infrastructure
    ├── docker-compose.prod.yml
    └── postgres/            # Database migrations
```

---

## 🚀 Quick Start

### Prerequisites

**Backend:**
- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 14+ with PostGIS
- Redis 7+
- MongoDB 7+
- (Optional) Ollama with Llama 3.1 8B for AI features

**Frontend:**
- Node.js 18+
- npm 9+

### Launch the Platform

The easiest way to start both backend and frontend services:

```bash
./launch.sh
```

This script:
1. Starts Docker containers (PostgreSQL, Redis, MongoDB, NATS)
2. Runs database migrations
3. Launches the backend game server
4. Starts the frontend development server
5. Creates the lobby world seed data

Once running, access the game at: **http://localhost:5173**

---

## 🏗️ Backend Architecture

### Technology Stack

- **Language**: Go 1.24
- **Web Framework**: Chi router
- **Real-time**: WebSocket (gorilla/websocket), NATS messaging
- **Databases**: 
  - PostgreSQL with PostGIS (spatial data, event store)
  - Redis (caching, sessions, rate limiting)
  - MongoDB (NPC memory and dialogue)
- **AI**: Ollama (LLM for world generation and NPC dialogue)
- **Monitoring**: Prometheus metrics, structured logging (zerolog)

### Core Services

#### 1. **game-server** (`cmd/game-server/`)
Main entry point that handles:
- HTTP API endpoints (auth, characters, worlds)
- WebSocket connections for real-time gameplay
- Command processing (movement, combat, inventory)
- Lobby system (world selection, character management)

**Key handlers:**
- `api/auth_handler.go` - Registration, login, JWT validation
- `api/character_handler.go` - Character creation/selection
- `api/interview_handler.go` - World creation interview with LLM
- `api/world_handler.go` - World CRUD operations
- `websocket/hub.go` - WebSocket connection management
- `websocket/client.go` - Client message handling

#### 2. **auth-service** (`cmd/auth-service/`)
Dedicated authentication microservice:
- JWT token generation with AES-256 encryption
- Argon2id password hashing (64MB, 3 iterations)
- Redis session management (24h TTL)
- Rate limiting (per-endpoint and per-IP)

#### 3. **ai-gateway** (`cmd/ai-gateway/`)
LLM integration gateway:
- Prompts Ollama for world creation interviews
- Generates dynamic area descriptions
- Creates NPC dialogue based on personality/mood
- Implements aggressive caching (15-minute TTL)

#### 4. **world-service** (`cmd/world-service/`)
World ticker and time management:
- Configurable time dilation (0.1x to 100x speed)
- Day/night cycles and seasonal progression
- Broadcasts `world.tick` events via NATS
- World pause/resume with catch-up fast-forwarding

### Internal Packages

#### Event Sourcing (`internal/eventstore/`)
- Append-only event log for complete audit trail
- CQRS pattern with optimized read models
- Event replay engine (rewind/fast-forward)
- Versioning strategy for schema migrations

#### Spatial System (`internal/spatial/`)
- PostGIS integration for 3D coordinates (X, Y, Z)
- Multi-world support (spherical, bounded cube, infinite)
- GIST spatial indexes for fast radius queries (<30ms)
- Collision detection and boundary validation

#### NPC System (`internal/npc/`)
- **Genetics** (`genetics/`): Mendelian inheritance, mutations
- **Memory** (`memory/`): MongoDB-backed memory with decay/rehearsal
- **Behavior** (`behavior/`): Desire engine, personality-driven actions
- **Dialogue** (`dialogue/`): LLM-enhanced conversations
- **Relationships** (`relationships/`): Affection, trust, fear tracking

#### Character System (`internal/character/`)
- Dual creation paths (inhabit NPC vs generate new)
- 15 attributes (5 physical, 5 mental, 5 sensory)
- Point-buy system with scaling costs
- Genetic variance (±5 per attribute)

#### Combat System (`internal/combat/`)
- Real-time action queue based on reaction time
- Weapon types (slashing, piercing, bludgeoning)
- Status effects (poison, stun, slow, bleed)
- Skill-based damage calculation

#### World Generation (`internal/worldgen/`)
- Tectonic plate simulation
- Heightmap generation (mountains, valleys, oceans)
- Weather simulation (evaporation, wind patterns, precipitation)
- Flora/fauna evolution with environmental pressures

### Database Schema

**Event Store:**
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

**Spatial Entities:**
```sql
entities (
  id UUID PRIMARY KEY,
  world_id UUID,
  position GEOMETRY(POINTZ, 4326),  -- X, Y, Z coordinates
  entity_type VARCHAR
)
```

**Worlds:**
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

### Running Backend Locally

```bash
cd tw-backend

# Start dependencies
docker-compose up -d postgres redis mongodb nats

# Run migrations
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/thousand_worlds?sslmode=disable" up

# Set environment variables
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/thousand_worlds?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export MONGODB_URL="mongodb://localhost:27017"
export JWT_SECRET="your-secret-key"
export PORT="8080"

# Run the server
go run cmd/game-server/main.go
```

### Testing Backend

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test suite
go test ./internal/character/...

# Run E2E tests
go test -v ./tests/e2e/...
```

---

## 🎨 Frontend Architecture

### Technology Stack

- **Framework**: SvelteKit 2.0
- **Language**: TypeScript 5.0
- **Styling**: Tailwind CSS 3.4
- **Real-time**: WebSocket + NATS.ws
- **PWA**: vite-plugin-pwa 0.17
- **Testing**: Vitest 3.2, Playwright 1.57

### Application Structure

#### Routes (`src/routes/`)

- **`+page.svelte`**: Landing page with login/registration
- **`game/+page.svelte`**: Main game interface
  - Command input area
  - Output/message log
  - Character stats display
  - Inventory/equipment UI

#### Components (`src/lib/components/`)

Core UI components:
- `CharacterCreation.svelte` - Character creation wizard
- `WorldInterview.svelte` - World creation interview interface
- `GameOutput.svelte` - Formatted game messages
- `CommandInput.svelte` - Command parser and history
- `InventoryPanel.svelte` - Item management
- `StatsPanel.svelte` - Character attributes/stats
- `MapView.svelte` - 2D top-down spatial visualization

#### Services (`src/lib/services/`)

**`api.ts`** - HTTP API client:
```typescript
// Authentication
login(username: string, password: string)
register(username: string, email: string, password: string)

// Characters
getCharacters()
createCharacter(data: CharacterData)

// Worlds
getWorlds()
createWorld(config: WorldConfig)

// Interview
getInterviewSession()
sendInterviewResponse(message: string)
```

**`websocket.ts`** - WebSocket connection:
```typescript
connect(characterId: string)
sendCommand(command: string)
onMessage(callback: (message: GameMessage) => void)
disconnect()
```

**`game-state.ts`** - Centralized game state management

#### Stores (`src/lib/stores/`)

**`gameStore.ts`** - Svelte stores for reactive state:
```typescript
{
  user: User | null,
  currentCharacter: Character | null,
  messages: GameMessage[],
  inventory: Item[],
  stats: CharacterStats,
  nearbyEntities: Entity[]
}
```

### PWA Features

Configured in `vite.config.ts`:
- **Service Worker**: Caches assets and API responses
- **Offline Mode**: Read-only access when disconnected
- **Install Prompt**: "Add to Home Screen" on mobile
- **Push Notifications**: Combat alerts, server events (planned)

### Mobile Optimization

- **Responsive Design**: Tailwind breakpoints for all screen sizes
- **Touch-Friendly**: Large tap targets, swipe gestures
- **Performance**: Code splitting, lazy loading
- **PWA**: Installable on iOS/Android home screen

### Running Frontend Locally

```bash
cd tw-frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Access at http://localhost:5173
```

### Testing Frontend

```bash
# Unit tests with Vitest
npm run test

# Test with UI
npm run test:ui

# Coverage report
npm run test:coverage

# E2E tests with Playwright
npm run test:e2e

# E2E with UI
npm run test:e2e:ui
```

### Building for Production

```bash
# Build optimized bundle
npm run build

# Preview production build
npm run preview
```

---

## 🔄 Development Workflow

### Test-Driven Development (TDD)

This project follows strict TDD principles:

1. **Write tests first** that validate acceptance criteria
2. **Implement code** to pass tests
3. **Refactor** with confidence
4. **Target**: 80%+ code coverage for all services

### Event Sourcing Architecture

All state changes are stored as immutable events:

```go
// Event examples
CharacterCreated{characterID, attributes, timestamp}
PlayerMoved{characterID, fromPos, toPos, staminaCost}
ItemPickedUp{characterID, itemID, weight}
```

Benefits:
- Complete audit trail
- Time-travel debugging (rewind to any point)
- Event replay for testing
- Future analytics and ML training data

### Microservices Communication

Services communicate via:
- **Synchronous**: HTTP/REST for request-response
- **Asynchronous**: NATS for event broadcasting
- **Real-time**: WebSocket for player connections

Example NATS subjects:
- `world.tick.{worldID}` - World ticker events
- `player.moved.{characterID}` - Movement events
- `combat.action.{battleID}` - Combat actions

---

## 📊 Monitoring & Operations

### Health Checks

Each service exposes `/health` endpoint:
```bash
curl http://localhost:8080/health
```

Returns:
```json
{
  "status": "healthy",
  "uptime": "2h34m12s",
  "version": "0.1.0",
  "dependencies": {
    "database": "connected",
    "redis": "connected",
    "nats": "connected"
  }
}
```

### Metrics

Prometheus metrics available at `/metrics`:
- Request latency (p50, p90, p95, p99)
- Error rates by endpoint
- Cache hit rates (L1 memory, L2 Redis)
- Active WebSocket connections
- Event store append rate

### Logging

Structured JSON logs with correlation IDs:
```json
{
  "level": "info",
  "service": "game-server",
  "requestID": "abc123",
  "userID": "user-456",
  "msg": "Player moved",
  "from": [100.5, 200.3, 10.0],
  "to": [101.5, 200.3, 10.0]
}
```

---

## 🗺️ Roadmap

See [`roadmap.md`](roadmap.md) for the complete development roadmap with phases 0-12.

**Current Progress:**
- ✅ Phase 0: Foundation & Infrastructure (Completed)
- ✅ Phase 1: World Ticker & Time System (Completed)
- ✅ Phase 2: Player Core Systems (Completed)
- 🚧 Phase 3: NPC Memory & Relationships (In Progress)

**Next Steps:**
- Phase 4: NPC Genetics & Appearance
- Phase 5: NPC AI & Desire Engine
- Phase 6: LLM Integration for Dialogue

---

## 🤝 Contributing

This is a personal project, but contributions are welcome! Please:

1. Follow TDD principles (write tests first)
2. Maintain 80%+ test coverage
3. Use structured logging with context
4. Document all public APIs
5. Run linters and tests before committing

---

## 📄 License

[MIT License](LICENSE) (or specify your license)

---

## 📞 Support

- **Issues**: Use GitHub Issues for bug reports
- **Documentation**: See [`features.md`](features.md) for detailed feature list
- **Roadmap**: See [`roadmap.md`](roadmap.md) for development plan
- **Optimization**: See [`optimizeandfix.md`](optimizeandfix.md) for improvement recommendations
- **Backend**: See [`tw-backend/README.md`](tw-backend/README.md) for backend documentation
- **Frontend**: See [`tw-frontend/README.md`](tw-frontend/README.md) for frontend documentation
