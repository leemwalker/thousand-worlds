# Thousand Worlds Development Roadmap

> Development Philosophy: Test-Driven Development (TDD)
> - Write tests FIRST that validate acceptance criteria
> - Implement code to pass tests
> - Refactor with confidence
> - Target: 80%+ code coverage for all services

---

# Phase 0: Foundation & Infrastructure (4-6 weeks)

## Phase 0.1: Event Sourcing & Data Layer (2 weeks)

### Status:  ✅ Completed

### Prompt:

Following TDD principles, implement Phase 0.1 Event Sourcing & Data Layer:

Core Requirements:
1. Design event schema with command/event separation
Commands: requests that MAY fail (CreateWorld, MovePlayer, AttackNPC)
Events: facts that HAVE happened (WorldCreated, PlayerMoved, NPCAttacked)
Include: eventID (UUID), eventType, aggregateID, aggregateType, version, timestamp, payload (JSON), metadata

2. Create PostgreSQL events table (append-only, immutable)
Schema: id, event_type, aggregate_id, aggregate_type, version, timestamp, payload, metadata
Indexes: (aggregate_id, version), (event_type), (timestamp)
Constraint: UNIQUE(aggregate_id, version) for optimistic locking

3. Build EventStore repository with methods:
AppendEvent(event) → error
GetEventsByAggregate(aggregateID, fromVersion) → []Event
GetEventsByType(eventType, fromTimestamp, toTimestamp) → []Event
GetAllEvents(fromTimestamp, limit) → []Event

4. Implement event replay engine:
ReplayEvents(aggregateID, fromVersion, toVersion) → []Event
RewindToTimestamp(aggregateID, timestamp) → aggregateState
FastForwardFrom(aggregateID, startVersion, endVersion) → aggregateState

5. Create event versioning strategy:
Support event schema migrations (V1 → V2 transformations)
Upcaster pattern for backward compatibility
Test migration of 1000 V1 events to V2 format

6. Implement CQRS read models:
WorldReadModel (current world state from events)
EventProjection pattern (listen to events, update read model)
Eventual consistency tests

Test Requirements (85%+ coverage):
AppendEvent stores event and returns no error
AppendEvent enforces append-only (cannot modify existing events)
AppendEvent respects version ordering (version N+1 after version N)
GetEventsByAggregate returns events in version order
GetEventsByType filters correctly by type and timestamp
ReplayEvents reconstructs aggregate state from event log
RewindToTimestamp returns state as of specific point in time
Event versioning migrates V1 → V2 without data loss
CQRS read model stays consistent with event stream
Concurrent appends handled with optimistic locking
Store 10k events and replay in < 5 seconds

Acceptance Criteria:
Can store/retrieve events with immutability guarantees
Event store can replay 10,000 events in < 5 seconds
Event versioning supports schema migrations
CQRS read models update from event stream
All tests pass with 85%+ coverage

Dependencies: PostgreSQL (existing in docker-compose)
Files to Create:
`internal/eventstore/types.go` - Event, Command types
`internal/eventstore/store.go` - EventStore repository
`internal/eventstore/replay.go` - Replay engine
`internal/eventstore/versioning.go` - Event upcasters
`internal/eventstore/projections.go` - CQRS projections
`internal/eventstore/_test.go` - Comprehensive test suite


---

## Phase 0.2: Authentication & Security (1 week)

### Status: ✅ Completed

### Prompt:

Core Requirements:
1. JWT token generation/validation with AES-256 encryption
Generate token with claims: userID, username, roles, exp (24h), iat
Encrypt token payload with AES-256-GCM
Validate signature, expiration, and decrypt payload
Rotate signing keys (support old + new key during rotation)

2. Argon2id password hashing
Parameters: 64MB memory, 3 iterations, 4 parallelism
Salt: 16 bytes random per password
Output: 32-byte hash
Timing-safe comparison for hash validation

3. Redis session management
Store: sessionID → {userID, username, loginTime, lastAccess}
TTL: 24 hours, extend on each request
Invalidate session on logout
Cleanup expired sessions

4. Rate limiting middleware
Per-endpoint limits: login (5/min), register (3/5min), world-create (1/hour)
Per-IP and per-user tracking
Use Redis for distributed rate limiting
Return 429 Too Many Requests with Retry-After header

5. Security tests:
Brute force protection (lock account after 10 failed attempts)
Token tampering detection (modified JWT rejected)
Session hijacking prevention (validate IP/User-Agent)
SQL injection resistance (parameterized queries only)

Test Requirements (80%+ coverage):
JWT generation creates valid, encrypted token
JWT validation rejects expired/tampered tokens
Argon2id hashing is deterministic (same password → same hash with same salt)
Password comparison is timing-safe (no timing attacks)
Sessions stored in Redis with correct TTL
Session extends TTL on access
Logout invalidates session
Rate limiter blocks after threshold exceeded
Rate limiter resets after time window
Brute force protection locks account
Token tampering detected and rejected

Acceptance Criteria:
JWT tokens generated and validated correctly
Passwords hashed with Argon2id (64MB, 3 iter, 4 parallel)
Sessions managed in Redis with 24h TTL
Rate limiting enforced per endpoint
Security tests pass (brute force, token tampering, session hijacking)
80%+ test coverage

Dependencies: Redis (existing in docker-compose)
Files to Create:
`internal/auth/jwt.go` - JWT generation/validation
`internal/auth/password.go` - Argon2id hashing
`internal/auth/session.go` - Redis session manager
`internal/auth/ratelimit.go` - Rate limiting middleware
`internal/auth/_test.go` - Security test suite
Refactor `cmd/auth-service/auth_handler.go` to use new auth package


---

## Phase 0.3: Monitoring & Observability (1 week)

### Status: ✅ Completed

### Prompt:

Following TDD principles, implement Phase 0.3 Monitoring & Observability:

Core Requirements:
1. Prometheus metrics collection
HTTP request latency (histogram: p50, p90, p95, p99)
Error rates (counter: by service, endpoint, error type)
Cache hit rates (gauge: L1 memory, L2 Redis)
NPC simulation FPS (gauge: ticks per second per world)
Event store append rate (counter: events/sec)
Active connections (gauge: WebSocket, database)

2. Grafana dashboard definitions (JSON)
System health overview (CPU, memory, goroutines)
Request performance (latency percentiles, throughput)
Error tracking (error rate, top errors)
Cache efficiency (hit rate, miss rate, evictions)

3. Health check endpoints
`/health` returns 200 OK if healthy, 503 if degraded
Check: database connectivity, Redis connectivity, NATS connectivity
Include: uptime, version, last event processed, memory usage
Kubernetes liveness/readiness probes

4. Structured logging with correlation IDs
Use zerolog (already imported)
Add requestID to all log entries in request chain
Log levels: Debug, Info, Warn, Error
Context: service, endpoint, userID, worldID, eventID

5. Performance benchmarks
Spatial queries: < 30ms for 1000 entities within 100m radius
Event replay: < 5s for 10k events
JWT validation: < 1ms per token
Cache lookup: < 100μs for L1, < 5ms for L2

Test Requirements (80%+ coverage):
Metrics registered and incremented correctly
Histogram records latency in correct buckets
Health check returns 200 when all dependencies healthy
Health check returns 503 when database unreachable
Correlation ID propagates through service calls
Benchmarks meet performance targets
Prometheus scrape endpoint returns valid metrics

Acceptance Criteria:
Prometheus metrics collected for all services
Grafana dashboards created for system health
All services expose `/health` endpoints
Structured logging with correlation IDs
Performance benchmarks pass (spatial < 30ms, replay < 5s)
80%+ test coverage

Dependencies: None (add Prometheus client library)
Files to Create:
`internal/metrics/prometheus.go` - Metric definitions
`internal/health/checker.go` - Health check logic
`internal/logging/logger.go` - Structured logger wrapper
`deploy/grafana/dashboards/.json` - Dashboard configs
`internal/metrics/_test.go` - Metrics tests
`_bench_test.go` - Performance benchmarks


---

## Phase 0.4: Spatial Foundation (1-2 weeks)

### Status: 🔴 Started

### Prompt:

Following TDD principles, implement Phase 0.4 Spatial Foundation:

Core Requirements:
1. PostgreSQL schema with PostGIS extensions
Enable PostGIS: CREATE EXTENSION IF NOT EXISTS postgis;
Table: entities (id, world_id, position GEOGRAPHY(POINTZ, 4326), ...)
Position: NUMERIC(10, 1) for X, Y, Z (decimeter precision)
Indexes: GIST(position), (world_id, position)

2. Coordinate system
X, Y, Z as NUMERIC(10, 1) - supports ±99999999.9 meters with 0.1m precision
Z represents elevation (positive = up, negative = down)
Store as PostGIS POINTZ in WGS84 (SRID 4326)

3. SpatialRepository methods:
CreateEntity(worldID, entityID, x, y, z) → error
UpdateEntityLocation(entityID, x, y, z) → error
GetEntity(entityID) → Entity
GetEntitiesNearby(worldID, x, y, z, radiusMeters) → []Entity
GetEntitiesInBounds(worldID, minX, minY, maxX, maxY) → []Entity
CalculateDistance(entity1, entity2) → float64

4. Coordinate-based movement in 10 directions
N (0, +1, 0), S (0, -1, 0), E (+1, 0, 0), W (-1, 0, 0)
NE (+1, +1, 0), SE (+1, -1, 0), SW (-1, -1, 0), NW (-1, +1, 0)
UP (0, 0, +1), DOWN (0, 0, -1)
Movement distance: 1.0 meter per step (configurable)

5. Collision detection and boundary validation
Check if target position is occupied
Check if target position is within world bounds
Check if target position is valid terrain (not inside solid object)
Return error if movement blocked

Test Requirements (80%+ coverage):
CreateEntity stores position with decimeter precision
UpdateEntityLocation moves entity to new position
GetEntitiesNearby returns only entities within radius
GetEntitiesNearby uses spatial index (query plan analysis)
GetEntitiesInBounds returns entities in bounding box
CalculateDistance returns correct 3D distance
Movement in all 10 directions works correctly
Collision detection prevents moving into occupied space
Boundary validation prevents leaving world bounds
Spatial query with 1000 entities completes in < 30ms

Acceptance Criteria:
Can store/query spatial data with decimeter (0.1m) precision
Spatial queries use GIST indexes efficiently
Radius queries, bounding box searches, distance calculations work
Basic coordinate-based movement in 10 directions
Collision detection and boundary validation
Spatial queries < 30ms with 1000 entities
80%+ test coverage

Dependencies: PostgreSQL with PostGIS extension
Files to Modify/Create:
`internal/repository/spatial_repository.go` - Expand existing stub
`internal/spatial/movement.go` - Movement logic
`internal/spatial/collision.go` - Collision detection
`internal/repository/spatial_repository_test.go` - Test suite
`deploy/postgres/init.sql` - PostGIS setup script


---

# Phase 1: World Ticker & Time System (2-3 weeks)

## Phase 1.1: World Service Core (1 week)

### Status: 🟡 In Progress (implementation plan approved)

### Prompt:

Following TDD principles, implement Phase 1.1 World Service Core:

Core Requirements:
1. World-service microservice
Main entrypoint: `cmd/world-service/main.go`
Connect to NATS and PostgreSQL
Initialize ticker manager
Graceful shutdown on SIGTERM/SIGINT

2. World registry (in-memory for now, event-sourced later)
Track active worlds: map[worldID]WorldState
WorldState: {ID, Name, Status (running/paused), TickCount, GameTime, DilationFactor, CreatedAt}
Thread-safe access (use sync.RWMutex)

3. Ticker manager
SpawnTicker(worldID, dilationFactor) → error
StopTicker(worldID) → error
GetTickerStatus(worldID) → (running bool, tickCount int64, gameTime time.Duration)
Manage goroutines per world (one ticker per world)

4. World state persistence (event-sourced)
Emit WorldCreated event on world creation
Emit WorldTicked event on each tick
Emit WorldPaused/WorldResumed events
Reconstruct world state from event log on restart

Test Requirements (80%+ coverage):
World registry stores and retrieves world state
World registry is thread-safe (concurrent access test)
SpawnTicker creates new ticker for world
SpawnTicker returns error if ticker already exists
StopTicker stops running ticker
StopTicker returns error if ticker not running
GetTickerStatus returns current tick count and game time
WorldCreated event emitted on world creation
WorldTicked event emitted on each tick
World state reconstructed from events on restart

Acceptance Criteria:
World registry tracks active/paused worlds
Ticker manager spawns/stops tickers per world
World state persisted using event sourcing
Tests for ticker lifecycle management
80%+ test coverage

Dependencies: Phase 0.1 (Event Sourcing), NATS, PostgreSQL
Files to Create:
`cmd/world-service/main.go`
`internal/world/registry.go`
`internal/world/ticker_manager.go`
`internal/world/types.go`
`internal/world/_test.go`


---

## Phase 1.2: Time Dilation & Tick Broadcast (1 week)

### Status: 🔴 Not Started

### Prompt:

Following TDD principles, implement Phase 1.2 Time Dilation & Tick Broadcast:

Core Requirements:
1. Configurable dilation factor per world
Range: 0.1 to 100.0
Default: 1.0 (real-time)
Examples: 10.0 = 10x faster, 0.5 = half speed

2. Tick loop implementation
Tick rate: 10 Hz (100ms real-time intervals)
Calculate: GameTime += RealDelta  DilationFactor
Example: 100ms real @ 10x dilation = 1000ms game time

3. Broadcast world.tick events to NATS
Subject: `world.tick.{worldID}`
Payload: {worldID, tickNumber, gameTime (ms), realTime (ms), dilationFactor}
Frequency: every tick (10 Hz default)

4. World pause/resume with catch-up
Pause: stop ticker, record pause time
Resume: calculate elapsed pause duration
Fast-forward: emit catch-up ticks at max speed (100x dilation)
Example: Paused for 10min @ 10x dilation = fast-forward 100min game time

5. Multi-world scenarios
Each world runs independently
Different dilation factors per world
No interference between world tickers
Graceful shutdown stops all tickers

Test Requirements (80%+ coverage):
Ticker emits ticks at 10 Hz
GameTime advances by RealDelta  DilationFactor
Dilation factor 10.0 makes game time advance 10x faster
Dilation factor 0.5 makes game time advance 0.5x slower
world.tick events published to correct NATS subject
Tick payload contains all required fields
Pause stops ticker and records pause time
Resume emits catch-up ticks at fast-forward speed
Multiple worlds run simultaneously with different dilation
Stopping one world doesn't affect others

Acceptance Criteria:
Multiple worlds run with independent tick rates
Time dilation accurately scales game time vs real time
Paused worlds fast-forward to catch up when resumed
world.tick events broadcast to NATS
80%+ test coverage

Dependencies: Phase 1.1 (World Service Core)
Files to Modify/Create:
`internal/world/ticker.go` - Ticker implementation
`internal/world/time_dilation.go` - Dilation logic
`internal/world/pause_resume.go` - Pause/resume with catch-up
`internal/world/ticker_test.go` - Comprehensive tests


---

## Phase 1.3: Day/Night & Seasons (1 week)

### Status: 🔴 Not Started

### Prompt:

Following TDD principles, implement Phase 1.3 Day/Night & Seasons:

Core Requirements:
1. Sun position calculation
24-hour cycle (configurable day length per world)
sunPosition: 0.0 (midnight) → 0.5 (noon) → 1.0 (midnight)
Calculate from gameTime: (gameTime % dayLength) / dayLength
Default day length: 24 game-hours = 86400 game-seconds

2. Seasonal changes
4 seasons: Spring, Summer, Autumn, Winter
Configurable cycle length (default: 90 game-days per season)
seasonProgress: 0.0 (start of season) → 1.0 (end of season)
Calculate from gameTime: ((gameTime % yearLength) / seasonLength) % 1.0

3. Time-of-day descriptors
Enum: Night (0.0-0.25), Dawn (0.25-0.3), Morning (0.3-0.45), Noon (0.45-0.55), Afternoon (0.55-0.7), Dusk (0.7-0.75), Evening (0.75-0.9), Night (0.9-1.0)
Map sunPosition to descriptor
User-friendly labels in tick events

4. Enhanced world.tick event payload
Add: timeOfDay (string), sunPosition (float64), currentSeason (string), seasonProgress (float64)
Existing: worldID, tickNumber, gameTime, realTime, dilationFactor

5. Time progression tests
Simulate 1 game-year (365 days) at high dilation (1000x)
Verify seasons cycle 4 times
Verify day/night cycles 365 times
Complete in < 5 seconds real time

Test Requirements (80%+ coverage):
Sun position cycles from 0.0 → 1.0 over 24 game-hours
Time-of-day descriptor changes correctly (Night → Dawn → Morning → etc.)
Seasons cycle Spring → Summer → Autumn → Winter → Spring
Season progress advances from 0.0 to 1.0 over 90 game-days
Tick event includes time-of-day and season data
1 game-year simulation completes in < 5s
Dilation factor affects day/night and seasons correctly
Configurable day length and season length work

Acceptance Criteria:
Multiple worlds can run with independent tick rates
Time dilation accurately scales game time vs real time
Paused worlds can fast-forward to catch up when resumed
Day/night and seasonal state correctly calculated
80%+ test coverage

Dependencies: Phase 1.2 (Time Dilation)
Files to Create:
`internal/world/time_of_day.go` - Sun position, descriptors
`internal/world/seasons.go` - Season calculation
`internal/world/time_test.go` - Time progression tests
Modify `internal/world/ticker.go` to include time data in events


---

# Phase 2: Player Core Systems (3-4 weeks)

## Phase 2.1: Character System (1 week)

### Status: 🔴 Not Started

### Prompt:

Following TDD principles, implement Phase 2.1 Character System:

Core Requirements:
1. Attribute schema
Core (7): Strength, Agility, Health, Wits, Wisdom, Charisma, Willpower (1-100 scale)
Secondary (4): MaxHP (Health10), MaxStamina (Health5 + Agility5), MaxFocus (Wits10), MaxMana (Wisdom10)
Sensory (5): Sight, Hearing, Smell, Taste, Touch (1-100 scale)

2. Character creation flow
Player provides: name, description
Attribute allocation: point-buy (300 points total for core)
OR dice rolling: roll 3d6 seven times, assign to attributes
Sensory attributes: roll 2d10 each (2-20 range, can be modified by race/traits later)

3. Validation
Core attributes: 1-100 each
Point-buy: exactly 300 points spent
Name: 3-20 characters, alphanumeric + spaces
Description: max 500 characters

4. Event-sourced persistence
CharacterCreated event: {characterID, playerID, name, attributes, timestamp}
AttributeModified event: {characterID, attribute, oldValue, newValue, reason, timestamp}
Reconstruct character state from events

5. Read model (CQRS)
CharacterReadModel: {id, playerID, name, attributes, createdAt, updatedAt}
Update from CharacterCreated/AttributeModified events
Query: GetCharacter(characterID), GetCharactersByPlayer(playerID)

Test Requirements (80%+ coverage):
Point-buy allocates exactly 300 points
Point-buy rejects if total != 300
Dice rolling generates valid attribute values (3-18 range for 3d6)
Secondary attributes calculated correctly from core attributes
Validation rejects invalid names (too short, too long, invalid chars)
CharacterCreated event emitted and stored
Character state reconstructed from events
Read model updated from events
GetCharacter returns correct character
Concurrent character creation handled correctly

Acceptance Criteria:
Players can create characters with valid attribute distributions
Attribute allocation (point-buy or rolling) works
Character persistence event-sourced
Read model for character queries
All validation rules enforced
80%+ test coverage

Dependencies: Phase 0.1 (Event Sourcing)
Files to Create:
`internal/player/character.go` - Character types
`internal/player/creation.go` - Creation flow
`internal/player/attributes.go` - Attribute logic
`internal/player/events.go` - Character events
`internal/player/character_test.go` - Test suite
`cmd/player-service/main.go` - New player service


---

## Phase 2.2: Stamina & Movement (1 week)

### Status: 🔴 Not Started

### Prompt:

Following TDD principles, implement Phase 2.2 Stamina & Movement:

Core Requirements:
1. Stamina pool
MaxStamina = (Health  5) + (Agility  5)
CurrentStamina: 0 to MaxStamina
Start at MaxStamina on character creation

2. Movement costs
Walk: 1 stamina per meter
Run: 2 stamina per meter (2x speed)
Sneak: 1.5 stamina per meter (0.5x speed, stealth bonus)
Sprint: 4 stamina per meter (3x speed, loud)

3. Stamina regeneration
BasRegenRate = (Health / 10) stamina per second
Regen only when not moving
Regen stops when taking damage

4. Movement prevention
Cannot move if CurrentStamina < movementCost
Return error: "Insufficient stamina"
UI should show stamina bar

5. Event-sourced movement
PlayerMoved event: {characterID, fromX, fromY, fromZ, toX, toY, toZ, movementType, staminaCost, timestamp}
StaminaChanged event: {characterID, oldValue, newValue, reason, timestamp}
Reconstruct stamina state from events

Test Requirements (80%+ coverage):
MaxStamina calculated correctly from attributes
Walk movement costs 1 stamina per meter
Run movement costs 2 stamina per meter
Sneak movement costs 1.5 stamina per meter
Movement rejected if insufficient stamina
Stamina regenerates at correct rate
Regeneration stops when moving
Regeneration stops when taking damage
PlayerMoved event emitted with correct data
StaminaChanged event emitted on regen/drain
Stamina state reconstructed from events

Acceptance Criteria:
Movement consumes stamina correctly
Stamina regenerates when resting
Cannot move with insufficient stamina
Movement types (walk/run/sneak) work with different costs
All stamina changes event-sourced
80%+ test coverage

Dependencies: Phase 0.4 (Spatial Foundation), Phase 2.1 (Character System)
Files to Create:
`internal/player/stamina.go` - Stamina logic
`internal/player/movement.go` - Movement with stamina
`internal/player/regeneration.go` - Stamina regen
`internal/player/stamina_test.go` - Test suite


---

## Phase 2.3: Inventory System (1-2 weeks)

### Status: 🔴 Not Started

### Prompt:

Following TDD principles, implement Phase 2.3 Inventory System:

Core Requirements:
1. Item schema
Item: {id, name, description, weight (kg), stackSize, durability (0-100), properties (JSON)}
Properties: {isEquippable, slot, damageType, armorValue, effects, etc.}

2. Weight-limited inventory
MaxCarryWeight = Strength  5 kg
CurrentCarryWeight = sum of all item weights
Cannot pick up if CurrentCarryWeight + itemWeight > MaxCarryWeight

3. Equipment slots
Slots: mainHand, offHand, head, chest, legs, feet, neck, ring1, ring2
Equip item: move from inventory to slot
Unequip item: move from slot to inventory
Slot restrictions: only appropriate items (e.g., weapon in mainHand)

4. Item durability
Durability: 0 (broken) to 100 (pristine)
Degrades with use (weapon: -1 per attack, armor: -1 per hit taken)
Broken items have reduced effectiveness (weapon: 50% damage, armor: 0 protection)
Repair: restore durability (requires resources/smith NPC)

5. Item commands
Pickup(itemID): add item to inventory
Drop(itemID): remove item from inventory, place in world
Use(itemID): consume item (food, potion, etc.)
Examine(itemID): get detailed description
Equip(itemID, slot): move to equipment slot
Unequip(slot): move from slot to inventory

6. Event-sourced inventory
ItemPickedUp: {characterID, itemID, weight, timestamp}
ItemDropped: {characterID, itemID, x, y, z, timestamp}
ItemUsed: {characterID, itemID, effect, timestamp}
ItemEquipped: {characterID, itemID, slot, timestamp}
ItemDurabilityChanged: {itemID, oldValue, newValue, reason, timestamp}

Test Requirements (80%+ coverage):
MaxCarryWeight calculated from Strength
Pickup succeeds when under weight limit
Pickup fails when over weight limit
Drop removes item from inventory and places in world
Use item consumes it and applies effect
Equip places item in correct slot
Equip fails if slot already occupied
Unequip moves item back to inventory
Durability decreases on use
Broken items have reduced effectiveness
Stacking items works correctly
Events emitted for all inventory actions
Inventory state reconstructed from events

Acceptance Criteria:
Inventory respects weight limits
Item stacking, equipment slots, durability work
Pickup/drop/use/examine/equip/unequip commands
All inventory changes event-sourced
Edge cases tested (overfilled, negative weight, stack overflow)
80%+ test coverage

Dependencies: Phase 2.1 (Character System)
Files to Create:
`internal/item/types.go` - Item schema
`internal/item/inventory.go` - Inventory logic
`internal/item/equipment.go` - Equipment slots
`internal/item/durability.go` - Durability system
`internal/item/commands.go` - Item commands
`internal/item/events.go` - Inventory events
`internal/item/_test.go` - Test suite


---

# Testing Reminders

## For Every Phase Section:

1. Write tests FIRST that validate acceptance criteria
2. Run tests frequently during implementation (`go test -v ./...`)
3. Measure coverage after completion (`go test -coverprofile=coverage.out ./...`)
4. Target 80%+ coverage for all new code
5. Write edge case tests (nil inputs, boundary values, concurrency)
6. Integration tests where services interact
7. Benchmark tests for performance-critical code

## Test Naming Conventions:

Unit tests: `TestFunctionName_Scenario_ExpectedBehavior`
Integration tests: `TestIntegration_FeatureName_Scenario`
Benchmarks: `BenchmarkFunctionName`

Example:
go
func TestStaminaRegen_WhenResting_RegensAtCorrectRate(t testing.T) { ... }
func TestInventory_PickupOverWeight_ReturnsError(t testing.T) { ... }
func BenchmarkSpatialQuery_1000Entities(b testing.B) { ... }


---

# Progress Tracking

Update this section as phases are completed:

Phase 0.1: Event Sourcing & Data Layer
Phase 0.2: Authentication & Security
Phase 0.3: Monitoring & Observability
Phase 0.4: Spatial Foundation
Phase 1.1: World Service Core
Phase 1.2: Time Dilation & Tick Broadcast
Phase 1.3: Day/Night & Seasons
Phase 2.1: Character System
Phase 2.2: Stamina & Movement
Phase 2.3: Inventory System

(Phases 3-12 will be added as we progress)
