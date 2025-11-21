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
- **Character Creation** - Customizable player characters with attributes, skills, and appearance
- **Coordinate-Based Movement** - Move in 10 directions (N, S, E, W, NE, SE, SW, NW, UP, DOWN) with stamina for repeated movements.
- **Stamina System** - Movement types (walk, run, sneak) drain stamina at different rates, regenerates when not moving and is based off the character's stats.
- **Inventory System** - Weight-limited inventory with item stacking, equipment slots, and durability.
- **Object Interaction** - Pick up, drop, use, examine items in the world.
- **Deep Crafting System** - Every item in the world was crafted, other than the first items spawned at creation, whether by a player or an NPC. Every item has at least one natural resource that must be gathered in order to craft it, unless it is the result of two or more items being combined.
- **Player Stats** - See the character attributes.
- **Skill Progression** - Skills improve with use and are used in combination with attributes to determine the success.
- **Perception System** - Player perception skill and natural attributes affects how much detail they see in area descriptions

### Attribugte System
- **Core Attributes** - Strength, Agility, Health, Wits, Wisdom, Charisma, Willpower
- **Secondary Attributes** - Health Points, Stamina, Focus, Mana
- **Sensory Attributes** - Sight, Hearing, Smell, Taste, Touch

### NPC (Non-Player Character) System
- **Genetic Trait Inheritance** - NPCs inherit traits from parents using Mendelian genetics with mutations (5% chance)
- **Physical Appearance Generation** - Traits determine how NPCs look (height, build, features) and NPCs attributes (core, secondary, and sensory)
- **Memory System** - NPCs remember conversations, events, observations, relationships, and knowledge (stored in MongoDB)
- **Memory Decay** - Memories lose clarity over time, simulating forgetting
- **Emotional Memory** - Memories have emotional weight (0.0-1.0) affecting how strongly they're remembered
- **Relationship Tracking** - NPCs track affection, trust, and fear toward players and other NPCs
- **Relationship Updates** - Interactions modify relationships (giving gifts increases affection, threats increase fear)
- **Desire Engine** - Rule-based AI prioritizes needs: Survival (hunger, thirst, sleep, safety) > Social (companionship, conversation) > Achievement (goals, tasks) > Pleasure/Exploration (hedonism or curiosity)
- **Personality-Driven Behavior** - Genetic traits and experiences shape personality, affecting decision-making
- **LLM Consultation** - NPCs use Ollama for dialogue, not decision making.
- **24/7 Simulation** - NPCs continue living and interacting with the world even when players are offline
- **NPC-to-NPC Interaction** - NPCs converse and interact with each other autonomously
- **Dialogue System** - NPC traits/desires/relationship/memories are used to determine their most likely dialogue, which is then passed to an LLM to enhance it to fit their personality and current mood.
- **Personality/Mood Speech Patterns** - Different personalities, relationships, and moods affect how NPCs speak (formal, casual, gruff, friendly, verbose, terse)

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
- **Local LLM (Ollama)** - Llama 3.1 8B for area descriptions and NPC dialogue (runs 24/7 on host machine)
- **Groq API** - Free tier Llama 3.1 70B for world creation interviews
- **$0 LLM Budget** - All ongoing gameplay uses free local LLM
- **Aggressive Caching** - 15-minute TTL on descriptions, 95%+ cache hit rate target
- **Fallback Templates** - If LLM fails, use template-based generation
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
- **Loki Logging** - Centralized structured logging with trace IDs
- **Jaeger Tracing** - Distributed request tracing across microservices
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

### Future Features (Roadiness)
- **Player-Driven Content** - Tools for players to create custom quests, items, and areas
- **Voice Commands** - Speech-to-text for command input
- **Procedural Quests** - LLM-generated dynamic quests based on world state
- **Advanced AI** - Vector similarity search for NPC memory retrieval
