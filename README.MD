# Thousand Worlds MUD Platform

This repo contains the back and front end logic for the Thousand Worlds MUD Platform. This game uses a combination of modern infrastructure, code, and design to build a game where the world feels alive as you watch NPCs interact with a world where flora and fauna are also alive and evolving.

Each world should feel unique based off the questions asked by the LLM during the world creation process. The LLM will interview the player to design a custom world (theme, tech level, geography, culture) and then use procedural algorithms to generate terrain, biomes, resources, and points of interest based off of the custom world. The LLM will also generate a set of rules for the world, which will be used to guide the behavior of NPCs and the environment. The world, flora, and fauna will all be generated at a primitive level based off the info provided by the player and then time will be sped up until the world is fleshed out to the point of the NPCs being introduced. Time will be sped up again to allow for NPCs to make the world lived in, the first histories to be written, and cultures showing up.

Players can lock their worlds down so only people they invite can visit. If a world is left open, other players are able to join it and interact with the world. When no players are present the world is paused until players re-enter, at which point the time is sped up until the world is caught up to the present. This allows the world to feel like it's truly living without resources being constantly used.

# MUD Platform Features

## Core Gameplay Features

### World System
- **LLM-Guided World Creation** - Players interview with an LLM to design custom worlds (theme, tech level, geography, culture)
- **Procedural Generation** - LLM directs procedural algorithms to generate terrain, biomes, resources, and points of interest based off of the custom world.
- **Multiple Persistent Worlds** - MMO-style shared worlds that persist indefinitely and NPCs/Flora/Fauna act as if its a living world.
- **Multi-World Coordinate System** - Three world types supported:
  - **Spherical Planets**: Earth-like worlds with radius-based spherical coordinates
  - **Bounded Cubes**: Finite rectangular spaces (continents, buildings, dungeons)
  - **Infinite Worlds**: Unbounded exploration spaces
- **Cartesian Coordinates** - Meter-based X, Y, Z positioning for precise spatial queries
- **Inter-World Travel** - High-level players can travel between worlds; world creators can lock their worlds to prevent travel.
- **Coordinate-Based Spatial System** - Replaces traditional MUD rooms with 3D coordinates (X, Y, Z) for precise positioning


### World Generation
- **Geographic Generation** - Tectonic plates and other geological processes are simulated to generate terrain such as ocean, mountain ranges, continents, and islands.
- **Realistic Weather System** - Weather patterns are simulated based off of the geography of the simulated world, including the evaporation of surface waters and planetary winds. This offers weather that is realistic for each simulated world, helping it feel deeper and more living.
- **Dynamic World Evolution** - Flora and fauna evolution simulated and sped up as part of world creation based on environmental pressures, predation, and competition.


### Spatial & Environment
- **PostGIS Spatial Queries** - Efficient radius searches, line-of-sight calculations, and pathfinding using PostgreSQL with PostGIS
- **Dynamic Area Descriptions** - LLM-generated descriptions based on location, weather, time of day, nearby entities, and player perception
- **Adaptive Time System** - Day/night cycle (By default it runs at a 1:1 ratio but time can be sped up or slowed down by higher level players). Seasonal changes happen regularly, based off the world generation.


### Player Systems
- **Dual Character Creation** - Choose to inhabit existing NPC or generate new adult character
- **Point-Buy Customization** - 100-point system with scaling costs for generated characters
- **Genetic Variance** - Random trait modifiers (±5) affecting starting attributes
- **Background Questionnaire** - 3-5 questions determine starting skill bonuses for new characters
- **Species Diversity** - Different species have different base attribute distributions
- **Character Creation** - Customizable player characters with attributes, skills, and appearance
- **Coordinate-Based Movement** - Move in 10 directions (N, S, E, W, NE, SE, SW, NW, UP, DOWN) with stamina for repeated movements.
- **Stamina System** - Movement types (walk, run, sneak) drain stamina at different rates, regenerates when not moving and is based off the character's stats.
- **Inventory System** - Weight-limited inventory with item stacking, equipment slots, and durability.
- **Object Interaction** - Pick up, drop, use, examine items in the world.
- **Deep Crafting System** - Every item in the world was crafted, other than the first items spawned at creation, whether by a player or an NPC. Every item has at least one natural resource that must be gathered in order to craft it, unless it is the result of two or more items being combined.
- **Player Stats** - See the character attributes.
- **Skill Progression** - Skills improve with use and are used in combination with attributes to determine the success.
- **Perception System** - Player perception skill and natural attributes affects how much detail they see in area descriptions
- **Behavioral Drift Detection** - System tracks when inhabited NPCs act against their nature
- **Drift Consequences** - NPCs notice and react to personality changes (relationship impacts)

### Attribute System
- **Physical Attributes (5)** - Might (power), Agility (speed/coordination), Endurance (stamina), Reflexes (reactions), Vitality (life force)
- **Mental Attributes (5)** - Intellect (learning), Cunning (tactics), Willpower (fortitude), Presence (influence), Intuition (instinct)
- **Secondary Attributes (5)** - Health Points (Vitality×10), Stamina Points ((Endurance×7)+(Might×3)), Focus Points ((Intellect×6)+(Willpower×4)), Mana Points ((Intuition×6)+(Willpower×4)), Nerve ((Willpower×5)+(Presence×3)+(Reflexes×2))
- **Sensory Attributes (5)** - Sight, Hearing, Smell, Taste, Touch (affect perception and gameplay mechanics)

### Skills System
- **Use-Based Progression** - Skills improve through actual use, not point allocation
- **Five Skill Categories** - Combat, Crafting, Gathering, Utility, Social
- **Skill Checks** - d100 + skill + attribute modifier vs difficulty threshold
- **Critical Rolls** - Natural 1-5 (critical failure) and 96-100 (critical success)
- **Diminishing Returns** - Prevents grinding by reducing XP for repetitive actions
- **Skill Synergies** - Related skills provide minor bonuses to checks
- **Soft Caps** - Attribute-based thresholds make advancing harder beyond natural aptitude
- **Skill Requirements** - Recipes, equipment, and abilities require minimum skill levels
- **Quality Scaling** - Higher skills produce better crafts, more damage, better yields

### NPC (Non-Player Character) System
- **Genetic Trait Inheritance** - NPCs inherit traits from parents using Mendelian genetics with mutations (5% chance)
- **Physical Appearance Generation** - Traits determine how NPCs look (height, build, features) and NPCs attributes (core, secondary, and sensory)
- **Memory System** - NPCs remember conversations, events, observations, relationships, and knowledge (stored in MongoDB)
- **Memory Decay** - Memories lose clarity over time, simulating forgetting
- **Emotional Memory** - Memories have emotional weight (0.0-1.0) affecting how strongly they're remembered
- **Relationship Tracking** - NPCs track affection, trust, and fear toward players and other NPCs
- **Relationship Updates** - Interactions modify relationships (giving gifts increases affection, threats increase fear)
- **Behavioral Baseline Tracking** - NPCs track patterns (aggression, generosity, honesty, sociability, recklessness, loyalty)
- **Personality Drift Detection** - System calculates deviation when inhabited NPCs act out of character
- **Drift Response Tiers** - Subtle (comments), Moderate (questioning), Severe (intervention/alarm)
- **Desire Engine** - Rule-based AI prioritizes needs: Survival (hunger, thirst, sleep, safety) > Social (companionship, conversation) > Achievement (goals, tasks) > Pleasure/Exploration (hedonism or curiosity)
- **Personality-Driven Behavior** - Genetic traits and experiences shape personality, affecting decision-making
- **LLM Consultation** - NPCs use Ollama for dialogue, not decision making.
- **24/7 Simulation** - NPCs continue living and interacting with the world even when players are offline
- **NPC-to-NPC Interaction** - NPCs converse and interact with each other autonomously
- **Dialogue System** - NPC traits/desires/relationship/memories are used to determine their most likely dialogue, which is then passed to an LLM to enhance it to fit their personality and current mood.
- **Personality/Mood Speech Patterns** - Different personalities, relationships, and moods affect how NPCs speak (formal, casual, gruff, friendly, verbose, terse)
- **NPC Inhabitation** - Players can take over existing NPCs, inheriting their full history
- **Inhabitation Consequences** - Acting against NPC's nature damages relationships and reputation
- **Post-Inhabitation Recovery** - When player leaves, NPC AI reviews actions and experiences confusion for conflicting behaviors
- **Relationship Sensitivity** - Close relationships (family, spouses) notice drift immediately and react strongly

### Combat System
- **Real-Time Combat** - Continuous action with reaction time mechanics (prevents spamming)
- **Action Queue System** - Actions queued based on reaction time, processed in order
- **Reaction Times** - These are placeholders and will be adapted once we start game testing. Quick attack (800ms), Normal attack (1000ms), Heavy attack (1500ms), Defend (500ms), Flee (2000ms)
- **Stamina-Based Actions** - Combat actions consume stamina
- **Damage Calculation** - Damage is based off the player's skill with the weapon, the related attribute, and how well they roll on a d100.
- **Weapon Types** - Slashing, piercing, bludgeoning with different armor effectiveness
- **Status Effects** - Poison (damage over time), Stun (can't act), Slow (increased reaction time), Bleed, Buffs, Debuffs
- **Effect Stacking** - Some effects stack (poison up to 5), others don't (stun extends duration)
- **Combat Logging** - Event sourcing records all combat actions for replay and analysis

### Economy & Crafting
- **World-Driven Economy** - NPCs harvest/forage/mine resources and trade them with players or NPCs who craft them and trade those items.
- **Currency System** - NPCs will accept trades in barter or whatever the currency is for that world.
- **NPC Merchants** - Actual inventory based off their local economy and trade.
- **Crafting System** - Recipes, resource gathering, skill-based quality
- **Crafting Stations** - Different stations for different crafts (forge, alchemy table, etc.)
- **Item Quality Tiers** - Quality depends on crafter skill level
- **Resource Distribution** - Procedurally placed resources (ores, plants, etc.) based on biome

## Technical Features

### Architecture
- **Microservices Backend** - Go-based services (Auth, World, Player, Spatial, NPC, Combat) communicating via NATS
- **Event Sourcing** - All state changes stored as immutable events in PostgreSQL
- **CQRS Pattern** - Separate read models optimized for queries
- **Progressive Web App (PWA)** - SvelteKit frontend optimized for mobile use
- **Service Worker Caching** - Cached assets and API responses
- **Kubernetes Orchestration** - k3s locally, EKS on AWS for production

### Security
- **Secure Authentication** - JWT tokens with AES-256 encryption, stored in memory (never localStorage)
- **Password Hashing** - Argon2id with 64MB memory, 3 iterations, 4 parallelism
- **Session Management** - Redis-backed sessions with 24-hour TTL
- **Rate Limiting** - Prevents abuse (5 login attempts/min, 1 world creation/hour, etc.)

### Data Storage
- **PostgreSQL with PostGIS** - Primary database for entities, events, spatial data
- **Redis** - Caching, session management, pub/sub for real-time updates
- **MongoDB** - NPC memory storage (conversations, experiences)
- **MinIO (S3-compatible)** - Data lake for cold storage and analytics
- **Event Store** - Append-only event log for complete audit trail and time-travel debugging

### AI Integration
- **Local LLM (Ollama)** - Llama 3.1 8B for world creation, area descriptions, and NPC dialogue (runs 24/7 on host machine)
- **Aggressive Caching** - 15-minute TTL on descriptions, 95%+ cache hit rate target
- **Context-Aware Prompts** - LLM receives location, weather, time, entities, player perception

### Performance & Scalability
- **Spatial Indexing** - PostGIS GIST indexes for fast nearby entity queries (< 30ms target)
- **Multi-Level Caching** - Memory (L1) → Redis (L2) → Database (L3)
- **Connection Pooling** - Efficient database connection management
- **Read Replicas** - Database read scaling for high player counts
- **WebSocket Support** - Real-time bidirectional communication (Phase 8 migration from HTTP polling)
- **Horizontal Scaling** - Stateless microservices can scale independently

### Monitoring & Operations
- **Prometheus Metrics** - Request latency, error rates, cache hit rates, NPC simulation FPS
- **Grafana Dashboards** - Real-time visualization of system health
- **Health Checks** - Each service exposes /health endpoint
- **Zero Downtime Deployments** - Rolling updates in Kubernetes

### Developer Experience
- **Test-Driven Development (TDD)** - Write tests first, then implementation
- **80%+ Code Coverage** - Required for all services
- **Infrastructure as Code** - Terraform for all production infrastructure
- **Docker Compose Fallback** - Alternative to k3s for simpler local development
- **Makefile Commands** - `make up`, `make down`, `make test`, `make deploy`
- **AI-Assisted Development** - Windsurf IDE with Cascade for pair programming with LLM
- **Comprehensive Documentation** - `.windsurf/instructions.md` files for each repo
- **Behavioral Drift Monitor UI** - Real-time display of personality changes and relationship impacts
- **Action Warning System** - Alerts players before taking actions that conflict with inhabited NPC nature
- **World Creator Drift Settings** - Configure drift tolerance (strict RP to no tracking) per world

### Future Features (Roadiness)
- **Player-Driven Content** - Tools for players to create custom quests, items, and areas
- **Voice Commands** - Speech-to-text for command input
- **Procedural Quests** - LLM-generated dynamic quests based on world state
- **Advanced AI** - Vector similarity search for NPC memory retrieval

# Development Roadmap

## Phase 0: Foundation & Infrastructure (4-6 weeks) - COMPLETED
### Goal: Establish core architecture, testing patterns, and data persistence before building features.

### 0.1 Event Sourcing & Data Layer (2 weeks) ✅
Design event schema (command/event separation)
Implement PostgreSQL event store with append-only events table
Build event replay engine (rewind, fast-forward, replay from point-in-time)
Create event versioning strategy for schema migrations
Implement CQRS read models for common queries
Write comprehensive tests for event persistence and replay (target: 85%+ coverage)

### 0.2 Authentication & Security (1 week) ✅
Implement JWT token generation/validation with AES-256 encryption
Add Argon2id password hashing (64MB memory, 3 iterations, 4 parallelism)
Build Redis session management with 24-hour TTL
Implement rate limiting middleware (5 login attempts/min, configurable per endpoint)
Add security tests (brute force, token tampering, session hijacking)

### 0.3 Monitoring & Observability (1 week) ✅
Set up Prometheus metrics collection (request latency, error rates, cache hits)
Create Grafana dashboards for system health
Add /health endpoints to all services
Implement structured logging with correlation IDs across services
Add performance benchmarks for critical paths (spatial queries < 30ms target)

### 0.4 Spatial Foundation & Multi-World System (1-2 weeks) ✅
Design PostgreSQL schema with PostGIS extensions
Migrated to Cartesian GEOMETRY (meter-based coordinates in X, Y, Z)
Created worlds table with shape types (spherical planets, bounded cubes, infinite worlds)
- World properties: shape, radius (for spherical), bounds (for cubes), independent coordinate spaces
Implement World Repository (Create, Get, List, Update, Delete) - 5/5 tests ✅
Updated Spatial Repository for Cartesian coordinates - 5/5 tests ✅
Build spatial repository with GIST indexes
Create tests for radius queries, bounding box searches, distance calculations
Implement basic coordinate-based movement in 10 directions
Add collision detection and boundary validation
Seed 3 example worlds (Earth-like sphere, continent-sized bounded cube, mansion-sized cube)
**Coverage: 69.6%**

**Deferred to Future Phases (Optional):**
- Spherical projection utilities (lat/lon ↔ Cartesian conversion) - See Phase 1.4
- Spherical wrapping logic (pole crossing, seamless movement) - See Phase 1.4

### Acceptance Criteria:
Can store/query spatial data with sub-meter precision ✅
Multiple world types supported (spherical, bounded, infinite) ✅
Event store can replay 10,000 events in < 5 seconds ✅
All services have health checks and metrics ✅
Authentication flow works end-to-end with tests ✅
World Repository full CRUD with tests ✅

## Phase 1: World Ticker & Time System (2-3 weeks)
### Goal: Establish the heartbeat of the game world with proper time dilation support.

### 1.1 World Service Core (1 week)
Implement world-service microservice
Create world registry (track active/paused worlds)
Build ticker manager (spawn/stop tickers per world)
Add world state persistence (running, paused, tick count, game time)
Write tests for ticker lifecycle management

### 1.2 Time Dilation & Tick Broadcast (1 week)
Implement configurable dilation factor (default: 1.0, range: 0.1-100.0)
Build tick loop: GameTime += RealDelta * DilationFactor
Broadcast world.tick events to NATS (includes worldID, tickNumber, gameTime, realTime)
Add tick rate configuration (default: 10 Hz)
Implement world pause/resume with catch-up fast-forwarding
Test multi-world scenarios (different dilation factors running simultaneously)

### 1.3 Day/Night & Seasons (1 week)
Calculate sun position based on game time
Implement seasonal changes (configurable cycle length per world)
Add day/night cycle state to tick events
Create time-of-day descriptors (dawn, morning, noon, dusk, night, etc.)
Test time progression over simulated months

### 1.4 Spherical World Utilities (Optional - As Needed)
Implement spherical projection utilities:
- Lat/Lon to Cartesian (X, Y, Z) conversion
- Cartesian to Lat/Lon conversion
- Great circle distance calculations
Build spherical wrapping logic:
- Pole crossing detection and seamless transitions
- Longitude wrapping at ±180°
- Maintain correct coordinate space when crossing boundaries
Add tests for projection accuracy and wrapping edge cases
**Note**: Only implement when first spherical world requires these features

### Acceptance Criteria:
Multiple worlds can run with independent tick rates
Time dilation accurately scales game time vs real time
Paused worlds can fast-forward to catch up when resumed
Day/night and seasonal state correctly calculated
(Optional) Spherical utilities work correctly if implemented

## Phase 2: Player Core Systems (4-5 weeks)
### Goal: Enable basic player interaction with the world.

### 2.1 Character System (1-2 weeks)
Design attribute schema (10 core + 5 sensory)
- **Physical Attributes (5)**: Might, Agility, Endurance, Reflexes, Vitality (1-100 scale)
- **Mental Attributes (5)**: Intellect, Cunning, Willpower, Presence, Intuition (1-100 scale)
- **Sensory Attributes (5)**: Sight, Hearing, Smell, Taste, Touch (1-100 scale)
- **Secondary Attributes (5)**: HP (Vitality×10), Stamina ((Endurance×7)+(Might×3)), Focus ((Intellect×6)+(Willpower×4)), Mana ((Intuition×6)+(Willpower×4)), Nerve ((Willpower×5)+(Presence×3)+(Reflexes×2))
Implement dual character creation paths:
- **Path 1: Inhabit Existing NPC** - Browse and select from eligible NPCs (adults with 5+ relationships, 1+ game-year lived)
- **Path 2: Generate New Adult** - Species selection, genetic baseline generation, point-buy customization (100 points with scaling costs), background questionnaire
Build species base attribute templates (Human, Dwarven, Elven, etc.)
Implement genetic variance system (±(1d10-5) per attribute)
Create point-buy system with cost scaling (1 point up to +10, 2 points +11 to +20, 3 points +21 to +30)
Add character persistence (event-sourced character state)
Create tests for attribute validation, point-buy mechanics, and constraints
Build NPC browser UI (filter by species, location, skills, behavioral baseline)
Implement behavioral baseline snapshot on NPC inhabitation
Create tests for both character creation paths

### 2.2 Stamina & Movement (1 week)
Implement stamina pool (derived from Endurance and Might)
Build movement system (walk/run/sneak with different stamina costs)
Add stamina regeneration (rate based on Endurance and Vitality)
Prevent movement when stamina depleted
Test stamina edge cases (negative stamina, rapid movement spam)

### 2.3 Inventory System (1-2 weeks)
Design item schema (weight, stack size, durability, properties)
Implement weight-limited inventory (capacity based on Might)
Add equipment slots (weapon, armor, accessories)
Build item durability system (degrades with use)
Create pickup/drop/use/examine item commands
Test inventory edge cases (overfilled, negative weight, stack overflow)

### Acceptance Criteria:
Players can create characters via both inhabit and generate paths
Point-buy system enforces costs and caps correctly
NPC inhabitation preserves full history and relationships
Movement consumes stamina correctly and regenerates when resting
Inventory respects weight limits and item properties
All systems emit events for audit trail

## Phase 2.5: Skills & Progression System (2-3 weeks)
### Goal: Implement use-based skill advancement that drives quality in crafting, combat, and perception.

### 2.5.1 Core Skills Framework (1 week)
Design skill schema (name, category, current value 0-100, experience points)
Implement skill categories:
- **Combat Skills**: Slashing, Piercing, Bludgeoning, Defense, Dodge
- **Crafting Skills**: Smithing, Alchemy, Carpentry, Tailoring, Cooking
- **Gathering Skills**: Mining, Herbalism, Logging, Hunting, Fishing
- **Utility Skills**: Perception, Stealth, Climbing, Swimming, Navigation
- **Social Skills**: Persuasion, Intimidation, Deception, Bartering
Create skill persistence (event-sourced skill state)
Build skill cap system (soft caps at attribute-based thresholds)
Test skill initialization and storage

### 2.5.2 Experience & Advancement (1 week)
Implement use-based experience gain (XP per action)
Calculate XP curves: `XP_needed = baseXP * (skillLevel^1.5)` (gets harder as you advance)
Add diminishing returns for repetitive actions (prevents grinding)
Build skill check system: `d100 + skill + (relevantAttribute/5)` vs difficulty threshold
Create critical success/failure on natural rolls (1-5 = crit fail, 96-100 = crit success)
Add skill synergy bonuses (related skills provide minor boosts)
Test advancement rates and balance

### 2.5.3 Skill Integration Points (1 week)
**Combat Integration**: Weapon skills modify damage calculation
**Crafting Integration**: Skill determines item quality tier and success rate
**Gathering Integration**: Skill affects yield quantity and rare resource chance
**Perception Integration**: Skill level determines area description detail depth
**Movement Integration**: Stealth skill reduces detection range, Climbing enables vertical movement
Add skill requirements for recipes/equipment (minimum skill to use effectively)
Implement skill-based unlocks (new recipes, abilities at milestone levels)
Test all integration points for balance

### Acceptance Criteria:
Skills increase through use with diminishing returns
Skill checks properly combine skill + attribute + random roll
All major systems (combat, crafting, gathering) respect skill levels
Skill progression feels rewarding without being grindable
Skills persist correctly via event sourcing

## Phase 3: NPC Memory & Relationships (5-7 weeks)
### Goal: Build the foundation for "living" NPCs with memory and emotions.

### 3.1 Memory Storage & Retrieval (2 weeks)
Design MongoDB memory schema (observation, conversation, event, relationship)
Implement memory creation/storage in NPCMemoryRepository
Add memory tagging (entities, locations, emotions, keywords)
Build memory retrieval by type, timeframe, entity, emotion
Add memory importance scoring (recency + emotional weight + access frequency)
Test memory CRUD operations and complex queries

### 3.2 Memory Decay & Rehearsal (1 week)
Implement linear decay over time (clarity degrades)
Add rehearsal bonus (accessed memories decay slower)
Build memory corruption system (details become fuzzy/altered on access)
Calculate memory retention: retention = baseRetention * (1 - decayRate * timeSinceCreation) * (1 + rehearsalBonus * accessCount)
Run decay simulation tests over simulated years

### 3.3 Relationship System (1-2 weeks)
Create relationship schema (affection, trust, fear scales -100 to +100)
Implement relationship updates based on interactions
Add relationship modifiers (gifts increase affection, threats increase fear)
Build relationship decay over time without interaction
Implement behavioral baseline tracking in relationship memories
- Track patterns: aggression, generosity, honesty, sociability, recklessness, loyalty (0.0-1.0)
- Store last 20 interactions for drift calculation
Add drift detection system (compare player actions to NPC's historical baseline)
Create NPC reactions to behavioral drift (subtle, moderate, severe)
Build relationship modifiers based on drift severity (-25 to +25)
Test relationship edge cases (max/min values, conflicting emotions)
Test drift calculation accuracy and NPC response appropriateness

### 3.4 Emotional Memory Weighting (1-2 weeks)
Add emotional intensity to memories (0.0-1.0 scale)
Boost memory importance for high-emotion events
Implement emotion-triggered memory recall (similar emotions trigger related memories)
Create emotion types (joy, anger, fear, sadness, surprise, disgust)
Test memory recall with emotional context

### Acceptance Criteria:
NPCs can store, retrieve, and decay memories over time
Rehearsal correctly prevents memory loss
Relationships update properly based on interactions
High-emotion memories are more persistent and easier to recall
Behavioral drift detection identifies when inhabited NPCs act out of character
NPCs respond appropriately to personality changes with dialogue and relationship updates

## Phase 4: NPC Genetics & Appearance (2-3 weeks)
### Goal: NPCs inherit traits from parents and have unique appearances/attributes.

### 4.1 Genetic System (1-2 weeks)
Design gene schema (dominant/recessive alleles for each trait)
Implement Mendelian inheritance (Punnett square calculations)
Add mutation system (5% chance per gene)
Create trait-to-attribute mapping (genes influence Strength, Agility, etc.)
Build trait-to-appearance mapping (height, build, hair/eye color, features)
Test inheritance patterns across multiple generations
### 4.2 Appearance Generation (1 week)
Generate physical description from genetic traits
Implement appearance variation within genetic constraints
Add age-based appearance changes (child → adult → elder)
Create appearance descriptors (tall/short, muscular/lean, handsome/plain)
Test appearance diversity and inheritance consistency

### Acceptance Criteria:
NPC children inherit traits from both parents following Mendelian genetics
Mutations occur at expected 5% rate
Appearance generation is deterministic from genes
Attributes are influenced by genetic traits

## Phase 5: NPC AI & Desire Engine (3-4 weeks)
### Goal: NPCs behave autonomously based on needs, personality, and context.

### 5.1 Desire Engine (2 weeks)
Implement need hierarchy (Survival > Social > Achievement > Pleasure)
Add survival needs (hunger, thirst, sleep, safety with 0-100 scales)
Build social needs (companionship, conversation, affection)
Create achievement goals (tasks, objectives, long-term plans)
Add pleasure/exploration drives (curiosity, hedonism)
Calculate desire priorities: priority = needUrgency * personalityWeight
Test desire switching based on need urgency
### 5.2 Personality System (1 week)
Define personality traits (extraversion, conscientiousness, openness, agreeableness, neuroticism)
Derive personality from genetic traits + life experiences
Add personality influence on decision-making
Implement mood system (temporary emotional state affecting behavior)
Test personality consistency across decisions
### 5.3 NPC-to-NPC Interaction (1-2 weeks)
Build conversation initiation logic (based on proximity, relationships, needs)
Implement basic conversation flow (greeting, topic selection, response)
Add relationship updates from conversations
Create memory formation during interactions
Test autonomous NPC interactions without player involvement

### Acceptance Criteria:
NPCs prioritize actions based on needs and personality
NPCs initiate conversations with each other
Conversations update relationships and create memories
NPC behavior is emergent and not scripted

## Phase 6: LLM Integration for Dialogue (2-3 weeks)
### Goal: Connect NPC decision-making to Ollama for natural dialogue generation.

### 6.1 Ollama Prompt Engineering (1 week)
Design dialogue prompts (include NPC personality, memories, relationships, mood, context)
Add behavioral drift information to prompts for inhabited NPCs
- Include original personality baseline
- Include current drift metrics
- Generate concerned/curious/alarmed dialogue based on drift severity
Implement prompt template system
Add response parsing and validation
Build dialogue caching (15-minute TTL)
Test prompt quality and response coherence
Test drift-aware dialogue generation

### 6.2 Dialogue Request Flow (1 week)
Integrate desire engine output with dialogue generation
Send NPC state + player input + drift data to ai-gateway
Process LLM response and extract dialogue + emotional reaction
Update NPC memory and relationships post-conversation (including drift observations)
Implement fallback responses if LLM fails

### 6.3 Performance Optimization (1 week)
Target 95%+ cache hit rate for area descriptions
Batch dialogue requests where possible
Monitor Ollama CPU/RAM usage (target: < 80% sustained)
Add request queueing if Ollama overloaded
Test concurrent dialogue performance

### Acceptance Criteria:
NPCs speak in character based on personality and mood
Dialogue reflects NPC's memories and relationships
NPCs acknowledge and react to behavioral changes in inhabited characters
Cache hit rate exceeds 90% during normal gameplay
Ollama can handle 5-10 concurrent players

## Phase 7: Combat System (2-3 weeks)
### Goal: Enable real-time combat with reaction times and status effects.

### 7.1 Action Queue System (1 week)
Implement action queue (FIFO based on reaction time)
Calculate reaction time from character stats: baseTime * (1 - (Agility/100) * 0.3)
Add action types (attack, defend, flee, use item)
Prevent action spam (enforce minimum reaction time)
Test queue ordering with multiple combatants
### 7.2 Damage & Weapons (1 week)
Implement damage calculation: baseDamage * skillModifier * attributeModifier * d100
Add weapon types (slashing, piercing, bludgeoning)
Build armor effectiveness vs weapon types
Create critical hit system (natural 95+ on d100)
Test damage variance and balance
### 7.3 Status Effects (1 week)
Implement poison (DoT, stacks up to 5)
Add stun (skip turns, extends duration)
Build slow effect (increased reaction time)
Create bleed (DoT, reduces on movement)
Add buff/debuff system (temporary stat modifications)
Test effect interactions and edge cases

### Acceptance Criteria:
Combat actions respect server-enforced reaction times
Damage calculations are balanced and skill-based
Status effects stack/extend correctly
All combat actions are event-sourced

## Phase 8: World Generation (6-8 weeks)
### Goal: Generate unique, procedurally-created worlds with LLM-guided customization.

### 8.1 LLM World Interview (2 weeks)
Design interview questions (theme, tech level, geography, culture)
Build conversation flow with Ollama
Extract structured data from LLM responses
Create world configuration schema
Validate and store world parameters
### 8.2 Geographic Generation (2-3 weeks)
Implement tectonic plate simulation (continental vs oceanic)
Generate heightmap (mountains, valleys, plains)
Create ocean/land distribution
Add river generation (erosion pathfinding)
Build biome assignment (based on elevation, latitude, moisture)
Test geographic realism and variation
### 8.3 Weather Simulation (1-2 weeks)
Calculate evaporation rates (temperature + water proximity)
Implement planetary wind patterns (Hadley cells)
Build precipitation simulation (moisture + elevation)
Add weather state (clear, cloudy, rain, snow, storm)
Test weather consistency with geography
### 8.4 Flora/Fauna Evolution (1-2 weeks)
Design species schema (traits, diet, habitat, population)
Implement survival mechanics (food chain, predation, competition)
Build evolutionary pressure system (environmental adaptation)
Time-accelerate evolution during world creation
Generate initial species diversity
Test ecosystem stability

### Acceptance Criteria:
Worlds feel unique based on LLM interview
Geography is realistic and varied
Weather patterns match biomes and seasons
Flora/fauna ecosystems are balanced

## Phase 9: Crafting & Economy (3-4 weeks)
### Goal: Enable resource gathering, crafting, and NPC-driven economy.

### 9.1 Resource Distribution (1 week)
Implement procedural resource placement (ores, plants, wood)
Add biome-specific resource types
Create resource node schema (quantity, regeneration rate)
Build harvesting mechanics (skill-based yield)
Test resource availability and balance
### 9.2 Tech Trees & Recipes (2 weeks)
Design generic tech trees (primitive → medieval → industrial → modern → futuristic)
Customize tech trees per world (based on world tech level)
Create recipe schema (inputs, outputs, required tools, skill level)
Implement recipe discovery system
Test crafting progression and dependencies
### 9.3 NPC Economy (1-2 weeks)
Implement NPC resource gathering (autonomous harvesting)
Build NPC crafting (produce goods for trade)
Add NPC merchant inventory (based on local economy)
Create dynamic pricing (supply/demand)
Implement barter system
Test economic simulation and inflation prevention

### Acceptance Criteria:
Resources are distributed realistically across biomes
Crafting recipes form logical tech trees
NPC merchants have realistic inventory
Economy feels dynamic and player actions matter

## Phase 10: Frontend UI (5-7 weeks)
### Goal: Build an intuitive, mobile-optimized interface for gameplay.

### 10.1 Core UI Components (2-3 weeks)
Design UI mockups (command input, output log, map, stats)
Build command parser (natural language → game commands)
Create output formatting (color-coded text, entity highlighting)
Implement map visualization (2D top-down with fog of war)
Add inventory/character sheet UI
Build behavioral drift monitor display:
- Show original personality baseline (bar graphs)
- Display current behavior patterns
- Highlight drift warnings (⚠️ indicators)
- List relationship changes due to drift
- Show recent NPC comments about behavior changes
Add warning system for actions that conflict with inhabited NPC nature
Test mobile responsiveness

### 10.2 Real-Time Updates (1 week)
Migrate from HTTP polling to WebSockets
Implement NATS event subscription in frontend
Add real-time combat updates
Build notification system (other players nearby, environmental changes, drift warnings)
Test WebSocket reconnection and reliability

### 10.3 PWA Features (1-2 weeks)
Implement service worker for offline caching
Add "Add to Home Screen" support
Create offline mode (read-only, sync when reconnected)
Build push notifications (combat alerts, server events, relationship warnings)
Test installation flow on iOS/Android

### 10.4 UX Polish (1-2 weeks)
Add command history (up/down arrows)
Implement auto-complete for commands
Create help system (contextual suggestions)
Add accessibility features (screen reader support, keyboard navigation)
Build character creation tutorial for both paths (inhabit vs generate)
Test user experience with real players

### Acceptance Criteria:
UI is responsive and works on mobile devices
Real-time updates are seamless
Behavioral drift monitor clearly communicates personality changes
Warning system helps players make informed decisions about NPC behavior
PWA installs successfully on mobile
User experience is intuitive for new players

## Phase 11: Multiplayer & Scaling (3-4 weeks)
### Goal: Support 10 players initially, design for 1000s.

### 11.1 Multi-Player Interactions (1-2 weeks)
Implement player-to-player visibility (proximity-based)
Add player chat (local, whisper, world)
Build party system (group invites, shared objectives)
Create PvP flag system (opt-in combat)
Test interactions with 10 concurrent players
### 11.2 Performance Optimization (1-2 weeks)
Add Redis caching for hot data (player positions, active entities)
Implement database connection pooling
Create read replicas for query offloading
Optimize spatial queries (batch nearby entity lookups)
Benchmark with 50+ simulated players
### 11.3 World Lock & Permissions (1 week)
Implement world ownership
Add world privacy settings (public/invite-only)
Build invite system (generate invite codes)
Create permission levels (owner, admin, guest)
Test access control

### Acceptance Criteria:
10 players can interact smoothly in the same world
Spatial queries remain < 50ms with 10 players
World privacy settings work correctly
System remains stable under load

## Phase 12: Testing & Launch Prep (2-3 weeks)
### Goal: Ensure stability and polish before initial launch.

### 12.1 Comprehensive Testing (1 week)
Run full integration test suite
Perform load testing (100 simulated players)
Test all event replay scenarios
Validate all 80%+ code coverage requirements
Fix critical bugs
### 12.2 Documentation (1 week)
Write player guide (getting started, commands, mechanics)
Create admin documentation (deploy, configure, monitor)
Document API endpoints for future integrations
Build troubleshooting guide
### 12.3 Initial Deployment (1 week)
Set up production k3s cluster
Configure monitoring and alerting
Create backup/restore procedures
Perform dry-run deployment
Launch with solo playtesting

### Acceptance Criteria:
All systems tested and stable
Documentation is comprehensive
Production environment is ready
Backup/restore procedures validated

## Future Enhancements (Post-Launch)
Player-Driven Content: Tools for custom quests, items, areas
Voice Commands: Speech-to-text integration
Procedural Quests: LLM-generated dynamic quests
Vector Memory Search: Advanced NPC memory retrieval using embeddings
AWS Migration: Lift-and-shift to EKS when scaling beyond 100 players
Advanced Graphics: Optional 3D visualization layer

## Estimated Timeline Summary

**Phase 0-2 (Foundation, Ticker, Player Core)**: 10-14 weeks

**Phase 2.5 (Skills)**: 2-3 weeks

**Phase 3-5 (NPC Systems)**: 10-14 weeks

**Phase 6-7 (Dialogue & Combat)**: 4-6 weeks

**Phase 8-9 (World Gen & Economy)**: 9-12 weeks

**Phase 10-11 (Frontend & Multiplayer)**: 8-11 weeks

**Phase 12 (Testing & Launch)**: 2-3 weeks

**Total: 45-63 weeks (11-16 months)** of focused development at 30-40 hrs/week with AI assistance.

**Milestone 1 (Solo Playable)**: End of Phase 7 (~7 months)

**Milestone 2 (10-Player Ready)**: End of Phase 11 (~12 months)

**Milestone 3 (Public Launch)**: End of Phase 12 (~14 months)

# THIS FILE SHOULD NEVER BE EDITED BY AI