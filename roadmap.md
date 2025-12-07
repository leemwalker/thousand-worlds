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
### Status: ✅ Complete (87.8% test coverage)
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
### Status: ✅ Complete (91.6% test coverage)
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
### Status: ✅ Complete (86.1% test coverage)
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

# Phase 6: LLM Integration for Dialogue (2-3 weeks)

## Phase 6.1: Ollama Prompt Engineering (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 6.1 Ollama Prompt Engineering with Drift-Aware Dialogue:

Core Requirements:
1. Design dialogue prompts with comprehensive NPC context:
   - Prompt structure template:
     ```
     You are {npcName}, a {age}-year-old {species} {occupation}.
     
     PERSONALITY:
     - Openness: {O}/100 (describe tendency)
     - Conscientiousness: {C}/100 (describe tendency)
     - Extraversion: {E}/100 (describe tendency)
     - Agreeableness: {A}/100 (describe tendency)
     - Neuroticism: {N}/100 (describe tendency)
     
     CURRENT STATE:
     - Mood: {currentMood}
     - Top Desire: {topDesire} (urgency: {urgency}/100)
     - Physical Condition: {hungerLevel}, {fatigueLevel}
     
     CONTEXT:
     - Location: {location}
     - Time: {timeOfDay}, {weather}
     - Nearby: {entitiesPresent}
     
     RELATIONSHIP WITH {speakerName}:
     - Affection: {affection}/100
     - Trust: {trust}/100
     - Fear: {fear}/100
     - Shared History: {recentMemoriesSummary}
     
     RECENT MEMORIES (last 24 hours):
     {topRelevantMemories}
     
     CONVERSATION TOPIC: {currentTopic}
     
     {speakerName} says: "{playerInput}"
     
     Respond as {npcName} would, considering your personality, mood, desires, and relationship. Keep response to 1-3 sentences. Stay in character.
     ```

2. Add behavioral drift information for inhabited NPCs:
   - Detect if NPC is player-inhabited: check `isPlayerControlled` flag
   - If inhabited, add drift context to prompt:
     ```
     PERSONALITY DRIFT DETECTED:
     - Original Baseline:
       - Aggression: {baselineAggression}/1.0
       - Generosity: {baselineGenerosity}/1.0
       - Honesty: {baselineHonesty}/1.0
       - Sociability: {baselineSociability}/1.0
       - Recklessness: {baselineRecklessness}/1.0
       - Loyalty: {baselineLoyalty}/1.0
     
     - Current Behavior (last 20 actions):
       - Aggression: {currentAggression}/1.0 (drift: {aggressionDrift})
       - Generosity: {currentGenerosity}/1.0 (drift: {generosityDrift})
       - Honesty: {currentHonesty}/1.0 (drift: {honestyDrift})
       - Sociability: {currentSociability}/1.0 (drift: {socialityDrift})
       - Recklessness: {currentRecklessness}/1.0 (drift: {recklessnessDrift})
       - Loyalty: {currentLoyalty}/1.0 (drift: {loyaltyDrift})
     
     - Drift Level: {driftLevel} (subtle/moderate/severe)
     
     You have noticed {npcName} acting differently recently. Show concern, 
     curiosity, or alarm based on drift level and relationship closeness.
     ```

3. Generate concerned/curious/alarmed dialogue based on drift severity:
   - **Subtle Drift (0.3-0.5)**:
     - Prompt addition: "You've noticed {npcName} seems slightly different. Mention it casually if appropriate to conversation flow."
     - Example generations: "You've been more quiet than usual lately.", "Something on your mind? You seem distracted."
   - **Moderate Drift (0.5-0.7)**:
     - Prompt addition: "You are genuinely concerned about {npcName}'s behavior change. Express this clearly."
     - Example generations: "What's gotten into you? That's not like you at all.", "I need to ask - are you alright? You've been acting strange."
   - **Severe Drift (0.7+)**:
     - Prompt addition: "You are alarmed by {npcName}'s drastic personality change. This is deeply unsettling to you."
     - Example generations: "You're not yourself. Something is very wrong.", "I barely recognize you anymore. What happened?"
   - Include specific behavioral examples in prompt:
     - "You used to be peaceful, but you've started {X} fights recently."
     - "You were always stingy, but now you're giving away {Y} items."

4. Implement prompt template system:
   - `PromptBuilder` struct with fluent interface:
     ```go
     type PromptBuilder struct {
       npc            *NPC
       speaker        *Entity
       conversation   *Conversation
       recentMemories []*Memory
       relationship   *Relationship
       driftMetrics   *DriftMetrics
     }
     
     func NewPromptBuilder() *PromptBuilder
     func (pb *PromptBuilder) WithNPC(npc *NPC) *PromptBuilder
     func (pb *PromptBuilder) WithSpeaker(speaker *Entity) *PromptBuilder
     func (pb *PromptBuilder) WithConversation(conv *Conversation) *PromptBuilder
     func (pb *PromptBuilder) WithMemories(memories []*Memory) *PromptBuilder
     func (pb *PromptBuilder) WithRelationship(rel *Relationship) *PromptBuilder
     func (pb *PromptBuilder) WithDriftMetrics(drift *DriftMetrics) *PromptBuilder
     func (pb *PromptBuilder) Build() (string, error)
     ```
   - Template sections as composable functions:
     - `buildPersonalitySection(npc) → string`
     - `buildStateSection(npc) → string`
     - `buildRelationshipSection(relationship) → string`
     - `buildMemoriesSection(memories) → string`
     - `buildDriftSection(drift, relationship) → string`

5. Add response parsing and validation:
   - Parse Ollama JSON response:
     ```go
     type OllamaResponse struct {
       Model     string `json:"model"`
       CreatedAt string `json:"created_at"`
       Response  string `json:"response"`
       Done      bool   `json:"done"`
     }
     ```
   - Extract dialogue text from response
   - Validate response:
     - Not empty
     - Length < 500 characters (ensure 1-3 sentence compliance)
     - Contains no system instructions or meta-text
     - In-character (doesn't reference being an AI)
   - Sanitize response:
     - Remove quotation marks if present
     - Trim whitespace
     - Remove markdown formatting

6. Build dialogue caching (15-minute TTL):
   - Cache key structure: `dialogue:{npcID}:{speakerID}:{topicHash}:{contextHash}`
   - Context hash includes:
     - NPC mood
     - Top desire
     - Relationship values (affection, trust, fear)
     - Drift level (if applicable)
   - Cache hit: return cached response immediately
   - Cache miss: call Ollama, store result with TTL=15min
   - Invalidation triggers:
     - Relationship changes by >10 points
     - NPC mood changes
     - Severe drift event
   - Target: 95%+ cache hit rate for repeated interactions

7. Test prompt quality and response coherence:
   - Unit tests for prompt generation:
     - Personality section includes all Big Five traits
     - Drift section only included for inhabited NPCs
     - Memories section includes top 3-5 relevant memories
     - Relationship section reflects current affinity values
   - Integration tests with Ollama:
     - High-affection NPC generates warm responses
     - Low-affection NPC generates cold/dismissive responses
     - Fearful NPC (fear >60) generates nervous/submissive responses
     - Drift-aware responses mention behavioral changes appropriately
   - Coherence validation:
     - Response matches NPC personality (high Openness = curious language)
     - Response reflects mood (angry mood = terse, aggressive tone)
     - Response stays in character (no AI meta-references)

8. Test drift-aware dialogue generation:
   - Set up inhabited NPC with severe drift (aggression +0.8)
   - Generate dialogue from close friend (affection >70)
   - Verify response includes concern about aggression change
   - Test moderate drift with acquaintance (affection 40-60)
   - Verify response shows curiosity but not alarm
   - Test subtle drift - verify minimal or no mention

Test Requirements (80%+ coverage):
- PromptBuilder fluent interface composes prompt sections correctly
- Personality section describes all Big Five traits accurately
- Current state section includes mood, desires, physical condition
- Relationship section includes affection, trust, fear values
- Recent memories section includes top 3-5 relevant memories
- Drift section only appears for inhabited NPCs with drift >0.3
- Drift section includes baseline vs current behavioral metrics
- Drift severity (subtle/moderate/severe) correctly determined
- Response parsing extracts dialogue text from Ollama JSON
- Response validation rejects empty, too-long, or meta responses
- Response sanitization removes quotes, trims whitespace
- Cache key generation includes all context factors
- Cache hit retrieves stored response without Ollama call
- Cache invalidation triggers on relationship/mood changes
- Prompt with high-affection relationship generates warm responses
- Prompt with drift generates concerned/alarmed responses
- Process 100 prompt generations in < 50ms (excluding Ollama calls)

Acceptance Criteria:
- Prompts include NPC personality, mood, desires, relationships, memories, and context
- Drift-aware prompts add behavioral baseline and drift metrics for inhabited NPCs
- Dialogue reflects drift severity (subtle → concern, moderate → questioning, severe → alarm)
- Template system allows composable prompt construction
- Response parsing and validation ensures quality dialogue output
- Caching reduces Ollama calls by >90% for repeated interactions
- All tests pass with 80%+ coverage

Dependencies:
- Phase 5.2 (Personality) - Big Five traits
- Phase 5.1 (Desire Engine) - current desires
- Phase 3.1 (Memory) - recent memories
- Phase 3.3 (Relationships) - affection, trust, fear, drift metrics
- Ollama (add to docker-compose: `ollama/ollama:latest`, model: `llama3.1:8b`)

Files to Create:
- `internal/ai/prompt/builder.go` - PromptBuilder with fluent interface
- `internal/ai/prompt/sections.go` - Individual section generators
- `internal/ai/prompt/drift.go` - Drift-aware section generation
- `internal/ai/prompt/templates.go` - Template strings
- `internal/ai/ollama/client.go` - Ollama HTTP client
- `internal/ai/ollama/parser.go` - Response parsing and validation
- `internal/ai/cache/dialogue.go` - Dialogue caching with Redis
- `internal/ai/cache/keys.go` - Cache key generation
- `internal/ai/prompt/builder_test.go` - Prompt generation tests
- `internal/ai/ollama/client_test.go` - Ollama integration tests
- `internal/ai/cache/dialogue_test.go` - Cache hit/miss tests
- `docker-compose.yml` - Add Ollama service

---

## Phase 6.2: Dialogue Request Flow (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 6.2 Dialogue Request Flow with Drift Integration:

Core Requirements:
1. Integrate desire engine output with dialogue generation:
   - Before generating dialogue, determine NPC's intent from top desire:
     - `hunger > 70` → Intent: "seeking food", tone: urgent
     - `companionship > 60` → Intent: "seeking connection", tone: friendly
     - `fear > 50` → Intent: "seeking safety", tone: nervous
     - `taskCompletion` high → Intent: "focused on goal", tone: distracted
   - Add intent to prompt:
     ```
     CURRENT INTENT: {intent}
     You are currently {intentDescription}. Your responses should reflect this priority.
     ```
   - Intent examples in prompt:
     - "You are desperately hungry and hoping to find food soon."
     - "You are feeling lonely and want to connect with someone."
     - "You are nervous and looking for reassurance."

2. Build DialogueService to coordinate request flow:
   - Service struct:
     ```go
     type DialogueService struct {
       npcRepo         NPCRepository
       memoryRepo      NPCMemoryRepository
       relationshipRepo NPCRelationshipRepository
       desireEngine    *DesireEngine
       driftDetector   *DriftDetector
       promptBuilder   *PromptBuilder
       ollamaClient    *OllamaClient
       dialogueCache   *DialogueCache
     }
     
     func (ds *DialogueService) GenerateDialogue(
       ctx context.Context,
       npcID uuid.UUID,
       speakerID uuid.UUID,
       input string,
     ) (*DialogueResponse, error)
     ```

3. Implement complete dialogue generation flow:
   - **Step 1**: Fetch NPC data
     - Get NPC (personality, mood, current state)
     - Check if player-controlled (inhabited)
   - **Step 2**: Analyze NPC desires
     - Calculate current desire priorities
     - Determine top desire and intent
   - **Step 3**: Retrieve relationship and memories
     - Get relationship with speaker (affection, trust, fear)
     - Get recent relevant memories (last 24 hours, high emotional weight)
     - Filter memories by relevance to current topic
   - **Step 4**: Detect behavioral drift (if inhabited)
     - Calculate drift metrics from last 20 actions
     - Determine drift level (subtle/moderate/severe)
     - Check if speaker would notice (relationship closeness)
   - **Step 5**: Build prompt
     - Use PromptBuilder with all gathered context
     - Include drift section if applicable
     - Add intent from desire engine
   - **Step 6**: Check cache
     - Generate cache key from context
     - Return cached response if hit
   - **Step 7**: Call Ollama
     - Send prompt to Ollama API
     - Parse and validate response
     - Handle errors with fallback responses
   - **Step 8**: Extract dialogue and emotional reaction
     - Parse dialogue text
     - Infer emotional reaction from response tone:
       - Exclamation marks → excited/angry
       - Question marks + concern words → worried
       - Short responses → dismissive/busy
       - Long responses → engaged/talkative
   - **Step 9**: Update NPC state
     - Create conversation memory (including drift observations if mentioned)
     - Update relationship based on interaction
     - Decrease conversation/companionship needs
   - **Step 10**: Cache and return
     - Store response in cache with TTL
     - Return DialogueResponse to caller

4. Send NPC state + player input + drift data to ai-gateway:
   - Create DialogueRequest payload:
     ```go
     type DialogueRequest struct {
       NPCID          uuid.UUID              `json:"npc_id"`
       SpeakerID      uuid.UUID              `json:"speaker_id"`
       Input          string                 `json:"input"`
       NPCState       NPCState               `json:"npc_state"`
       Relationship   Relationship           `json:"relationship"`
       RecentMemories []*Memory              `json:"recent_memories"`
       DriftMetrics   *DriftMetrics          `json:"drift_metrics,omitempty"`
       Intent         string                 `json:"intent"`
     }
     ```
   - Drift metrics only included if inhabited and drift >0.3

5. Process LLM response and extract dialogue + emotional reaction:
   - Parse Ollama response text
   - Infer emotional reaction using keyword analysis:
     - Joy keywords: "wonderful", "great", "happy", "love" → emotion: joy (0.7)
     - Anger keywords: "damn", "furious", "unacceptable" → emotion: anger (0.8)
     - Fear keywords: "terrifying", "worried", "scared" → emotion: fear (0.7)
     - Sadness keywords: "unfortunate", "sad", "heartbreaking" → emotion: sadness (0.6)
   - Punctuation analysis:
     - Multiple exclamation marks → intensity +0.2
     - Multiple question marks → confusion/concern +0.3
   - Response length analysis:
     - <20 words → disinterest (emotionalWeight 0.2)
     - 20-50 words → engaged (emotionalWeight 0.5)
     - >50 words → very engaged (emotionalWeight 0.7)

6. Update NPC memory and relationships post-conversation (including drift observations):
   - Create conversation memory for NPC:
     ```go
     memory := &Memory{
       NPCID:           npcID,
       MemoryType:      "conversation",
       Timestamp:       time.Now(),
       Clarity:         1.0,
       EmotionalWeight: calculatedEmotion,
       Content: ConversationContent{
         Participants: []uuid.UUID{npcID, speakerID},
         Dialogue: []DialogueTurn{
           {Speaker: speakerID, Text: input, Emotion: "neutral"},
           {Speaker: npcID, Text: response, Emotion: inferredEmotion},
         },
         Topic: extractTopic(input),
         Outcome: determineOutcome(relationship, response),
         RelationshipImpact: calculateImpact(relationship, response),
       },
       Tags: []string{"conversation", "player_interaction"},
     }
     ```
   - If drift mentioned in response, add tag: `"drift_observation"`
   - If drift severe and response alarmed, add tag: `"personality_concern"`
   - Update relationship:
     - Positive response → affection +2, trust +1
     - Negative response → affection -3, trust -2
     - Drift concern expressed → trust -5 (uncertainty about change)
   - Decrease needs:
     - Conversation need: -20
     - Companionship need: -15 (if friendly interaction)

7. Implement fallback responses if LLM fails:
   - Fallback triggers:
     - Ollama timeout (>10 seconds)
     - Ollama returns error
     - Response validation fails
     - Response parsing fails
   - Fallback response templates by context:
     - High affection: "{npcName} smiles warmly but seems distracted."
     - Low affection: "{npcName} grunts noncommittally."
     - High fear: "{npcName} looks away nervously."
     - High companionship need: "{npcName} seems eager to talk but struggles to find words."
   - Log fallback usage for monitoring
   - Fallbacks don't create memories (insufficient context)

Test Requirements (80%+ coverage):
- DialogueService.GenerateDialogue fetches all required data
- Desire engine determines correct intent from top desire
- Relationship and recent memories retrieved correctly
- Drift detection calculates metrics for inhabited NPCs
- Drift detection skipped for non-inhabited NPCs
- PromptBuilder receives all context components
- Cache check returns cached response when available
- Cache miss triggers Ollama call
- Ollama client sends properly formatted request
- Response parser extracts dialogue text correctly
- Emotional reaction inferred from keywords and punctuation
- Conversation memory created with full dialogue turns
- Drift observations tagged when mentioned in response
- Relationship updates based on response sentiment
- Companionship/conversation needs decrease after interaction
- Fallback responses used when Ollama times out
- Fallback responses match relationship context
- Process complete dialogue flow in < 2 seconds (with Ollama)
- Process 10 concurrent dialogue requests without blocking

Acceptance Criteria:
- Dialogue generation integrates NPC desires, personality, mood, memories, and relationships
- Drift data included in requests for inhabited NPCs
- LLM responses parsed and emotional reactions extracted
- Conversation memories created with drift observations when applicable
- Relationships updated post-conversation including drift-based trust impacts
- Fallback system handles Ollama failures gracefully
- All tests pass with 80%+ coverage

Dependencies:
- Phase 6.1 (Prompt Engineering) - PromptBuilder, OllamaClient
- Phase 5.1 (Desire Engine) - intent determination
- Phase 3.1 (Memory) - memory creation
- Phase 3.3 (Relationships) - relationship updates, drift detection
- Redis (caching)

Files to Create:
- `internal/ai/dialogue/service.go` - DialogueService orchestration
- `internal/ai/dialogue/types.go` - Request/Response types
- `internal/ai/dialogue/flow.go` - Step-by-step flow implementation
- `internal/ai/dialogue/emotion.go` - Emotional reaction inference
- `internal/ai/dialogue/fallback.go` - Fallback response templates
- `internal/ai/dialogue/memory.go` - Post-conversation memory creation
- `internal/ai/dialogue/relationship.go` - Relationship update logic
- `internal/ai/dialogue/service_test.go` - Service orchestration tests
- `internal/ai/dialogue/flow_test.go` - Complete flow tests
- `internal/ai/dialogue/emotion_test.go` - Emotion inference tests
- `internal/ai/dialogue/fallback_test.go` - Fallback tests

---

## Phase 6.3: Performance Optimization (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 6.3 Performance Optimization for AI System:

Core Requirements:
1. Target 95%+ cache hit rate for area descriptions:
   - Area descriptions change less frequently than dialogue
   - Cache key structure: `area:{worldID}:{x}:{y}:{z}:{weather}:{timeOfDay}:{season}`
   - Context factors:
     - Location coordinates
     - Weather state
     - Time of day (dawn/morning/noon/dusk/evening/night)
     - Season (spring/summer/fall/winter)
     - Player perception skill (bucketed: 0-25, 26-50, 51-75, 76-100)
   - TTL: 60 minutes (descriptions stable longer than dialogue)
   - Invalidation: Only on major world events (building constructed, terrain changed)

2. Implement area description generation:
   - `GenerateAreaDescription(ctx context.Context, location Location, observer *Entity) → string, error`
   - Prompt structure:
     ```
     Generate a description of this location:
     
     COORDINATES: {x}, {y}, {z} in {worldName}
     BIOME: {biome}
     TERRAIN: {terrainType}
     WEATHER: {weatherState} ({temperature}°C)
     TIME: {timeOfDay}, {season}
     
     NEARBY ENTITIES:
     {entitiesList}
     
     NEARBY STRUCTURES:
     {structuresList}
     
     OBSERVER PERCEPTION: {perceptionSkill}/100
     
     Describe what the observer sees, hears, and smells. Adjust detail level based on perception skill:
     - Low (0-25): Basic, vague description
     - Medium (26-75): Standard detail
     - High (76-100): Rich, nuanced detail
     
     Keep description to 3-5 sentences.
     ```
   - Perception-based detail tiers:
     - **Low (0-25)**: "A forest path. Trees surround you. Birds chirp."
     - **Medium (26-75)**: "A narrow forest path winds between oak trees. Bird calls echo from the canopy. Damp earth smells fill the air."
     - **High (76-100)**: "A deer trail meanders between ancient oaks, their gnarled roots breaking through the loamy soil. Crow alarm calls from the northwest suggest a predator nearby. The rich scent of decomposing leaves mingles with pine resin."

3. Batch dialogue requests where possible:
   - Identify batchable scenarios:
     - Multiple NPCs responding to same player action (e.g., entering room with 5 NPCs)
     - NPC-to-NPC conversations (both NPCs need responses)
   - Implement BatchDialogueRequest:
     ```go
     type BatchDialogueRequest struct {
       Requests []*DialogueRequest `json:"requests"`
     }
     
     type BatchDialogueResponse struct {
       Responses []*DialogueResponse `json:"responses"`
       Errors    map[int]error       `json:"errors"` // Index to error mapping
     }
     ```
   - Send multiple prompts to Ollama in single request
   - Process responses in parallel
   - Target: 3x throughput improvement for batch scenarios

4. Monitor Ollama CPU/RAM usage (target: < 80% sustained):
   - Add metrics collection:
     ```go
     type OllamaMetrics struct {
       CPUPercent       float64
       RAMUsedMB        int64
       RAMTotalMB       int64
       ActiveRequests   int32
       QueuedRequests   int32
       AvgResponseTime  time.Duration
       RequestsPerMin   int
     }
     ```
   - Poll Ollama container stats every 10 seconds
   - Expose metrics to Prometheus:
     - `ollama_cpu_percent`
     - `ollama_ram_used_mb`
     - `ollama_active_requests`
     - `ollama_queued_requests`
     - `ollama_response_time_ms`
   - Alert if CPU >80% for 5+ minutes or RAM >90%

5. Add request queueing if Ollama overloaded:
   - Implement queue with priority levels:
     - **Critical** (priority 1): Combat dialogue, safety-critical responses
     - **High** (priority 2): Player-facing dialogue
     - **Normal** (priority 3): NPC-to-NPC dialogue
     - **Low** (priority 4): Area descriptions, background processing
   - Queue parameters:
     - Max queue size: 1000 requests
     - Max concurrent Ollama requests: 10
     - Timeout: 30 seconds in queue
   - Request processing:
     ```go
     type DialogueQueue struct {
       critical chan *DialogueRequest
       high     chan *DialogueRequest
       normal   chan *DialogueRequest
       low      chan *DialogueRequest
       semaphore chan struct{} // Limit concurrent requests
     }
     
     func (dq *DialogueQueue) Enqueue(req *DialogueRequest, priority Priority) error
     func (dq *DialogueQueue) process() // Goroutine worker
     ```
   - Process priority queues in order: critical → high → normal → low
   - If queue full, reject low-priority requests with error

6. Implement graceful degradation:
   - If Ollama unavailable or overloaded:
     - Area descriptions: use template-based fallbacks
     - Player dialogue: use template responses (Phase 5.3 style)
     - NPC-to-NPC: defer until capacity available
   - Fallback quality tiers:
     - Tier 1 (Ollama available): Full LLM-generated content
     - Tier 2 (Ollama slow): Cached responses only, templates for new
     - Tier 3 (Ollama down): All templates, log for later regeneration
   - Track degradation state in metrics

7. Test concurrent dialogue performance:
   - Load test: 10 concurrent players, each triggering 1 dialogue/second
   - Expected: 10 requests/second sustained for 5 minutes
   - Verify:
     - Cache hit rate >90%
     - Ollama CPU stays <80%
     - Response time P95 <2 seconds
     - No request failures
   - Stress test: 50 concurrent requests
   - Expected: Queue handles overflow, no crashes
   - Verify:
     - Low-priority requests queued or rejected gracefully
     - Critical requests always processed
     - System recovers when load decreases

Test Requirements (80%+ coverage):
- Area description cache key includes all context factors
- Area descriptions cached with 60-minute TTL
- Cache hit returns description without Ollama call
- Perception skill determines description detail level
- Low perception generates basic descriptions
- High perception generates rich, detailed descriptions
- BatchDialogueRequest sends multiple prompts efficiently
- Batch responses map back to original requests correctly
- Ollama metrics collected and exposed to Prometheus
- CPU/RAM alerts trigger at correct thresholds
- DialogueQueue enqueues requests by priority correctly
- Critical priority requests processed before others
- Low priority requests rejected when queue full
- Request semaphore limits concurrent Ollama calls to 10
- Graceful degradation uses templates when Ollama unavailable
- Load test: 10 concurrent players sustained for 5 minutes
- Stress test: 50 concurrent requests handled without crashes
- Cache hit rate measured and >90% after warmup period

Acceptance Criteria:
- Area descriptions cached with 95%+ hit rate
- Perception skill affects description detail appropriately
- Batch requests improve throughput by 3x
- Ollama resource usage monitored with Prometheus metrics
- Request queue prevents Ollama overload
- Priority system ensures critical requests processed first
- Graceful degradation maintains functionality when Ollama stressed
- All tests pass with 80%+ coverage including load tests

Dependencies:
- Phase 6.1 (Prompt Engineering) - prompt generation
- Phase 6.2 (Dialogue Flow) - dialogue service
- Redis (caching)
- Prometheus (metrics)
- Ollama (running in docker-compose)

Files to Create:
- `internal/ai/area/description.go` - Area description generation
- `internal/ai/area/cache.go` - Area description caching
- `internal/ai/batch/processor.go` - Batch request processing
- `internal/ai/metrics/collector.go` - Ollama metrics collection
- `internal/ai/metrics/prometheus.go` - Prometheus exporters
- `internal/ai/queue/dialogue_queue.go` - Priority queue implementation
- `internal/ai/queue/worker.go` - Queue worker goroutines
- `internal/ai/degradation/fallback.go` - Graceful degradation logic
- `internal/ai/area/description_test.go` - Area description tests
- `internal/ai/batch/processor_test.go` - Batch processing tests
- `internal/ai/queue/dialogue_queue_test.go` - Queue tests
- `internal/ai/load_test.go` - Load and stress tests

---

# Phase 7: Combat System (2-3 weeks)

## Phase 7.1: Action Queue System (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 7.1 Action Queue System with Reaction Times:

Core Requirements:
1. Implement action queue (FIFO based on reaction time):
   - Combat is turn-based with reaction time delays
   - Action struct:
     ```go
     type CombatAction struct {
       ActionID       uuid.UUID
       ActorID        uuid.UUID
       TargetID       uuid.UUID
       ActionType     ActionType // attack, defend, flee, useItem
       ReactionTime   time.Duration
       QueuedAt       time.Time
       ExecuteAt      time.Time // QueuedAt + ReactionTime
       Resolved       bool
     }
     ```
   - CombatQueue maintains ordered list by ExecuteAt time
   - Process actions in chronological order (earliest ExecuteAt first)

2. Calculate reaction time from character stats:
   - Base reaction times by action type:
     - Quick attack: 800ms base
     - Normal attack: 1000ms base
     - Heavy attack: 1500ms base
     - Defend: 500ms base
     - Flee: 2000ms base
     - Use item: 700ms base
   - Agility modifier formula:
     ```
     finalReactionTime = baseTime × (1 - (Agility / 100) × 0.3)
     ```
   - Examples:
     - Agility 50, Normal attack: 1000ms × (1 - 0.15) = 850ms
     - Agility 90, Quick attack: 800ms × (1 - 0.27) = 584ms
     - Agility 20, Heavy attack: 1500ms × (1 - 0.06) = 1410ms
   - Minimum reaction time: 200ms (prevents instant actions)
   - Status effects modify reaction time:
     - Slowed: ×1.5
     - Hasted: ×0.7
     - Stunned: action fails, no queue entry

3. Add action types with different characteristics:
   - **Attack** (normal):
     - Base damage, standard reaction time
     - Stamina cost: 15
     - Success depends on weapon skill + Might
   - **Quick Attack**:
     - 70% base damage, faster reaction time
     - Stamina cost: 10
     - Higher chance to interrupt opponent's action
   - **Heavy Attack**:
     - 150% base damage, slower reaction time
     - Stamina cost: 25
     - Lower accuracy but higher critical chance
   - **Defend**:
     - No damage, fast reaction time
     - Reduces incoming damage by 50% until next action
     - Stamina cost: 5
     - Defending state cleared when next action executes
   - **Flee**:
     - Attempt to escape combat
     - Slow reaction time (vulnerable during)
     - Success chance: `(Agility / 100) × (1 - (enemyAgility / 200))`
     - Stamina cost: 20
   - **Use Item**:
     - Consume item (potion, bandage, etc.)
     - Fast reaction time
     - Stamina cost: 5
     - Item effects apply immediately

4. Prevent action spam (enforce minimum reaction time):
   - Track last action time per combatant:
     ```go
     type Combatant struct {
       EntityID         uuid.UUID
       LastActionTime   time.Time
       CurrentAction    *CombatAction
       DefendingUntil   time.Time
       StatusEffects    []*StatusEffect
     }
     ```
   - Validation before queueing action:
     ```go
     func (cs *CombatService) CanQueueAction(combatant *Combatant, now time.Time) error {
       if combatant.CurrentAction != nil && !combatant.CurrentAction.Resolved {
         return ErrActionInProgress
       }
       
       timeSinceLastAction := now.Sub(combatant.LastActionTime)
       if timeSinceLastAction < MinReactionTime {
         return ErrActionTooSoon
       }
       
       return nil
     }
     ```
   - Reject action if:
     - Previous action not yet resolved
     - Less than 200ms since last action
     - Combatant is stunned

5. Test queue ordering with multiple combatants:
   - Scenario: 3 combatants in combat
     - Player A queues Normal Attack at T=0 (Agility 60, reaction time 820ms)
     - NPC B queues Quick Attack at T=100ms (Agility 40, reaction time 704ms)
     - NPC C queues Heavy Attack at T=50ms (Agility 70, reaction time 1185ms)
   - Expected execution order:
     1. T=754ms: NPC B's Quick Attack (queued at 100ms + 704ms - 50ms early queue advantage)
     2. T=820ms: Player A's Normal Attack
     3. T=1235ms: NPC C's Heavy Attack
   - Verify actions execute in correct chronological order
   - Verify each action respects its calculated reaction time

6. Implement combat state machine:
   - Combat states:
     - `Idle`: No combat active
     - `InCombat`: Active combat, actions being queued/processed
     - `Fleeing`: Attempting to escape
     - `Defeated`: Combatant HP ≤ 0
   - State transitions:
     - Idle → InCombat: when attacked or initiating attack
     - InCombat → Fleeing: when flee action succeeds
     - InCombat → Defeated: when HP reaches 0
     - Fleeing → InCombat: if flee fails
     - InCombat → Idle: when all opponents defeated/fled

7. Build action resolution system:
   - Process queue every 50ms (game tick)
   - For each action with ExecuteAt ≤ now:
     - Validate combatant still alive and not stunned
     - Check stamina available
     - Execute action (resolve in Phase 7.2)
     - Mark action as resolved
     - Update LastActionTime
     - Remove from queue
   - Handle interrupted actions:
     - If combatant takes heavy damage (>30% max HP) before action executes
     - Chance to interrupt: `damagePercent × 0.5`
     - Interrupted action fails, reaction time wasted

Test Requirements (80%+ coverage):
- CombatAction struct stores all required fields
- CombatQueue maintains actions ordered by ExecuteAt time
- Reaction time calculation applies Agility modifier correctly
- Minimum reaction time (200ms) enforced
- Quick attack has shorter reaction time than normal attack
- Heavy attack has longer reaction time than normal attack
- Defend action has fastest reaction time
- CanQueueAction rejects spam attempts (<200ms since last action)
- CanQueueAction rejects if previous action unresolved
- CanQueueAction rejects if combatant stunned
- Status effects (slowed, hasted) modify reaction times correctly
- Multiple combatants' actions queue and execute in correct order
- Action resolution processes queue every tick (50ms)
- Resolved actions removed from queue
- LastActionTime updated after action execution
- Interrupted actions fail and are removed from queue
- Combat state machine transitions correctly
- Process 100 concurrent combatants' actions in < 100ms per tick

Acceptance Criteria:
- Actions queue with calculated reaction times based on Agility
- Queue processes actions in chronological order (ExecuteAt)
- Action spam prevented by minimum reaction time enforcement
- Multiple action types supported with different characteristics
- Combat state machine manages combat lifecycle
- All tests pass with 80%+ coverage

Dependencies:
- Phase 2.1 (Character System) - Agility attribute
- Phase 2.2 (Stamina System) - stamina costs
- Phase 1 (Time System) - game time for action timing

Files to Create:
- `internal/combat/types.go` - CombatAction, Combatant, ActionType enums
- `internal/combat/queue.go` - CombatQueue implementation
- `internal/combat/reaction_time.go` - Reaction time calculations
- `internal/combat/validation.go` - Action validation (spam prevention)
- `internal/combat/state_machine.go` - Combat state management
- `internal/combat/resolution.go` - Action resolution loop
- `internal/combat/queue_test.go` - Queue ordering tests
- `internal/combat/reaction_time_test.go` - Reaction time tests
- `internal/combat/validation_test.go` - Spam prevention tests
- `internal/combat/state_machine_test.go` - State transition tests

---

## Phase 7.2: Damage & Weapons (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 7.2 Damage Calculation and Weapon System:

Core Requirements:
1. Implement damage calculation formula:
   - Base damage formula:
     ```
     rawDamage = weaponBaseDamage × skillModifier × attributeModifier × (roll / 100)
     ```
   - Skill modifier:
     ```
     skillModifier = 1.0 + (weaponSkill / 200)
     - Skill 0: 1.0x damage
     - Skill 50: 1.25x damage
     - Skill 100: 1.5x damage
     ```
   - Attribute modifier (by weapon type):
     - **Slashing/Bludgeoning**: `1.0 + (Might / 200)`
     - **Piercing**: `1.0 + ((Might + Agility) / 400)`
     - **Ranged**: `1.0 + (Agility / 200)`
   - Roll: d100 (1-100)
   - Final damage after armor reduction:
     ```
     finalDamage = rawDamage × (1 - armorReduction)
     ```

2. Add weapon types with base damage values:
   - **Slashing weapons**:
     - Short sword: 15 base damage
     - Longsword: 22 base damage
     - Greatsword: 35 base damage
   - **Piercing weapons**:
     - Dagger: 10 base damage
     - Spear: 18 base damage
     - Rapier: 20 base damage
   - **Bludgeoning weapons**:
     - Club: 12 base damage
     - Mace: 20 base damage
     - Warhammer: 30 base damage
   - **Ranged weapons**:
     - Short bow: 18 base damage
     - Longbow: 28 base damage
     - Crossbow: 32 base damage
   - Weapon struct:
     ```go
     type Weapon struct {
       WeaponID     uuid.UUID
       Name         string
       Type         WeaponType // slashing, piercing, bludgeoning, ranged
       BaseDamage   int
       Durability   int
       MaxDurability int
       SkillRequired int // Minimum skill to use effectively
     }
     ```

3. Build armor effectiveness vs weapon types:
   - Armor types:
     - **Leather**: Light armor
       - vs Slashing: 15% reduction
       - vs Piercing: 10% reduction
       - vs Bludgeoning: 5% reduction
     - **Chain mail**: Medium armor
       - vs Slashing: 35% reduction
       - vs Piercing: 20% reduction
       - vs Bludgeoning: 15% reduction
     - **Plate armor**: Heavy armor
       - vs Slashing: 50% reduction
       - vs Piercing: 40% reduction
       - vs Bludgeoning: 25% reduction
   - Armor degrades with hits:
     - Each hit reduces durability by 1
     - Damage reduction scales with durability: `baseReduction × (currentDurability / maxDurability)`
     - Broken armor (0 durability): 0% reduction

4. Create critical hit system (natural 95+ on d100):
   - Critical hit conditions:
     - Natural roll of 95-100 on d100
     - Additional chance from high Cunning: `+(Cunning / 50)%`
     - Heavy attacks: +5% critical chance
   - Critical hit effects:
     - Damage multiplier: ×2.0
     - Ignore 50% of armor
     - Apply bonus status effect chance (bleed, stun)
   - Critical failure (natural 1-5):
     - Miss completely (0 damage)
     - Lose extra stamina (×1.5 stamina cost)
     - Weapon durability -2 (fumble)

5. Test damage variance and balance:
   - Run 1000 attack simulations per weapon type
   - Track damage distributions:
     - Minimum damage (roll=1, no critical)
     - Average damage (roll=50)
     - Maximum damage (roll=100 or critical)
   - Verify balance:
     - Greatsword avg > Longsword avg > Short sword avg
     - Heavy weapons: high variance (high max, low min)
     - Light weapons: low variance (consistent damage)
     - Skill 100 deals ~50% more damage than skill 0
   - Validate critical hit rate ~5-10% depending on Cunning

6. Implement weapon durability degradation:
   - Durability reduces on:
     - Normal hit: -1 durability
     - Heavy attack: -2 durability
     - Critical failure (fumble): -3 durability
     - Blocked by armor: -1 durability
   - Durability affects performance:
     - <50% durability: -10% damage
     - <25% durability: -25% damage
     - 0% durability: weapon breaks, 0 damage
   - Repair mechanics:
     - NPCs/players can repair weapons at crafting stations
     - Repair cost: materials + time
     - Cannot exceed original MaxDurability

7. Add damage type effectiveness chart:
   - Track weapon type vs armor type matchups
   - Bludgeoning weapons bypass heavy armor better (only 25% reduction)
   - Slashing weapons effective vs light armor (only 15% reduction)
   - Piercing weapons balanced (moderate reduction across all armor types)
   - Store effectiveness multipliers in configuration:
     ```go
     var ArmorEffectiveness = map[WeaponType]map[ArmorType]float64{
       Slashing: {
         Leather: 0.15,
         ChainMail: 0.35,
         Plate: 0.50,
       },
       Piercing: {
         Leather: 0.10,
         ChainMail: 0.20,
         Plate: 0.40,
       },
       Bludgeoning: {
         Leather: 0.05,
         ChainMail: 0.15,
         Plate: 0.25,
       },
     }
     ```

Test Requirements (80%+ coverage):
- Damage calculation applies skill modifier correctly
- Attribute modifier scales with Might/Agility appropriately
- d100 roll generates values 1-100 uniformly
- Critical hits (95+) deal 2x damage
- Critical failures (1-5) deal 0 damage
- Weapon base damage values match specifications
- Armor reduction percentages applied correctly
- Plate armor reduces slashing damage by 50%
- Leather armor reduces bludgeoning damage by 5%
- Durability reduces by 1 per normal hit
- Durability reduces by 2 per heavy attack
- Durability reduces by 3 on critical failure
- Damage penalty applied when durability <50%
- Weapon breaks (0 damage) when durability = 0
- Skill 100 deals ~50% more damage than skill 0 (1000 simulations)
- Critical hit rate ~5% base + Cunning bonus
- Heavy weapons show higher damage variance than light weapons
- Bludgeoning weapons bypass heavy armor better than slashing
- Process 1000 damage calculations in < 50ms

Acceptance Criteria:
- Damage calculation combines weapon, skill, attribute, and roll correctly
- Weapon types have distinct base damage and characteristics
- Armor effectiveness varies by weapon type realistically
- Critical hit system adds excitement (2x damage, ignore partial armor)
- Weapon durability degrades with use and affects performance
- Damage balance validated through statistical testing
- All tests pass with 80%+ coverage

Dependencies:
- Phase 2.1 (Character System) - Might, Agility, Cunning attributes
- Phase 2.5 (Skills) - weapon skill values
- Phase 7.1 (Action Queue) - attack actions to resolve

Files to Create:
- `internal/combat/damage/calculator.go` - Damage calculation logic
- `internal/combat/damage/types.go` - Weapon, Armor structs
- `internal/combat/damage/critical.go` - Critical hit/failure system
- `internal/combat/damage/durability.go` - Weapon/armor durability
- `internal/combat/damage/effectiveness.go` - Armor effectiveness chart
- `internal/combat/damage/calculator_test.go` - Damage calculation tests
- `internal/combat/damage/critical_test.go` - Critical system tests
- `internal/combat/damage/durability_test.go` - Durability tests
- `internal/combat/damage/balance_test.go` - Statistical balance tests (1000+ simulations)

---

## Phase 7.3: Status Effects (1 week)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 7.3 Status Effects System:

Core Requirements:
1. Implement poison (DoT, stacks up to 5):
   - Poison effect:
     ```go
     type PoisonEffect struct {
       EffectID      uuid.UUID
       TargetID      uuid.UUID
       Stacks        int // 1-5
       DamagePerTick int // 3 damage per stack
       TickInterval  time.Duration // 5 seconds
       Duration      time.Duration // 30 seconds
       AppliedAt     time.Time
       LastTickAt    time.Time
     }
     ```
   - Poison application:
     - Chance to apply on hit: weapon-specific (daggers 20%, poisoned arrows 40%)
     - Each application adds 1 stack (max 5)
     - Stacks increase damage: `totalDamage = basePoison × stacks`
   - Poison tick:
     - Every 5 seconds, apply damage
     - Damage: `3 × stacks`
     - Duration: 30 seconds per stack
     - Example: 3 stacks = 9 damage per 5 seconds for 90 seconds
   - Cure methods:
     - Antidote item: remove all stacks
     - Time: duration expires naturally
     - Vitality check: 10% chance per tick to reduce 1 stack if Vitality >70

2. Add stun (can't act, extends duration):
   - Stun effect:
     ```go
     type StunEffect struct {
       EffectID   uuid.UUID
       TargetID   uuid.UUID
       Duration   time.Duration // Base 2 seconds
       AppliedAt  time.Time
       EndsAt     time.Time
     }
     ```
   - Stun application:
     - Heavy attacks: 15% chance to stun
     - Bludgeoning weapons: +10% stun chance
     - Critical hits: +20% stun chance
   - Stun mechanics:
     - Stunned combatant cannot queue actions
     - Current action interrupted and fails
     - Defending state cleared
   - Duration extension:
     - Each new stun application adds to existing duration
     - Max duration: 10 seconds total
     - Formula: `newEndsAt = max(existingEndsAt, now) + newDuration`

3. Build slow effect (increased reaction time):
   - Slow effect:
     ```go
     type SlowEffect struct {
       EffectID      uuid.UUID
       TargetID      uuid.UUID
       Multiplier    float64 // 1.5 = 50% slower
       Duration      time.Duration // 15 seconds
       AppliedAt     time.Time
     }
     ```
   - Slow application:
     - Ice/frost weapons: 25% chance
     - Leg attacks: 15% chance
     - Applies 1.5x reaction time multiplier
   - Effect on combat:
     - All action reaction times multiplied by 1.5
     - Example: Normal attack 1000ms → 1500ms when slowed
     - Stacks with base reaction time calculation
   - Duration: 15 seconds, doesn't stack (refresh instead)

4. Create bleed (DoT, reduces on movement):
   - Bleed effect:
     ```go
     type BleedEffect struct {
       EffectID       uuid.UUID
       TargetID       uuid.UUID
       DamagePerTick  int // 5 damage
       TickInterval   time.Duration // 3 seconds
       Duration       time.Duration // 20 seconds
       AppliedAt      time.Time
       MovementCounter int
     }
     ```
   - Bleed application:
     - Slashing weapons: 20% chance on critical hits
     - Applies 5 damage per 3 seconds
     - Duration: 20 seconds
   - Movement interaction:
     - Each movement action increments MovementCounter
     - Every 3 movements: reduce damage by 1
     - At 0 damage: bleed ends early
     - Encourages stationary combat to avoid worsening bleed
   - Cure methods:
     - Bandage item: stop immediately
     - Time: duration expires
     - Standing still: damage reduces over time

5. Add buff/debuff system (temporary stat modifications):
   - Generic buff/debuff struct:
     ```go
     type StatModifier struct {
       EffectID    uuid.UUID
       TargetID    uuid.UUID
       Stat        Stat // Might, Agility, Vitality, etc.
       Modifier    int  // +10, -15, etc.
       IsPercent   bool // True = percentage, false = flat
       Duration    time.Duration
       AppliedAt   time.Time
     }
     ```
   - Buff examples:
     - Battle Cry: +15 Might for 60 seconds
     - Bless: +10% max HP for 120 seconds
     - Haste: -30% reaction time for 30 seconds
   - Debuff examples:
     - Curse: -20 Vitality for 90 seconds
     - Weakness: -25% damage for 45 seconds
     - Exhaustion: +50% stamina cost for 60 seconds
   - Stat calculation:
     - Apply all modifiers when calculating effective stats
     - Order: base stat → flat modifiers → percentage modifiers
     - Example: Might 60, +10 flat, +20% = (60 + 10) × 1.2 = 84

6. Test effect stacking and interactions:
   - **Poison stacking**:
     - Apply poison 5 times
     - Verify stacks = 5 (max)
     - Verify damage = 15 per tick (3 × 5)
     - Apply 6th poison: verify stacks still 5
   - **Stun extension**:
     - Apply 2-second stun
     - Wait 1 second
     - Apply another 2-second stun
     - Verify total duration = 3 seconds (1 remaining + 2 new)
   - **Slow non-stacking**:
     - Apply 1.5x slow
     - Apply 2.0x slow
     - Verify only 2.0x active (newest replaces)
   - **Multiple effects**:
     - Apply poison + slow + bleed simultaneously
     - Verify all active and independent
     - Verify damage ticks occur correctly
   - **Buff/debuff cancellation**:
     - Apply +20 Might buff
     - Apply -15 Might debuff
     - Verify net effect: +5 Might

7. Implement effect expiration and cleanup:
   - Background job checks effects every second
   - For each active effect:
     - If `now >= expiresAt`: remove effect
     - If DoT effect: check if tick needed
     - If duration-based: calculate remaining time
   - Clean up expired effects from combatant state
   - Log effect applications and removals for combat log

Test Requirements (80%+ coverage):
- Poison applies and stacks up to 5 correctly
- Poison damage scales with stacks (3 × stacks per tick)
- Poison ticks occur every 5 seconds
- Poison duration extends with each stack
- Stun prevents action queueing
- Stun interrupts current action
- Stun duration extends when reapplied
- Stun max duration capped at 10 seconds
- Slow multiplies reaction time by 1.5x
- Slow doesn't stack, newest replaces oldest
- Bleed applies 5 damage per 3 seconds
- Bleed damage reduces with movement
- Bleed ends when damage reaches 0
- StatModifier applies flat and percentage correctly
- Multiple buffs/debuffs combine additively
- Effect expiration removes at correct time
- DoT effects tick at correct intervals
- Poison + bleed + slow can all be active simultaneously
- Process 100 active effects in < 50ms per tick

Acceptance Criteria:
- Poison stacks up to 5 and deals scaling DoT
- Stun prevents actions and extends duration on reapplication
- Slow increases reaction times by multiplier
- Bleed deals DoT that reduces with movement
- Buff/debuff system modifies stats temporarily
- All effects expire correctly at duration end
- All tests pass with 80%+ coverage

Dependencies:
- Phase 7.1 (Action Queue) - stun interrupts actions
- Phase 7.2 (Damage) - status effects applied on hits
- Phase 2.1 (Character System) - stat modifiers affect attributes
- Phase 1 (Time System) - effect durations and ticks

Files to Create:
- `internal/combat/effects/types.go` - All status effect structs
- `internal/combat/effects/poison.go` - Poison implementation
- `internal/combat/effects/stun.go` - Stun implementation
- `internal/combat/effects/slow.go` - Slow implementation
- `internal/combat/effects/bleed.go` - Bleed implementation
- `internal/combat/effects/modifiers.go` - Buff/debuff system
- `internal/combat/effects/manager.go` - Effect application/removal
- `internal/combat/effects/expiration.go` - Effect expiration job
- `internal/combat/effects/poison_test.go` - Poison tests
- `internal/combat/effects/stun_test.go` - Stun tests
- `internal/combat/effects/slow_test.go` - Slow tests
- `internal/combat/effects/bleed_test.go` - Bleed tests
- `internal/combat/effects/modifiers_test.go` - Buff/debuff tests
- `internal/combat/effects/interactions_test.go` - Effect stacking/interaction tests

---

# Phase 8: World Generation (6-8 weeks)

## Phase 8.1: LLM World Interview (2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 8.1 LLM World Interview for Custom World Creation:

Core Requirements:
1. Design interview questions covering key world parameters:
   - **Theme** (5-7 questions):
     - "What kind of world do you want to create?" (fantasy, sci-fi, post-apocalyptic, historical, surreal)
     - "What's the overall tone?" (grim/dark, hopeful, comedic, mysterious, epic)
     - "Any specific inspirations?" (Lord of the Rings, Dune, Mad Max, etc.)
     - "What makes this world unique?"
     - "What conflicts or tensions exist in this world?"
   - **Tech Level** (3-4 questions):
     - "What's the technological advancement level?" (stone age, medieval, renaissance, industrial, modern, futuristic, mixed)
     - "Is magic present?" (none, subtle, common, dominant)
     - "What's the most advanced technology available?"
     - "How does magic/technology affect daily life?"
   - **Geography** (4-5 questions):
     - "What's the planet size?" (moon-sized, Earth-sized, super-Earth, multiple planets)
     - "What's the climate range?" (frozen, temperate, tropical, desert, varied)
     - "Any unique geographical features?" (floating islands, underground cities, ocean worlds)
     - "How much land vs water?"
     - "Any extreme environments?"
   - **Culture** (5-6 questions):
     - "What sentient species exist?" (humans only, elves/dwarves, aliens, custom species)
     - "What's the political structure?" (kingdoms, democracy, tribal, corporate, anarchic)
     - "What are the main cultural values?" (honor, knowledge, wealth, freedom, tradition)
     - "What's the economic system?" (barter, currency, post-scarcity, feudal)
     - "What religions or belief systems exist?"
     - "What's considered taboo or forbidden?"

2. Build conversation flow with Ollama:
   - Conversational interview structure (not form-based):
     - LLM asks one question at a time
     - Player answers naturally
     - LLM acknowledges answer and asks follow-up
     - LLM can ask clarifying questions based on previous answers
   - Prompt template for interviewer:
     ```
     You are a world-building assistant helping a player create a custom game world.
     
     INTERVIEW PROGRESS:
     {questionsAnswered} / {totalQuestions} questions answered
     
     INFORMATION GATHERED SO FAR:
     {existingAnswers}
     
     CURRENT CATEGORY: {currentCategory}
     
     Ask the next question about {currentCategory}. Be conversational and enthusiastic.
     If the player's previous answer was vague, ask a clarifying follow-up.
     Keep questions concise (1-2 sentences).
     
     Previous player response: "{playerResponse}"
     ```
   - Track conversation state:
     ```go
     type WorldInterview struct {
       InterviewID      uuid.UUID
       PlayerID         uuid.UUID
       CurrentCategory  string
       QuestionsAsked   []string
       Answers          map[string]string
       LLMConversation  []Message
       Progress         float64 // 0.0 to 1.0
       Completed        bool
     }
     ```

3. Extract structured data from LLM responses:
   - After each answer, extract key parameters:
     - Parse theme keywords: "fantasy", "dark", "magic"
     - Parse tech level: "medieval", "no guns", "swords"
     - Parse geography: "mountainous", "cold", "islands"
     - Parse culture: "tribal", "honor-based", "multiple species"
   - Use secondary LLM call for extraction:
     ```
     Extract structured world parameters from this conversation:
     
     {fullConversationHistory}
     
     Return JSON:
     {
       "theme": "high fantasy",
       "tone": "epic and hopeful",
       "inspirations": ["Lord of the Rings", "Wheel of Time"],
       "uniqueAspect": "magic is dying out",
       "techLevel": "medieval",
       "magicLevel": "rare and fading",
       "planetSize": "Earth-sized",
       "climateRange": "varied, mostly temperate",
       "landWaterRatio": "60% land, 40% water",
       "sentientSpecies": ["humans", "elves", "dwarves"],
       "politicalStructure": "feudal kingdoms",
       "culturalValues": ["honor", "tradition", "courage"],
       "economicSystem": "barter and coin",
       "religions": "multiple pantheons",
       "conflicts": "kingdoms at war, magic fading"
     }
     ```
   - Validate extracted JSON schema
   - Store in WorldConfiguration

4. Create world configuration schema:
   - Comprehensive configuration struct:
     ```go
     type WorldConfiguration struct {
       ConfigID         uuid.UUID
       WorldID          uuid.UUID
       CreatedBy        uuid.UUID
       
       // Theme
       Theme            string
       Tone             string
       Inspirations     []string
       UniqueAspect     string
       MajorConflicts   []string
       
       // Tech Level
       TechLevel        string // "stone_age", "medieval", "industrial", etc.
       MagicLevel       string // "none", "rare", "common", "dominant"
       AdvancedTech     string
       MagicImpact      string
       
       // Geography
       PlanetSize       string
       ClimateRange     string
       LandWaterRatio   string
       UniqueFeatures   []string
       ExtremeEnvironments []string
       
       // Culture
       SentientSpecies  []string
       PoliticalStructure string
       CulturalValues   []string
       EconomicSystem   string
       Religions        []string
       Taboos           []string
       
       // Generation Parameters (derived)
       BiomeWeights     map[string]float64
       ResourceDistribution map[string]float64
       SpeciesStartAttributes map[string]AttributeSet
       
       CreatedAt        time.Time
     }
     ```

5. Validate and store world parameters:
   - Validation rules:
     - Required fields: Theme, TechLevel, PlanetSize, SentientSpecies
     - Tech level must match predefined options
     - Magic level compatible with tech level
     - At least one sentient species defined
     - Land/water ratio must sum to 100%
   - Store in PostgreSQL `world_configurations` table
   - Link to world: `worlds.configuration_id` foreign key

6. Implement interview resumption:
   - Save interview state after each question
   - If player disconnects, resume from last question
   - Show progress indicator: "Question 8 of 20 (40% complete)"
   - Allow editing previous answers:
     - "Actually, let me change my answer about the climate..."
     - LLM acknowledges and updates stored answer

7. Test conversation flow end-to-end:
   - Mock interview with predefined answers
   - Verify all categories covered
   - Verify follow-up questions asked when answers vague
   - Verify structured data extracted correctly
   - Verify WorldConfiguration populated with all fields

Test Requirements (80%+ coverage):
- WorldInterview tracks conversation state correctly
- Questions cover all categories (theme, tech, geography, culture)
- LLM generates conversational, contextual questions
- Player responses stored in Answers map
- Follow-up questions generated for vague answers
- Progress tracking shows 0.0 to 1.0 completion
- Structured data extraction parses conversation correctly
- JSON schema validation catches missing required fields
- WorldConfiguration stored in database with all fields
- Interview resumption loads previous state
- Interview completion marks Completed=true
- Process 20-question interview in < 5 minutes (with user interaction)

Acceptance Criteria:
- Interview covers theme, tech level, geography, and culture comprehensively
- LLM generates natural, conversational questions (not rigid forms)
- Player responses extracted into structured WorldConfiguration
- All required fields validated before world generation proceeds
- Interview state persists and can resume after disconnection
- All tests pass with 80%+ coverage

Dependencies:
- Phase 6.1 (Ollama Integration) - LLM for conversational interview
- PostgreSQL - store world_configurations table

Files to Create:
- `internal/worldgen/interview/types.go` - WorldInterview, WorldConfiguration structs
- `internal/worldgen/interview/service.go` - Interview orchestration
- `internal/worldgen/interview/questions.go` - Question templates by category
- `internal/worldgen/interview/extraction.go` - Structured data extraction from conversation
- `internal/worldgen/interview/validation.go` - Configuration validation
- `internal/worldgen/interview/state.go` - Interview state persistence
- `internal/worldgen/interview/repository.go` - Database storage
- `internal/worldgen/interview/service_test.go` - Interview flow tests
- `internal/worldgen/interview/extraction_test.go` - Data extraction tests
- `internal/worldgen/interview/validation_test.go` - Validation tests
- `migrations/postgres/XXX_world_configurations.sql` - Configuration table schema

---

## Phase 8.2: Geographic Generation (2-3 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 8.2 Geographic Generation with Tectonic Simulation:

Core Requirements:
1. Implement tectonic plate simulation:
   - Generate tectonic plates:
     - Number of plates based on planet size:
       - Moon-sized: 4-6 plates
       - Earth-sized: 8-12 plates
       - Super-Earth: 15-20 plates
     - Plate types:
       - **Continental**: thick, buoyant, forms land
       - **Oceanic**: thin, dense, forms ocean floor
     - Plate struct:
       ```go
       type TectonicPlate struct {
         PlateID      uuid.UUID
         Type         PlateType // Continental, Oceanic
         BoundaryPoints []geo.Point // Polygon vertices
         Centroid     geo.Point
         MovementVector geo.Vector // Direction and speed
         Thickness    float64 // km
         Age          float64 // million years
       }
       ```
   - Plate generation algorithm:
     - Randomly place plate centroids on sphere
     - Use Voronoi tessellation to create boundaries
     - Assign 30% continental, 70% oceanic (matches Earth)
     - Calculate movement vectors (diverge/converge/slide)

2. Generate heightmap from tectonic interactions:
   - Plate boundary types and effects:
     - **Divergent** (plates moving apart):
       - Oceanic-oceanic: mid-ocean ridge (+500m elevation)
       - Continental-continental: rift valley (-200m elevation)
     - **Convergent** (plates colliding):
       - Oceanic-oceanic: deep ocean trench (-8000m)
       - Oceanic-continental: coastal mountains (+4000m) + trench
       - Continental-continental: major mountain range (+6000m)
     - **Transform** (plates sliding):
       - Fault lines, minimal elevation change
       - Earthquake zones
   - Heightmap generation:
     - Start with base elevation: oceanic = -4000m, continental = 100m
     - For each boundary:
       - Calculate boundary type (divergent/convergent/transform)
       - Apply elevation modifier in 50km radius
       - Use gradient falloff: `elevation × (1 - distance/radius)²`
     - Add fractal noise for natural variation (Perlin noise)
     - Smooth with Gaussian blur (10km radius)

3. Create ocean/land distribution:
   - Sea level determination:
     - Based on land/water ratio from interview
     - 60% land: sea level at -200m elevation
     - 50% land (Earth): sea level at 0m
     - 40% land: sea level at +150m
   - Land mass formation:
     - All elevation > sea level = land
     - All elevation ≤ sea level = ocean
     - Coastal shelves: -200m to 0m (shallow ocean)
   - Validate against target ratio:
     - Calculate actual land percentage
     - Adjust sea level iteratively if >5% off target
     - Max 10 iterations, then accept closest

4. Add river generation (erosion pathfinding):
   - River source placement:
     - Spawn sources at high elevation (>2000m)
     - Density: 1 river per 100km² of mountainous terrain
     - Avoid ocean proximity (<50km from coast)
   - River pathfinding:
     - Use A* algorithm with elevation as cost
     - Path always flows downhill (cannot go uphill)
     - Path ends at ocean or lake (local minimum)
     - Width based on flow accumulation:
       - Headwaters: 5m width
       - Tributary: 20m width
       - Major river: 100m width
       - Delta: 500m width
   - Erosion effects:
     - Rivers carve valleys: -50m elevation in 20m radius
     - Sediment deposition at river mouth: +10m elevation
     - Create floodplains: flat terrain along river (±5m elevation)

5. Build biome assignment (elevation, latitude, moisture):
   - Biome determination factors:
     - **Elevation**:
       - <0m: Ocean
       - 0-200m: Lowland
       - 200-1000m: Highland
       - 1000-3000m: Mountain
       - >3000m: High mountain
     - **Latitude** (distance from equator):
       - 0-15°: Tropical
       - 15-35°: Subtropical
       - 35-55°: Temperate
       - 55-70°: Subarctic
       - >70°: Polar
     - **Moisture** (from Phase 8.3 weather simulation):
       - <200mm/year: Arid
       - 200-600mm: Semi-arid
       - 600-1200mm: Moderate
       - >1200mm: Humid
   - Biome lookup table:
     ```
     Tropical + Lowland + Humid = Rainforest
     Tropical + Lowland + Arid = Desert
     Temperate + Lowland + Moderate = Grassland
     Temperate + Highland + Humid = Deciduous Forest
     Subarctic + Lowland + Moderate = Taiga
     Polar + Any + Any = Tundra
     Mountain + Any + Any = Alpine
     ```
   - Biome struct:
     ```go
     type Biome struct {
       BiomeID      uuid.UUID
       Name         string
       Temperature  TemperatureRange
       Precipitation PrecipitationRange
       Vegetation   []string
       NativeSpecies []string
       Resources    []ResourceType
     }
     ```

6. Test geographic realism and variation:
   - Realism checks:
     - Mountain ranges follow plate boundaries
     - Rivers don't flow uphill
     - Oceans connected (no isolated ocean pockets)
     - Biomes transition smoothly (no desert next to tundra)
     - Land/water ratio within 5% of target
   - Variation tests:
     - Generate 10 worlds with same config
     - Verify each has different geography
     - Verify tectonic plates vary (different count, positions)
     - Verify river networks unique each time
   - Performance:
     - Generate Earth-sized world (510M km²) in < 5 minutes
     - Heightmap resolution: 1km per cell (510,000 cells)

7. Implement world shape support (spherical/bounded/infinite):
   - **Spherical planets**:
     - Use spherical coordinates (latitude, longitude, altitude)
     - Tectonic plates wrap around sphere
     - Rivers flow based on spherical surface
     - Biomes calculated with spherical distance
   - **Bounded cubes**:
     - Use Cartesian coordinates (X, Y, Z)
     - Tectonic simulation optional (can use pure heightmap)
     - Rivers flow within boundaries
     - Edges are hard boundaries (no wrapping)
   - **Infinite worlds**:
     - Use chunked generation (generate on-demand)
     - Seed-based Perlin noise for consistency
     - No tectonic simulation (pure procedural)
     - Biomes determined by noise functions

Test Requirements (80%+ coverage):
- Tectonic plate generation creates correct number for planet size
- Plate types assigned with 30/70 continental/oceanic ratio
- Voronoi tessellation creates valid plate boundaries
- Divergent boundaries create ridges/rifts at correct elevations
- Convergent boundaries create mountains/trenches
- Continental-continental collision creates highest mountains
- Heightmap elevations within realistic ranges (-11km to +9km)
- Sea level adjustment achieves target land/water ratio within 5%
- River sources placed at high elevation only
- River pathfinding always flows downhill
- River width scales with flow accumulation
- Rivers end at ocean or lake (no dead ends in land)
- Biome assignment considers elevation + latitude + moisture
- Tropical + humid = rainforest, tropical + arid = desert
- Biome transitions are smooth (no sudden jumps)
- Mountain ranges align with plate boundaries
- Generate 10 worlds: verify geographic variation
- Generate Earth-sized world in < 5 minutes
- Support spherical, bounded, and infinite world shapes

Acceptance Criteria:
- Tectonic simulation creates realistic plate interactions
- Heightmap shows mountains, valleys, plains, trenches
- Ocean/land distribution matches interview-specified ratio
- Rivers generated with realistic pathfinding and erosion
- Biomes assigned based on elevation, latitude, and moisture
- Geographic variation ensures unique worlds each generation
- All tests pass with 80%+ coverage

Dependencies:
- Phase 8.1 (World Interview) - WorldConfiguration with geography parameters
- Phase 0.4 (Spatial Foundation) - coordinate system and world shapes
- PostGIS - spatial queries for proximity/distance calculations
- Perlin noise library: `github.com/aquilax/go-perlin`

Files to Create:
- `internal/worldgen/geography/types.go` - TectonicPlate, Heightmap, Biome structs
- `internal/worldgen/geography/tectonics.go` - Tectonic plate generation and simulation
- `internal/worldgen/geography/heightmap.go` - Heightmap generation from tectonics
- `internal/worldgen/geography/ocean.go` - Ocean/land distribution
- `internal/worldgen/geography/rivers.go` - River generation and pathfinding
- `internal/worldgen/geography/biomes.go` - Biome assignment logic
- `internal/worldgen/geography/noise.go` - Perlin noise wrapper
- `internal/worldgen/geography/shapes.go` - World shape handling (spherical/bounded/infinite)
- `internal/worldgen/geography/tectonics_test.go` - Tectonic simulation tests
- `internal/worldgen/geography/heightmap_test.go` - Heightmap generation tests
- `internal/worldgen/geography/rivers_test.go` - River pathfinding tests
- `internal/worldgen/geography/biomes_test.go` - Biome assignment tests
- `internal/worldgen/geography/realism_test.go` - Geographic realism validation tests

---

## Phase 8.2b: Mineral Distribution (integrated into Phase 8.2)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 8.2b Mineral Distribution as part of Geographic Generation:
Core Requirements:

Implement mineral deposit generation during tectonic simulation:

Mineral formation tied to geological processes:

Igneous minerals (formed from molten rock):

Gold, silver, copper (hydrothermal deposits near volcanoes)
Diamonds (deep mantle, brought up by volcanic activity)
Basalt, granite (common igneous rock)


Sedimentary minerals (formed from layered deposits):

Coal (ancient plant matter in swamps/forests)
Limestone, sandstone (ocean floor deposits)
Salt deposits (dried ancient seas)

Metamorphic minerals (formed under pressure):

Marble (metamorphosed limestone in mountain ranges)
Gemstones (ruby, sapphire, emerald) in high-pressure zones
Iron ore (concentrated by tectonic pressure)

Mineral deposit struct:

go     type MineralDeposit struct {
       DepositID       uuid.UUID
       MineralType     MineralType // gold, iron, coal, diamond, etc.
       FormationType   FormationType // igneous, sedimentary, metamorphic
       Location        geo.Point // X, Y, Z (depth matters)
       Depth           float64 // Meters below surface (0 to -11000m)
       Quantity        int // Total extractable units
       Concentration   float64 // 0.0 to 1.0 (ore grade)
       VeinSize        VeinSize // small, medium, large, massive
       GeologicalAge   float64 // Million years old
       
       // Spatial extent
       VeinShape       VeinShape // spherical, linear, planar
       VeinOrientation geo.Vector // Direction of vein (for linear)
       VeinLength      float64 // Meters
       VeinWidth       float64 // Meters
       
       // Discovery
       SurfaceVisible  bool // Can be seen without mining
       RequiredDepth   float64 // Min mining depth to reach
     }

Generate mineral deposits based on tectonic plate boundaries:

Convergent boundaries (mountain formation):

High-pressure metamorphic minerals:

Iron ore veins: large deposits (5000-20000 units)
Gemstones (ruby, sapphire): rare, small deposits (50-200 units)
Marble: medium deposits (1000-5000 units)


Spawn along mountain ranges (elevation >2000m)
Depth: 500m to 3000m below surface
Vein shape: linear, following fault lines


Divergent boundaries (rifts, mid-ocean ridges):

Hydrothermal/igneous minerals:

Copper deposits: medium (1000-8000 units)
Gold veins: rare, small to medium (200-2000 units)
Silver: rare, small (100-1000 units)


Spawn at rift zones and volcanic areas
Depth: surface to 2000m depth
Often associated with hot springs/geothermal activity


Sedimentary basins (ancient oceans, swamps):

Coal deposits: very large (10000-50000 units)
Salt: massive deposits (50000-200000 units)
Limestone: abundant (20000-100000 units)
Spawn in low-elevation areas (<500m) with ancient water coverage
Depth: 100m to 1000m
Vein shape: planar (horizontal layers)


Subduction zones (oceanic plate under continental):

Rare, deep minerals:

Platinum: very rare (50-300 units)
Rare earth elements: rare (100-800 units)


Depth: 2000m to 5000m
Difficult to access, high value




Implement mineral vein formation with realistic geometry:

Vein generation algorithm:



go     func GenerateMineralVein(
       tectonicContext *TectonicContext,
       mineralType MineralType,
       epicenter geo.Point,
     ) *MineralDeposit {
       
       // 1. Determine vein size based on mineral type and geological conditions
       veinSize := DetermineVeinSize(mineralType, tectonicContext)
       
       // 2. Calculate quantity based on size and concentration
       concentration := CalculateConcentration(mineralType, tectonicContext)
       baseQuantity := GetBaseQuantity(mineralType, veinSize)
       quantity := int(baseQuantity * concentration)
       
       // 3. Determine vein shape based on formation type
       var veinShape VeinShape
       var veinOrientation geo.Vector
       var veinLength, veinWidth float64
       
       switch mineralType.FormationType {
       case Igneous:
         // Follows magma intrusion paths
         veinShape = Linear
         veinOrientation = tectonicContext.MagmaFlowDirection
         veinLength = 500 + rand.Float64() * 2000 // 500m to 2500m
         veinWidth = 10 + rand.Float64() * 50 // 10m to 60m
         
       case Sedimentary:
         // Horizontal layers
         veinShape = Planar
         veinOrientation = geo.Vector{X: 0, Y: 0, Z: 1} // Horizontal
         veinLength = 1000 + rand.Float64() * 5000 // 1km to 6km
         veinWidth = 500 + rand.Float64() * 2000 // 500m to 2.5km
         
       case Metamorphic:
         // Follows fault lines and stress patterns
         veinShape = Linear
         veinOrientation = tectonicContext.FaultLineDirection
         veinLength = 200 + rand.Float64() * 1500 // 200m to 1700m
         veinWidth = 5 + rand.Float64() * 30 // 5m to 35m
       }
       
       // 4. Determine depth based on formation conditions
       depth := CalculateDepositDepth(mineralType, tectonicContext, epicenter)
       
       // 5. Surface visibility (only if depth < 50m and in mountainous/eroded area)
       surfaceVisible := depth < 50 && tectonicContext.ErosionLevel > 0.7
       
       return &MineralDeposit{
         MineralType:     mineralType,
         FormationType:   mineralType.FormationType,
         Location:        epicenter,
         Depth:           depth,
         Quantity:        quantity,
         Concentration:   concentration,
         VeinSize:        veinSize,
         VeinShape:       veinShape,
         VeinOrientation: veinOrientation,
         VeinLength:      veinLength,
         VeinWidth:       veinWidth,
         SurfaceVisible:  surfaceVisible,
         RequiredDepth:   depth,
         GeologicalAge:   tectonicContext.Age,
       }
     }
```

4. Add mineral concentration and ore grade:
   - Concentration calculation:
```
     concentration = baseConcentration × geologicalModifier × randomVariation
     
     baseConcentration by mineral:
     - Common (iron, coal): 0.6 to 0.9 (60-90% ore grade)
     - Uncommon (copper, tin): 0.4 to 0.7 (40-70%)
     - Rare (gold, silver): 0.1 to 0.4 (10-40%)
     - Very rare (diamonds, platinum): 0.01 to 0.1 (1-10%)
     
     geologicalModifier:
     - Optimal formation conditions: 1.0 to 1.5
     - Average conditions: 0.7 to 1.0
     - Poor conditions: 0.3 to 0.7
     
     randomVariation: 0.8 to 1.2 (±20%)
```
   - Ore grade affects extraction:
     - High grade (>50%): easy extraction, high yield per ton
     - Medium grade (20-50%): standard extraction
     - Low grade (5-20%): requires processing, lower yield
     - Trace grade (<5%): not economically viable with primitive/medieval tech

5. Implement mineral discovery mechanics:
   - Surface deposits:
     - 5-10% of mineral deposits visible on surface
     - Found in eroded mountains, riverbeds, exposed cliffs
     - Easily discoverable, often depleted first
   - Subsurface deposits:
     - Require mining/excavation to discover
     - Detection methods:
       - **Visual**: Follow visible veins into ground
       - **Geological survey**: High Perception + Mining skill to detect signs
       - **Prospecting**: Systematic searching in geologically favorable areas
       - **Magic/Tech**: Advanced detection (varies by world tech level)
   - Discovery probability:
```
     discoveryChance = baseChance × geologicalKnowledge × prospectingEffort
     
     baseChance = surfaceVisibility (0.0 to 1.0)
     
     geologicalKnowledge = (MiningSkill + Perception) / 200
     
     prospectingEffort = timeSpent / optimalTime
     - Quick search (1 hour): 0.3x
     - Standard search (4 hours): 1.0x
     - Thorough search (8+ hours): 1.5x
```

6. Create mineral cluster formation:
   - Clustering rules:
     - Primary vein spawning at optimal geological location
     - 60% chance of secondary veins nearby (50-500m radius)
     - 30% chance of tertiary veins (500-2000m radius)
     - Same mineral type in cluster (follows geological consistency)
   - Vein richness distribution in cluster:
     - Primary vein: largest, highest concentration
     - Secondary veins: 50-80% size of primary
     - Tertiary veins: 20-50% size of primary
   - Example iron deposit cluster:
```
     Primary vein: 15000 units, 75% concentration, 1500m long
     Secondary vein 1: 10000 units, 60% concentration, 800m long
     Secondary vein 2: 8000 units, 55% concentration, 600m long
     Tertiary vein 1: 3000 units, 40% concentration, 300m long

Test mineral distribution realism:

Geological consistency:

Gold deposits near volcanic/rift zones (not in sedimentary plains)
Coal deposits in ancient forest/swamp regions (not in deserts)
Iron ore in metamorphic mountain ranges (convergent boundaries)
Salt deposits in dried ocean basins (not in high mountains)

Quantity validation:

Earth-like planet should have:

Iron: 50-100 major deposits (10000+ units each)
Coal: 30-60 major deposits (20000+ units each)
Copper: 40-80 medium deposits (5000+ units each)
Gold: 20-40 small deposits (500-2000 units each)
Diamonds: 5-15 tiny deposits (50-300 units each)

Accessibility distribution:

5-10% surface visible (easy early game)
40-50% shallow depth (<500m)
30-40% medium depth (500-2000m)
10-20% deep deposits (>2000m)


Generate 10 Earth-sized worlds:

Verify mineral counts within expected ranges
Verify geological placement consistency
Verify no deserts with coal (geological impossibility)

Add mineral depletion tracking:

Depletion mechanics:

Each unit extracted reduces deposit quantity
When quantity reaches 0, deposit is exhausted
Exhausted deposits remain as markers (historical sites)


Depletion struct:

go     type DepletionHistory struct {
       DepositID       uuid.UUID
       OriginalQuantity int
       CurrentQuantity  int
       ExtractedBy      []uuid.UUID // NPCs/Players who mined
       FirstExtracted   time.Time
       DepletedAt       *time.Time // null if not depleted
       ExtractionRate   float64 // Units per day
     }

No regeneration for minerals (geological timescale: millions of years)

Test Requirements (80%+ coverage):

Mineral deposits generated during tectonic simulation
Igneous minerals (gold, copper) spawn near volcanoes and rifts
Sedimentary minerals (coal, salt) spawn in ancient basins
Metamorphic minerals (iron, gems) spawn in mountain ranges
Convergent boundaries have iron ore and gemstone deposits
Divergent boundaries have copper and gold deposits
Sedimentary basins have coal and salt deposits
Mineral vein geometry follows geological context
Igneous veins: linear, following magma flow
Sedimentary veins: planar, horizontal layers
Metamorphic veins: linear, following fault lines
Concentration calculation includes base + geological + random factors
High-grade ore (>50%) spawns in optimal conditions
Low-grade ore (<20%) spawns in poor conditions
Surface deposits: 5-10% of total deposits visible
Discovery chance scales with Mining skill + Perception
Mineral clusters: 60% secondary veins, 30% tertiary veins
Secondary veins 50-80% size of primary
Gold deposits near volcanic zones (not sedimentary plains)
Coal deposits in ancient swamps (not deserts or mountains)
Iron ore in metamorphic mountains (convergent boundaries)
Earth-sized world: 50-100 iron deposits, 30-60 coal deposits
Accessibility: 5-10% surface, 40-50% shallow, 30-40% medium, 10-20% deep
Generate 10 worlds: verify mineral distribution consistency
Depletion tracking reduces quantity with each extraction
Exhausted deposits (quantity=0) marked as depleted
Generate 1000 mineral deposits in < 5 seconds

Acceptance Criteria:

Mineral deposits generated during tectonic/geographic phase (Phase 8.2)
Mineral types tied to geological formation processes (igneous/sedimentary/metamorphic)
Mineral distribution follows plate tectonics (convergent/divergent boundaries)
Vein geometry realistic based on formation type
Ore grade and concentration affect extraction difficulty
Surface deposits discoverable, subsurface requires mining
Mineral clusters form realistically around primary veins
Geological consistency validated (no coal in deserts, etc.)
Depletion tracking for finite mineral resources
All tests pass with 80%+ coverage

Dependencies:

Phase 8.2 (Geographic Generation) - tectonic plates, heightmap, biomes
Phase 8.3 (Weather Simulation) - ancient climate for coal formation
PostGIS - spatial queries for mineral proximity

Files to Create:

internal/worldgen/minerals/types.go - MineralDeposit, FormationType, VeinShape structs
internal/worldgen/minerals/formation.go - Mineral formation tied to tectonics
internal/worldgen/minerals/veins.go - Vein geometry generation
internal/worldgen/minerals/concentration.go - Ore grade calculations
internal/worldgen/minerals/clusters.go - Mineral clustering algorithm
internal/worldgen/minerals/discovery.go - Surface visibility and discovery mechanics
internal/worldgen/minerals/depletion.go - Depletion tracking
internal/worldgen/minerals/repository.go - Mineral deposit persistence
internal/worldgen/minerals/formation_test.go - Formation type tests
internal/worldgen/minerals/veins_test.go - Vein geometry tests
internal/worldgen/minerals/clusters_test.go - Clustering tests
internal/worldgen/minerals/consistency_test.go - Geological consistency validation
migrations/postgres/XXX_mineral_deposits.sql - Mineral deposits table

---

## Phase 8.3: Weather Simulation (1-2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 8.3 Weather Simulation with Atmospheric Dynamics:

Core Requirements:
1. Calculate evaporation rates (temperature + water proximity):
   - Evaporation formula:
     ```
     evaporation = baseRate × temperatureFactor × waterProximity × sunlight
     ```
   - Temperature factor:
     ```
     temperatureFactor = max(0, (temperature - 5°C) / 30°C)
     - Below 5°C: no evaporation (ice)
     - 5-35°C: linear increase
     - Above 35°C: capped at 1.0
     ```
   - Water proximity:
     ```
     waterProximity = 1.0 if ocean cell
     waterProximity = riverWidth / 100m if river cell
     waterProximity = 0.1 if adjacent to water
     waterProximity = 0.01 if land >10km from water
     ```
   - Sunlight (latitude-based):
     ```
     sunlight = cos(latitude) × seasonalModifier
     seasonalModifier ranges 0.7 (winter) to 1.3 (summer)
     ```
   - Base evaporation rate: 5mm/day in optimal conditions

2. Implement planetary wind patterns (Hadley cells):
   - Atmospheric circulation model:
     - **Hadley cells** (0-30° latitude):
       - Rising air at equator (low pressure)
       - Descending air at 30° (high pressure)
       - Trade winds: easterlies near equator
     - **Ferrel cells** (30-60° latitude):
       - Rising air at 60° (low pressure)
       - Descending air at 30° (high pressure)
       - Westerlies in mid-latitudes
     - **Polar cells** (60-90° latitude):
       - Descending air at poles (high pressure)
       - Rising air at 60° (low pressure)
       - Polar easterlies
   - Wind vector calculation:
     ```go
     func CalculateWind(latitude float64, longitude float64, season Season) geo.Vector {
       // Determine cell
       absLat := math.Abs(latitude)
       
       var windDirection float64
       var windSpeed float64
       
       if absLat < 30 {
         // Hadley cell: Trade winds (easterly)
         windDirection = -90 // Easterly (from east)
         windSpeed = 5 + (30 - absLat) / 6 // 5-10 m/s
       } else if absLat < 60 {
         // Ferrel cell: Westerlies
         windDirection = 90 // Westerly (from west)
         windSpeed = 8 + (absLat - 30) / 6 // 8-13 m/s
       } else {
         // Polar cell: Polar easterlies
         windDirection = -90 // Easterly
         windSpeed = 3 + (90 - absLat) / 10 // 3-6 m/s
       }
       
       // Add Coriolis effect deflection
       // Northern hemisphere: deflect right, Southern: deflect left
       coriolisDeflection := 15 * math.Copysign(1, latitude)
       windDirection += coriolisDeflection
       
       return geo.Vector{
         Direction: windDirection,
         Speed: windSpeed,
       }
     }
     ```

3. Build precipitation simulation (moisture + elevation):
   - Moisture transport:
     - Wind carries moisture from evaporation sources
     - Moisture content (humidity): 0-100%
     - Moisture accumulates as air passes over water
     - Moisture released as precipitation
   - Orographic precipitation (mountain effect):
     ```
     when air encounters elevation increase:
       - Air forced to rise
       - Temperature drops (adiabatic cooling): -6.5°C per 1000m
       - Cooler air holds less moisture
       - Excess moisture precipitates
       
     precipitationRate = moistureContent × elevationGradient × 0.1
     ```
   - Rain shadow effect:
     - Windward side (facing wind): heavy precipitation
     - Leeward side (behind mountain): dry (rain shadow)
     - Moisture depleted after crossing mountains
   - Precipitation calculation per cell:
     ```go
     func CalculatePrecipitation(
       cell *GeographyCell,
       upwindCells []*GeographyCell,
       wind geo.Vector,
     ) float64 {
       // Accumulate moisture from upwind
       moisture := 0.0
       for _, upwind := range upwindCells {
         if upwind.IsWater() {
           moisture += wind.Speed * 0.05 // 5% moisture per m/s over water
         }
       }
       
       // Orographic effect
       if cell.Elevation > upwindCells[0].Elevation {
         elevationGain := cell.Elevation - upwindCells[0].Elevation
         precipMm := moisture * elevationGain * 0.001 // 0.1% per meter gain
         return min(precipMm, moisture * 10) // Cap at 10mm per event
       }
       
       // Flat land precipitation
       if moisture > 50 { // 50% humidity threshold
         return (moisture - 50) * 0.5 // Light rain
       }
       
       return 0
     }
     ```

4. Add weather state (clear, cloudy, rain, snow, storm):
   - Weather states based on conditions:
     - **Clear**: moisture <30%, no precipitation
     - **Cloudy**: moisture 30-60%, no precipitation yet
     - **Rain**: temperature >0°C, precipitation >2mm/day
     - **Snow**: temperature ≤0°C, precipitation >2mm/day
     - **Storm**: precipitation >20mm/day OR wind >15m/s
   - Weather state struct:
     ```go
     type WeatherState struct {
       CellID        uuid.UUID
       Timestamp     time.Time
       State         WeatherType // clear, cloudy, rain, snow, storm
       Temperature   float64 // °C
       Precipitation float64 // mm/day
       Wind          geo.Vector
       Humidity      float64 // 0-100%
       Visibility    float64 // km
     }
     ```
   - Visibility affected by weather:
     - Clear: 50km visibility
     - Cloudy: 30km
     - Rain: 10km
     - Snow: 5km
     - Storm: 2km

5. Test weather consistency with geography:
   - Consistency checks:
     - Tropical regions (0-15° latitude): high precipitation (>2000mm/year)
     - Desert regions (20-30° latitude, leeward of mountains): low precipitation (<250mm/year)
     - Polar regions (>70° latitude): low precipitation (<400mm/year)
     - Coastal areas: higher precipitation than inland
     - Mountains: wet windward side, dry leeward side
   - Seasonal variation:
     - Summer: higher temperatures, more evaporation
     - Winter: lower temperatures, less evaporation
     - Temperature swing: 20-30°C difference between seasons at mid-latitudes
   - Validate against Earth patterns:
     - Equatorial rainforests: 2000-4000mm/year
     - Sahara desert: <100mm/year
     - Seattle (coastal, mid-latitude): 950mm/year
     - Antarctic interior: <200mm/year

6. Implement weather updates over game time:
   - Weather update frequency: every 6 hours (game time)
   - Update process:
     - Recalculate wind patterns (change gradually with seasons)
     - Update temperature (diurnal cycle + seasonal)
     - Calculate evaporation from water bodies
     - Transport moisture via wind
     - Calculate precipitation
     - Update weather state per cell
   - Persistence:
     - Store current weather state in `weather_states` table
     - Keep 30-day history for patterns/analysis
     - Aggregate annual precipitation for biome validation

7. Add extreme weather events:
   - **Hurricanes/Cyclones**:
     - Form over warm ocean (>26°C) in tropics
     - Require low pressure system + wind shear
     - Frequency: 10-15 per year per ocean basin
     - Effects: 50+ m/s winds, 200+ mm/day precipitation
   - **Blizzards**:
     - Cold temperatures (<-5°C) + high precipitation
     - Frequency: 5-10 per winter in subarctic/polar
     - Effects: <100m visibility, travel impossible
   - **Droughts**:
     - Precipitation <50% normal for 90+ consecutive days
     - More common in semi-arid regions
     - Effects: crop failure, water scarcity
   - **Heat waves**:
     - Temperature >10°C above normal for 7+ days
     - More common in summer at mid-latitudes
     - Effects: increased water consumption, health risks

Test Requirements (80%+ coverage):
- Evaporation rate scales with temperature correctly
- Evaporation higher over ocean than land
- Wind patterns follow Hadley/Ferrel/Polar cell model
- Trade winds (easterlies) at equator, westerlies at mid-latitudes
- Coriolis effect deflects wind correctly (right NH, left SH)
- Orographic precipitation occurs on windward slopes
- Rain shadow effect creates dry leeward areas
- Precipitation calculation considers moisture + elevation gradient
- Weather states assigned correctly (clear/cloudy/rain/snow/storm)
- Tropical regions receive >2000mm/year precipitation
- Desert regions receive <250mm/year precipitation
- Coastal areas wetter than inland
- Mountains create wet/dry divide
- Seasonal temperature variation 20-30°C at mid-latitudes
- Weather updates every 6 hours (game time)
- Hurricanes form over warm tropical oceans
- Blizzards occur in cold regions with precipitation
- Process global weather update (510M km²) in < 30 seconds

Acceptance Criteria:
- Evaporation rates calculated based on temperature and water proximity
- Planetary wind patterns follow realistic atmospheric circulation
- Precipitation simulated with orographic effects and rain shadows
- Weather states assigned appropriately per conditions
- Weather patterns consistent with geographic features (deserts, rainforests, etc.)
- Seasonal and diurnal temperature variations implemented
- All tests pass with 80%+ coverage

Dependencies:
- Phase 8.2 (Geographic Generation) - elevation, water bodies, biomes
- Phase 1 (Time System) - seasonal cycles, game time progression
- PostGIS - spatial queries for upwind cells

Files to Create:
- `internal/worldgen/weather/types.go` - WeatherState, Wind, Precipitation structs
- `internal/worldgen/weather/evaporation.go` - Evaporation calculations
- `internal/worldgen/weather/wind.go` - Atmospheric circulation model
- `internal/worldgen/weather/precipitation.go` - Precipitation simulation
- `internal/worldgen/weather/states.go` - Weather state assignment
- `internal/worldgen/weather/updates.go` - Weather update loop
- `internal/worldgen/weather/extremes.go` - Extreme weather events
- `internal/worldgen/weather/repository.go` - Weather state persistence
- `internal/worldgen/weather/evaporation_test.go` - Evaporation tests
- `internal/worldgen/weather/wind_test.go` - Wind pattern tests
- `internal/worldgen/weather/precipitation_test.go` - Precipitation tests
- `internal/worldgen/weather/consistency_test.go` - Geographic consistency validation
- `migrations/postgres/XXX_weather_states.sql` - Weather state table

---

## Phase 8.4: Flora/Fauna Evolution (1-2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 8.4 Flora/Fauna Evolution Simulation:

Core Requirements:
1. Design species schema (traits, diet, habitat, population):
   - Species struct:
     ```go
     type Species struct {
       SpeciesID    uuid.UUID
       Name         string
       Type         SpeciesType // flora, herbivore, carnivore, omnivore
       Generation   int // Evolutionary generation number
       
       // Physical Traits
       Size         float64 // kg for fauna, height in m for flora
       Speed        float64 // m/s for fauna, 0 for flora
       Armor        float64 // 0-100, defense rating
       Camouflage   float64 // 0-100, stealth rating
       
       // Dietary Needs
       Diet         DietType // herbivore, carnivore, omnivore, photosynthesis
       CaloriesPerDay int
       PreferredPrey []uuid.UUID // For carnivores
       PreferredPlants []uuid.UUID // For herbivores
       
       // Habitat
       PreferredBiomes []string
       TemperatureTolerance TemperatureRange
       MoistureTolerance    MoistureRange
       ElevationTolerance   ElevationRange
       
       // Reproduction
       ReproductionRate float64 // Offspring per year
       MaturityAge      int     // Years to adulthood
       Lifespan         int     // Years
       
       // Population
       Population       int
       PopulationDensity float64 // Per km²
       ExtinctionRisk   float64 // 0-1, higher = more at risk
       
       // Evolution
       MutationRate     float64
       FitnessScore     float64
       ParentSpeciesID  *uuid.UUID // null for initial species
     }
     ```

2. Implement survival mechanics (food chain, predation, competition):
   - **Food chain dynamics**:
     - **Primary producers (flora)**: Generate biomass from sunlight + water
       ```
       biomassProduction = baseRate × sunlight × moisture × temperature
       ```
     - **Herbivores**: Consume flora for energy
       ```
       energyGained = floraConsumed × 0.1 (10% energy transfer efficiency)
       ```
     - **Carnivores**: Consume herbivores
       ```
       energyGained = preyConsumed × 0.1
       ```
   - **Predation mechanics**:
     - Predator success rate:
       ```
       successRate = (predatorSpeed / preySpeed) × (1 - preyCamouflage/100) × predatorHunger
       ```
     - Prey consumption:
       ```
       preyKilled = predatorPopulation × successRate × huntsPerDay
       ```
   - **Competition**:
     - Intraspecific (same species):
       ```
       if population > carryingCapacity:
         survivalRate = carryingCapacity / population
       ```
     - Interspecific (different species, same niche):
       ```
       competitionPressure = overlap(niche1, niche2) × populationDensity
       fitnessReduction = competitionPressure × 0.1
       ```

3. Build evolutionary pressure system (environmental adaptation):
   - **Selection pressures**:
     - **Climate pressure**: Species outside temperature/moisture tolerance lose fitness
       ```
       climateFitness = 1.0 - (abs(actualTemp - optimalTemp) / toleranceRange)
       ```
     - **Predation pressure**: High predation favors speed, camouflage, armor
       ```
       if predationRate > 0.3:
         favor traits: +speed, +camouflage, +armor
       ```
     - **Food scarcity**: Favors efficient metabolism, wider diet
       ```
       if foodAvailability < 0.5:
         favor traits: -caloriesPerDay, +dietFlexibility
       ```
     - **Competition pressure**: Favors niche specialization
       ```
       if competition > 0.7:
         favor traits: narrower niche, unique adaptations
       ```
   - **Fitness calculation**:
     ```
     totalFitness = climateFitness × foodFitness × predationFitness × competitionFitness
     survivalProbability = totalFitness × (1 - extinctionRisk)
     ```

4. Time-accelerate evolution during world creation:
   - Evolution simulation loop:
     ```
     for generation := 0; generation < targetGenerations; generation++ {
       // 1. Calculate fitness for all species
       for each species {
         species.FitnessScore = CalculateFitness(species, environment)
       }
       
       // 2. Predation (population reduction)
       for each predator species {
         ConsumePreyPopulation(predator)
       }
       
       // 3. Natural selection (remove unfit)
       for each species {
         if species.FitnessScore < 0.3 {
           species.Population *= 0.5 // Die-off
         }
       }
       
       // 4. Reproduction (population growth)
       for each species {
         if species.Population > 0 {
           offspring = species.Population × species.ReproductionRate × species.FitnessScore
           species.Population += offspring
         }
       }
       
       // 5. Mutation (trait changes)
       for each species {
         if random() < species.MutationRate {
           MutateSpecies(species) // Create new variant
         }
       }
       
       // 6. Extinction (remove dead species)
       for each species {
         if species.Population < minViablePopulation {
           MarkExtinct(species)
         }
       }
       
       // 7. Time step: 100 years per generation
       advanceTime(100 years)
     }
     ```
   - Target generations: 1000-5000 (100k-500k years of evolution)
   - Time acceleration: simulate 1000 years per second

5. Generate initial species diversity:
   - Starting species (seeded based on biomes):
     - **Flora**:
       - Grass (temperate/tropical grasslands): fast reproduction, low calories
       - Trees (forests): slow reproduction, high calories
       - Shrubs (arid/semi-arid): medium reproduction, moderate calories
       - Aquatic plants (wetlands/coasts): fast reproduction, moderate calories
     - **Herbivores**:
       - Small herbivores (rabbits): fast, high reproduction, low food needs
       - Medium herbivores (deer): moderate speed, moderate reproduction
       - Large herbivores (bison): slow, low reproduction, high food needs
     - **Carnivores**:
       - Small carnivores (foxes): fast, moderate reproduction, hunt small prey
       - Medium carnivores (wolves): pack hunters, moderate speed
       - Large carnivores (bears): slow, low reproduction, hunt large prey
   - Spawn 5-10 species per major biome
   - Diversity target: 50-100 species per Earth-sized world

6. Test ecosystem stability:
   - Stability metrics:
     - **Population oscillations**: Predator-prey cycles should stabilize
       - Initial: wild swings (boom-bust)
       - After 1000 generations: smooth cycles
     - **Species diversity**: Should maintain 70-90% of initial species
       - Some extinction expected (20-30%)
       - New species from mutation (10-20%)
     - **Biomass balance**: Total biomass should stabilize
       - Flora: 80-90% of biomass
       - Herbivores: 8-15%
       - Carnivores: 2-5%
   - Run 5000-generation simulation
   - Verify no runaway extinction (>50% species lost)
   - Verify no runaway growth (population → infinity)

7. Implement speciation (new species from mutations):
   - Mutation triggers speciation when trait changes exceed threshold:
     ```
     if mutatedTrait > parentTrait × 1.5 OR mutatedTrait < parentTrait × 0.5 {
       CreateNewSpecies(mutatedTraits, parentSpeciesID)
     }
     ```
   - New species inherits:
     - 90% of parent traits
     - 10% mutated traits
     - Separate population counter
     - Generation = parent.Generation + 1
   - Speciation rate: ~5-10% of mutations create new species

Test Requirements (80%+ coverage):
- Species schema stores all traits, diet, habitat, population correctly
- Food chain energy transfer (10% efficiency) calculated correctly
- Predation success rate considers speed, camouflage, hunger
- Intraspecific competition reduces population when over capacity
- Interspecific competition reduces fitness for overlapping niches
- Climate fitness penalizes species outside tolerance
- Predation pressure favors defensive traits (speed, camouflage, armor)
- Food scarcity favors efficient metabolism
- Fitness calculation combines all pressures multiplicatively
- Evolution loop processes 1000 generations correctly
- Predation reduces prey populations
- Reproduction increases populations based on fitness
- Mutations create trait variations
- Extinction removes species below minimum viable population
- Initial species diversity: 50-100 species generated
- 5000-generation simulation reaches stability
- Species diversity maintained at 70-90% of initial
- Biomass balance: flora 80-90%, herbivores 8-15%, carnivores 2-5%
- Speciation creates new species from significant mutations
- Simulate 5000 generations in < 2 minutes

Acceptance Criteria:
- Species schema captures physical traits, diet, habitat, and population
- Survival mechanics simulate food chains, predation, and competition
- Evolutionary pressures drive trait adaptation (climate, predation, food, competition)
- Time-accelerated evolution simulates 100k-500k years of development
- Initial species diversity appropriate for world biomes (50-100 species)
- Ecosystem stability achieved after 1000+ generations (no runaway growth/extinction)
- Speciation creates new species from significant mutations
- All tests pass with 80%+ coverage

Dependencies:
- Phase 8.2 (Geographic Generation) - biomes for habitat placement
- Phase 8.3 (Weather Simulation) - climate for environmental pressures
- Phase 1 (Time System) - time acceleration for evolution

Files to Create:
- `internal/worldgen/evolution/types.go` - Species, Trait structs
- `internal/worldgen/evolution/species.go` - Species generation and management
- `internal/worldgen/evolution/food_chain.go` - Food chain dynamics
- `internal/worldgen/evolution/predation.go` - Predation mechanics
- `internal/worldgen/evolution/competition.go` - Competition calculations
- `internal/worldgen/evolution/fitness.go` - Fitness scoring
- `internal/worldgen/evolution/mutation.go` - Trait mutation and speciation
- `internal/worldgen/evolution/simulation.go` - Evolution simulation loop
- `internal/worldgen/evolution/stability.go` - Ecosystem stability metrics
- `internal/worldgen/evolution/repository.go` - Species persistence
- `internal/worldgen/evolution/species_test.go` - Species generation tests
- `internal/worldgen/evolution/food_chain_test.go` - Food chain tests
- `internal/worldgen/evolution/fitness_test.go` - Fitness calculation tests
- `internal/worldgen/evolution/simulation_test.go` - Evolution simulation tests
- `internal/worldgen/evolution/stability_test.go` - Ecosystem stability tests
- `migrations/postgres/XXX_species.sql` - Species table schema

# Phase 9: Crafting & Economy (3-4 weeks)

## Phase 9.1: Resource Distribution (1 week)
### Status: Completed
### Prompt:
Following TDD principles, implement Phase 9.1 Resource Distribution with Biome-Based Placement:

Core Requirements:
1. **Reference pre-generated mineral deposits from Phase 8.2b**:
   - Minerals (ore, coal, gems) are NOT generated in Phase 9.1
   - Phase 9.1 queries mineral deposits created during world generation (Phase 8.2b)
   - Create ResourceNode references pointing to MineralDeposit locations:
     ```go
     func CreateResourceNodesFromMinerals(worldID uuid.UUID) error {
       // Query mineral deposits from Phase 8.2b
       mineralDeposits := GetMineralDeposits(worldID)
       
       for _, deposit := range mineralDeposits {
         // Create ResourceNode reference (doesn't duplicate mineral data)
         resourceNode := &ResourceNode{
           NodeID:           uuid.New(),
           ResourceType:     MapMineralToResourceType(deposit.MineralType),
           Location:         deposit.Location,
           MineralDepositID: &deposit.DepositID, // Links to Phase 8.2b deposit
           Quantity:         deposit.Quantity,
           MaxQuantity:      deposit.Quantity,
           RegenRate:        0.0, // Minerals don't regenerate
           RequiredSkill:    "mining",
           MinSkillLevel:    DetermineSkillRequirement(deposit.Depth, deposit.Concentration),
           Discoverable:     deposit.SurfaceVisible,
           Depth:            deposit.Depth,
         }
         
         CreateResourceNode(resourceNode)
       }
       
       return nil
     }
     ```
   - Mineral ResourceNode properties inherited from MineralDeposit:
     - Quantity: from deposit.Quantity
     - Depth: from deposit.Depth (affects mining difficulty)
     - Concentration: from deposit.Concentration (affects yield)
     - VeinGeometry: from deposit vein shape for mining planning

2. **Implement procedural placement for renewable resources**:
   - Resource types by category (non-mineral):
     - **Vegetation**: Wood (oak, pine, birch), herbs, medicinal plants, fibers (cotton, hemp)
     - **Animal products**: Hide, meat, bones, wool, feathers (tied to Phase 8.4 species)
     - **Special**: Rare crystals, magical reagents, ancient artifacts
   - Resource struct:
     ```go
     type ResourceNode struct {
       NodeID            uuid.UUID
       Name              string
       Type              ResourceType // mineral, vegetation, animal, special
       Rarity            Rarity       // common, uncommon, rare, very_rare, legendary
       Location          geo.Point    // X, Y, Z coordinates
       Quantity          int          // Current available amount
       MaxQuantity       int          // Maximum capacity
       RegenRate         float64      // Units per day
       RegenCooldown     time.Duration // Time before regen starts after harvest
       LastHarvested     time.Time
       BiomeAffinity     []string     // Biomes where this resource appears
       RequiredSkill     string       // Skill needed to harvest
       MinSkillLevel     int          // Minimum skill level required
       
       // Mineral-specific (only for mineral type)
       MineralDepositID  *uuid.UUID   // Links to Phase 8.2b MineralDeposit
       Depth             float64      // Mining depth required
       
       // Animal-specific (only for animal type)
       SpeciesID         *uuid.UUID   // Links to Phase 8.4 Species
     }
     ```

3. Add biome-specific resource types (vegetation and animal only):
   - **Forest biomes**:
     - Wood (oak, birch, maple): very common, high density (10-20 per km²)
     - Herbs (medicinal plants): common, medium density (5-10 per km²)
     - Wild game (deer): uncommon, low density (1-3 per km²) - tied to Species population
   - **Grassland biomes**:
     - Wild grains: very common, high density (15-25 per km²)
     - Grazing animals (bison, sheep): common, medium density (5-10 per km²) - tied to Species
     - Fibers (cotton, hemp): common, medium density (8-12 per km²)
   - **Desert biomes**:
     - Cacti (water source): uncommon, low density (2-4 per km²)
     - Rare crystals (special): rare, low density (0.3-0.8 per km²)
   - **Ocean/Coastal biomes**:
     - Fish: very common, high density (20-30 per km²) - tied to aquatic Species
     - Pearls: rare, very low density (0.2-0.5 per km²)
     - Kelp/Seaweed: common, high density (10-15 per km²)
   - **Tundra biomes**:
     - Fur-bearing animals: uncommon, low density (2-4 per km²) - tied to Species
     - Ice crystals: uncommon, low density (1-3 per km²)
   - **Note**: Mountain biome ore deposits are queried from Phase 8.2b, not generated here

4. Create resource node schema (quantity, regeneration rate):
   - Resource node properties:
     - **Quantity tiers** (for vegetation/animal resources):
       - Small node: 10-50 units
       - Medium node: 51-200 units
       - Large node: 201-500 units
       - Rich source: 501-2000 units
     - **Regeneration rates by type**:
       - **Minerals**: 0 units/day (non-renewable, finite) - from Phase 8.2b
       - **Vegetation (fast-growing)**: 10-20 units/day
       - **Vegetation (slow-growing trees)**: 1-3 units/day
       - **Animal resources**: 5-10 units/day (animal reproduction)
       - **Special resources**: 0.1-1 unit/day (very slow, magical)
     - **Regeneration cooldown**:
       - After harvest, no regen for cooldown period
       - Fast resources: 1 day cooldown
       - Medium resources: 3 days cooldown
       - Slow resources: 7 days cooldown
       - Minerals: N/A (non-renewable, from Phase 8.2b)

5. Build harvesting mechanics (skill-based yield):
   - Harvest yield formula:
     ```
     baseYield = nodeQuantity × harvestEfficiency
     
     harvestEfficiency = skillModifier × toolModifier × randomFactor
     
     skillModifier = 0.5 + (gathererSkill / 200)
     - Skill 0: 50% efficiency
     - Skill 50: 75% efficiency
     - Skill 100: 100% efficiency
     
     toolModifier = toolQuality / 100
     - No tool: 0.5x (50%)
     - Basic tool: 0.7x (70%)
     - Good tool: 1.0x (100%)
     - Excellent tool: 1.3x (130%)
     
     randomFactor = 0.8 to 1.2 (±20% variance)
     
     finalYield = round(baseYield × harvestEfficiency)
     ```
   - **Mining-specific modifiers** (for mineral resources):
     - Depth penalty: `depthModifier = 1.0 - (depth / 10000m) × 0.5` (deeper = harder)
     - Concentration bonus: `concentrationModifier = oreGrade` (from Phase 8.2b)
     - Final mineral yield: `baseYield × harvestEfficiency × depthModifier × concentrationModifier`
   - Skill requirements:
     - Common resources: skill 0-20 required
     - Uncommon resources: skill 20-40 required
     - Rare resources: skill 40-60 required
     - Very rare resources: skill 60-80 required
     - Legendary resources: skill 80-100 required
   - Harvest failure:
     - If gathererSkill < requiredSkill - 20: 50% chance to fail completely
     - If gathererSkill < requiredSkill: 25% chance to fail
     - Failure yields 0 resources, still triggers cooldown

6. Test resource availability and balance:
   - Balance checks per biome:
     - **Resource density**: Total resources per km² should match biome tier
       - Rich biomes (forest, grassland): 30-50 resources/km² (vegetation + animal)
       - Moderate biomes (mountains, tundra): 15-30 resources/km² (includes mineral nodes from Phase 8.2b)
       - Poor biomes (desert, polar): 5-15 resources/km²
     - **Rarity distribution** (vegetation + animal resources only):
       - Common: 60-70% of resources
       - Uncommon: 20-25%
       - Rare: 8-12%
       - Very rare: 2-4%
       - Legendary: <1%
     - **Renewable vs non-renewable**:
       - Renewable (vegetation, animals): generated in Phase 9.1
       - Non-renewable (minerals): queried from Phase 8.2b
       - Target ratio: 70-80% renewable, 20-30% non-renewable
   - Generate 1000 km² test world
   - Validate resource counts match expected densities
   - Verify rarity distribution follows target percentages
   - Verify mineral nodes reference Phase 8.2b deposits (no duplication)

7. Implement resource respawn system (renewable resources only):
   - Background job: `RegenerateResources()` runs hourly (game time)
   - For each renewable resource node:
     ```go
     func RegenerateResource(node *ResourceNode, timePassed time.Duration) {
       // Skip non-renewable minerals
       if node.Type == MineralResource {
         return // Handled by Phase 8.2b depletion system
       }
       
       // Check if cooldown expired
       if time.Since(node.LastHarvested) < node.RegenCooldown {
         return // Still in cooldown
       }
       
       // Calculate regen amount
       hoursPasssed := timePassed.Hours()
       regenAmount := node.RegenRate * (hoursPasssed / 24.0) // Daily rate
       
       // Apply regen
       node.Quantity = min(node.Quantity + int(regenAmount), node.MaxQuantity)
       
       // Special case: Animal resources (tied to species population from Phase 8.4)
       if node.Type == AnimalResource {
         speciesPopulation := GetSpeciesPopulation(node.SpeciesID)
         node.Quantity = min(node.Quantity, speciesPopulation * 2) // Max 2 units per animal
       }
     }
     ```
   - Mineral resources:
     - Depletion tracked in Phase 8.2b MineralDeposit
     - Once depleted (Quantity = 0), remain at 0
     - Mark as `Depleted` status in MineralDeposit
     - No regeneration (geological timescale)

8. Add resource clusters (vegetation and special only):
   - Clustering algorithm:
     - Primary resource placement: random with biome affinity
     - Secondary clustering: 60% chance to place 2-5 additional nodes nearby
     - Cluster radius: 50-200m from primary node
     - **Note**: Mineral veins already clustered in Phase 8.2b
   - Rich deposit spawning (vegetation/special):
     - 5% of vegetation nodes are "rich sources"
     - Rich sources: 3x normal quantity, 2x regen rate
     - Always uncommon or higher rarity
     - Spawn near geographic features (rivers, ancient sites, magical locations)

Test Requirements (80%+ coverage):
- ResourceNode struct stores all properties correctly
- CreateResourceNodesFromMinerals queries Phase 8.2b deposits successfully
- Mineral ResourceNodes have MineralDepositID linking to Phase 8.2b
- Mineral ResourceNodes inherit quantity, depth, concentration from deposits
- No duplicate mineral generation (query only, don't create new deposits)
- Vegetation resources generated independently of minerals
- Biome-specific vegetation spawns in correct biomes only
- Forest biomes have high wood density (10-20 per km²)
- Grassland biomes have high grain density (15-25 per km²)
- Mountain biomes query mineral deposits from Phase 8.2b (5-8 iron nodes per km²)
- Desert biomes have low overall density (5-15 per km²)
- Rarity distribution (vegetation): common 60-70%, uncommon 20-25%, rare 8-12%, very rare 2-4%, legendary <1%
- Harvest yield calculation applies skill + tool + random modifiers correctly
- Mining harvest applies depth penalty and concentration bonus
- Skill 0 yields 50% efficiency, skill 100 yields 100% efficiency
- Tool modifier ranges from 0.5x (no tool) to 1.3x (excellent)
- Harvest failure occurs when skill < required - 20 (50% chance)
- Regeneration cooldown prevents immediate regen after harvest
- RegenerateResources adds correct amount based on regen rate
- RegenerateResources skips mineral resources (handled by Phase 8.2b)
- Animal resources capped by species population × 2 (from Phase 8.4)
- Vegetation clusters spawn 2-5 additional nodes within 50-200m
- Rich vegetation sources have 3x quantity and 2x regen rate
- Generate 1000 km² world: validate resource density per biome
- Verify 70-80% renewable, 20-30% non-renewable ratio
- Process 10,000 resource regenerations in < 1 second
- Mineral depletion updates Phase 8.2b MineralDeposit, not ResourceNode

Acceptance Criteria:
- Mineral ResourceNodes reference Phase 8.2b MineralDeposits (no duplication)
- Vegetation and animal resources distributed procedurally based on biome affinity
- Biome-specific resource types spawn with appropriate densities
- Resource nodes have quantity, regen rate, and cooldown properties
- Harvest yield scales with gatherer skill and tool quality
- Mining harvest considers depth and ore concentration
- Resource regeneration system replenishes renewable resources only
- Non-renewable minerals deplete via Phase 8.2b system
- Vegetation/special resource clusters create realistic distribution patterns
- All tests pass with 80%+ coverage

Dependencies:
- Phase 8.2b (Mineral Distribution) - mineral deposits for query/reference
- Phase 8.2 (Geographic Generation) - biomes for vegetation placement
- Phase 8.4 (Flora/Fauna Evolution) - species populations for animal resources
- Phase 2.5 (Skills) - gathering and mining skill levels
- PostGIS - spatial queries for resource proximity

Files to Create:
- `internal/economy/resources/types.go` - ResourceNode struct (with MineralDepositID field)
- `internal/economy/resources/minerals.go` - Query and reference Phase 8.2b deposits
- `internal/economy/resources/placement.go` - Procedural vegetation/animal placement
- `internal/economy/resources/biomes.go` - Biome-specific resource mappings (vegetation/animal)
- `internal/economy/resources/harvest.go` - Harvesting mechanics and yield calculation
- `internal/economy/resources/mining.go` - Mining-specific modifiers (depth, concentration)
- `internal/economy/resources/regeneration.go` - Resource regeneration system (renewable only)
- `internal/economy/resources/clusters.go` - Vegetation clustering algorithm
- `internal/economy/resources/repository.go` - Resource persistence
- `internal/economy/resources/minerals_test.go` - Mineral reference tests (no duplication)
- `internal/economy/resources/placement_test.go` - Vegetation placement algorithm tests
- `internal/economy/resources/harvest_test.go` - Harvest yield tests
- `internal/economy/resources/mining_test.go` - Mining-specific modifier tests
- `internal/economy/resources/regeneration_test.go` - Regeneration tests (skip minerals)
- `internal/economy/resources/balance_test.go` - Resource density validation tests
- `migrations/postgres/XXX_resources.sql` - Resources table schema

---

## Phase 9.2: Tech Trees & Recipes (2 weeks)
### Status: Completed
### Prompt:
Following TDD principles, implement Phase 9.2 Tech Trees and Crafting Recipes:

Core Requirements:
1. Design generic tech trees (primitive → medieval → industrial → modern → futuristic):
   - Tech tree structure:
     ```go
     type TechTree struct {
       TreeID       uuid.UUID
       WorldID      uuid.UUID
       Name         string
       TechLevel    TechLevel // primitive, medieval, industrial, modern, futuristic
       Nodes        []*TechNode
     }
     
     type TechNode struct {
       NodeID          uuid.UUID
       Name            string
       Description     string
       TechLevel       TechLevel
       Tier            int // 1-4 within each tech level
       Prerequisites   []*TechNode // Parent nodes required
       UnlocksRecipes  []uuid.UUID
       ResearchCost    map[string]int // Resource costs to unlock
       ResearchTime    time.Duration
       IconPath        string
     }
     ```
   - **Primitive tech tree** (Stone Age → Bronze Age):
     - Tier 1: Basic tools (stone axe, wooden spear, fire starting)
       - Stone Axe: no prerequisites, unlocks wood harvesting
       - Fire Starting: no prerequisites, unlocks cooking recipes
       - Flint Knapping: no prerequisites, unlocks stone tools
     - Tier 2: Basic crafting (pottery, weaving, leather working)
       - Pottery: requires Fire Starting, unlocks storage containers
       - Weaving: requires Plant Fiber Gathering, unlocks cloth
       - Leather Working: requires Hunting, unlocks leather armor
     - Tier 3: Agriculture (farming, animal domestication)
       - Basic Farming: requires Pottery, unlocks crop planting
       - Animal Domestication: requires Hunting, unlocks livestock
       - Food Preservation: requires Pottery + Fire, unlocks preserved food
     - Tier 4: Basic metallurgy (copper tools, bronze)
       - Copper Working: requires Fire + Mining, unlocks copper tools
       - Bronze Alloy: requires Copper Working + Tin, unlocks bronze items
       - Advanced Tools: requires Bronze Alloy, unlocks improved efficiency
   - **Medieval tech tree** (Iron Age → High Medieval):
     - Tier 1: Iron working (iron tools, weapons, armor)
       - Iron Smelting: requires Bronze Alloy + Advanced Fire, unlocks iron
       - Blacksmithing: requires Iron Smelting, unlocks iron weapons/tools
       - Advanced Forging: requires Blacksmithing, unlocks steel precursors
     - Tier 2: Advanced agriculture (windmills, irrigation)
       - Windmill Technology: requires Basic Farming + Engineering, unlocks grinding
       - Irrigation Systems: requires Basic Farming, unlocks crop yield boost
       - Crop Rotation: requires Advanced Farming, unlocks soil fertility
     - Tier 3: Architecture (stone buildings, fortifications)
       - Masonry: requires Stone Cutting, unlocks stone buildings
       - Fortification Design: requires Masonry, unlocks defensive structures
       - Advanced Architecture: requires Masonry, unlocks large buildings
     - Tier 4: Steel production (advanced weapons, plate armor)
       - Steel Production: requires Iron Smelting + Carbon, unlocks steel
       - Advanced Weaponsmithing: requires Steel + Blacksmithing, unlocks steel weapons
       - Plate Armor Crafting: requires Steel + Advanced Forging, unlocks plate armor
   - **Industrial tech tree** (Steam Age → Electrical Age):
     - Tier 1: Steam power (engines, factories)
       - Steam Engine: requires Steel + Thermodynamics, unlocks mechanical power
       - Factory System: requires Steam Engine, unlocks mass production
       - Railway Technology: requires Steam Engine + Steel, unlocks trains
     - Tier 2: Advanced chemistry (gunpowder, explosives)
       - Gunpowder: requires Sulfur + Charcoal + Saltpeter, unlocks firearms
       - Explosives: requires Gunpowder + Chemistry, unlocks dynamite
       - Advanced Chemistry: requires Explosives, unlocks pharmaceuticals
     - Tier 3: Mass production (assembly lines, standardization)
       - Interchangeable Parts: requires Factory System, unlocks efficiency
       - Assembly Line: requires Factory + Interchangeable Parts, unlocks mass goods
       - Quality Control: requires Assembly Line, unlocks consistent quality
     - Tier 4: Early electricity (generators, telegraphs)
       - Electrical Generation: requires Steam Engine + Magnets, unlocks electricity
       - Telegraph: requires Electricity + Wire, unlocks long-distance communication
       - Electric Lighting: requires Electricity, unlocks illumination
   - **Modern tech tree** (Information Age):
     - Tier 1: Electronics (computers, telecommunications)
       - Transistor: requires Electricity + Silicon, unlocks computing
       - Computer: requires Transistor, unlocks automation
       - Internet: requires Computer + Telecommunications, unlocks networking
     - Tier 2: Advanced materials (plastics, composites)
       - Plastic Production: requires Petroleum + Chemistry, unlocks synthetic materials
       - Composite Materials: requires Plastics + Engineering, unlocks lightweight structures
       - Advanced Alloys: requires Metallurgy + Chemistry, unlocks high-performance metals
     - Tier 3: Automation (robotics, AI assistants)
       - Robotics: requires Computer + Engineering, unlocks automated labor
       - AI Programming: requires Advanced Computers, unlocks intelligent systems
       - Automated Manufacturing: requires Robotics + AI, unlocks unmanned production
     - Tier 4: Renewable energy (solar, wind, nuclear)
       - Solar Power: requires Photovoltaics + Electricity, unlocks clean energy
       - Wind Turbines: requires Advanced Engineering, unlocks wind power
       - Nuclear Fission: requires Physics + Uranium, unlocks nuclear power
   - **Futuristic tech tree** (Post-Scarcity):
     - Tier 1: Nanotechnology (molecular assembly)
       - Molecular Assembly: requires Advanced Materials + Quantum, unlocks nanotech
       - Smart Materials: requires Nanotech, unlocks self-repairing items
       - Medical Nanobots: requires Nanotech + Medicine, unlocks healing tech
     - Tier 2: Energy weapons (lasers, plasma)
       - Laser Weapons: requires Optics + High Energy, unlocks directed energy
       - Plasma Technology: requires Fusion + Magnetics, unlocks plasma weapons
       - Force Fields: requires Energy Manipulation, unlocks barriers
     - Tier 3: Quantum computing (advanced AI, simulations)
       - Quantum Computing: requires Quantum Physics, unlocks supercomputing
       - Advanced AI: requires Quantum + Neural Networks, unlocks sentient AI
       - Virtual Reality: requires Quantum + Neural Interface, unlocks immersion
     - Tier 4: Exotic physics (anti-gravity, FTL travel)
       - Anti-Gravity: requires Quantum Gravity, unlocks flight/levitation
       - Warp Drive: requires Exotic Matter, unlocks FTL travel
       - Dimensional Tech: requires Quantum + Relativity, unlocks portals

2. Customize tech trees per world (based on world tech level):
   - World tech level configuration:
     - `maxTechLevel`: Highest tier available in world (from Phase 8.1 interview)
     - `startingTechLevel`: Where NPCs/players begin
     - `magicTechInteraction`: How magic affects technology
       - `"none"`: Magic and tech separate, both function normally
       - `"enhancing"`: Magic improves tech (magitech), +25% efficiency
       - `"suppressing"`: Magic disrupts tech (high fantasy), tech unreliable in high-magic areas
       - `"replacing"`: Magic replaces tech functions (arcane instead of machines)
   - Tech tree customization:
     ```go
     func CustomizeTechTree(worldConfig *WorldConfiguration) *TechTree {
       tree := &TechTree{
         WorldID:   worldConfig.WorldID,
         TechLevel: worldConfig.TechLevel,
       }
       
       // Add base nodes for starting tech level and all prerequisites
       switch worldConfig.TechLevel {
       case Primitive:
         tree.Nodes = LoadPrimitiveTech()
       case Medieval:
         tree.Nodes = LoadPrimitiveTech() // Include lower tiers
         tree.Nodes = append(tree.Nodes, LoadMedievalTech()...)
       case Industrial:
         tree.Nodes = LoadUpToIndustrial() // All lower tiers
       case Modern:
         tree.Nodes = LoadUpToModern()
       case Futuristic:
         tree.Nodes = LoadAllTech()
       }
       
       // Modify based on magic level
       switch worldConfig.MagicLevel {
       case "dominant":
         AddMagicAlternatives(tree) // Add magical equivalents
         RemoveTechDependencies(tree, []string{"electricity", "combustion"})
         ReplaceWithMagic(tree, []string{"steam_engine", "computer"})
       case "common":
         AddMagitechNodes(tree) // Hybrid tech/magic
       case "rare":
         // Tech-focused, magic as rare enhancement
       case "none":
         // Pure technology tree
       }
       
       // Apply world-specific modifiers
       if worldConfig.UniqueAspect == "dying magic" {
         MarkNodesAsUnstable(tree, MagicNodes)
       }
       
       return tree
     }
     ```
   - Magic-tech interaction examples:
     - **Enhancing**: "Mana-Powered Forge" (steel production +25% speed)
     - **Replacing**: "Teleportation Circle" replaces "Railway Technology"
     - **Suppressing**: Electronics fail in high-magic zones (>80% ambient magic)

3. Create recipe schema (inputs, outputs, required tools, skill level):
   - Recipe struct:
     ```go
     type Recipe struct {
       RecipeID      uuid.UUID
       Name          string
       Description   string
       Category      RecipeCategory // weapon, armor, tool, consumable, building, component
       TechNodeID    *uuid.UUID // null if no tech requirement (basic recipes)
       
       // Inputs
       Ingredients   []Ingredient
       RequiredTool  *ToolRequirement
       RequiredStation *CraftingStation // forge, anvil, alchemy_table, workbench
       
       // Outputs
       Output        ItemOutput
       ByProducts    []ItemOutput // Secondary products (e.g., leather scraps, slag)
       
       // Requirements
       RequiredSkill string // "smithing", "alchemy", "carpentry", etc.
       MinSkillLevel int
       CraftingTime  time.Duration
       SuccessRate   SuccessRateFormula
       
       // Quality
       QualityTiers  []QualityTier
       
       // Economy
       BaseValue     int // Base economic value
       Difficulty    Difficulty // trivial, easy, medium, hard, very_hard, masterwork
     }
     
     type Ingredient struct {
       ResourceID uuid.UUID
       Quantity   int
       Quality    *ItemQuality // null if any quality accepted
       Substitute []uuid.UUID // Alternative resources (oak OR birch wood)
     }
     
     type ToolRequirement struct {
       ToolType      string // "hammer", "saw", "chisel", "tongs"
       MinToolQuality ItemQuality
     }
     
     type CraftingStation struct {
       StationType   string // "forge", "anvil", "alchemy_table", "loom"
       MinStationTier int // 1-5 (basic to legendary station quality)
     }
     
     type ItemOutput struct {
       ItemID   uuid.UUID
       Quantity int
       Quality  ItemQuality // Determined by crafter skill
     }
     
     type SuccessRateFormula struct {
       BaseRate      float64 // 0.0 to 1.0
       SkillModifier float64 // Per skill point above minimum
       ToolModifier  float64 // Per tool quality level
       StationModifier float64 // Per station tier
     }
     
     type QualityTier struct {
       Name          string // "poor", "common", "good", "excellent", "masterwork"
       MinSkillLevel int
       StatModifier  float64 // Multiplier for item stats
       Probability   float64 // Base chance at this tier
       SkillInfluence float64 // How much skill increases this tier's probability
     }
     ```

4. Implement recipe discovery system:
   - Discovery methods:
     - **Automatic unlock**: When tech node researched
       ```go
       func UnlockTechNode(npcID uuid.UUID, nodeID uuid.UUID) error {
         node := GetTechNode(nodeID)
         
         // Validate prerequisites
         if !HasPrerequisites(npcID, node.Prerequisites) {
           return ErrPrerequisitesNotMet
         }
         
         // Consume research costs
         if !ConsumeResources(npcID, node.ResearchCost) {
           return ErrInsufficientResources
         }
         
         // Unlock node
         MarkTechUnlocked(npcID, nodeID)
         
         // Auto-discover all recipes unlocked by this node
         for _, recipeID := range node.UnlocksRecipes {
           DiscoverRecipe(npcID, recipeID, "research")
         }
         
         return nil
       }
       ```
     - **Experimentation**: Try combining ingredients at crafting station
       ```go
       func ExperimentWithIngredients(
         npcID uuid.UUID,
         ingredients []Ingredient,
         station *CraftingStation,
       ) (*Recipe, error) {
         // Calculate compatibility
         compatibility := CalculateIngredientCompatibility(ingredients)
         
         // Get crafter's skill
         crafterSkill := GetSkill(npcID, station.RelevantSkill)
         
         // Success chance
         successChance := compatibility × (crafterSkill / 100) × 0.3 // Max 30%
         
         if rand.Float64() < successChance {
           // Discover new recipe
           recipe := GenerateRecipeFromIngredients(ingredients, station)
           DiscoverRecipe(npcID, recipe.RecipeID, "experimentation")
           return recipe, nil
         }
         
         // Failure: consume ingredients, no output
         ConsumeIngredients(npcID, ingredients)
         return nil, ErrExperimentFailed
       }
       ```
     - **Teaching**: NPC with recipe teaches player/other NPC
       ```go
       func TeachRecipe(teacherID uuid.UUID, studentID uuid.UUID, recipeID uuid.UUID) error {
         // Validate teacher knows recipe
         if !KnowsRecipe(teacherID, recipeID) {
           return ErrTeacherDoesntKnowRecipe
         }
         
         // Check relationship (only teach friends/family)
         relationship := GetRelationship(teacherID, studentID)
         if relationship.Affection < 40 {
           return ErrRelationshipTooLow
         }
         
         // Teaching takes time (based on recipe difficulty)
         recipe := GetRecipe(recipeID)
         teachingTime := recipe.CraftingTime × 2
         
         // Success based on teacher's proficiency
         teacherProficiency := GetRecipeProficiency(teacherID, recipeID)
         successChance := 0.7 + (teacherProficiency / 200) // 70-100%
         
         if rand.Float64() < successChance {
           DiscoverRecipe(studentID, recipeID, "taught")
           ImproveRelationship(teacherID, studentID, 5) // Teaching improves bond
           return nil
         }
         
         return ErrTeachingFailed
       }
       ```
     - **Found**: Recipe scrolls/books as loot or treasure
       - Spawn recipe books in world (libraries, dungeons, treasure chests)
       - Reading recipe book: `DiscoverRecipe(readerID, recipeID, "found")`
   - Discovery tracking:
     ```go
     type RecipeKnowledge struct {
       EntityID    uuid.UUID // Player or NPC
       RecipeID    uuid.UUID
       Proficiency float64 // 0-100, improves with use
       TimesUsed   int
       Discovered  time.Time
       Source      string // "research", "experiment", "taught", "found"
       TeacherID   *uuid.UUID // If taught, who taught it
     }
     ```
   - Experimentation ingredient compatibility:
     - Same-category items: 0.6 (wood + wood)
     - Complementary items: 0.8 (metal + wood for handles)
     - Traditional combinations: 1.0 (flour + water)
     - Nonsensical combinations: 0.1 (rock + feather)

5. Test crafting progression and dependencies:
   - Dependency chain validation:
     - Cannot craft iron sword without "Blacksmithing" tech unlocked
     - Cannot craft steel armor without "Steel Production" tech unlocked
     - Cannot craft health potion without "Alchemy" tech unlocked
     - Cannot research Tier 2 tech without completing Tier 1 prerequisites
   - Progressive complexity validation:
     - **Early recipes** (Tier 1):
       - 1-2 ingredients, common resources
       - Example: Stone Axe (Stone + Stick)
       - Crafting time: 5-15 minutes
       - Skill requirement: 0-10
     - **Mid recipes** (Tier 2-3):
       - 3-5 ingredients, mix of common/uncommon
       - Example: Iron Sword (Iron Bar × 3 + Wood + Leather Grip)
       - Crafting time: 1-4 hours
       - Skill requirement: 30-50
     - **Late recipes** (Tier 4):
       - 5-10 ingredients, rare resources, complex dependencies
       - Example: Steel Plate Armor (Steel Bar × 15 + Leather × 8 + Cloth × 5 + Rivets × 50)
       - Crafting time: 8-24 hours
       - Skill requirement: 60-80
   - Example progression test:
     ```
     Primitive Tier 1:
       Stone + Stick → Stone Axe (no prereqs, 5 min, skill 0)
     
     Primitive Tier 4:
       Copper Ore + Coal → Copper Bar (requires "Copper Working", 30 min, skill 15)
       Copper Bar + Stick → Copper Axe (requires "Copper Working", 45 min, skill 20)
     
     Medieval Tier 1:
       Iron Ore + Coal → Iron Bar (requires "Iron Smelting", 1 hour, skill 30)
       Iron Bar × 3 + Stick → Iron Axe (requires "Blacksmithing", 2 hours, skill 35)
     
     Medieval Tier 4:
       Iron Bar + Carbon → Steel Bar (requires "Steel Production", 4 hours, skill 60)
       Steel Bar × 2 + Advanced Handle → Steel Sword (requires "Advanced Weaponsmithing", 6 hours, skill 70)
     ```

6. Build recipe repository with search/filter:
   - Repository methods:
     ```go
     type RecipeRepository interface {
       CreateRecipe(recipe *Recipe) error
       GetRecipe(recipeID uuid.UUID) (*Recipe, error)
       GetRecipesByCategory(category RecipeCategory) ([]*Recipe, error)
       GetRecipesByTechLevel(techLevel TechLevel) ([]*Recipe, error)
       GetRecipesByTechNode(nodeID uuid.UUID) ([]*Recipe, error)
       GetRecipesBySkill(skill string, maxLevel int) ([]*Recipe, error)
       GetCraftableRecipes(entityID uuid.UUID, availableResources map[uuid.UUID]int) ([]*Recipe, error)
       GetKnownRecipes(entityID uuid.UUID) ([]*Recipe, error)
       SearchRecipes(query string, filters RecipeFilters) ([]*Recipe, error)
       UpdateRecipe(recipe *Recipe) error
       DeleteRecipe(recipeID uuid.UUID) error
     }
     
     type RecipeFilters struct {
       Category      *RecipeCategory
       TechLevel     *TechLevel
       MaxSkillLevel *int
       MaxCraftTime  *time.Duration
       Difficulty    *Difficulty
       RequiredStation *string
       HasIngredients bool // Only show craftable with current inventory
     }
     ```
   - Advanced search capabilities:
     - Filter by output item type ("show me all weapon recipes")
     - Filter by available ingredients ("what can I make with iron and wood?")
     - Filter by tech level unlocked ("show medieval recipes I can craft")
     - Filter by skill requirements ("show recipes I have skill for")
     - Sort by crafting time, complexity, value, difficulty

7. Create sample recipes for each tech tier:
   - **Primitive tier recipes** (50+ recipes):
     - Tools:
       - Stone Axe: Stone (3) + Stick (1) → Stone Axe (skill: 0, time: 5 min)
       - Wooden Spear: Stick (2) + Flint (1) → Wooden Spear (skill: 5, time: 10 min)
       - Flint Knife: Flint (2) + Leather Strip (1) → Flint Knife (skill: 8, time: 15 min)
     - Survival:
       - Campfire: Wood (5) + Flint (1) → Campfire (skill: 0, time: 2 min)
       - Torch: Stick (1) + Resin (1) + Cloth (1) → Torch (skill: 3, time: 5 min)
       - Waterskin: Leather (3) → Waterskin (skill: 10, time: 20 min)
     - Armor:
       - Leather Armor: Hide (8) → Leather Armor (skill: 15, time: 30 min, requires: drying rack)
       - Bone Armor: Bone (12) + Leather (4) → Bone Armor (skill: 20, time: 1 hour)
     - Food:
       - Cooked Meat: Raw Meat (1) + Fire → Cooked Meat (skill: 0, time: 5 min)
       - Dried Meat: Raw Meat (3) + Salt (1) → Dried Meat (3) (skill: 10, time: 8 hours)
   - **Medieval tier recipes** (80+ recipes):
     - Weapons:
       - Iron Sword: Iron Bar (3) + Wood (1) → Iron Sword (skill: 30, time: 2 hours, requires: forge + anvil)
       - Steel Longsword: Steel Bar (4) + Leather Grip (1) → Steel Longsword (skill: 65, time: 6 hours, requires: forge + anvil)
       - Crossbow: Wood (8) + Iron Bar (2) + String (2) → Crossbow (skill: 50, time: 4 hours, requires: workbench)
     - Armor:
       - Chainmail: Iron Ring (200) → Chainmail (skill: 45, time: 10 hours, requires: forge)
       - Plate Armor: Steel Bar (20) + Leather (10) → Plate Armor (skill: 75, time: 20 hours, requires: forge + anvil)
     - Consumables:
       - Health Potion: Herbs (3) + Water (1) → Health Potion (skill: 25, time: 20 min, requires: alchemy table)
       - Mana Potion: Magic Herb (2) + Moonwater (1) → Mana Potion (skill: 40, time: 30 min, requires: alchemy table)
     - Food:
       - Bread: Flour (2) + Water (1) + Yeast (1) → Bread (3) (skill: 10, time: 1 hour, requires: oven)
       - Feast: Meat (5) + Vegetables (5) + Spices (2) → Feast (10 servings) (skill: 35, time: 3 hours, requires: kitchen)
   - **Industrial tier recipes** (60+ recipes):
     - Weapons:
       - Rifle: Steel Bar (5) + Wood (2) + Gunpowder (1) → Rifle (skill: 60, time: 8 hours, requires: machine shop)
       - Revolver: Steel Bar (3) + Wood Grip (1) + Gunpowder Mechanism (1) → Revolver (skill: 65, time: 6 hours, requires: machine shop)
     - Tools:
       - Steam Engine: Steel Bar (20) + Iron Bar (15) + Copper Tubing (10) → Steam Engine (skill: 70, time: 40 hours, requires: factory)
       - Mechanical Clock: Brass Gears (20) + Steel Spring (5) + Glass (1) → Clock (skill: 55, time: 12 hours, requires: workshop)
     - Explosives:
       - Dynamite: Nitroglycerin (2) + Stabilizer (1) → Dynamite (5) (skill: 70, time: 30 min, requires: chemistry lab)
       - Grenade: Steel Casing (1) + Gunpowder (2) + Fuse (1) → Grenade (skill: 60, time: 45 min, requires: workshop)
   - **Modern tier recipes** (40+ recipes):
     - Electronics:
       - Computer: Transistors (500) + Circuit Board (1) + Power Supply (1) → Computer (skill: 75, time: 20 hours, requires: electronics lab)
       - Smartphone: Microchips (10) + Screen (1) + Battery (1) → Smartphone (skill: 70, time: 8 hours, requires: clean room)
     - Advanced Materials:
       - Carbon Fiber: Carbon Nanotubes (100) + Resin (5) → Carbon Fiber Sheet (skill: 65, time: 4 hours, requires: materials lab)
       - Kevlar Vest: Kevlar Fabric (20) + Ceramic Plates (4) → Body Armor (skill: 60, time: 6 hours, requires: textile factory)
   - **Futuristic tier recipes** (30+ recipes):
     - Nanotech:
       - Nanite Swarm: Raw Nanites (1000) + Control Chip (1) + Power Cell (1) → Nanite Swarm (skill: 85, time: 12 hours, requires: nanoforge)
       - Self-Repair Module: Smart Materials (10) + Nanites (500) + AI Core (1) → Repair Module (skill: 80, time: 8 hours, requires: nanoforge)
     - Energy Weapons:
       - Laser Rifle: Focusing Crystal (1) + Power Cell (2) + Targeting Computer (1) → Laser Rifle (skill: 90, time: 16 hours, requires: weapons fab)
       - Plasma Sword: Plasma Generator (1) + Magnetic Containment (1) + Energy Cell (1) → Plasma Sword (skill: 95, time: 20 hours, requires: advanced forge)
   - **Total: 260+ recipes** across all tiers

Test Requirements (80%+ coverage):
- TechTree struct stores nodes with prerequisites correctly
- Tech node unlocking validates all prerequisites met before unlock
- Primitive tech tree has 15-20 basic nodes
- Medieval tech tree extends primitive with 20-30 additional nodes
- Industrial tech tree includes all lower tiers + 25-35 new nodes
- Modern tech tree has 20-30 nodes
- Futuristic tech tree has 15-25 nodes
- Tech tree customization applies world config correctly (max tech level)
- Magic-dominant worlds add magic alternatives or replace tech nodes
- Magic-suppressing worlds mark tech as unreliable in high-magic areas
- Magic-enhancing worlds boost tech efficiency by 25%
- Recipe struct stores all inputs, outputs, requirements correctly
- Ingredient requirements specify resource + quantity + optional substitutes
- Tool requirements specify type + min quality
- Station requirements specify type + min tier
- Recipe unlocking via tech research works correctly
- Recipe discovery via experimentation succeeds based on skill + compatibility
- Experimentation failure consumes ingredients without output
- Recipe teaching requires affection >40 and teacher knows recipe
- Recipe knowledge tracks proficiency (increases with use)
- Dependency chain validation prevents crafting without tech unlocked
- Cannot craft iron sword without "Blacksmithing" unlocked
- Cannot craft steel armor without "Steel Production" unlocked
- Crafting progression increases complexity (Tier 1: 1-2 ingredients, Tier 4: 5-10 ingredients)
- Early recipes have short craft times (5-15 min), late recipes long (8-24 hours)
- GetCraftableRecipes filters by available resources correctly
- GetKnownRecipes returns only discovered recipes
- SearchRecipes finds recipes by name/description/ingredients
- Recipe quality tiers: poor/common/good/excellent/masterwork
- Quality probability scales with crafter skill above minimum
- Generate 260+ recipes: validate all have valid inputs/outputs/requirements
- Verify all recipes reference valid resource IDs from Phase 9.1
- Verify all tech nodes have valid prerequisite chains (no circular dependencies)
- Process 1000 recipe lookups in < 100ms

Acceptance Criteria:
- Generic tech trees defined for 5 tech levels (primitive → futuristic)
- Each tech level has 4 tiers with progressive difficulty
- Tech trees customized per world based on Phase 8.1 interview configuration
- Magic-tech interaction properly modifies tech tree (enhancing/suppressing/replacing)
- Recipe schema captures inputs, outputs, tools, stations, skill requirements
- Recipe discovery system supports research, experimentation, teaching, finding
- Crafting progression shows logical dependencies (copper → iron → steel → advanced steel)
- Recipe repository provides comprehensive search and filter capabilities
- 260+ sample recipes created across all tech tiers
- All tests pass with 80%+ coverage

Dependencies:
- Phase 8.1 (World Interview) - world tech level and magic level configuration
- Phase 9.1 (Resource Distribution) - recipe ingredients reference resources
- Phase 2.5 (Skills) - crafting skill requirements
- Phase 3.3 (Relationships) - recipe teaching requires relationship check

Files to Create:
- `internal/economy/crafting/types.go` - TechTree, TechNode, Recipe, RecipeKnowledge structs
- `internal/economy/crafting/tech_trees.go` - Tech tree definitions for all 5 levels
- `internal/economy/crafting/customization.go` - World-specific tech tree customization
- `internal/economy/crafting/magic_interaction.go` - Magic-tech interaction logic
- `internal/economy/crafting/recipes.go` - Recipe definitions and samples
- `internal/economy/crafting/discovery.go` - Recipe discovery system (research/experiment/teach/find)
- `internal/economy/crafting/experimentation.go` - Ingredient compatibility and experimentation
- `internal/economy/crafting/teaching.go` - Recipe teaching mechanics
- `internal/economy/crafting/repository.go` - Recipe and tech tree repository
- `internal/economy/crafting/validation.go` - Dependency chain validation
- `internal/economy/crafting/search.go` - Recipe search and filtering
- `internal/economy/crafting/tech_trees_test.go` - Tech tree structure tests
- `internal/economy/crafting/customization_test.go` - Customization tests
- `internal/economy/crafting/recipes_test.go` - Recipe validation tests
- `internal/economy/crafting/discovery_test.go` - Discovery system tests
- `internal/economy/crafting/progression_test.go` - Crafting progression validation tests
- `migrations/postgres/XXX_tech_trees.sql` - Tech trees and nodes tables
- `migrations/postgres/XXX_recipes.sql` - Recipes table
- `migrations/postgres/XXX_recipe_knowledge.sql` - Recipe knowledge tracking table
- `data/tech_trees/primitive.json` - Primitive tech tree definition
- `data/tech_trees/medieval.json` - Medieval tech tree definition
- `data/tech_trees/industrial.json` - Industrial tech tree definition
- `data/tech_trees/modern.json` - Modern tech tree definition
- `data/tech_trees/futuristic.json` - Futuristic tech tree definition
- `data/recipes/primitive.json` - 50+ primitive tier recipe definitions
- `data/recipes/medieval.json` - 80+ medieval tier recipe definitions
- `data/recipes/industrial.json` - 60+ industrial tier recipe definitions
- `data/recipes/modern.json` - 40+ modern tier recipe definitions
- `data/recipes/futuristic.json` - 30+ futuristic tier recipe definitions

---

## Phase 9.3: NPC Economy (1-2 weeks)
### Status: ✅ Completed
### Prompt:
Following TDD principles, implement Phase 9.3 NPC-Driven Economy:

Core Requirements:
1. Implement NPC resource gathering (autonomous harvesting):
   - NPC gathering behavior:
     - Driven by desires: `resourceAcquisition` need from Phase 5.1
     - Occupation-based: farmers gather crops, miners gather ore, woodcutters gather wood
     - Skill-based: NPCs gather resources matching their gathering skills
   - Autonomous gathering loop:
     ```go
     func (npc *NPC) GatherResources(ctx context.Context) error {
       // 1. Determine gathering need
       if npc.Desires.ResourceAcquisition < 50 {
         return nil // Not urgent enough
       }
       
       // 2. Find nearby harvestable resources
       resources := FindNearbyResourceNodes(npc.Location, radius=500m)
       
       // 3. Filter by skill capability and occupation preference
       capableResources := FilterBySkill(resources, npc.Skills)
       preferredResources := FilterByOccupation(capableResources, npc.Occupation)
       
       // 4. Prioritize by value, need, and distance
       target := PrioritizeResource(
         preferredResources,
         npc.Inventory,
         npc.Occupation,
         npc.Wealth,
       )
       
       if target == nil {
         return ErrNoSuitableResources
       }
       
       // 5. Path to resource
       if err := PathToResource(npc, target); err != nil {
         return err
       }
       
       // 6. Harvest (uses Phase 9.1 harvest mechanics)
       yield, err := HarvestResource(target, npc.Skills, npc.Equipment)
       if err != nil {
         return err
       }
       
       // 7. Add to inventory
       if err := npc.Inventory.Add(yield); err != nil {
         // Inventory full, return home to unload
         return ErrInventoryFull
       }
       
       // 8. Decrease resource acquisition need
       npc.Desires.ResourceAcquisition -= 20
       
       // 9. Create memory of gathering
       CreateGatheringMemory(npc.ID, target, yield)
       
       return nil
     }
     ```
   - Occupation-based priorities:
     ```go
     type Occupation struct {
       Name               string
       PrimaryResources   []ResourceType // Main focus
       SecondaryResources []ResourceType // Opportunistic gathering
       PreferredSkills    []string
       GatheringRadius    float64 // How far they travel to gather
     }
     
     var Occupations = map[string]Occupation{
       "farmer": {
         PrimaryResources:   []ResourceType{Grain, Vegetables, Fruit},
         SecondaryResources: []ResourceType{Herbs, Fiber},
         PreferredSkills:    []string{"farming", "herbalism"},
         GatheringRadius:    200, // Stay near farm
       },
       "miner": {
         PrimaryResources:   []ResourceType{IronOre, CopperOre, Coal, Gold},
         SecondaryResources: []ResourceType{Gemstones, Stone},
         PreferredSkills:    []string{"mining"},
         GatheringRadius:    1000, // Travel to mines
       },
       "woodcutter": {
         PrimaryResources:   []ResourceType{Wood, Bark},
         SecondaryResources: []ResourceType{Resin, Herbs},
         PreferredSkills:    []string{"logging"},
         GatheringRadius:    500, // Work in forest
       },
       "hunter": {
         PrimaryResources:   []ResourceType{Meat, Hide, Bones},
         SecondaryResources: []ResourceType{Feathers, Horns},
         PreferredSkills:    []string{"hunting", "tracking"},
         GatheringRadius:    2000, // Track far for game
       },
       "herbalist": {
         PrimaryResources:   []ResourceType{MedicinalHerbs, Flowers, Roots},
         SecondaryResources: []ResourceType{Mushrooms, Berries},
         PreferredSkills:    []string{"herbalism", "alchemy"},
         GatheringRadius:    800, // Forage widely
       },
     }
     ```
   - Gathering efficiency by occupation:
     - Primary resources: 100% normal speed
     - Secondary resources: 70% speed (opportunistic)
     - Non-related resources: 40% speed (not specialized)

2. Build NPC crafting (produce goods for trade):
   - NPC crafting behavior:
     - Driven by occupation and skill proficiency
     - Crafts items they know recipes for (via Phase 9.2 discovery)
     - Prioritizes valuable items with high local demand
   - Autonomous crafting loop:
     ```go
     func (npc *NPC) CraftGoods(ctx context.Context) error {
       // 1. Check if crafting desire active
       if npc.Desires.TaskCompletion < 40 {
         return nil // Not motivated to craft
       }
       
       // 2. Get known recipes
       knownRecipes, err := GetRecipeKnowledge(npc.ID)
       if err != nil {
         return err
       }
       
       // 3. Filter craftable (have ingredients + tools + station access)
       craftable := FilterCraftableRecipes(
         knownRecipes,
         npc.Inventory,
         npc.Location, // Check for nearby crafting stations
       )
       
       if len(craftable) == 0 {
         // Need to gather more resources
         return ErrNoIngredients
       }
       
       // 4. Prioritize by profit margin and demand
       localMarket := GetLocalMarketData(npc.Location)
       bestRecipe := SelectMostProfitableRecipe(craftable, localMarket, npc.Skills)
       
       // 5. Move to appropriate crafting station
       station := FindNearestCraftingStation(npc.Location, bestRecipe.RequiredStation)
       if station == nil {
         return ErrNoCraftingStation
       }
       
       if err := PathToStation(npc, station); err != nil {
         return err
       }
       
       // 6. Craft item (uses Phase 9.2 crafting mechanics)
       result, err := CraftItem(
         bestRecipe,
         npc.Skills[bestRecipe.RequiredSkill],
         npc.Equipment,
         station,
       )
       
       if err != nil {
         // Crafting failed, ingredients consumed
         CreateCraftingMemory(npc.ID, bestRecipe, false, nil)
         return err
       }
       
       // 7. Add crafted item to inventory
       npc.Inventory.Add(result.Item)
       
       // 8. Handle byproducts
       for _, byproduct := range result.ByProducts {
         npc.Inventory.Add(byproduct)
       }
       
       // 9. Improve recipe proficiency
       ImproveRecipeProficiency(npc.ID, bestRecipe.ID, 1)
       
       // 10. Decrease task completion desire
       npc.Desires.TaskCompletion -= 30
       
       // 11. Create crafting memory
       CreateCraftingMemory(npc.ID, bestRecipe, true, result.Item)
       
       return nil
     }
     ```
   - Crafting specialization over time:
     - NPCs track recipe proficiency (0-100)
     - Higher proficiency: faster crafting, higher quality, lower failure rate
     - Proficiency gain formula:
       ```
       proficiencyGain = baseGain × difficultyModifier × successModifier
       
       baseGain = 1.0
       difficultyModifier = recipeDifficulty / 100 (harder recipes teach more)
       successModifier = 2.0 if success, 0.5 if failure
       ```
     - After 50+ uses, NPC becomes "specialist" in that recipe category
     - Specialists: +15% crafting speed, +10% quality tier chance

3. Add NPC merchant inventory (based on local economy):
   - Merchant inventory dynamics:
     - **Stock sources**:
       - Crafted by merchant (if has crafting skills)
       - Purchased from other NPCs/players (wholesale)
       - Looted/found during travels
       - Commissioned from artisan NPCs
     - **Stock limits**:
       - Based on merchant wealth and storage capacity
       - Wealth tiers:
         - Poor merchant: 10-20 unique items, 500-2000 gold
         - Average merchant: 20-50 unique items, 2000-10000 gold
         - Wealthy merchant: 50-150 unique items, 10000-50000 gold
         - Trading house: 150-500 unique items, 50000+ gold
     - **Inventory updates**:
       - Restock daily (game time)
       - Sell popular items, discontinue slow movers
       - Adjust quantities based on 7-day sales history
   - Merchant struct:
     ```go
     type Merchant struct {
       NPCID          uuid.UUID
       ShopName       string
       Specialization MerchantType // blacksmith, general_goods, alchemist, jeweler, tavern
       Location       geo.Point // Shop location
       Inventory      *Inventory
       Wealth         int // Current gold
       PriceModifier  float64 // 0.8 to 1.5 (80% to 150% of base value)
       
       // Reputation
       Reputation     float64 // 0-100, affects prices and customer willingness
       CustomerCount  int // Total unique customers served
       
       // Trade stats
       SalesHistory     []SaleRecord // Last 30 days
       PurchaseHistory  []PurchaseRecord // Items bought from others
       
       // Business hours
       OpeningHour    int // 6 = 6am
       ClosingHour    int // 20 = 8pm
       DaysOpen       []DayOfWeek // Which days merchant operates
       
       // Relationships
       Suppliers      []uuid.UUID // NPCs who supply goods
       Competitors    []uuid.UUID // Other merchants in area
     }
     
     type SaleRecord struct {
       Timestamp    time.Time
       ItemID       uuid.UUID
       Quantity     int
       PricePer     int
       CustomerID   uuid.UUID
       Profit       int
     }
     
     type PurchaseRecord struct {
       Timestamp    time.Time
       ItemID       uuid.UUID
       Quantity     int
       PricePer     int
       SupplierID   uuid.UUID
     }
     ```
   - Restocking algorithm:
     ```go
     func (merchant *Merchant) Restock(ctx context.Context) error {
       // 1. Analyze sales from past 7 days
       salesData := AnalyzeSalesHistory(merchant.SalesHistory, days=7)
       
       popularItems := salesData.TopSellers(limit=10)
       slowMovers := salesData.SlowMovers(soldLessThan=3)
       outOfStock := salesData.OutOfStock()
       
       // 2. Restock popular items
       for _, item := range popularItems {
         targetStock := CalculateTargetStock(item, salesData)
         currentStock := merchant.Inventory.Count(item.ItemID)
         needed := targetStock - currentStock
         
         if needed > 0 {
           if merchant.CanCraft(item) {
             // Craft it themselves
             merchant.CraftGoods() // Uses crafting loop above
           } else {
             // Purchase from supplier or wholesale
             merchant.PurchaseFromSupplier(item, needed)
           }
         }
       }
       
       // 3. Restock out-of-stock items
       for _, item := range outOfStock {
         if item.SalesLast7Days > 0 {
           // Was selling, need to restock
           merchant.PurchaseFromSupplier(item, 5) // Buy small batch
         }
       }
       
       // 4. Remove slow-moving items (free up capital)
       for _, item := range slowMovers {
         quantity := merchant.Inventory.Count(item.ItemID)
         if quantity > 0 {
           // Sell to wholesale at 60% value
           merchant.SellToWholesale(item, quantity)
         }
       }
       
       // 5. Acquire new variety if wealth allows
       if merchant.Wealth > merchant.RestockThreshold() {
         newItems := DiscoverNewProducts(merchant.Specialization, merchant.Location)
         for _, newItem := range newItems {
           merchant.PurchaseFromSupplier(newItem, 3) // Buy trial quantity
         }
       }
       
       // 6. Adjust prices based on stock levels
       merchant.AdjustPrices()
       
       return nil
     }
     
     func (merchant *Merchant) CalculateTargetStock(item *Item, salesData *SalesAnalysis) int {
       avgDailySales := salesData.AverageDailySales(item.ItemID)
       
       // Stock for 7 days of average sales
       targetStock := int(avgDailySales * 7)
       
       // Min 3, max based on storage capacity
       return max(3, min(targetStock, merchant.MaxStockPerItem()))
     }
     ```

4. Create dynamic pricing (supply/demand):
   - Price calculation:
     ```
     finalPrice = baseValue × supplyDemandModifier × merchantModifier × qualityModifier × relationshipModifier
     
     supplyDemandModifier = sqrt(localDemand / localSupply)
     - High demand, low supply: 2.5x (250%)
     - Balanced (demand ≈ supply): 1.0x (100%)
     - Low demand, high supply: 0.4x (40%)
     - Sqrt dampens extreme swings
     
     merchantModifier = merchant.PriceModifier × reputationFactor
     - merchant.PriceModifier: 0.8 (cheap) to 1.5 (expensive)
     - reputationFactor = 0.9 + (reputation / 1000)
       - Low reputation (20): 0.92x (discount to attract customers)
       - High reputation (80): 0.98x (slight premium for quality)
     
     qualityModifier = itemQuality.StatModifier
     - Poor: 0.5x
     - Common: 1.0x
     - Good: 1.5x
     - Excellent: 2.5x
     - Masterwork: 5.0x
     
     relationshipModifier (for selling to player/NPC):
     - Friend (affection >60): 0.85x (15% discount)
     - Neutral: 1.0x
     - Disliked (affection <20): 1.2x (20% markup, begrudging sale)
     ```
   - Supply/demand tracking:
     ```go
     type MarketData struct {
       LocationID       uuid.UUID
       ItemID           uuid.UUID
       LocalSupply      int // Total quantity available from all merchants
       LocalDemand      int // Purchase attempts in past 7 days
       AveragePrice     float64
       PriceHistory     []PricePoint // Daily prices for trend analysis
       LastUpdated      time.Time
       
       // Market health indicators
       ShortageLevel    float64 // 0.0 (surplus) to 1.0 (severe shortage)
       InflationRate    float64 // % change in prices over 30 days
     }
     
     type PricePoint struct {
       Date  time.Time
       Price float64
     }
     
     func CalculateSupplyDemandModifier(locationID uuid.UUID, itemID uuid.UUID) float64 {
       market := GetMarketData(locationID, itemID)
       
       if market.LocalSupply == 0 {
         return 2.5 // Extreme scarcity, max price
       }
       
       ratio := float64(market.LocalDemand) / float64(market.LocalSupply)
       modifier := math.Sqrt(ratio)
       
       // Clamp between 0.4 and 2.5
       return math.Max(0.4, math.Min(2.5, modifier))
     }
     
     func (market *MarketData) UpdateSupply(locationID uuid.UUID, itemID uuid.UUID) {
       // Query all merchants in area
       merchants := GetMerchantsInRadius(locationID, radius=1000m)
       
       totalSupply := 0
       for _, merchant := range merchants {
         quantity := merchant.Inventory.Count(itemID)
         totalSupply += quantity
       }
       
       market.LocalSupply = totalSupply
       market.LastUpdated = time.Now()
     }
     
     func (market *MarketData) UpdateDemand(locationID uuid.UUID, itemID uuid.UUID) {
       // Count purchase attempts (successful + failed) in past 7 days
       cutoff := time.Now().AddDate(0, 0, -7)
       
       purchases := CountPurchaseAttempts(locationID, itemID, cutoff)
       market.LocalDemand = purchases
       market.LastUpdated = time.Now()
     }
     ```
   - Dynamic price adjustment by merchants:
     ```go
     func (merchant *Merchant) AdjustPrices() {
       for itemID, quantity := range merchant.Inventory.Items {
         // Get market data
         market := GetMarketData(merchant.Location, itemID)
         
         // Adjust based on stock level
         stockRatio := float64(quantity) / float64(merchant.TargetStock(itemID))
         
         if stockRatio < 0.3 {
           // Low stock, increase price
           merchant.SetItemPrice(itemID, market.AveragePrice * 1.2)
         } else if stockRatio > 2.0 {
           // Overstocked, decrease price
           merchant.SetItemPrice(itemID, market.AveragePrice * 0.8)
         } else {
           // Normal stock, use market average
           merchant.SetItemPrice(itemID, market.AveragePrice)
         }
       }
     }
     ```

5. Implement barter system:
   - Barter exchange:
     - NPCs accept trades if value approximately equal
     - Value threshold depends on relationship
     - Merchants prefer currency but will barter if necessary
   - Barter struct:
     ```go
     type BarterOffer struct {
       OfferID             uuid.UUID
       OfferedBy           uuid.UUID // Player or NPC
       OfferedTo           uuid.UUID // NPC merchant
       OfferedItems        []ItemStack
       RequestedItems      []ItemStack
       TotalOfferedValue   int
       TotalRequestedValue int
       Status              BarterStatus // pending, accepted, rejected, countered
       CounterOffer        *BarterOffer
       CreatedAt           time.Time
       ExpiresAt           time.Time // Offers expire after 5 minutes
     }
     
     type ItemStack struct {
       ItemID   uuid.UUID
       Quantity int
       Quality  ItemQuality
     }
     
     type BarterStatus string
     const (
       BarterPending   BarterStatus = "pending"
       BarterAccepted  BarterStatus = "accepted"
       BarterRejected  BarterStatus = "rejected"
       BarterCountered BarterStatus = "countered"
       BarterExpired   BarterStatus = "expired"
     )
     ```
   - Barter evaluation:
     ```go
     func (merchant *Merchant) EvaluateBarter(offer *BarterOffer) (BarterStatus, *BarterOffer) {
       // 1. Calculate value ratio
       ratio := float64(offer.TotalOfferedValue) / float64(offer.TotalRequestedValue)
       
       // 2. Get relationship
       relationship := GetRelationship(merchant.NPCID, offer.OfferedBy)
       
       // 3. Determine acceptable range based on affection
       minRatio, maxRatio := GetAcceptableRange(relationship.Affection)
       // Friend (>60): 0.7 to 1.3 (accept 70-130%)
       // Neutral (20-60): 0.8 to 1.2 (accept 80-120%)
       // Disliked (<20): 1.1 to 1.5 (require 110-150%)
       
       // 4. Check if merchant wants offered items
       wantFactor := CalculateWantFactor(merchant, offer.OfferedItems)
       // wantFactor: 0.0 (don't want at all) to 1.5 (desperately need)
       
       adjustedMinRatio := minRatio / wantFactor
       adjustedMaxRatio := maxRatio / wantFactor
       
       // 5. Evaluate
       if ratio >= adjustedMinRatio && ratio <= adjustedMaxRatio {
         return BarterAccepted, nil
       } else if ratio > 0.6 && ratio < adjustedMinRatio {
         // Close enough, make counteroffer
         counter := merchant.GenerateCounterOffer(offer, adjustedMinRatio)
         return BarterCountered, counter
       } else {
         return BarterRejected, nil
       }
     }
     
     func GetAcceptableRange(affection int) (float64, float64) {
       if affection > 60 {
         return 0.7, 1.3 // Friends: generous range
       } else if affection > 20 {
         return 0.8, 1.2 // Neutral: fair range
       } else {
         return 1.1, 1.5 // Disliked: demand premium
       }
     }
     
     func CalculateWantFactor(merchant *Merchant, items []ItemStack) float64 {
       totalWant := 0.0
       
       for _, item := range items {
         // Check if merchant can resell this item
         market := GetMarketData(merchant.Location, item.ItemID)
         demand := market.LocalDemand
         supply := market.LocalSupply
         
         if demand > supply * 2 {
           // High demand item, want it
           totalWant += 1.5
         } else if demand > supply {
           // Some demand
           totalWant += 1.0
         } else {
           // Low demand, don't really want it
           totalWant += 0.5
         }
       }
       
       return totalWant / float64(len(items))
     }
     
     func (merchant *Merchant) GenerateCounterOffer(original *BarterOffer, targetRatio float64) *BarterOffer {
       // Keep requested items the same
       // Adjust offered items to match target ratio
       
       targetValue := int(float64(original.TotalRequestedValue) * targetRatio)
       
       counter := &BarterOffer{
         OfferID:             uuid.New(),
         OfferedBy:           merchant.NPCID,
         OfferedTo:           original.OfferedBy,
         OfferedItems:        original.RequestedItems, // Merchant gives what was requested
         RequestedItems:      AdjustItemsToValue(original.OfferedItems, targetValue),
         TotalOfferedValue:   original.TotalRequestedValue,
         TotalRequestedValue: targetValue,
         Status:              BarterPending,
         CreatedAt:           time.Now(),
         ExpiresAt:           time.Now().Add(5 * time.Minute),
       }
       
       return counter
     }
     ```

6. Test economic simulation and inflation prevention:
   - Economic balance tests:
     - **Money supply**: Track total currency in circulation
       ```go
       func (sim *EconomicSimulation) GetMoneySupply() int {
         totalMoney := 0
         
         // Sum all NPC wealth
         npcs := GetAllNPCs(sim.WorldID)
         for _, npc := range npcs {
           totalMoney += npc.Wealth
         }
         
         // Sum all player wealth
         players := GetAllPlayers(sim.WorldID)
         for _, player := range players {
           totalMoney += player.Wealth
         }
         
         return totalMoney
       }
       ```
     - **Price stability**: Monitor average prices over time
       ```go
       func (sim *EconomicSimulation) GetInflationRate(days int) float64 {
         markets := GetAllMarketData(sim.WorldID)
         
         totalPriceChange := 0.0
         count := 0
         
         for _, market := range markets {
           if len(market.PriceHistory) >= days {
             oldPrice := market.PriceHistory[len(market.PriceHistory)-days].Price
             currentPrice := market.AveragePrice
             priceChange := (currentPrice - oldPrice) / oldPrice
             totalPriceChange += priceChange
             count++
           }
         }
         
         if count == 0 {
           return 0.0
         }
         
         return (totalPriceChange / float64(count)) * 100 // Percentage
       }
       ```
     - **Wealth distribution**: Measure Gini coefficient
       ```go
       func (sim *EconomicSimulation) CalculateGiniCoefficient() float64 {
         // Get all entity wealth values
         var wealthValues []int
         
         npcs := GetAllNPCs(sim.WorldID)
         for _, npc := range npcs {
           wealthValues = append(wealthValues, npc.Wealth)
         }
         
         players := GetAllPlayers(sim.WorldID)
         for _, player := range players {
           wealthValues = append(wealthValues, player.Wealth)
         }
         
         // Sort ascending
         sort.Ints(wealthValues)
         
         // Calculate Gini
         n := len(wealthValues)
         sumOfDifferences := 0.0
         sumOfWealth := 0.0
         
         for i, wealth := range wealthValues {
           sumOfWealth += float64(wealth)
           sumOfDifferences += float64(i+1) * float64(wealth)
         }
         
         if sumOfWealth == 0 {
           return 0.0
         }
         
         gini := (2.0 * sumOfDifferences) / (float64(n) * sumOfWealth) - (float64(n+1) / float64(n))
         
         return gini
         // Target: 0.3-0.5 (moderate inequality)
         // >0.7 indicates extreme wealth concentration
       }
       ```
     - **Trade volume**: NPCs should trade actively
       ```go
       func (sim *EconomicSimulation) GetAverageTradesPerNPC(days int) float64 {
         cutoff := time.Now().AddDate(0, 0, -days)
         
         npcs := GetAllNPCs(sim.WorldID)
         totalTrades := 0
         
         for _, npc := range npcs {
           if merchant := npc.AsMerchant(); merchant != nil {
             trades := CountTradesSince(merchant.NPCID, cutoff)
             totalTrades += trades
           }
         }
         
         return float64(totalTrades) / float64(len(npcs)) / float64(days)
       }
       ```
   - Inflation prevention mechanisms:
     - **Currency sinks**: Remove money from economy
       ```go
       var CurrencySinks = []CurrencySink{
         {Name: "Taxes", Rate: 0.05}, // 5% of transactions
         {Name: "Repair Costs", Rate: 0.02}, // 2% of item value per repair
         {Name: "Rent", Amount: 100}, // Fixed 100 gold per week
         {Name: "Tool Maintenance", Amount: 50}, // 50 gold per week
         {Name: "Food Consumption", Amount: 20}, // 20 gold per day
       }
       ```
     - **Resource sinks**: Remove items from economy
       - Durability loss (Phase 7.2, Phase 9.2)
       - Food consumption/spoilage
       - Ammunition consumption (arrows, bullets)
     - **Regulated markets**: Merchant price adjustments prevent runaway inflation
     - **Production limits**: Resource regeneration rates (Phase 9.1) cap supply growth
   - Run 365-day economic simulation:
     ```go
     func TestEconomicStability(t *testing.T) {
       // Setup: Spawn 100 NPCs in test world
       world := CreateTestWorld()
       npcs := SpawnNPCs(world.ID, count=100)
       
       // Assign occupations
       AssignOccupations(npcs, distribution=map[Occupation]int{
         "farmer": 30,
         "miner": 15,
         "woodcutter": 10,
         "merchant": 20,
         "craftsman": 25,
       })
       
       // Track metrics over 365 days
       var metrics []DailyMetrics
       
       for day := 0; day < 365; day++ {
         // Run daily simulation
         SimulateDay(world)
         
         // Collect metrics
         metrics = append(metrics, DailyMetrics{
           Day:             day,
           MoneySupply:     GetMoneySupply(world.ID),
           AveragePrice:    GetAveragePriceLevel(world.ID),
           GiniCoefficient: CalculateGiniCoefficient(world.ID),
           TradesPerNPC:    GetTradesPerNPC(world.ID, days=1),
         })
       }
       
       // Verify stability
       yearInflation := CalculateInflation(metrics[0].AveragePrice, metrics[364].AveragePrice)
       assert.Less(t, yearInflation, 0.10, "Inflation should be <10% per year")
       
       finalGini := metrics[364].GiniCoefficient
       assert.InRange(t, finalGini, 0.3, 0.5, "Gini coefficient should show moderate inequality")
       
       avgTrades := AverageMetric(metrics, func(m DailyMetrics) float64 { return m.TradesPerNPC })
       assert.Greater(t, avgTrades, 10.0, "NPCs should average 10+ trades/day")
       
       // No NPC should have >50% of wealth
       maxWealth := GetMaxNPCWealth(world.ID)
       totalWealth := GetMoneySupply(world.ID)
       wealthConcentration := float64(maxWealth) / float64(totalWealth)
       assert.Less(t, wealthConcentration, 0.5, "No single NPC should have >50% of wealth")
     }
     ```

7. Add trade routes and merchant travel:
   - Trade route system:
     - Merchant NPCs travel between settlements
     - Buy low in one location, sell high in another
     - Arbitrage opportunities drive travel decisions
   - Route calculation:
     ```go
     type TradeRoute struct {
       RouteID       uuid.UUID
       MerchantID    uuid.UUID
       Origin        geo.Point
       Destination   geo.Point
       Distance      float64
       TravelTime    time.Duration
       CargoCapacity int
       
       // Economics
       EstimatedProfit int
       TravelCost      int
       NetProfit       int
       
       // Items to trade
       BuyItems      []TradeItem // Buy at origin
       SellItems     []TradeItem // Sell at destination
       
       // Status
       Status        RouteStatus // planning, traveling, trading, returning
       StartedAt     time.Time
     }
     
     type TradeItem struct {
       ItemID        uuid.UUID
       Quantity      int
       BuyPriceEach  int
       SellPriceEach int
       Margin        int
     }
     
     func (merchant *Merchant) PlanTradeRoute(ctx context.Context) *TradeRoute {
       currentLocation := merchant.Location
       
       // 1. Survey nearby settlements (within 50km)
       settlements := FindNearbySettlements(currentLocation, radius=50km)
       
       if len(settlements) == 0 {
         return nil // No trade opportunities
       }
       
       // 2. Compare prices across settlements
       var bestRoute *TradeRoute
       bestProfit := 0.0
       
       for _, settlement := range settlements {
         // Get market data for both locations
         homeMarket := GetMarketDataForLocation(currentLocation)
         foreignMarket := GetMarketDataForLocation(settlement.Location)
         
         // Find arbitrage opportunities
         opportunities := FindArbitrageOpportunities(homeMarket, foreignMarket)
         
         if len(opportunities) == 0 {
           continue
         }
         
         // Calculate potential profit
         route := &TradeRoute{
           RouteID:     uuid.New(),
           MerchantID:  merchant.NPCID,
           Origin:      currentLocation,
           Destination: settlement.Location,
           Distance:    CalculateDistance(currentLocation, settlement.Location),
         }
         
         route.TravelTime = CalculateTravelTime(route.Distance, merchant.TravelSpeed)
         route.TravelCost = CalculateTravelCost(route.Distance, route.TravelTime)
         route.CargoCapacity = merchant.CargoCapacity()
         
         // Select best items to trade within cargo capacity
         cargo := SelectOptimalCargo(opportunities, route.CargoCapacity, merchant.Wealth)
         route.BuyItems = cargo.BuyAtOrigin
         route.SellItems = cargo.SellAtDestination
         
         route.EstimatedProfit = cargo.TotalProfit
         route.NetProfit = route.EstimatedProfit - route.TravelCost
         
         if route.NetProfit > bestProfit {
           bestProfit = route.NetProfit
           bestRoute = route
         }
       }
       
       // 3. Only travel if profit > 1.5x travel cost (worthwhile margin)
       if bestRoute != nil && bestRoute.NetProfit > bestRoute.TravelCost * 1.5 {
         return bestRoute
       }
       
       return nil // Stay home, no profitable routes
     }
     
     func FindArbitrageOpportunities(homeMarket, foreignMarket *LocationMarket) []ArbitrageOpportunity {
       var opportunities []ArbitrageOpportunity
       
       // Compare prices for each item type
       for itemID, homePrice := range homeMarket.Prices {
         if foreignPrice, exists := foreignMarket.Prices[itemID]; exists {
           // Check supply at home (need to buy) and demand abroad (need to sell)
           homeSupply := homeMarket.Supply[itemID]
           foreignDemand := foreignMarket.Demand[itemID]
           
           if homeSupply > 0 && foreignDemand > 0 {
             priceDifference := foreignPrice - homePrice
             margin := float64(priceDifference) / float64(homePrice)
             
             // Only worthwhile if >20% margin
             if margin > 0.2 {
               opportunities = append(opportunities, ArbitrageOpportunity{
                 ItemID:         itemID,
                 BuyPrice:       homePrice,
                 SellPrice:      foreignPrice,
                 Margin:         margin,
                 AvailableSupply: homeSupply,
                 ForeignDemand:  foreignDemand,
               })
             }
           }
         }
       }
       
       // Sort by margin (highest first)
       sort.Slice(opportunities, func(i, j int) bool {
         return opportunities[i].Margin > opportunities[j].Margin
       })
       
       return opportunities
     }
     
     func SelectOptimalCargo(
       opportunities []ArbitrageOpportunity,
       cargoCapacity int,
       availableWealth int,
     ) *TradeCargo {
       cargo := &TradeCargo{
         BuyAtOrigin: []TradeItem{},
         SellAtDestination: []TradeItem{},
         TotalProfit: 0,
       }
       
       usedCapacity := 0
       spentWealth := 0
       
       // Greedy algorithm: take highest margin items first
       for _, opp := range opportunities {
         item := GetItem(opp.ItemID)
         maxQuantity := min(
           (cargoCapacity - usedCapacity) / item.Weight,
           (availableWealth - spentWealth) / opp.BuyPrice,
           opp.AvailableSupply,
         )
         
         if maxQuantity <= 0 {
           continue
         }
         
         tradeItem := TradeItem{
           ItemID:        opp.ItemID,
           Quantity:      maxQuantity,
           BuyPriceEach:  opp.BuyPrice,
           SellPriceEach: opp.SellPrice,
           Margin:        opp.SellPrice - opp.BuyPrice,
         }
         
         cargo.BuyAtOrigin = append(cargo.BuyAtOrigin, tradeItem)
         cargo.SellAtDestination = append(cargo.SellAtDestination, tradeItem)
         cargo.TotalProfit += tradeItem.Margin * maxQuantity
         
         usedCapacity += item.Weight * maxQuantity
         spentWealth += opp.BuyPrice * maxQuantity
         
         if usedCapacity >= cargoCapacity || spentWealth >= availableWealth {
           break
         }
       }
       
       return cargo
     }
     
     func (merchant *Merchant) ExecuteTradeRoute(route *TradeRoute) error {
       // Phase 1: Purchase goods at origin
       for _, item := range route.BuyItems {
         if err := merchant.PurchaseItem(item.ItemID, item.Quantity, item.BuyPriceEach); err != nil {
           return fmt.Errorf("failed to purchase %s: %w", item.ItemID, err)
         }
       }
       
       // Phase 2: Travel to destination
       route.Status = RouteTraveling
       route.StartedAt = time.Now()
       
       if err := merchant.TravelTo(route.Destination, route.TravelTime); err != nil {
         return fmt.Errorf("travel failed: %w", err)
       }
       
       // Phase 3: Sell goods at destination
       route.Status = RouteTrading
       actualProfit := 0
       
       for _, item := range route.SellItems {
         soldFor, err := merchant.SellItem(item.ItemID, item.Quantity, item.SellPriceEach)
         if err != nil {
           return fmt.Errorf("failed to sell %s: %w", item.ItemID, err)
         }
         actualProfit += soldFor - (item.BuyPriceEach * item.Quantity)
       }
       
       // Phase 4: Return home
       route.Status = RouteReturning
       if err := merchant.TravelTo(route.Origin, route.TravelTime); err != nil {
         return fmt.Errorf("return travel failed: %w", err)
       }
       
       // Record trade route success
       merchant.Wealth += actualProfit - route.TravelCost
       merchant.RecordSuccessfulRoute(route, actualProfit)
       
       return nil
     }
     ```
   - Travel mechanics:
     - Merchants travel at walking speed: 5 km/hour (game time)
     - Carry cargo limited by strength/pack animals
     - Random encounters during travel (bandits, wildlife, other merchants)
     - Weather affects travel time (rain, snow slows by 30-50%)

Test Requirements (80%+ coverage):
- NPC gathering autonomously finds and harvests nearby resources
- Occupation-based priorities: farmers prioritize crops, miners prioritize ore
- NPC gathering uses Phase 9.1 harvest mechanics (skill-based yield)
- Resource acquisition need decreases by 20 after successful gathering
- NPC crafting selects recipes from known recipes
- Crafting prioritizes profitable items based on local market prices
- Recipe proficiency improves with repeated crafting (+1 per use)
- Merchant inventory restocks based on 7-day sales history
- Popular items restocked, items with <3 sales in 14 days removed
- Merchant wealth limits inventory size (poor: 10-20 items, wealthy: 50-150 items)
- Price calculation includes supply/demand modifier correctly
- High demand + low supply → prices increase (up to 2.5x)
- Low demand + high supply → prices decrease (down to 0.4x)
- Quality modifier affects price (masterwork 5x base value)
- Relationship modifier affects price (friends get 15% discount)
- Barter evaluation accepts trades within relationship-based threshold
- Friends accept 70-130% value ratio, neutral 80-120%, disliked 110-150%
- Want factor increases acceptance for high-demand items
- Merchant generates counteroffer when trade is close but not acceptable
- Money supply tracking sums all NPC and player wealth
- Inflation rate calculation compares prices over time (30-day periods)
- Gini coefficient calculation measures wealth inequality correctly
- Trade volume tracking counts trades per NPC per day
- Currency sinks (taxes, repairs, rent) remove money from economy
- Economic simulation runs 365 days without crashes
- Inflation stays <10% per year over simulation
- Gini coefficient remains 0.3-0.5 (moderate inequality)
- Trade volume: NPCs average 10+ trades/day
- No single NPC accumulates >50% of total wealth
- Trade route planning finds profitable arbitrage opportunities
- Trade routes only executed when profit >1.5x travel cost
- Arbitrage opportunities identified with >20% price difference
- Cargo selection maximizes profit within capacity and wealth limits
- Trade route execution: buy → travel → sell → return
- Process 1000 NPC economic actions in < 2 seconds
- Process 100 NPCs gathering/crafting/trading simultaneously

Acceptance Criteria:
- NPCs autonomously gather resources based on occupation and skills
- NPCs craft goods using available resources and known recipes
- Recipe proficiency increases with use, creating specialization over time
- Merchant inventories dynamic, restocking based on sales data
- Dynamic pricing responds to supply, demand, quality, and relationships
- Barter system allows non-currency trade with relationship-based evaluation
- Economic simulation remains stable over 365+ days (no runaway inflation)
- Wealth distribution remains moderate (Gini 0.3-0.5)
- Currency and resource sinks prevent infinite wealth accumulation
- Trade routes enable profitable inter-settlement commerce
- All tests pass with 80%+ coverage

Dependencies:
- Phase 5.1 (Desire Engine) - resource acquisition and task completion needs
- Phase 9.1 (Resource Distribution) - resource harvesting mechanics
- Phase 9.2 (Tech Trees & Recipes) - crafting mechanics and recipe knowledge
- Phase 3.3 (Relationships) - barter evaluation and price modifiers
- Phase 2.5 (Skills) - gathering, crafting, and mining skills
- Phase 8.2 (Geographic Generation) - settlements for trade routes

Files to Create:
- `internal/economy/npc/types.go` - Occupation, Merchant, TradeRoute structs
- `internal/economy/npc/gathering.go` - NPC autonomous gathering
- `internal/economy/npc/crafting.go` - NPC autonomous crafting
- `internal/economy/npc/specialization.go` - Recipe proficiency and specialization
- `internal/economy/npc/merchant.go` - Merchant inventory and behavior
- `internal/economy/npc/restocking.go` - Merchant restocking algorithm
- `internal/economy/market/types.go` - MarketData, PricePoint structs
- `internal/economy/market/pricing.go` - Dynamic pricing calculations
- `internal/economy/market/supply_demand.go` - Supply/demand tracking and updates
- `internal/economy/trade/barter.go` - Barter system and evaluation
- `internal/economy/trade/routes.go` - Trade route planning and execution
- `internal/economy/trade/arbitrage.go` - Arbitrage opportunity detection
- `internal/economy/trade/cargo.go` - Optimal cargo selection
- `internal/economy/simulation/types.go` - EconomicSimulation, DailyMetrics structs
- `internal/economy/simulation/runner.go` - Economic simulation loop
- `internal/economy/simulation/inflation.go` - Inflation prevention and monitoring
- `internal/economy/simulation/metrics.go` - Economic health metrics (Gini, money supply, trade volume)
- `internal/economy/simulation/sinks.go` - Currency and resource sink implementations
- `internal/economy/npc/gathering_test.go` - Gathering behavior tests
- `internal/economy/npc/crafting_test.go` - Crafting behavior tests
- `internal/economy/npc/merchant_test.go` - Merchant restocking tests
- `internal/economy/market/pricing_test.go` - Dynamic pricing tests
- `internal/economy/trade/barter_test.go` - Barter evaluation tests
- `internal/economy/trade/routes_test.go` - Trade route planning tests
- `internal/economy/simulation/stability_test.go` - 365-day economic simulation tests
- `internal/economy/simulation/metrics_test.go` - Gini coefficient and inflation tests
- `migrations/postgres/XXX_merchants.sql` - Merchants table
- `migrations/postgres/XXX_market_data.sql` - Market data and price history tables
- `migrations/postgres/XXX_trade_routes.sql` - Trade routes table
- `migrations/postgres/XXX_barter_offers.sql` - Barter offers table

---

# Phase 10: Frontend UI (3-4 weeks)

> **Architecture Principle: Thin Client**
> The frontend is a "dumb terminal" - it sends raw text to the backend and renders structured responses.
> All command parsing, validation, context tracking, and output formatting logic resides on the backend.

## Phase 10.1: Core UI Components (2 weeks)
### Status: ⏳ Not Started
### Prompt:
Following TDD principles and **thin-client architecture**, implement Phase 10.1 Core UI Components optimized for mobile with desktop support:

Core Requirements:

1. **Mobile-first UI layout with desktop responsive scaling**:
   - Mobile layout (primary target):
     ```
     ┌─────────────────────────┐
     │  Status Bar (HP/Stam)   │ ← Fixed top, 60px
     ├─────────────────────────┤
     │                         │
     │   Main Text Display     │ ← Scrollable, auto-scroll
     │   (Game Output)         │   to bottom on new content
     │                         │
     ├─────────────────────────┤
     │  Command Input          │ ← Fixed bottom, 80px
     │  [Text box] [Send]      │   NO client-side parsing
     ├─────────────────────────┤
     │ [Quick Buttons]         │ ← Customizable, 60px
     └─────────────────────────┘
     ```
   - Desktop layout (responsive):
     ```
     ┌──────────┬──────────────────────┬──────────┐
     │  Left    │   Status Bar         │  Right   │
     │  Panel   ├──────────────────────┤  Panel   │
     │  (Map)   │   Main Text Display  │  (Stats) │
     │  200px   ├──────────────────────┤  250px   │
     │          │   Command Input      │          │
     └──────────┴──────────────────────┴──────────┘
     ```
   - Breakpoints: Mobile (0-768px), Tablet (769-1024px), Desktop (1025px+)

2. **Simple command input (NO client-side parsing)**:
   - Frontend responsibility: ONLY send raw text via WebSocket
   - Backend handles: alias resolution, target extraction, fuzzy matching, context memory
   - Implementation:
     ```svelte
     <script lang="ts">
       import { gameSocket } from '$lib/network/websocket';
       
       let inputText = '';
       let commandHistory: string[] = [];
       let historyIndex = -1;
       
       function sendCommand() {
         if (!inputText.trim()) return;
         
         // Send raw text to backend - NO PARSING
         gameSocket.send({ text: inputText.trim() });
         
         // Track history for up/down navigation
         commandHistory.unshift(inputText);
         if (commandHistory.length > 50) commandHistory.pop();
         
         inputText = '';
         historyIndex = -1;
       }
       
       function handleKeydown(e: KeyboardEvent) {
         if (e.key === 'Enter') {
           sendCommand();
         } else if (e.key === 'ArrowUp') {
           // Navigate command history
           if (historyIndex < commandHistory.length - 1) {
             historyIndex++;
             inputText = commandHistory[historyIndex];
           }
           e.preventDefault();
         } else if (e.key === 'ArrowDown') {
           if (historyIndex > 0) {
             historyIndex--;
             inputText = commandHistory[historyIndex];
           } else if (historyIndex === 0) {
             historyIndex = -1;
             inputText = '';
           }
           e.preventDefault();
         }
       }
     </script>
     
     <div class="command-input">
       <input
         type="text"
         bind:value={inputText}
         on:keydown={handleKeydown}
         placeholder="Enter command..."
         autocomplete="off"
       />
       <button on:click={sendCommand}>Send</button>
     </div>
     ```

3. **Render structured responses from backend**:
   - Backend sends structured `GameMessage` with formatting data:
     ```typescript
     interface GameMessage {
       type: 'movement' | 'area_description' | 'combat' | 'dialogue' | 
             'item_acquired' | 'crafting' | 'error' | 'system';
       text: string;
       
       // Pre-computed display properties from backend
       segments?: TextSegment[];  // Color-coded segments
       entities?: EntityRef[];     // Clickable entity references
       
       // Type-specific data
       direction?: string;
       damage?: number;
       speakerName?: string;
       itemName?: string;
       itemRarity?: string;
       quality?: string;
     }
     
     interface TextSegment {
       text: string;
       color: string;        // CSS class from backend
       bold?: boolean;
       italic?: boolean;
       entityId?: string;    // For click interactions
       entityType?: string;
     }
     ```
   - Frontend simply renders what backend provides:
     ```svelte
     <script lang="ts">
       export let message: GameMessage;
       
       function handleEntityClick(entityId: string, entityType: string) {
         // Send interaction command to backend
         gameSocket.send({ text: `look ${entityId}` });
       }
     </script>
     
     <div class="message message-{message.type}">
       {#if message.segments}
         {#each message.segments as segment}
           <span
             class={segment.color}
             class:font-bold={segment.bold}
             class:italic={segment.italic}
             class:clickable={segment.entityId}
             on:click={() => segment.entityId && 
               handleEntityClick(segment.entityId, segment.entityType)}
           >
             {segment.text}
           </span>
         {/each}
       {:else}
         {message.text}
       {/if}
     </div>
     ```

4. **Map visualization (render backend-provided data)**:
   - Backend sends `visible_tiles` with pre-computed data:
     ```typescript
     interface VisibleTile {
       x: number;
       y: number;
       biome: string;
       biomeColor: string;    // Backend provides CSS color
       elevation: number;
       isVisible: boolean;    // Fog of war computed by backend
       entities: MapEntity[];
     }
     ```
   - Frontend renders without computing visibility or colors

5. **Status bars (display backend-provided values)**:
   - Backend sends current/max values for HP, Stamina, Focus
   - Frontend calculates percentage and renders bars
   - Color thresholds can be defined on frontend (presentation only)

6. **Quick action buttons**:
   - Pre-defined buttons that send fixed commands
   - Example: "Look" → `gameSocket.send({ text: 'look' })`
   - No parsing, just send raw command strings

Test Requirements (80%+ coverage):
- CommandInput sends raw text via WebSocket without modification
- Command history navigation (up/down arrows) works correctly
- History limited to 50 entries
- TextDisplay renders segments with correct CSS classes
- TextDisplay handles clicks on entity segments
- Entity clicks send "look {entityId}" command to backend
- MapCanvas renders tiles with backend-provided colors
- StatusBar calculates percentage from backend values
- Mobile layout renders single column at <768px
- Desktop layout renders three columns at ≥1025px
- Quick buttons send correct command strings
- WebSocket connection handles reconnection gracefully
- Component renders in <100ms on mobile devices

Acceptance Criteria:
- **Frontend sends raw text only** - no command parsing on client
- **Backend provides all formatting** - colors, entity highlights, segments
- UI responsive across mobile/tablet/desktop breakpoints
- Command history navigable with up/down arrows
- Clickable entities send interaction commands
- All tests pass with 80%+ coverage

Dependencies:
- Phase 10.2 (Backend Output Formatting) - structured message format
- WebSocket connection (existing)
- SvelteKit (existing)

Files to Create:
- `src/lib/components/Input/CommandInput.svelte` - Raw text input, NO parsing
- `src/lib/components/Output/TextDisplay.svelte` - Renders backend segments (exists, verify)
- `src/lib/components/Output/FormattedText.svelte` - Segment renderer (exists, verify)
- `src/lib/components/Layout/MobileLayout.svelte` - Mobile layout (exists)
- `src/lib/components/Layout/DesktopLayout.svelte` - Desktop layout (exists)
- `src/lib/components/Layout/GameContainer.svelte` - Responsive container (exists)
- `src/lib/components/Map/MapCanvas.svelte` - Canvas map (exists)
- `src/lib/components/Character/StatusBar.svelte` - HP/Stamina bars (exists)
- `src/lib/components/Input/QuickButtons.svelte` - Pre-defined command buttons
- `src/lib/stores/commandHistory.ts` - Command history store
- `tests/components/CommandInput.test.ts` - Input tests
- `tests/components/TextDisplay.test.ts` - Output tests
- `tests/e2e/ui_interaction.spec.ts` - E2E UI tests

---

## Phase 10.2: Backend Output Formatting (1 week)
### Status: ⏳ Not Started
### Prompt:
Following TDD principles, implement Phase 10.2 Backend Output Formatting to support thin-client architecture:

Core Requirements:

1. **Enhance GameMessage struct with formatting data**:
   ```go
   type GameMessage struct {
       Type     string        `json:"type"`
       Text     string        `json:"text"`
       Segments []TextSegment `json:"segments,omitempty"`
       Entities []EntityRef   `json:"entities,omitempty"`
       
       // Type-specific fields
       Direction   string `json:"direction,omitempty"`
       Damage      int    `json:"damage,omitempty"`
       SpeakerName string `json:"speaker_name,omitempty"`
       SpeakerID   string `json:"speaker_id,omitempty"`
       ItemName    string `json:"item_name,omitempty"`
       ItemRarity  string `json:"item_rarity,omitempty"`
       ItemQuality string `json:"item_quality,omitempty"`
   }
   
   type TextSegment struct {
       Text       string `json:"text"`
       Color      string `json:"color"`      // CSS class: "text-red-500"
       Bold       bool   `json:"bold,omitempty"`
       Italic     bool   `json:"italic,omitempty"`
       EntityID   string `json:"entity_id,omitempty"`
       EntityType string `json:"entity_type,omitempty"`
   }
   
   type EntityRef struct {
       ID   string `json:"id"`
       Name string `json:"name"`
       Type string `json:"type"` // npc, item, resource, location
   }
   ```

2. **Implement OutputFormatter service**:
   ```go
   type OutputFormatter struct {
       rarityColors  map[string]string
       qualityColors map[string]string
       messageColors map[string]string
   }
   
   func NewOutputFormatter() *OutputFormatter {
       return &OutputFormatter{
           rarityColors: map[string]string{
               "common":    "text-gray-100",
               "uncommon":  "text-green-400",
               "rare":      "text-blue-400",
               "very_rare": "text-purple-400",
               "legendary": "text-orange-500",
           },
           qualityColors: map[string]string{
               "poor":       "text-gray-400",
               "common":     "text-gray-100",
               "good":       "text-green-400",
               "excellent":  "text-blue-400",
               "masterwork": "text-purple-500",
           },
           messageColors: map[string]string{
               "combat":   "text-red-400",
               "dialogue": "text-green-300",
               "system":   "text-purple-400",
               "error":    "text-red-500",
           },
       }
   }
   
   func (f *OutputFormatter) FormatCombat(
       attacker string, target string, damage int, targetID uuid.UUID,
   ) *GameMessage {
       return &GameMessage{
           Type:   "combat",
           Text:   fmt.Sprintf("%s attacks %s for %d damage!", attacker, target, damage),
           Damage: damage,
           Segments: []TextSegment{
               {Text: attacker + " attacks ", Color: "text-gray-300"},
               {Text: target, Color: "text-yellow-400", Bold: true, 
                EntityID: targetID.String(), EntityType: "npc"},
               {Text: " for ", Color: "text-gray-300"},
               {Text: strconv.Itoa(damage), Color: "text-orange-500", Bold: true},
               {Text: " damage!", Color: "text-red-400"},
           },
       }
   }
   
   func (f *OutputFormatter) FormatDialogue(
       speaker string, speakerID uuid.UUID, text string,
   ) *GameMessage {
       return &GameMessage{
           Type:        "dialogue",
           Text:        fmt.Sprintf("%s says: \"%s\"", speaker, text),
           SpeakerName: speaker,
           SpeakerID:   speakerID.String(),
           Segments: []TextSegment{
               {Text: speaker, Color: "text-cyan-400", Bold: true,
                EntityID: speakerID.String(), EntityType: "npc"},
               {Text: " says: \"", Color: "text-gray-300"},
               {Text: text, Color: "text-green-300", Italic: true},
               {Text: "\"", Color: "text-gray-300"},
           },
       }
   }
   
   func (f *OutputFormatter) FormatItemAcquired(
       item *Item, quantity int,
   ) *GameMessage {
       return &GameMessage{
           Type:       "item_acquired",
           Text:       fmt.Sprintf("You obtained %s (×%d)", item.Name, quantity),
           ItemName:   item.Name,
           ItemRarity: item.Rarity,
           Segments: []TextSegment{
               {Text: "You obtained ", Color: "text-gray-300"},
               {Text: item.Name, Color: f.rarityColors[item.Rarity], Bold: true,
                EntityID: item.ID.String(), EntityType: "item"},
               {Text: fmt.Sprintf(" (×%d)", quantity), Color: "text-gray-400"},
           },
       }
   }
   ```

3. **Add fuzzy matching to backend CommandParser**:
   ```go
   func (p *CommandParser) ParseTextWithFuzzy(text string) *websocket.CommandData {
       cmd := p.ParseText(text)
       
       // If unknown command, try fuzzy matching
       if cmd.Action == text && !p.isKnownCommand(cmd.Action) {
           fuzzyMatch := p.tryFuzzyMatch(cmd.Action)
           if fuzzyMatch != "" {
               cmd.Action = fuzzyMatch
               cmd.WasCorrected = true
               cmd.OriginalInput = text
           }
       }
       
       return cmd
   }
   
   func (p *CommandParser) tryFuzzyMatch(input string) string {
       bestMatch := ""
       bestDistance := 3 // Max Levenshtein distance
       
       for action := range p.aliases {
           distance := levenshtein(input, action)
           if distance < bestDistance {
               bestDistance = distance
               bestMatch = action
           }
           
           // Also check aliases
           for _, alias := range p.aliases[action] {
               distance := levenshtein(input, alias)
               if distance < bestDistance {
                   bestDistance = distance
                   bestMatch = action
               }
           }
       }
       
       return bestMatch
   }
   
   func levenshtein(s1, s2 string) int {
       // Standard Levenshtein distance implementation
       // ... (full implementation in code)
   }
   ```

4. **Add context memory to backend**:
   ```go
   type CommandContext struct {
       LastTarget    string
       LastNPC       string
       LastDirection string
       LastRoom      uuid.UUID
   }
   
   // Store context per client session
   type ClientState struct {
       // ... existing fields
       CommandContext *CommandContext
   }
   
   func (p *GameProcessor) ProcessCommand(ctx context.Context, client *Client, cmd *CommandData) error {
       // Use context for missing targets
       if cmd.Target == nil && p.requiresTarget(cmd.Action) {
           if client.State.CommandContext.LastTarget != "" {
               cmd.Target = &client.State.CommandContext.LastTarget
           }
       }
       
       // ... process command
       
       // Update context after successful command
       if cmd.Target != nil {
           client.State.CommandContext.LastTarget = *cmd.Target
       }
   }
   ```

Test Requirements (80%+ coverage):
- OutputFormatter.FormatCombat returns correct segments with colors
- OutputFormatter.FormatDialogue includes speaker entity reference
- OutputFormatter.FormatItemAcquired uses correct rarity color
- All rarity colors map correctly (common→gray, legendary→orange)
- All quality colors map correctly (poor→gray, masterwork→purple)
- Fuzzy matching corrects "loook" to "look" (distance 1)
- Fuzzy matching corrects "atack" to "attack" (distance 1)
- Fuzzy matching does NOT correct "xyz" (distance >2)
- Context memory stores last target after successful command
- Context memory provides last target when target missing
- GameMessage JSON serializes all segments correctly
- EntityRef clickable references serialize with IDs
- Process 1000 format operations in <100ms

Acceptance Criteria:
- All game output includes pre-formatted segments with CSS classes
- Entity references include IDs for frontend click handling
- Fuzzy matching corrects typos with Levenshtein distance ≤2
- Context memory tracks last target, NPC, direction per session
- Frontend receives ready-to-render formatted messages
- All tests pass with 80%+ coverage

Dependencies:
- Existing WebSocket message system
- Existing CommandParser

Files to Create:
- `internal/game/output/formatter.go` - OutputFormatter service
- `internal/game/output/types.go` - GameMessage, TextSegment, EntityRef
- `internal/game/output/colors.go` - Color mappings
- `internal/game/processor/fuzzy.go` - Levenshtein + fuzzy matching
- `internal/game/processor/context.go` - Command context memory
- `internal/game/output/formatter_test.go` - Formatter tests
- `internal/game/processor/fuzzy_test.go` - Fuzzy matching tests
- `internal/game/processor/context_test.go` - Context memory tests

---

## Phase 10.3: E2E UI Testing (1 week)
### Status: ⏳ Not Started
### Prompt:
Following TDD principles, implement Phase 10.3 End-to-End UI Testing for the thin-client architecture:

Core Requirements:

1. **WebSocket command flow E2E tests**:
   ```go
   func TestE2E_CommandFlow(t *testing.T) {
       // Setup: Create user, connect WebSocket
       client := setupTestClient(t)
       defer client.Close()
       
       // Test: Send raw text command
       client.Send(`{"text": "look"}`)
       
       // Verify: Receive formatted response with segments
       response := client.WaitForMessage(t, 5*time.Second)
       
       assert.Equal(t, "area_description", response.Type)
       assert.NotEmpty(t, response.Segments)
       assert.NotEmpty(t, response.Text)
       
       // Verify segments have colors
       for _, segment := range response.Segments {
           assert.NotEmpty(t, segment.Color)
       }
   }
   
   func TestE2E_FuzzyMatchingCorrection(t *testing.T) {
       client := setupTestClient(t)
       defer client.Close()
       
       // Send typo command
       client.Send(`{"text": "loook"}`)
       
       // Should be corrected to "look" and processed
       response := client.WaitForMessage(t, 5*time.Second)
       
       assert.Equal(t, "area_description", response.Type)
       // Optionally: system message about correction
   }
   
   func TestE2E_ContextMemory(t *testing.T) {
       client := setupTestClient(t)
       defer client.Close()
       
       // First: target a specific NPC
       client.Send(`{"text": "look guard"}`)
       response1 := client.WaitForMessage(t, 5*time.Second)
       assert.Contains(t, response1.Text, "guard")
       
       // Second: use implicit target
       client.Send(`{"text": "talk"}`) // No target specified
       response2 := client.WaitForMessage(t, 5*time.Second)
       
       // Should talk to "guard" from context
       assert.Equal(t, "dialogue", response2.Type)
       assert.Contains(t, response2.SpeakerName, "guard")
   }
   ```

2. **UI component E2E tests (Playwright)**:
   ```typescript
   // tests/e2e/ui_interaction.spec.ts
   import { test, expect } from '@playwright/test';
   
   test.describe('Thin Client UI', () => {
     test.beforeEach(async ({ page }) => {
       // Login and navigate to game
       await page.goto('/login');
       await page.fill('[data-testid="email"]', 'test@example.com');
       await page.fill('[data-testid="password"]', 'password123');
       await page.click('[data-testid="login-button"]');
       await page.waitForURL('/game');
     });
     
     test('sends raw text command on Enter', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       // Type and send command
       await input.fill('look');
       await input.press('Enter');
       
       // Verify input cleared
       await expect(input).toHaveValue('');
       
       // Verify response received
       const output = page.locator('[data-testid="game-output"]');
       await expect(output).toContainText(/You see|area/i, { timeout: 5000 });
     });
     
     test('navigates command history with arrow keys', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       // Send multiple commands
       await input.fill('look');
       await input.press('Enter');
       await input.fill('north');
       await input.press('Enter');
       await input.fill('inventory');
       await input.press('Enter');
       
       // Navigate up through history
       await input.press('ArrowUp');
       await expect(input).toHaveValue('inventory');
       
       await input.press('ArrowUp');
       await expect(input).toHaveValue('north');
       
       await input.press('ArrowUp');
       await expect(input).toHaveValue('look');
       
       // Navigate back down
       await input.press('ArrowDown');
       await expect(input).toHaveValue('north');
     });
     
     test('renders formatted text with colors', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       // Trigger combat (requires setup)
       await input.fill('attack goblin');
       await input.press('Enter');
       
       // Verify colored segments
       const damageText = page.locator('.text-orange-500');
       await expect(damageText).toBeVisible({ timeout: 5000 });
     });
     
     test('clickable entity sends look command', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       // Look to get entities in view
       await input.fill('look');
       await input.press('Enter');
       
       // Click on an entity
       const entity = page.locator('[data-entity-type="npc"]').first();
       if (await entity.isVisible()) {
         await entity.click();
         
         // Verify look command was sent for that entity
         const output = page.locator('[data-testid="game-output"]');
         await expect(output).toContainText(/You examine|looking at/i, { timeout: 5000 });
       }
     });
     
     test('responsive layout switches at breakpoints', async ({ page }) => {
       // Mobile
       await page.setViewportSize({ width: 375, height: 667 });
       const mobileLayout = page.locator('[data-testid="mobile-layout"]');
       await expect(mobileLayout).toBeVisible();
       
       // Desktop
       await page.setViewportSize({ width: 1920, height: 1080 });
       const desktopLayout = page.locator('[data-testid="desktop-layout"]');
       await expect(desktopLayout).toBeVisible();
     });
     
     test('quick buttons send raw commands', async ({ page }) => {
       const lookButton = page.locator('[data-testid="quick-look"]');
       await lookButton.click();
       
       const output = page.locator('[data-testid="game-output"]');
       await expect(output).toContainText(/You see|area/i, { timeout: 5000 });
     });
   });
   ```

3. **Mobile gesture E2E tests**:
   ```typescript
   test.describe('Mobile Gestures', () => {
     test.beforeEach(async ({ page }) => {
       await page.setViewportSize({ width: 375, height: 667 });
       // Login...
     });
     
     test('touch input works for command entry', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       await input.tap();
       await page.keyboard.type('look');
       await page.locator('[data-testid="send-button"]').tap();
       
       const output = page.locator('[data-testid="game-output"]');
       await expect(output).toContainText(/You see/i, { timeout: 5000 });
     });
     
     test('output scrolls to bottom on new content', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       // Send multiple commands to generate content
       for (let i = 0; i < 10; i++) {
         await input.fill('look');
         await input.press('Enter');
         await page.waitForTimeout(500);
       }
       
       // Verify scrolled to bottom
       const output = page.locator('[data-testid="game-output"]');
       const scrollTop = await output.evaluate(el => el.scrollTop);
       const scrollHeight = await output.evaluate(el => el.scrollHeight);
       const clientHeight = await output.evaluate(el => el.clientHeight);
       
       expect(scrollTop + clientHeight).toBeGreaterThanOrEqual(scrollHeight - 50);
     });
   });
   ```

4. **Performance E2E tests**:
   ```typescript
   test.describe('Performance', () => {
     test('renders new message in <100ms', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       const startTime = Date.now();
       await input.fill('look');
       await input.press('Enter');
       
       // Wait for response to appear
       const output = page.locator('[data-testid="game-output"] .message').last();
       await output.waitFor({ state: 'visible' });
       
       const endTime = Date.now();
       expect(endTime - startTime).toBeLessThan(2000); // Including network
     });
     
     test('handles rapid commands without lag', async ({ page }) => {
       const input = page.locator('[data-testid="command-input"]');
       
       // Send 10 commands in quick succession
       for (let i = 0; i < 10; i++) {
         await input.fill(`look ${i}`);
         await input.press('Enter');
       }
       
       // All should be processed
       const messages = page.locator('[data-testid="game-output"] .message');
       await expect(messages).toHaveCount(10, { timeout: 10000 });
     });
   });
   ```

Test Requirements (80%+ coverage):
- E2E: Raw text sent via WebSocket without client parsing
- E2E: Backend returns formatted segments with colors
- E2E: Fuzzy matching corrects typos server-side
- E2E: Context memory auto-fills missing targets
- E2E: Command history navigation with arrow keys
- E2E: Formatted text renders with correct CSS classes
- E2E: Entity clicks trigger look commands
- E2E: Responsive layout switches at breakpoints
- E2E: Quick buttons send correct commands
- E2E: Mobile touch interactions work
- E2E: Output auto-scrolls to bottom
- E2E: New messages render in <100ms (excluding network)
- E2E: Rapid commands processed without lag

Acceptance Criteria:
- All E2E tests verify thin-client architecture (no client parsing)
- Tests cover command flow, formatting, context, history
- Mobile and desktop layouts tested
- Performance benchmarks met
- All tests pass reliably in CI

Dependencies:
- Phase 10.1 (Core UI Components)
- Phase 10.2 (Backend Output Formatting)
- Playwright test framework
- Go test framework

Files to Create:
- `tests/e2e/command_flow_test.go` - Backend command flow E2E
- `tests/e2e/fuzzy_matching_test.go` - Fuzzy correction E2E
- `tests/e2e/context_memory_test.go` - Context memory E2E
- `mud-platform-client/tests/e2e/ui_interaction.spec.ts` - Playwright UI tests
- `mud-platform-client/tests/e2e/mobile_gestures.spec.ts` - Mobile E2E tests
- `mud-platform-client/tests/e2e/performance.spec.ts` - Performance E2E tests
- `mud-platform-client/playwright.config.ts` - Playwright configuration

---
## Phase 10.4: iPhone UX/UI Polish (1 week)
### Status: ⏳ Not Started
### Prompt:
Optimize the frontend UI for iPhone devices following iOS Human Interface Guidelines and mobile UX best practices. Focus on creating a premium, native-feeling experience for iPhone users while maintaining cross-platform compatibility.

> **IMPORTANT: iPhone-First Optimization**
> All UX/UI improvements must prioritize the iPhone experience. Test on real iPhone devices (iPhone 12-15 series) to ensure optimal performance and interaction quality.

Core Requirements:

1. **iOS Safe Area Handling** (Critical for notched iPhones):
   - Add proper viewport meta configuration:
     ```html
     <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover, maximum-scale=1.0, user-scalable=no">
     <meta name="apple-mobile-web-app-capable" content="yes">
     <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
     ```
   - Implement CSS safe area insets for notch/Dynamic Island:
     ```css
     .app-container {
       padding-top: env(safe-area-inset-top);
       padding-left: env(safe-area-inset-left);
       padding-right: env(safe-area-inset-right);
       padding-bottom: env(safe-area-inset-bottom);
     }
     
     /* For fixed bottom elements */
     .command-input-container {
       padding-bottom: calc(env(safe-area-inset-bottom) + 16px);
     }
     ```
   - Test on iPhone 12 Pro (notch), iPhone 14 Pro (Dynamic Island), iPhone SE (no notch)

2. **Touch Target Optimization** (iOS HIG: minimum 44×44pt):
   - Audit all interactive elements:
     ```svelte
     <!-- BEFORE: Too small for finger -->
     <button class="w-6 h-6">×</button>
     
     <!-- AFTER: Meeting iOS minimum -->
     <button class="min-w-[44px] min-h-[44px] flex items-center justify-center">
       <span class="text-xl">×</span>
     </button>
     ```
   - Add adequate spacing between tappable elements (minimum 8px)
   - Implement visual feedback for all touch interactions (`:active` states)

3. **Keyboard Management** (Critical for text input):
   - Prevent viewport zoom on input focus:
     ```css
     input, textarea, select {
       font-size: 16px; /* Prevents iOS auto-zoom */
     }
     ```
   - Auto-scroll focused input into view:
     ```typescript
     function handleInputFocus(element: HTMLElement) {
       // Wait for keyboard to appear
       setTimeout(() => {
         element.scrollIntoView({ behavior: 'smooth', block: 'center' });
       }, 300);
     }
     ```
   - Dim background content when keyboard is active
   - Add toolbar above keyboard for quick commands:
     ```svelte
     <div class="keyboard-accessory-bar" 
          style="position: fixed; bottom: 0; left: 0; right: 0; z-index: 1000;">
       <button>North</button>
       <button>South</button>
       <button>Look</button>
       <button>Inventory</button>
     </div>
     ```

4. **Haptic Feedback** (Vibration API):
   - Add tactile feedback for interactions:
     ```typescript
     function triggerHaptic(type: 'light' | 'medium' | 'heavy' | 'selection' | 'error' | 'success') {
       if (!('vibrate' in navigator)) return;
       
       const patterns = {
         light: [10],
         medium: [15],
         heavy: [20],
         selection: [5],
         error: [10, 50, 10],
         success: [10, 50, 10, 50, 10]
       };
       
       navigator.vibrate(patterns[type]);
     }
     
     // Usage:
     // - Command sent: triggerHaptic('light')
     // - Error message: triggerHaptic('error')
     // - Success action: triggerHaptic('success')
     // - Button tap: triggerHaptic('selection')
     ```
   - Add user preference toggle for haptic feedback

5. **iOS Scroll Behavior**:
   - Enable momentum scrolling:
     ```css
     .scrollable-area {
       -webkit-overflow-scrolling: touch; /* iOS momentum scrolling */
       overscroll-behavior: contain; /* Prevent pull-to-refresh */
     }
     ```
   - Prevent body scroll when modal is open:
     ```typescript
     // Lock body scroll (for modals, overlays)
     function lockScroll() {
       document.body.style.overflow = 'hidden';
       document.body.style.position = 'fixed';
       document.body.style.width = '100%';
     }
     
     function unlockScroll() {
       document.body.style.overflow = '';
       document.body.style.position = '';
       document.body.style.width = '';
     }
     ```
   - Disable pull-to-refresh in PWA mode:
     ```css
     body {
       overscroll-behavior-y: none;
     }
     ```

6. **Performance Optimization for 60 FPS**:
   - Use CSS transforms instead of top/left for animations:
     ```css
     /* SLOW: causes reflow */
     .animated { left: 100px; top: 50px; }
     
     /* FAST: GPU-accelerated */
     .animated { transform: translate(100px, 50px); }
     ```
   - Debounce expensive operations:
     ```typescript
     import { debounce } from 'lodash-es';
     
     const onInputChange = debounce((value: string) => {
       // Expensive operation
     }, 150);
     ```
   - Use `requestAnimationFrame` for smooth animations
   - Lazy load images and components:
     ```svelte
     <script>
       import { onMount } from 'svelte';
       
       let visible = false;
       onMount(() => {
         const observer = new IntersectionObserver((entries) => {
           if (entries[0].isIntersecting) {
             visible = true;
             observer.disconnect();
           }
         });
         observer.observe(element);
       });
     </script>
     ```

7. **PWA Installation Prompt** (iOS-specific):
   - Add iOS installation instructions:
     ```svelte
     {#if isIOS && !isStandalone}
       <div class="install-prompt">
         <p>Install Thousand Worlds for the best experience:</p>
         <ol>
           <li>Tap the Share button <svg>...</svg></li>
           <li>Select "Add to Home Screen"</li>
           <li>Tap "Add"</li>
         </ol>
       </div>
     {/if}
     
     <script>
       const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
       const isStandalone = window.matchMedia('(display-mode: standalone)').matches;
     </script>
     ```
   - Add apple-touch-icon with proper sizing:
     ```html
     <link rel="apple-touch-icon" sizes="180x180" href="/icons/apple-touch-icon.png">
     <link rel="apple-touch-startup-image" href="/icons/launch-1125x2436.png" 
           media="(device-width: 375px) and (device-height: 812px) and (-webkit-device-pixel-ratio: 3)">
     ```

8. **Dark Mode Adaptation**:
   - Respect iOS dark mode preference:
     ```css
     @media (prefers-color-scheme: dark) {
       :root {
         --bg-color: #000;
         --text-color: #fff;
         --border-color: #1c1c1e;
       }
     }
     ```
   - Use native iOS colors when possible:
     ```css
     :root {
       --ios-blue: #007AFF;
       --ios-green: #34C759;
       --ios-red: #FF3B30;
       --ios-orange: #FF9500;
       --ios-yellow: #FFCC00;
       --ios-gray: #8E8E93;
     }
     ```

9. **Accessibility Enhancements**:
   - Support Dynamic Type (iOS font scaling):
     ```css
     body {
       font-size: 16px;
       /* Scale with iOS text size preference */
       font: -apple-system-body;
     }
     ```
   - Add ARIA labels for all interactive elements
   - Support VoiceOver gestures:
     ```html
     <button aria-label="Send command" role="button">
       <svg aria-hidden="true">...</svg>
     </button>
     ```
   - High contrast mode support:
     ```css
     @media (prefers-contrast: high) {
       .button {
         border: 2px solid;
       }
     }
     ```

10. **Network Resilience** (for iPhone's background tab management):
    - Auto-reconnect WebSocket when app returns to foreground:
      ```typescript
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') {
          // Reconnect WebSocket
          gameWebSocket.reconnect();
        }
      });
      ```
    - Show connection status indicator
    - Cache last 50 commands for offline review:
      ```typescript
      const commandHistory = writable<string[]>([]);
      
      // Store in localStorage
      commandHistory.subscribe(value => {
        localStorage.setItem('commandHistory', JSON.stringify(value.slice(-50)));
      });
      ```

Test Requirements (100% coverage on real iPhone devices):
- iPhone SE (2022): All touch targets ≥44×44pt
- iPhone 12/13: Safe area insets properly applied
- iPhone 14 Pro/15 Pro: Dynamic Island clearance verified
- Input focus: Keyboard doesn't obscure text field
- Scroll performance: Maintains 60 FPS during rapid scrolling
- Haptic feedback: Works on all supported interactions
- PWA installation: Adds to home screen correctly
- Offline mode: Commands cached and viewable
- Dark mode: Switches automatically with iOS setting
- VoiceOver: All elements properly labeled
- Text zoom: UI doesn't break at 200% text size
- Safari: No zoom on input focus (font-size ≥16px)
- Momentum scroll: Works smoothly in all scrollable areas
- Pull-to-refresh: Disabled in standalone mode

Acceptance Criteria:
- All interactive elements meet iOS 44×44pt minimum touch target
- App uses safe area insets on all iPhone models (SE, 12-15 series)
- Keyboard management prevents UI obstruction
- Haptic feedback enhances user interactions
- PWA installs correctly with proper iOS styling
- UI remains responsive at 60 FPS on iPhone 12 and newer
- Dark mode switches automatically with iOS system preference
- VoiceOver navigation works for all screens
- No viewport zoom occurs on input focus
- Connection status visible and accurate
- App auto-reconnects when returning from background

Dependencies:
- Phase 10.1 (Core UI Components)
- iOS 14+ for safe area support
- Safari 14+ for PWA features

Files to Modify:
- `mud-platform-client/src/app.html` - Add meta tags and iOS-specific config
- `mud-platform-client/src/app.css` - Add safe area variables and iOS styles
- `mud-platform-client/static/manifest.json` - Add iOS-specific PWA config
- `mud-platform-client/src/lib/components/Layout/MobileLayout.svelte` - Safe area padding
- `mud-platform-client/src/lib/components/Input/CommandInput.svelte` - Keyboard management
- `mud-platform-client/src/lib/stores/haptic.ts` - Haptic feedback utilities
- `mud-platform-client/src/lib/stores/pwa.ts` - PWA install prompt logic
- `mud-platform-client/src/lib/services/websocket.ts` - Auto-reconnect on visibility change

Files to Create:
- `mud-platform-client/static/icons/apple-touch-icon-180x180.png` - iOS home screen icon
- `mud-platform-client/static/icons/apple-launch-1125x2436.png` - iPhone X/11 Pro splash
- `mud-platform-client/static/icons/apple-launch-1242x2688.png` - iPhone 11 Pro Max splash
- `mud-platform-client/static/icons/apple-launch-828x1792.png` - iPhone 11 splash
- `mud-platform-client/src/lib/utils/ios.ts` - iOS detection and utilities
- `mud-platform-client/tests/e2e/iphone_ux.spec.ts` - iPhone-specific E2E tests