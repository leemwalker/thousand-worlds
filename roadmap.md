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

### Status: ✅ Completed

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

### Status: ✅ Completed

### Prompt:

Following TDD principles, implement Phase 2.4 Skills & Progression System:

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

# Phase 3: NPC Memory & Relationships (5-7 weeks)

## Phase 3.1: Memory Storage & Retrieval (2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 3.1 NPC Memory Storage & Retrieval:

Core Requirements:
1. Design MongoDB memory schema with document types:
   - **Observation**: NPCs witnessing events
     ```
     {
       npcID: UUID,
       memoryType: "observation",
       timestamp: ISODate,
       clarity: 0.0-1.0,
       emotionalWeight: 0.0-1.0,
       accessCount: int,
       lastAccessed: ISODate,
       content: {
         event: string,
         location: {x, y, z, worldID},
         entitiesPresent: [UUID],
         weatherConditions: string,
         timeOfDay: string
       },
       tags: [string],
       relatedMemories: [memoryID]
     }
     ```
   - **Conversation**: Dialogue with players/NPCs
     ```
     {
       npcID: UUID,
       memoryType: "conversation",
       timestamp: ISODate,
       clarity: 0.0-1.0,
       emotionalWeight: 0.0-1.0,
       accessCount: int,
       lastAccessed: ISODate,
       content: {
         participants: [UUID],
         dialogue: [{speaker: UUID, text: string, emotion: string}],
         location: {x, y, z, worldID},
         outcome: string,
         relationshipImpact: {entityID: UUID, affinityDelta: int}
       },
       tags: [string],
       relatedMemories: [memoryID]
     }
     ```
   - **Event**: Significant personal experiences
     ```
     {
       npcID: UUID,
       memoryType: "event",
       timestamp: ISODate,
       clarity: 0.0-1.0,
       emotionalWeight: 0.0-1.0,
       accessCount: int,
       lastAccessed: ISODate,
       content: {
         eventType: string,
         description: string,
         location: {x, y, z, worldID},
         participants: [UUID],
         consequences: string,
         emotionalResponse: string
       },
       tags: [string],
       relatedMemories: [memoryID]
     }
     ```
   - **Relationship**: Connections with entities
     ```
     {
       npcID: UUID,
       memoryType: "relationship",
       targetEntityID: UUID,
       timestamp: ISODate (first meeting),
       lastInteraction: ISODate,
       clarity: 0.0-1.0,
       emotionalWeight: 0.0-1.0,
       accessCount: int,
       content: {
         affection: -100 to +100,
         trust: -100 to +100,
         fear: -100 to +100,
         firstImpression: string,
         sharedExperiences: [memoryID],
         relationshipType: string
       },
       tags: [string]
     }
     ```

2. Implement NPCMemoryRepository with methods:
   - `CreateMemory(memory) → memoryID, error`
   - `GetMemory(memoryID) → Memory, error`
   - `GetMemoriesByNPC(npcID, limit, offset) → []Memory, error`
   - `GetMemoriesByType(npcID, memoryType, limit) → []Memory, error`
   - `GetMemoriesByTimeframe(npcID, startTime, endTime) → []Memory, error`
   - `GetMemoriesByEntity(npcID, entityID) → []Memory, error`
   - `GetMemoriesByEmotion(npcID, minEmotionalWeight, limit) → []Memory, error`
   - `GetMemoriesByTags(npcID, tags, matchAll bool) → []Memory, error`
   - `UpdateMemory(memoryID, updates) → error`
   - `DeleteMemory(memoryID) → error`

3. Add memory tagging system:
   - Auto-tag entities mentioned (player_123, npc_456)
   - Auto-tag locations (coordinates, world)
   - Auto-tag emotions (joy, anger, fear, sadness, surprise, disgust)
   - Manual keywords from content analysis
   - Compound tags: "combat_with_player_123_at_forest"

4. Build memory retrieval with relevance scoring:
   - `GetRelevantMemories(npcID, context, limit) → []Memory`
   - Scoring formula: `score = (recency × 0.3) + (emotionalWeight × 0.4) + (accessCount × 0.1) + (contextMatch × 0.2)`
   - Context matching: tags present in current situation boost score
   - Return memories sorted by relevance score

5. Implement memory importance calculation:
   - Base importance: `emotionalWeight × clarity`
   - Recency bonus: `1.0 - (daysSinceCreation / 365)` capped at 0.0
   - Access frequency bonus: `min(accessCount / 10, 1.0)`
   - Final importance: `baseImportance × (1 + recencyBonus + accessBonus)`

Test Requirements (80%+ coverage):
- CreateMemory stores all memory types correctly
- GetMemory retrieves exact memory by ID
- GetMemoriesByType filters correctly (observation vs conversation vs event)
- GetMemoriesByTimeframe returns memories in date range
- GetMemoriesByEntity finds all memories involving specific entity
- GetMemoriesByEmotion filters by emotional weight threshold
- GetMemoriesByTags supports AND/OR tag matching
- Auto-tagging extracts entities, locations, emotions correctly
- GetRelevantMemories ranks by composite relevance score
- Memory importance calculation considers all factors
- UpdateMemory modifies fields without losing data
- Concurrent memory creation handles race conditions
- Store 10,000 memories and retrieve relevant subset in < 100ms
- Complex queries (tags + emotion + timeframe) perform in < 200ms

Acceptance Criteria:
- NPCs can store all four memory types with full schemas
- Memory retrieval supports filtering by type, time, entity, emotion, tags
- Relevance scoring prioritizes important and recent memories
- Auto-tagging extracts meaningful keywords
- Query performance meets targets (< 100ms simple, < 200ms complex)
- All tests pass with 80%+ coverage

Dependencies:
- MongoDB (add to docker-compose if not present)
- Event sourcing (Phase 0.1) for memory creation events

Files to Create:
- `internal/npc/memory/types.go` - Memory structs, enums
- `internal/npc/memory/repository.go` - NPCMemoryRepository
- `internal/npc/memory/tagging.go` - Auto-tagging logic
- `internal/npc/memory/relevance.go` - Relevance scoring
- `internal/npc/memory/repository_test.go` - Repository tests
- `internal/npc/memory/tagging_test.go` - Tagging tests
- `internal/npc/memory/relevance_test.go` - Scoring tests
- `migrations/mongodb/001_memory_indexes.js` - MongoDB indexes

---

## Phase 3.2: Memory Decay & Rehearsal (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 3.2 Memory Decay & Rehearsal:

Core Requirements:
1. Implement linear clarity decay over time:
   - Formula: `currentClarity = initialClarity × (1 - decayRate × daysSinceCreation)`
   - Base decay rate: `0.001` per day (0.1% daily decay)
   - Decay accelerates for low-emotion memories: `decayRate × (1 + (1 - emotionalWeight))`
   - High-emotion memories decay slower: `decayRate × emotionalWeight`
   - Minimum clarity floor: `0.1` (memories never completely vanish, just become fuzzy)

2. Add rehearsal bonus system:
   - Each memory access increments `accessCount`
   - Updates `lastAccessed` timestamp
   - Rehearsal protection: `rehearsalBonus = min(accessCount / 20, 0.5)`
   - Modified decay: `currentClarity = initialClarity × (1 - decayRate × daysSinceCreation × (1 - rehearsalBonus))`
   - Frequently accessed memories decay 50% slower at max

3. Build memory corruption system:
   - On access, chance to corrupt details: `corruptionChance = 0.05 × (1 - clarity)`
   - Corruption types:
     - Location drift (coordinates shift slightly)
     - Entity confusion (wrong participant remembered)
     - Detail loss (content fields become vague)
     - Emotional shift (emotionalWeight changes ±10%)
   - Corrupted memories tagged with `corrupted: true` flag
   - Original memory preserved in `originalContent` field for debugging

4. Implement background decay job:
   - Cron job runs daily: `DecayAllMemories()`
   - Batch process all memories created > 1 day ago
   - Update clarity values using decay formula
   - Apply corruption checks to accessed memories
   - Log statistics: memories decayed, corrupted, fallen below clarity threshold
   - Performance target: process 100k memories in < 5 minutes

5. Calculate memory retention score:
   - Formula: `retention = baseRetention × (1 - decayRate × timeSinceCreation) × (1 + rehearsalBonus × accessCount)`
   - `baseRetention = clarity × emotionalWeight`
   - Used to determine if memory surfaces in recall queries
   - Threshold for "forgetting": retention < 0.15

Test Requirements (80%+ coverage):
- Linear decay reduces clarity correctly over simulated days
- High-emotion memories decay slower than low-emotion
- Rehearsal bonus correctly reduces decay rate
- AccessCount increments on memory retrieval
- LastAccessed timestamp updates on access
- Corruption probability scales with low clarity
- Corrupted memories preserve original content
- DecayAllMemories processes batch efficiently
- Memory retention score combines all factors
- Memories below retention threshold excluded from queries
- Run decay simulation over 365 simulated days
- Verify memory half-life: 50% clarity at expected timepoint
- Concurrent access during decay job doesn't corrupt data

Acceptance Criteria:
- Clarity decays linearly over time with emotion-based adjustments
- Rehearsal (access frequency) protects memories from decay
- Corruption introduces realistic "fuzzy" details at low clarity
- Background job processes 100k memories in < 5 minutes
- Memory retention formula accurately predicts recall likelihood
- All tests pass with 80%+ coverage including time-series simulation

Dependencies:
- Phase 3.1 (Memory Storage) - requires memory schema
- Cron scheduler (add dependency: `github.com/robfig/cron/v3`)

Files to Create:
- `internal/npc/memory/decay.go` - Decay calculations
- `internal/npc/memory/rehearsal.go` - Rehearsal bonus logic
- `internal/npc/memory/corruption.go` - Memory corruption
- `internal/npc/memory/retention.go` - Retention scoring
- `internal/npc/memory/jobs.go` - Background decay job
- `internal/npc/memory/decay_test.go` - Decay simulation tests
- `internal/npc/memory/corruption_test.go` - Corruption tests
- `internal/npc/memory/jobs_test.go` - Batch processing tests

---

## Phase 3.3: Relationship System (1-2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 3.3 Relationship System with Behavioral Drift Detection:

Core Requirements:
1. Create relationship schema (extends memory type "relationship"):
   - Core affinity metrics (all -100 to +100):
     - `affection`: emotional bond (positive/negative feelings)
     - `trust`: reliability and honesty perception
     - `fear`: threat level and intimidation
   - Behavioral baseline tracking (all 0.0 to 1.0):
     - `aggression`: frequency of hostile actions
     - `generosity`: frequency of helpful/gift actions
     - `honesty`: pattern of truthful vs deceptive behavior
     - `sociability`: conversation initiation rate
     - `recklessness`: frequency of dangerous decisions
     - `loyalty`: consistency in supporting relationship
   - Recent interaction tracking:
     - `recentInteractions`: array of last 20 interactions
     - Each interaction: `{timestamp, actionType, affinityDelta, behavioralContext}`

2. Implement NPCRelationshipRepository methods:
   - `CreateRelationship(npcID, targetEntityID) → Relationship, error`
   - `GetRelationship(npcID, targetEntityID) → Relationship, error`
   - `GetAllRelationships(npcID) → []Relationship, error`
   - `UpdateAffinity(npcID, targetEntityID, affinityType, delta) → error`
   - `RecordInteraction(npcID, targetEntityID, interaction) → error`
   - `GetBehavioralBaseline(npcID, targetEntityID) → BehavioralBaseline, error`
   - `CalculateDrift(npcID, targetEntityID, recentActions) → DriftMetrics, error`

3. Add relationship update modifiers:
   - **Gift giving**: `affection += giftValue / 10`, `trust += 5`
   - **Threats**: `fear += 20`, `trust -= 10`, `affection -= 15`
   - **Helping**: `affection += 10`, `trust += 8`
   - **Lying (caught)**: `trust -= 25`, `affection -= 10`
   - **Violence**: `fear += 30`, `affection -= 40`, `trust -= 20`
   - **Betrayal**: `trust -= 50`, `affection -= 60`, `loyalty -= 0.3`
   - **Consistent support**: `trust += 5`, `loyalty += 0.05`, `affection += 5`
   - Modifiers respect bounds (-100 to +100 for affinity, 0.0 to 1.0 for behavioral)

4. Build relationship decay system:
   - Decay occurs when no interaction for extended period
   - Formula: `affinityDecay = 0.5 per 30 days without interaction`
   - Only affects affection and trust (not fear)
   - Strong relationships (affection > 75) decay 50% slower
   - Negative relationships (affection < -50) decay toward neutral faster
   - Background job: `DecayInactiveRelationships()` runs weekly

5. Implement behavioral baseline tracking:
   - Calculate baseline from first 20 interactions or 30-day history
   - Baseline values: `sum(behavioralMetric) / interactionCount`
   - Update baseline slowly: `newBaseline = (oldBaseline × 0.9) + (recentAverage × 0.1)`
   - Snapshot baseline when NPC is inhabited by player

6. Add drift detection system:
   - Calculate current behavior average from last 20 actions
   - Drift formula: `drift = |currentBehavior - historicalBaseline|`
   - Drift thresholds:
     - `0.0-0.3`: No significant drift
     - `0.3-0.5`: Subtle drift (NPC comments)
     - `0.5-0.7`: Moderate drift (NPC questions behavior)
     - `0.7+`: Severe drift (NPC alarmed, relationship impacts)
   - Per-trait drift tracking (separate drift scores for aggression, generosity, etc.)

7. Create NPC reactions to behavioral drift:
   - **Subtle Drift (0.3-0.5)**:
     - Generate concerned comments: "You seem different lately"
     - Relationship modifier: `affection += -5 to +5` depending on change direction
     - Memory created: observation type tagged with "personality_change"
   - **Moderate Drift (0.5-0.7)**:
     - Direct questioning: "That's not like you. What's going on?"
     - Relationship modifier: `trust -= 10`, `fear += 5` if negative change
     - Family intervention triggered if close relationship
     - Memory created with high emotional weight (0.7)
   - **Severe Drift (0.7+)**:
     - Alarmed response: "You're not yourself. Something is very wrong."
     - Relationship modifier: `trust -= 25`, possible relationship break
     - Community response if public figure (quest trigger)
     - Supernatural suspicion in magic worlds
     - Memory created with max emotional weight (1.0)

8. Build relationship modifier based on drift:
   - Positive drift (coward → brave): `affection += 15`, `trust += 10`
   - Negative drift (honest → deceptive): `affection -= 20`, `trust -= 30`
   - Calculate modifier: `affinityDelta = driftMagnitude × 50 × directionMultiplier`
   - Direction: +1 if personality improves by NPC's values, -1 if worsens

Test Requirements (80%+ coverage):
- CreateRelationship initializes with neutral affinity values
- GetRelationship retrieves by NPC-target pair
- UpdateAffinity respects bounds (-100 to +100)
- Gift/threat/help modifiers update affinity correctly
- Relationship decay reduces affection/trust over time (simulated days)
- Strong relationships decay slower than weak ones
- Behavioral baseline calculates correctly from interaction history
- Baseline updates slowly with new interactions
- Drift calculation compares current vs baseline accurately
- Drift thresholds trigger appropriate NPC responses
- Subtle drift generates comments, moderate triggers questioning
- Severe drift causes relationship penalties
- RecordInteraction stores last 20 interactions (FIFO)
- Concurrent relationship updates handle race conditions
- Drift detection works for all 6 behavioral traits
- Positive drift improves relationships, negative worsens
- Process 10k relationship updates in < 2 seconds

Acceptance Criteria:
- Relationships track affection, trust, fear accurately
- Behavioral baseline captures NPC's normal personality patterns
- Drift detection identifies when inhabited NPCs act out of character
- NPC reactions scale appropriately (subtle → moderate → severe)
- Relationship modifiers based on drift feel balanced
- Decay system maintains realistic relationship dynamics
- All tests pass with 80%+ coverage

Dependencies:
- Phase 3.1 (Memory Storage) - relationship memories
- Phase 3.2 (Memory Decay) - relationship decay follows similar patterns
- Event sourcing (Phase 0.1) for interaction events

Files to Create:
- `internal/npc/relationship/types.go` - Relationship, BehavioralBaseline structs
- `internal/npc/relationship/repository.go` - NPCRelationshipRepository
- `internal/npc/relationship/modifiers.go` - Affinity update logic
- `internal/npc/relationship/decay.go` - Relationship decay
- `internal/npc/relationship/baseline.go` - Behavioral baseline tracking
- `internal/npc/relationship/drift.go` - Drift detection and scoring
- `internal/npc/relationship/reactions.go` - NPC response generation
- `internal/npc/relationship/repository_test.go` - Repository tests
- `internal/npc/relationship/modifiers_test.go` - Modifier tests
- `internal/npc/relationship/drift_test.go` - Drift detection tests
- `internal/npc/relationship/reactions_test.go` - Reaction tests

---

## Phase 3.4: Emotional Memory Weighting (1-2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 3.4 Emotional Memory Weighting:

Core Requirements:
1. Add emotional intensity to memory schema:
   - `emotionalWeight`: 0.0 (neutral) to 1.0 (peak emotion)
   - Calculated from event context:
     - Combat: `0.7 + (damageTaken / maxHP) × 0.3`
     - First meeting: `0.5`
     - Gift received: `0.3 + (giftValue / wealthLevel) × 0.4`
     - Betrayal: `0.9`
     - Witnessed death: `0.8`
     - Casual conversation: `0.1 - 0.3`
     - Achievement: `0.6`
     - Threat to life: `0.95`

2. Boost memory importance for high-emotion events:
   - Modified importance: `importance = baseImportance × (1 + emotionalWeight)`
   - High-emotion memories (>0.7) get 1.7x to 2.0x importance boost
   - Prioritized in relevance scoring for recall
   - Resist decay better: `decayRate × (1 - emotionalWeight × 0.5)`

3. Implement emotion-triggered memory recall:
   - When NPC experiences similar emotion, related memories surface
   - Emotion similarity: `similarity = 1 - |currentEmotion - memoryEmotion|`
   - Threshold for triggering: similarity > 0.6
   - Returns top 5 most similar memories by combined score:
     - `score = (emotionalSimilarity × 0.5) + (importance × 0.3) + (recency × 0.2)`
   - Examples:
     - Current fear → recalls past threats/combat
     - Current joy → recalls celebrations/gifts
     - Current anger → recalls betrayals/injustices

4. Create emotion types system:
   - Six primary emotions (Ekman model):
     - `joy`: 0.0-1.0 intensity
     - `anger`: 0.0-1.0 intensity
     - `fear`: 0.0-1.0 intensity
     - `sadness`: 0.0-1.0 intensity
     - `surprise`: 0.0-1.0 intensity
     - `disgust`: 0.0-1.0 intensity
   - Complex emotions as combinations:
     - `anticipation = joy × 0.5 + surprise × 0.5`
     - `contempt = anger × 0.6 + disgust × 0.4`
     - `anxiety = fear × 0.7 + surprise × 0.3`
   - Memory tagged with dominant emotion (highest intensity)

5. Build EmotionEngine for event analysis:
   - `AnalyzeEvent(event) → EmotionProfile`
   - Event types map to emotion intensities:
     - Player gives gift → `{joy: 0.6, surprise: 0.3}`
     - Player attacks → `{fear: 0.8, anger: 0.5}`
     - Friend dies → `{sadness: 0.9, anger: 0.3}`
     - Discover treasure → `{joy: 0.7, surprise: 0.8}`
     - Betrayed by ally → `{anger: 0.8, sadness: 0.6}`
   - Personality modifies emotions:
     - Neurotic NPCs: fear/sadness +20%
     - Aggressive NPCs: anger +30%
     - Optimistic NPCs: joy +20%, sadness -20%

6. Implement emotional memory consolidation:
   - During sleep/rest cycles, high-emotion memories reinforced
   - Background job: `ConsolidateEmotionalMemories(npcID)`
   - Selects memories with emotionalWeight > 0.6 from past 24 hours
   - Increases clarity by `0.1 × emotionalWeight`
   - Links related emotional memories via `relatedMemories` field
   - Simulates "replaying" important events during rest

7. Add emotional memory decay resistance:
   - Formula: `effectiveDecayRate = baseDecayRate × (1 - emotionalWeight × 0.5)`
   - Peak emotion (1.0) provides 50% decay resistance
   - Neutral memories (0.0) get no protection
   - Test over simulated years: high-emotion memories persist longer

Test Requirements (80%+ coverage):
- EmotionalWeight assigned correctly based on event type
- High-emotion memories have boosted importance scores
- Memory importance calculation includes emotional multiplier
- Decay resistance formula reduces decay for emotional memories
- Emotion-triggered recall finds similar emotional memories
- Similarity scoring prioritizes close emotional matches
- Six primary emotions map correctly to event types
- Complex emotions combine primary emotions correctly
- Personality traits modify emotional intensities
- AnalyzeEvent produces consistent EmotionProfiles
- ConsolidateEmotionalMemories reinforces recent high-emotion memories
- Consolidated memories show increased clarity
- Related emotional memories linked correctly
- Simulate 365 days: emotional memories outlast neutral ones
- Process 10k memories with emotional scoring in < 200ms

Acceptance Criteria:
- Memories tagged with emotional intensity (0.0-1.0)
- High-emotion memories prioritized in recall and resist decay
- Emotion-triggered recall surfaces contextually relevant memories
- Six primary emotions supported with event-based assignment
- Personality influences emotional responses realistically
- Emotional memory consolidation simulates realistic sleep processing
- All tests pass with 80%+ coverage

Dependencies:
- Phase 3.1 (Memory Storage) - requires memory schema with emotionalWeight
- Phase 3.2 (Memory Decay) - integrates decay resistance
- Phase 3.3 (Relationships) - emotional events affect relationships

Files to Create:
- `internal/npc/emotion/types.go` - Emotion, EmotionProfile structs
- `internal/npc/emotion/engine.go` - EmotionEngine for event analysis
- `internal/npc/emotion/recall.go` - Emotion-triggered memory recall
- `internal/npc/emotion/consolidation.go` - Memory consolidation job
- `internal/npc/emotion/decay.go` - Emotional decay resistance
- `internal/npc/memory/emotional_scoring.go` - Importance boosting
- `internal/npc/emotion/engine_test.go` - EmotionEngine tests
- `internal/npc/emotion/recall_test.go` - Recall tests
- `internal/npc/emotion/consolidation_test.go` - Consolidation tests

---

# Phase 4: NPC Genetics & Appearance (2-3 weeks)

## Phase 4.1: Genetic System (1-2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 4.1 Genetic System with Mendelian Inheritance:

Core Requirements:
1. Design gene schema with dominant/recessive alleles:
   - Each trait has two alleles (one from each parent)
   - Allele types:
     - `Dominant` (capital letter): A, B, C...
     - `Recessive` (lowercase letter): a, b, c...
   - Genotype: pair of alleles (AA, Aa, aa)
   - Phenotype: expressed trait (Aa expresses A if dominant)
   - Gene struct:
     ```go
     type Gene struct {
       TraitName    string   // "height", "strength", "eyeColor"
       Allele1      string   // From parent1
       Allele2      string   // From parent2
       IsDominant1  bool     // Is allele1 dominant?
       IsDominant2  bool
       Phenotype    string   // Expressed trait
     }
     ```

2. Implement Mendelian inheritance with Punnett squares:
   - `InheritGene(parent1Gene, parent2Gene) → childGene, error`
   - Randomly select one allele from each parent
   - Determine phenotype from genotype:
     - If either allele is dominant: express dominant trait
     - If both recessive: express recessive trait
   - Punnett square logic for probability:
     - AA × AA → 100% AA
     - AA × Aa → 50% AA, 50% Aa
     - Aa × Aa → 25% AA, 50% Aa, 25% aa
     - Aa × aa → 50% Aa, 50% aa
   - Support co-dominance for some traits (both express partially)

3. Add mutation system (5% chance per gene):
   - `MutateGene(gene) → mutatedGene, error`
   - Mutation chance: 5% per gene during inheritance
   - Mutation types:
     - Allele flip: A → a or a → A (50% of mutations)
     - New allele: introduce rare variant (30% of mutations)
     - Amplification: strengthen existing trait (20% of mutations)
   - Beneficial vs neutral vs harmful mutations (weighted random)
   - Track mutation history in NPC lineage

4. Create trait-to-attribute mapping:
   - Physical attributes influenced by genes:
     - `Might`: strength gene (S/s), muscle gene (M/m)
       - SS = +10, Ss = +5, ss = +0
       - MM = +8, Mm = +4, mm = +0
       - Final: baseStrength + strengthBonus + muscleBonus
     - `Agility`: reflex gene (R/r), coordination gene (C/c)
     - `Endurance`: stamina gene (E/e), resilience gene (L/l)
     - `Vitality`: health gene (H/h), recovery gene (V/v)
   - Mental attributes influenced by genes:
     - `Intellect`: cognition gene (I/i), learning gene (K/k)
     - `Cunning`: perception gene (P/p), analysis gene (A/a)
   - Sensory attributes influenced by genes:
     - `Sight`: vision gene (Vi/vi), color gene (Co/co)
     - `Hearing`: auditory gene (Au/au), range gene (Ra/ra)
   - Each gene contributes ±15 max to attribute

5. Build trait-to-appearance mapping:
   - Physical appearance from genes:
     - **Height**: height gene (T/t)
       - TT = 180-200cm, Tt = 165-185cm, tt = 150-170cm
     - **Build**: build gene (B/b), muscle gene (M/m)
       - BB/MM = muscular, Bb/Mm = average, bb/mm = lean
     - **Hair color**: hair gene (Hr/hr), pigment gene (Pi/pi)
       - HrHr/PiPi = black, HrHr/Pipi = brown, hrhr/pipi = blonde
     - **Eye color**: eye gene (Ey/ey), melanin gene (Me/me)
       - EyEy/MeMe = brown, Eyey/Meme = hazel, eyey/meme = blue
     - **Facial features**: nose gene (N/n), jaw gene (J/j), cheek gene (Ch/ch)
   - Generate appearance string from genotype:
     - "Tall, muscular human with black hair, brown eyes, strong jaw"

6. Implement genetic diversity system:
   - Starting population needs diverse gene pool
   - Initialize with random genotypes ensuring variety
   - Avoid inbreeding depression: track common ancestors
   - Genetic compatibility check for breeding
   - Minimum genetic distance: `geneticSimilarity < 0.8` for healthy offspring

7. Test inheritance across multiple generations:
   - Simulate 5 generations of breeding
   - Verify Mendelian ratios (3:1, 9:3:3:1, etc.)
   - Track recessive trait emergence
   - Validate mutation accumulation rates
   - Check attribute distributions remain balanced

Test Requirements (80%+ coverage):
- InheritGene correctly implements Punnett square logic
- Dominant alleles express in heterozygous genotypes
- Recessive traits only express in homozygous recessive
- Mutation occurs at ~5% rate across large sample
- Mutation types distributed correctly (50/30/20 split)
- Trait-to-attribute mapping adds correct bonuses
- Multiple genes combine additively for attributes
- Appearance generation produces consistent descriptions
- Height ranges match genotype (TT > Tt > tt)
- Hair/eye color inheritance follows gene rules
- Genetic diversity check prevents excessive inbreeding
- Simulate 5 generations: verify Mendelian ratios
- Process 1000 inheritances in < 1 second
- Appearance strings are unique and descriptive

Acceptance Criteria:
- NPCs inherit genes from both parents following Mendelian genetics
- Dominant/recessive alleles express correctly in phenotype
- Mutations occur at 5% rate with appropriate variety
- Attributes are influenced by multiple genes additively
- Appearance is deterministically generated from genotype
- Genetic diversity maintained in breeding populations
- All tests pass with 80%+ coverage including multi-generational simulations

Dependencies:
- Phase 2.1 (Character System) - requires attribute schema
- Random number generator (crypto/rand for genetic randomness)

Files to Create:
- `internal/npc/genetics/types.go` - Gene, Genotype structs
- `internal/npc/genetics/inheritance.go` - Mendelian inheritance
- `internal/npc/genetics/mutation.go` - Mutation system
- `internal/npc/genetics/traits.go` - Trait-to-attribute mappings
- `internal/npc/genetics/appearance.go` - Appearance generation
- `internal/npc/genetics/diversity.go` - Genetic diversity checks
- `internal/npc/genetics/inheritance_test.go` - Punnett square tests
- `internal/npc/genetics/mutation_test.go` - Mutation rate tests
- `internal/npc/genetics/traits_test.go` - Attribute mapping tests
- `internal/npc/genetics/appearance_test.go` - Appearance generation tests
- `internal/npc/genetics/simulation_test.go` - Multi-generation tests

---

## Phase 4.2: Appearance Generation (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 4.2 Appearance Generation:

Core Requirements:
1. Generate physical description from genetic traits:
   - `GenerateAppearance(genotype, age, species) → AppearanceDescription, error`
   - Combine all physical trait genes into coherent description
   - Template structure:
     ```
     [Height descriptor], [build descriptor] [species] with [hair descriptor], 
     [eye descriptor], [facial feature descriptors]. [Age-specific details].
     ```
   - Example outputs:
     - "Tall, muscular human with straight black hair, piercing brown eyes, strong jawline and high cheekbones. In their prime with weathered features."
     - "Average height, lean elf with wavy golden hair, bright blue eyes, delicate features and pointed ears. Youthful with smooth skin."
     - "Short, stocky dwarf with thick red beard, deep-set green eyes, prominent nose and broad shoulders. Middle-aged with some gray in beard."

2. Implement appearance variation within genetic constraints:
   - Add minor randomization (±5%) to prevent clones
   - Variation sources:
     - Environmental factors: sun exposure, scars, weathering
     - Lifestyle: well-fed vs malnourished affects build
     - Occupation: calloused hands, muscular development
     - Random minor features: moles, freckles, birthmarks
   - Variation doesn't override genetics: tall gene → tall outcome always
   - Generate unique identifier: `appearanceID = hash(genotype + randomSeed)`

3. Add age-based appearance changes:
   - Age categories by species lifecycle:
     - **Child** (0-20% lifespan): "youthful, smooth skin, bright eyes"
     - **Young Adult** (20-40%): "in their prime, clear features"
     - **Adult** (40-60%): "mature, some weathering"
     - **Middle-Aged** (60-80%): "middle-aged, lines forming, some gray"
     - **Elder** (80-100%): "elderly, deeply lined, gray/white hair, stooped"
   - Age progression modifiers:
     - Hair color: gradual graying after 60% lifespan
     - Skin: wrinkles, age spots after 70% lifespan
     - Posture: slight stoop after 85% lifespan
     - Eyes: clouding after 90% lifespan
   - Store base appearance at birth, apply age modifiers dynamically

4. Create appearance descriptors by trait value:
   - **Height** (from height gene):
     - 190+ cm: "very tall", "towering"
     - 175-189: "tall", "above average height"
     - 160-174: "average height"
     - 145-159: "short", "below average height"
     - <145: "very short", "diminutive"
   - **Build** (from build + muscle genes):
     - High muscle, low fat: "muscular", "athletic", "powerful"
     - High muscle, high fat: "stocky", "heavyset", "burly"
     - Low muscle, low fat: "lean", "slender", "wiry"
     - Low muscle, high fat: "soft", "portly", "rotund"
   - **Hair texture** (from hair genes):
     - "straight", "wavy", "curly", "kinky"
   - **Facial features** (from feature genes):
     - Jaw: "strong jawline", "weak chin", "square jaw", "pointed chin"
     - Nose: "prominent nose", "button nose", "aquiline nose", "flat nose"
     - Cheeks: "high cheekbones", "hollow cheeks", "full cheeks"
     - Brow: "heavy brow", "delicate brow", "furrowed brow"

5. Add species-specific appearance templates:
   - **Human**: Standard template, highest variation
   - **Elf**: Add "pointed ears", "angular features", tendency toward "graceful", "elegant"
   - **Dwarf**: Add "thick beard" (males), "broad shoulders", "stocky" build default, "weathered" skin common
   - **Orc**: Add "prominent tusks", "green/gray skin tone", "muscular" default, "intimidating" presence
   - **Halfling**: Height capped at 120cm, "youthful" appearance even when aged, "curly hair" common
   - Each species has genetic variants within their phenotype range

6. Build appearance comparison system:
   - `CompareAppearances(appearance1, appearance2) → SimilarityScore`
   - Similarity metrics:
     - Height difference: `1 - (abs(height1 - height2) / maxHeightRange)`
     - Build similarity: compare muscle/fat distributions
     - Coloring match: hair + eyes + skin
     - Feature similarity: jawline, nose, cheeks
   - Overall similarity: weighted average (height 20%, build 20%, coloring 30%, features 30%)
   - Used for family resemblance: siblings should score 0.6-0.8 similarity
   - Identical twins from same genotype: 0.95+ (not 1.0 due to variation)

7. Implement appearance caching and updates:
   - Cache generated appearance string to avoid regeneration
   - Recalculate only when:
     - Age category changes (child → young adult)
     - Major event: scarring, injury, mutation
   - Store in NPCs table: `appearance TEXT, appearanceLastUpdated TIMESTAMP`
   - Background job: `UpdateAgedAppearances()` runs weekly

Test Requirements (80%+ coverage):
- GenerateAppearance produces consistent output for same genotype + age
- Appearance variation stays within genetic constraints
- Height descriptor matches height gene value ranges
- Build descriptor combines muscle + build genes correctly
- Hair/eye color matches genetic hair/eye genes
- Age category correctly determined from lifespan percentage
- Age modifiers apply at correct life stages (graying, wrinkles, etc.)
- Species templates add appropriate features (elf ears, dwarf beards)
- Species height ranges respected (halflings max 120cm)
- CompareAppearances scores family members 0.6-0.8
- Identical genotypes score 0.95+ similarity
- Minor variation prevents exact clones (different random seeds)
- Appearance caching retrieves without regeneration
- UpdateAgedAppearances only recalculates when age category changes
- Generate 1000 appearances in < 500ms

Acceptance Criteria:
- Physical descriptions generated deterministically from genotype
- Age-based changes applied correctly throughout lifespan
- Species-specific features included in templates
- Appearance variation creates diversity without breaking genetics
- Family resemblance detectable via similarity scoring
- All tests pass with 80%+ coverage

Dependencies:
- Phase 4.1 (Genetic System) - requires genotype with all trait genes
- Phase 2.1 (Character System) - species definitions

Files to Create:
- `internal/npc/appearance/generator.go` - GenerateAppearance function
- `internal/npc/appearance/descriptors.go` - Trait-to-descriptor mappings
- `internal/npc/appearance/aging.go` - Age-based appearance changes
- `internal/npc/appearance/species.go` - Species-specific templates
- `internal/npc/appearance/comparison.go` - Similarity scoring
- `internal/npc/appearance/cache.go` - Appearance caching logic
- `internal/npc/appearance/jobs.go` - Background update job
- `internal/npc/appearance/generator_test.go` - Generation tests
- `internal/npc/appearance/aging_test.go` - Age progression tests
- `internal/npc/appearance/comparison_test.go` - Similarity tests

---

# Phase 5: NPC AI & Desire Engine (3-4 weeks)

## Phase 5.1: Desire Engine (2 weeks)
### Status: In Progress
### Prompt:
Following TDD principles, implement Phase 5.1 Desire Engine with Need Hierarchy:

Core Requirements:
1. Implement need hierarchy (Maslow-inspired, game-adapted):
   - **Tier 1 - Survival** (highest priority):
     - `hunger`: 0-100 scale (100 = starving, 0 = full)
     - `thirst`: 0-100 scale (100 = parched, 0 = hydrated)
     - `sleep`: 0-100 scale (100 = exhausted, 0 = well-rested)
     - `safety`: 0-100 scale (100 = extreme danger, 0 = completely safe)
   - **Tier 2 - Social**:
     - `companionship`: 0-100 scale (100 = very lonely, 0 = socially fulfilled)
     - `conversation`: 0-100 scale (need to talk)
     - `affection`: 0-100 scale (need for positive relationships)
   - **Tier 3 - Achievement**:
     - `taskCompletion`: 0-100 scale (desire to finish goals)
     - `skillImprovement`: 0-100 scale (desire to practice/learn)
     - `resourceAcquisition`: 0-100 scale (desire for wealth/items)
   - **Tier 4 - Pleasure/Exploration**:
     - `curiosity`: 0-100 scale (desire to explore/discover)
     - `hedonism`: 0-100 scale (desire for enjoyment/indulgence)
     - `creativity`: 0-100 scale (desire to create/express)

2. Add survival needs with realistic progression:
   - **Hunger**:
     - Increases by `1.0 per hour` (game time)
     - Eating reduces by `foodValue` (bread = 30, meat = 50, feast = 100)
     - At 70+: "hungry" status, seek food
     - At 85+: "starving", prioritize above all else, -10% to all attributes
     - At 95+: take damage (1 HP per hour)
   - **Thirst**:
     - Increases by `1.5 per hour` (faster than hunger)
     - Drinking reduces by `drinkValue` (water = 50, ale = 40, wine = 35)
     - At 60+: "thirsty", seek water
     - At 80+: "parched", -15% to attributes
     - At 95+: take damage (2 HP per hour, more critical than hunger)
   - **Sleep**:
     - Increases by `1.0 per hour awake`
     - Sleeping reduces by `10 per hour of sleep`
     - At 75+: "tired", -10% to Focus and Reflexes
     - At 90+: "exhausted", -25% to all attributes, chance to fall asleep mid-action
   - **Safety**:
     - Calculated from context:
       - Combat: `100`
       - Nearby hostile entities: `50 + (hostileCount × 10)`
       - Dangerous area (wilderness at night): `30`
       - Town/safe zone: `5`
       - Home: `0`
     - High safety need drives fleeing, seeking shelter

3. Build social needs system:
   - **Companionship**:
     - Increases by `0.5 per hour` when alone
     - Decreases by `5 per hour` when with friendly NPCs/players
     - Personality modifiers:
       - Extraverted: +50% increase rate (loneliness comes faster)
       - Introverted: -50% increase rate (comfortable alone longer)
     - At 60+: seek out social interactions
     - At 80+: "lonely", -10 to Presence, mood becomes melancholy
   - **Conversation**:
     - Increases by `1.0 per hour` without talking
     - Decreases by `20` per meaningful conversation
     - Extraverted: +100% increase rate
     - At 50+: initiate conversations with nearby entities
   - **Affection**:
     - Increases by `0.2 per hour` (slow, long-term need)
     - Decreases through positive relationship interactions
     - At 70+: seek to strengthen bonds (give gifts, help others)

4. Create achievement goals system:
   - **Task Completion**:
     - Increases while tasks are pending
     - Each active task adds `10` to need
     - Decreases by `30` when task completed
     - Conscientiousness personality trait multiplies urgency: `needValue × (1 + conscientiousness)`
   - **Skill Improvement**:
     - Increases by `0.3 per hour` for NPCs with high Openness
     - Decreases when skill XP gained
     - Drives practice behaviors (crafting, combat training)
   - **Resource Acquisition**:
     - Based on wealth comparison to peers
     - Wealth percentile < 50%: need increases
     - Drives trading, looting, resource gathering

5. Add pleasure/exploration drives:
   - **Curiosity**:
     - High Openness NPCs: base value 40-60
     - Low Openness NPCs: base value 5-15
     - Increases near unexplored areas, new NPCs, unknown items
     - Drives exploration, asking questions, examining objects
   - **Hedonism**:
     - Varies by personality: high Extraversion + low Conscientiousness = high hedonism
     - Increases when bored (no stimulation for 2+ hours)
     - Drives seeking entertainment, drinking, gambling, leisure
   - **Creativity**:
     - High in NPCs with artist/craftsman occupations
     - Drives crafting, building, artistic expression

6. Calculate desire priorities with personality weighting:
   - Priority formula: `priority = needUrgency × personalityWeight × tierMultiplier`
   - Tier multipliers:
     - Tier 1 (Survival): `4.0` (always critical)
     - Tier 2 (Social): `2.0`
     - Tier 3 (Achievement): `1.5`
     - Tier 4 (Pleasure): `1.0`
   - Personality weights (0.5 to 2.0 based on relevant trait):
     - Hunger/Thirst: affected by Neuroticism (anxious about needs)
     - Social needs: affected by Extraversion
     - Achievement: affected by Conscientiousness
     - Curiosity: affected by Openness
   - Top priority need determines NPC's next action goal

7. Implement desire switching based on urgency:
   - NPCs can interrupt current activity if higher-priority need emerges
   - Interruption threshold: new priority must be `2x` current priority
   - Examples:
     - Chatting (social, priority 30) → interrupted by hunger at 90 (survival, priority 360)
     - Crafting (achievement, priority 45) → not interrupted by curiosity at 60 (pleasure, priority 60, only 1.3x)
   - Critical survival needs (95+) always interrupt immediately

8. Build desire fulfillment actions:
   - Map desires to actionable behaviors:
     - `hunger > 70` → `SeekFood()`, `EatMeal()`
     - `thirst > 60` → `SeekWater()`, `Drink()`
     - `sleep > 75` → `SeekBed()`, `Sleep()`
     - `safety > 50` → `Flee()`, `SeekShelter()`
     - `companionship > 60` → `SeekCompany()`, `JoinGroup()`
     - `conversation > 50` → `InitiateDialogue()`
     - `taskCompletion` → `ContinueTask()`, `CompleteGoal()`
     - `curiosity > 40` → `Explore()`, `Examine()`
   - NPCs autonomously select appropriate action based on top desire

Test Requirements (80%+ coverage):
- Hunger/thirst/sleep increase at correct rates over simulated time
- Eating/drinking/sleeping reduce needs by correct amounts
- Safety calculated correctly from context (combat, hostiles, location)
- Survival needs at high values (85+) apply attribute penalties
- Companionship need increases faster for extraverted NPCs
- Social needs decrease through appropriate interactions
- Task completion need scales with number of active tasks
- Curiosity higher in high-Openness NPCs
- Priority calculation combines urgency × personality × tier correctly
- Top-priority need correctly identified from all needs
- Desire switching interrupts when new priority > 2x current
- Critical survival (95+) always interrupts immediately
- Personality weights modify need priorities appropriately
- Conscientiousness increases task completion urgency
- Extraversion increases social need rates
- Simulate 24-hour cycle: verify realistic need progression
- Process 1000 NPCs' desire calculations in < 100ms

Acceptance Criteria:
- NPCs prioritize needs following Maslow-like hierarchy (Survival > Social > Achievement > Pleasure)
- Survival needs drive behavior when critical (hunger, thirst, sleep, safety)
- Social needs influenced by personality (extraversion, neuroticism)
- Achievement goals motivate task-oriented NPCs
- Pleasure/exploration drives emerge in stable, safe contexts
- Desire switching allows dynamic priority changes
- All tests pass with 80%+ coverage

Dependencies:
- Phase 2.1 (Character System) - requires attributes for penalties
- Phase 3.3 (Relationships) - social needs reference relationship states
- Time system (Phase 1) - needs progress with game time

Files to Create:
- `internal/npc/desire/types.go` - Need, Desire, Priority structs
- `internal/npc/desire/survival.go` - Hunger, thirst, sleep, safety
- `internal/npc/desire/social.go` - Companionship, conversation, affection
- `internal/npc/desire/achievement.go` - Task completion, skill improvement
- `internal/npc/desire/pleasure.go` - Curiosity, hedonism, creativity
- `internal/npc/desire/priority.go` - Priority calculation and switching
- `internal/npc/desire/fulfillment.go` - Desire-to-action mapping
- `internal/npc/desire/survival_test.go` - Survival need tests
- `internal/npc/desire/social_test.go` - Social need tests
- `internal/npc/desire/priority_test.go` - Priority calculation tests
- `internal/npc/desire/simulation_test.go` - 24-hour cycle simulation

---

## Phase 5.2: Personality System (1 week)
### Status: ⏳ Not Started
### Prompt:
Following TDD principles, implement Phase 5.2 Personality System with Big Five Model:

Core Requirements:
1. Define personality traits (Big Five OCEAN model):
   - Each trait scored 0-100:
     - **Openness** (0-100):
       - High: curious, creative, adventurous, seeks novelty
       - Low: practical, conventional, prefers routine
     - **Conscientiousness** (0-100):
       - High: organized, disciplined, goal-oriented, reliable
       - Low: spontaneous, flexible, disorganized
     - **Extraversion** (0-100):
       - High: outgoing, energetic, talkative, seeks company
       - Low: reserved, solitary, introspective
     - **Agreeableness** (0-100):
       - High: cooperative, empathetic, trusting, helpful
       - Low: competitive, skeptical, assertive
     - **Neuroticism** (0-100):
       - High: anxious, moody, sensitive to stress, emotional
       - Low: calm, stable, resilient, even-tempered

2. Derive personality from genetic traits + life experiences:
   - **Genetic baseline** (50% of personality):
     - Map genetic traits to personality:
       - Neurotic gene (Ne/ne): NeNe = 70-90, Nene = 50-70, nene = 20-50
       - Extraversion gene (Ex/ex): ExEx = 70-90, Exex = 50-70, exex = 20-50
       - Similar mappings for O/C/A
     - Genetic component: `geneticScore = rollFromGeneRange()`
   - **Life experiences** (50% of personality):
     - Childhood experiences (first 20% of lifespan):
       - Trauma: -20 to Neuroticism (more anxious)
       - Nurtured: +15 to Agreeableness
       - Challenged: +10 to Conscientiousness
     - Adult experiences:
       - Social success: +10 Extraversion
       - Repeated failures: +10 Neuroticism, -10 Conscientiousness
       - Exploration: +10 Openness
     - Track experience events in MongoDB: `personalityFormingEvents: [{event, traitImpact, timestamp}]`

3. Add personality influence on decision-making:
   - Decision context: `{options: [action1, action2...], stakes: low/medium/high}`
   - Personality modifiers per option:
     - **Openness** affects preference for novel vs familiar:
       - High Openness: +20 score to "explore cave", +5 to "visit tavern again"
       - Low Openness: +20 to "return home", +5 to "try new place"
     - **Conscientiousness** affects planning:
       - High: +25 to "plan carefully", -15 to "act impulsively"
       - Low: +20 to "wing it", -10 to "prepare thoroughly"
     - **Extraversion** affects social choices:
       - High: +30 to "join party", +15 to "initiate conversation"
       - Low: +25 to "stay home", +10 to "work alone"
     - **Agreeableness** affects cooperation:
       - High: +20 to "help stranger", +15 to "share resources"
       - Low: +15 to "ignore request", +20 to "negotiate hard"
     - **Neuroticism** affects risk assessment:
       - High: +30 to "avoid danger", -20 to "risky option"
       - Low: +15 to "take calculated risk", -5 to "panic response"
   - Final option score: `baseScore + sum(personalityModifiers) + random(0, 20)`
   - NPC selects highest-scoring option

4. Implement mood system (temporary emotional state):
   - Current mood affects behavior short-term (minutes to hours)
   - Mood types:
     - `cheerful`: +10 to social interactions, +5 Extraversion temporarily
     - `melancholy`: -10 to social, +10 to introspective actions
     - `anxious`: +15 Neuroticism, +20 to safety-seeking
     - `angry`: -15 Agreeableness, +20 to aggressive actions
     - `calm`: baseline personality, no modifiers
     - `excited`: +10 Openness, +15 to impulsive actions
   - Mood duration: `baseDuration × (1 + Neuroticism/100)`
     - High Neuroticism = moods last longer
   - Mood triggers:
     - Positive event: cheerful (30 min - 2 hours)
     - Threat: anxious (1-3 hours)
     - Betrayal: angry (2-6 hours)
     - Loss: melancholy (4-12 hours)
     - Achievement: excited (1-4 hours)
   - Mood decays back to neutral over time

5. Test personality consistency across decisions:
   - Run decision scenarios 100 times per personality archetype
   - High Openness NPCs should choose novel options > 70% of time
   - High Conscientiousness NPCs should plan > 75% of time
   - High Extraversion NPCs should seek social > 70% of time
   - Verify personality modifiers actually influence outcomes statistically

6. Build personality archetypes for testing:
   - **The Adventurer**: O=90, C=40, E=75, A=60, N=30
   - **The Scholar**: O=85, C=90, E=30, A=50, N=40
   - **The Leader**: O=60, C=80, E=90, A=70, N=25
   - **The Hermit**: O=50, C=60, E=10, A=40, N=70
   - **The Merchant**: O=50, C=85, E=65, A=45, N=35
   - Use for consistent testing and validation

Test Requirements (80%+ coverage):
- Personality traits initialize within 0-100 bounds
- Genetic baseline contributes ~50% to trait values
- Life experiences modify traits correctly (+/- values)
- Openness modifier increases score for novel options
- Conscientiousness increases score for planned actions
- Extraversion increases score for social choices
- Agreeableness increases score for cooperative actions
- Neuroticism increases score for safe options, decreases risky
- Mood modifiers apply temporarily to decisions
- Mood duration scales with Neuroticism correctly
- Mood decays back to neutral over simulated time
- High Openness archetype chooses exploration >70% (100 trials)
- High Conscientiousness chooses planning >75% (100 trials)
- Personality archetypes behave consistently across scenarios
- Process 1000 personality-influenced decisions in < 200ms

Acceptance Criteria:
- NPCs have Big Five personality traits (0-100 scale each)
- Personality derived from 50% genetics + 50% life experiences
- Personality traits influence decision-making with appropriate modifiers
- Mood system provides temporary emotional state changes
- Personality consistency validated statistically across many decisions
- All tests pass with 80%+ coverage

Dependencies:
- Phase 4.1 (Genetics) - personality partially inherited
- Phase 3.1 (Memory) - life experiences stored as memories
- Phase 5.1 (Desire Engine) - personality weights need priorities

Files to Create:
- `internal/npc/personality/types.go` - Personality, Mood structs
- `internal/npc/personality/genetics.go` - Genetic personality derivation
- `internal/npc/personality/experiences.go` - Life experience modifiers
- `internal/npc/personality/decisions.go` - Decision-making influence
- `internal/npc/personality/mood.go` - Mood system
- `internal/npc/personality/archetypes.go` - Predefined archetypes for testing
- `internal/npc/personality/genetics_test.go` - Genetic derivation tests
- `internal/npc/personality/decisions_test.go` - Decision influence tests
- `internal/npc/personality/mood_test.go` - Mood system tests
- `internal/npc/personality/consistency_test.go` - Statistical consistency tests

---

## Phase 5.3: NPC-to-NPC Interaction (1-2 weeks)
### Status: ⏳ Not Started
### Prompt:
Following TDD principles, implement Phase 5.3 NPC-to-NPC Autonomous Interaction:

Core Requirements:
1. Build conversation initiation logic:
   - NPCs autonomously decide when to start conversations
   - Initiation triggers:
     - **Proximity**: Within 5 meters of another NPC
     - **Companionship need**: Companionship > 60
     - **Relationship**: Existing positive relationship (affection > 40)
     - **Shared location**: Both NPCs idle in same area for > 2 minutes
     - **News to share**: NPC has high-emotion memory < 1 hour old
   - Initiation probability formula:
     ```
     probability = baseChance × extraversionMultiplier × relationshipBonus × needUrgency
     baseChance = 0.1 (10% per tick when conditions met)
     extraversionMultiplier = 1.0 + (extraversion / 100)
     relationshipBonus = 1.0 + (affection / 200)
     needUrgency = 1.0 + (companionship / 100)
     ```
   - Check every 30 seconds (game time) for idle NPCs

2. Implement basic conversation flow:
   - **Greeting Phase**:
     - Initiator: Select greeting based on relationship + personality
       - Affection > 60: "Hello friend!", "Good to see you!"
       - Affection 20-60: "Greetings.", "Hello there."
       - Affection < 20: "Oh, it's you.", *nods curtly*
     - Responder: Match greeting tone or deflect if busy/antisocial
   - **Topic Selection Phase**:
     - Initiator proposes topic from:
       - Recent high-emotion memory (emotionalWeight > 0.6)
       - Shared experience (both NPCs have related memories)
       - Current desire (if hunger high: "I'm famished, where's food?")
       - Random small talk (weather, local news)
     - Topic relevance scoring:
       ```
       score = (emotionalWeight × 0.4) + (recency × 0.3) + (shared × 0.3)
       ```
     - Select highest-scoring topic
   - **Response Phase**:
     - Responder retrieves related memories
     - Generates response based on:
       - Own experience with topic
       - Relationship with initiator (empathy if high affection)
       - Personality (Agreeableness affects supportiveness)
     - Response types:
       - Agreeable: supportive, empathetic
       - Neutral: acknowledgment, minimal engagement
       - Disagreeable: dismissive, argumentative
   - **Continuation or End**:
     - Continue if both NPCs engaged (response != dismissive)
     - Exchange 2-5 dialogue turns before natural end
     - End conversation: update relationship, create memories

3. Add relationship updates from conversations:
   - **Positive outcomes** (agreed, shared positive emotion):
     - `affection += 3`, `trust += 2`
     - Memory tagged "positive_interaction"
   - **Negative outcomes** (disagreed, conflict):
     - `affection -= 5`, `trust -= 3`
     - Memory tagged "argument"
   - **Neutral outcomes** (small talk, no conflict):
     - `affection += 1`
     - Memory tagged "casual_conversation"
   - Companionship need decreased by `15` for both participants
   - Conversation need decreased by `20`

4. Create memory formation during interactions:
   - Both NPCs create conversation memory:
     ```
     {
       memoryType: "conversation",
       timestamp: now,
       participants: [initiatorID, responderID],
       dialogue: [{speaker, text, emotion}...],
       topic: topicString,
       outcome: "positive" | "neutral" | "negative",
       emotionalWeight: calculateFromOutcome(),
       relationshipImpact: {targetID, affinityDelta}
     }
     ```
   - Emotional weight based on conversation outcome:
     - Positive deep conversation (shared trauma, celebration): 0.7-0.9
     - Conflict/argument: 0.6-0.8
     - Casual pleasant chat: 0.2-0.4
     - Boring small talk: 0.1-0.2

5. Test autonomous NPC interactions without player involvement:
   - Spawn 10 NPCs in same area
   - Run simulation for 1 hour (game time)
   - Verify:
     - NPCs initiate conversations based on proximity + needs
     - High-Extraversion NPCs initiate more frequently
     - Conversations create memories for both participants
     - Relationships update correctly after interactions
     - Companionship needs decrease after socializing
     - NPCs don't spam conversations (cooldown between initiations)

6. Implement conversation cooldown system:
   - After conversation ends, both NPCs have cooldown: `lastConversationTime`
   - Minimum time before next initiation: `5 minutes` (game time)
   - Prevents endless conversation loops
   - Cooldown reduced for high-Extraversion NPCs: `5min × (1 - extraversion/200)`

7. Build dialogue content generation (placeholder for Phase 6):
   - For Phase 5.3, use templated dialogue:
     - Greeting templates: "{greeting}, {name}! {opening_line}"
     - Topic templates: "Did you hear about {topic}?" / "I experienced {event}"
     - Response templates: "That sounds {adjective}!"/ "I {reaction} when that happened to me."
   - Emotion-appropriate language selection:
     - Joy: "wonderful", "exciting", "delightful"
     - Anger: "infuriating", "unacceptable", "outrageous"
     - Fear: "terrifying", "worrying", "dangerous"
     - Sadness: "heartbreaking", "unfortunate", "tragic"
   - Phase 6 will replace templates with LLM-generated dialogue

Test Requirements (80%+ coverage):
- Conversation initiation triggered by proximity + need conditions
- Initiation probability scales with extraversion correctly
- Relationship bonus increases initiation chance
- High companionship need increases initiation urgency
- Greeting tone matches relationship affection level
- Topic selection chooses highest-scoring recent memory
- Shared experiences score higher than solo memories
- Response type determined by personality (Agreeableness)
- Positive conversations increase affection and trust
- Negative conversations decrease affection and trust
- Both participants create conversation memories
- Memories include full dialogue array
- Emotional weight assigned based on outcome
- Companionship need decreases after conversation
- Conversation cooldown prevents spam initiations
- High-Extraversion NPCs have shorter cooldowns
- Simulate 10 NPCs for 1 hour: verify autonomous interactions
- Process 100 concurrent conversations in < 500ms

Acceptance Criteria:
- NPCs autonomously initiate conversations based on proximity, personality, needs, and relationships
- Conversation flow follows greeting → topic → response → end structure
- Topic selection prioritizes emotionally significant recent memories
- Responses reflect personality traits and relationship state
- Conversations update relationships appropriately (positive/neutral/negative)
- Both participants create memories with dialogue content
- Cooldown system prevents conversation spam
- All tests pass with 80%+ coverage including multi-NPC simulations

Dependencies:
- Phase 3.1 (Memory) - conversation memories
- Phase 3.3 (Relationships) - relationship updates
- Phase 5.1 (Desire Engine) - companionship need
- Phase 5.2 (Personality) - personality-based responses
- Phase 1 (Time System) - game time for cooldowns and need progression

Files to Create:
- `internal/npc/interaction/types.go` - Conversation, DialogueTurn structs
- `internal/npc/interaction/initiation.go` - Conversation trigger logic
- `internal/npc/interaction/flow.go` - Greeting, topic selection, response
- `internal/npc/interaction/topics.go` - Topic scoring and selection
- `internal/npc/interaction/responses.go` - Response generation
- `internal/npc/interaction/outcomes.go` - Relationship updates
- `internal/npc/interaction/memory.go` - Conversation memory creation
- `internal/npc/interaction/cooldown.go` - Conversation cooldown system
- `internal/npc/interaction/templates.go` - Template dialogue (temporary)
- `internal/npc/interaction/initiation_test.go` - Initiation logic tests
- `internal/npc/interaction/flow_test.go` - Conversation flow tests
- `internal/npc/interaction/outcomes_test.go` - Relationship update tests
- `internal/npc/interaction/simulation_test.go` - Multi-NPC autonomous tests

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
Phase 2.4: Skills & Progression System

(Phases 3-12 will be added as we progress)
