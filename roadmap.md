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

### Status: ✅ Completed

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

### Status: ✅ Completed

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

### Status: ✅ Completed

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

### Status: ✅ Completed

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

## Phase 1.4: Spherical World Utilities (Optional - As Needed)

### Status: ✅ Completed

### Prompt:

Following TDD principles, implement Phase 1.4 Spherical World Utilities (only when needed):

Core Requirements:
1. Spherical projection utilities
Lat/Lon to Cartesian (X, Y, Z) conversion
Formula: X = R × cos(lat) × cos(lon), Y = R × cos(lat) × sin(lon), Z = R × sin(lat)
Cartesian to Lat/Lon conversion
Formula: lat = arcsin(Z/R), lon = arctan2(Y, X)
Support multiple world radii (Earth = 6.371M meters, custom planets)

2. Great circle distance calculations
Calculate shortest distance between two points on sphere
Haversine formula: a = sin²(Δlat/2) + cos(lat1) × cos(lat2) × sin²(Δlon/2)
Distance = 2 × R × arcsin(√a)
Use for proximity queries on spherical worlds

3. Spherical wrapping logic
Pole crossing detection and seamless transitions
North pole: lat approaches +90°, longitude wraps
South pole: lat approaches -90°, longitude wraps
Longitude wrapping at ±180° (crossing international date line)
Maintain correct coordinate space when crossing boundaries

4. Movement validation for spherical worlds
Check if movement crosses pole
Adjust longitude correctly when crossing poles
Wrap longitude when crossing ±180°
Ensure position stays on sphere surface (constant radius)

Test Requirements (80%+ coverage):
Lat/Lon to Cartesian conversion is accurate to 1cm
Cartesian to Lat/Lon conversion is accurate to 0.001 degrees
Round-trip conversion (Lat/Lon → Cartesian → Lat/Lon) preserves values
Great circle distance matches expected values for known points
Pole crossing correctly adjusts longitude
Longitude wraps correctly at ±180°
Movement stays on sphere surface (radius constant)
Works for multiple sphere sizes (Earth-sized, moon-sized, custom)

Acceptance Criteria:
Spherical projections work correctly for any radius
Great circle distances accurate for proximity queries
Pole crossing and longitude wrapping handle edge cases
Movement validation ensures positions stay valid
80%+ test coverage

Dependencies: Phase 0.4 (Spatial Foundation), Phase 1.3 (Time System)
Files to Create:
`internal/spatial/spherical_projection.go` - Projection utilities
`internal/spatial/great_circle.go` - Distance calculations
`internal/spatial/spherical_wrapping.go` - Pole and longitude wrapping
`internal/spatial/spherical_test.go` - Comprehensive tests


---

# Phase 2: Player Core Systems (4-5 weeks)

## Phase 2.1: Character System (1-2 weeks)

### Status: ✅ Completed

### Prompt:

Following TDD principles, implement Phase 2.1 Character System:

Core Requirements:
1. Attribute schema (updated system)
**Physical Attributes (5)**: Might, Agility, Endurance, Reflexes, Vitality (1-100 scale)
**Mental Attributes (5)**: Intellect, Cunning, Willpower, Presence, Intuition (1-100 scale)
**Sensory Attributes (5)**: Sight, Hearing, Smell, Taste, Touch (1-100 scale)
**Secondary Attributes (5 - calculated)**:
- HP = Vitality × 10
- Stamina = (Endurance × 7) + (Might × 3)
- Focus = (Intellect × 6) + (Willpower × 4)
- Mana = (Intuition × 6) + (Willpower × 4)
- Nerve = (Willpower × 5) + (Presence × 3) + (Reflexes × 2)

2. Dual character creation paths
**Path 1: Inhabit Existing NPC**
- Browse eligible NPCs (adults with 5+ relationships, 1+ game-year lived)
- Filter by species, location, skills, behavioral baseline
- Select NPC and take over their identity
- Inherit full history: relationships, memories, reputation, skills
- Snapshot behavioral baseline at time of inhabitation

**Path 2: Generate New Adult Character**
- Species selection (Human, Dwarven, Elven, etc.)
- Genetic baseline generation with variance (±(1d10-5) per attribute)
- Point-buy customization (100 points total)
  - Cost scaling: 1 point for +1 to +10, 2 points for +11 to +20, 3 points for +21 to +30
  - Cannot reduce attributes below species baseline
- Background questionnaire (3-5 questions for starting skill bonuses)
- Generate adult with randomized history

3. Species base attribute templates
Define templates: Human (balanced), Dwarven (Might+10, Endurance+15, Agility-10), Elven (Agility+15, Intuition+10, Might-10), etc.
Each species has different baseline distributions
Sensory attributes vary by species (Elven Sight+20, Dwarven Hearing+15, etc.)

4. Genetic variance system
Apply random modifier: ±(1d10-5) to each attribute
Range: -5 to +5 variance per attribute
Ensures no two generated characters are identical
Applied BEFORE point-buy

5. Point-buy validation
Total points: 100
Cost per point: tier 1 (0-10 above baseline) = 1 point, tier 2 (11-20) = 2 points, tier 3 (21-30) = 3 points
Cannot buy above baseline + 30
Cannot reduce below baseline (after variance)

6. NPC browser UI
Filter: species, location, age range, skill levels
Sort: by relationship count, game years lived, skills
Display: name, age, species, location, primary relationships, behavioral baseline preview
Preview mode: view full character sheet before committing

7. Behavioral baseline snapshot
On inhabitation, record: aggression, generosity, honesty, sociability, recklessness, loyalty (0.0-1.0 each)
Baseline calculated from last 20 NPC actions
Stored immutably for drift detection later

8. Event-sourced persistence
CharacterCreatedViaInhabitance: {characterID, playerID, npcID, baselineSnapshot, timestamp}
CharacterCreatedViaGeneration: {characterID, playerID, species, attributes, variance, pointBuyChoices, timestamp}
AttributeModified: {characterID, attribute, oldValue, newValue, reason, timestamp}
Reconstruct character state from events

Test Requirements (80%+ coverage):
Species templates define valid baselines
Genetic variance applies ±5 correctly
Point-buy enforces cost scaling (1/2/3 points)
Point-buy prevents exceeding baseline + 30
Point-buy enforces 100-point budget
Secondary attributes calculated correctly from primaries
Inhabitation preserves NPC history and relationships
Inhabitation snapshots behavioral baseline
NPC browser filters and sorts correctly
CharacterCreated events emitted for both paths
Character state reconstructed from events

Acceptance Criteria:
Players can create characters via both inhabit and generate paths
Point-buy system enforces costs and caps correctly
NPC inhabitation preserves full history and relationships
Species provide different attribute distributions
Genetic variance ensures no identical generated characters
All systems emit events for audit trail
80%+ test coverage

Dependencies: Phase 0.1 (Event Sourcing), Phase 3.1 (NPC Memory for inhabitation)
Files to Create/Modify:
`internal/character/types.go` - Attribute types, species templates
`internal/character/creation.go` - Dual creation paths
`internal/character/point_buy.go` - Point-buy system
`internal/character/genetic_variance.go` - Variance system
`internal/character/npc_browser.go` - NPC selection
`internal/character/baseline_snapshot.go` - Behavioral baseline
`internal/character/repository.go` - Character CRUD
`internal/character/creation_test.go` - Test suite
`cmd/player-service/main.go` - New player service


---

## Phase 2.2: Stamina & Movement (1 week)

### Status: ✅ Completed

### Prompt:

Following TDD principles, implement Phase 2.2 Stamina & Movement:

Core Requirements:
1. Stamina pool
MaxStamina = ((Endurance×7)+(Might×3))
CurrentStamina: 0 to MaxStamina
Start at MaxStamina on character creation

2. Movement costs
Walk: 1 stamina per meter
Run: 2 stamina per meter (2x speed)
Sneak: 1.5 stamina per meter (0.5x speed, stealth bonus)
Sprint: 4 stamina per meter (3x speed, loud)

3. Stamina regeneration
BaseRegenRate = (Endurance / 10) stamina per second
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

## Phase 2.4: Skills & Progression System (2-3 weeks)

### Status: 🔴 Not Started

### Prompt:

Following TDD principles, implement Phase 2.5 Skills & Progression System:

Core Requirements:
1. Core skills framework
Design skill schema: {name, category, currentValue (0-100), experiencePoints}
5 skill categories:
- **Combat Skills**: Slashing, Piercing, Bludgeoning, Defense, Dodge
- **Crafting Skills**: Smithing, Alchemy, Carpentry, Tailoring, Cooking
- **Gathering Skills**: Mining, Herbalism, Logging, Hunting, Fishing
-  **Utility Skills**: Perception, Stealth, Climbing, Swimming, Navigation
- **Social Skills**: Persuasion, Intimidation, Deception, Bartering

2. Use-based experience gain
XP awarded per skill use (varies by action difficulty)
XP curve formula: `XP_needed = baseXP × (skillLevel^1.5)` (exponential growth)
Diminishing returns for repetitive actions (prevents grinding)
Example: Mining same rock 10 times = full XP first time, 10% XP by 10th time

3. Skill check system
Roll: `d100 + skill + (relevantAttribute/5)`
Compare vs difficulty threshold (Easy=30, Medium=50, Hard=70, VeryHard=90)
Critical success: natural 96-100 (automatic success + bonus effect)
Critical failure: natural 1-5 (automatic failure + penalty)

4. Skill synergies
Related skills provide minor bonuses (+5 to rolls)
Example: Smithing+5 when Metalworking > 50
Define synergy pairs in configuration

5. Soft caps (attribute-based thresholds)
Skill advancement harder after reaching (relevantAttribute × 1.5)
Example: Slashing soft cap at Might × 1.5 (Might=50 → soft cap at 75)
XP required doubles after soft cap

6. Skill requirements for content
Recipes require minimum skill levels (Cooking 40 for advanced meals)
Equipment has skill requirements (Legendary Sword requires Slashing 80)
Abilities unlock at skill milestones (Dodge 50 unlocks Counter-Attack)

7. Quality scaling with skill level
Crafting: skill determines quality tier (Poor < Common < Fine < Masterwork < Legendary)
Quality formula: `quality = floor(skill / 20)` (0-4 scale)
Combat: higher skill increases damage/accuracy
Gathering: higher skill increases yield and rare resource chance

8. Event-sourced persistence
SkillIncreased: {characterID, skillName, oldValue, newValue, xpGained, timestamp}
SkillUsed: {characterID, skillName, context, xpGained, diminishingReturnFactor, timestamp}
Reconstruct skill state from events

Test Requirements (80%+ coverage):
Skill categories properly defined
XP curve calculates correctly (exponential growth)
Diminishing returns reduces XP for repetitive actions
Skill checks combine skill + attribute + d100 correctly
Critical success/failure triggers on natural 1-5 and 96-100
Soft caps apply after attribute-based threshold
Skill synergies grant correct bonuses
Quality tiers scale with skill level
SkillIncreased/SkillUsed events emitted
Skill state reconstructed from events

Acceptance Criteria:
Skills increase through use with diminishing returns
Skill checks properly combine skill + attribute + random roll
All major systems (combat, crafting, gathering) respect skill levels
Skill progression feels rewarding without being grindable
Skills persist correctly via event sourcing
Quality scaling provides meaningful rewards for skill advancement
80%+ test coverage

Dependencies: Phase 2.1 (Character System)
Files to Create:
`internal/skills/types.go` - Skill schema, categories
`internal/skills/progression.go` - XP gain, leveling
`internal/skills/checks.go` - Skill check system
`internal/skills/synergies.go` - Skill synergy logic
`internal/skills/soft_caps.go` - Soft cap system
`internal/skills/quality.go` - Quality scaling
`internal/skills/repository.go` - Skill persistence
`internal/skills/diminishing_returns.go` - Anti-grinding
`internal/skills/progression_test.go` - Test suite


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
