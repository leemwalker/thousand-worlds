# Project: Local LLM MUD Backend
# Tech Stack: Go 1.23+, NATS JetStream, PostgreSQL (pgx/v5), MongoDB, Redis, Ollama.

## Core Architecture
1. **Microservices:** Independent services. NO synchronous HTTP calls between services. Use NATS Async Events.
2. **Event Sourcing:** State changes (Move, Attack, Speak) must be emitted as NATS events.
3. **AI Gateway:** All LLM interaction MUST go through the `ai-gateway` service. Never call Ollama directly from other services. Queue limit: 1 concurrent request (Local Hardware Constraint).


## Database Rules
1. **PostgreSQL (World State):** Use `pgx`. All spatial queries (radius, distance) must use PostGIS SQL functions (`ST_3DDWithin`).
2. **MongoDB (NPC Memory):** Use standard mongo-driver. Memories never delete; they compress from Text -> Summary -> Fact.
3. **Redis:** Use for player session tokens and short-term spatial caching. Also stores `WorldClock` state and `TimeMultiplier`.

## NPC Memory Rules (Critical)
- Memories NEVER delete. They "crystallize".
- **Decay Algo:** Memories fade from "Vivid" (Raw Text) -> "Fading" (Summary) -> "Crystallized" (Fact).
- **Persistence:** Interactions with REAL players decay 5x slower than NPC-NPC interactions.

## Coding Standards
- **Error Handling:** Use `fmt.Errorf("service.op: %w", err)` for wrapping.
- **Config:** All env vars must be loaded via a `config` package.
- **Concurrency:** Use `errgroup` for parallel tasks.
- **Logging:** Use `zerolog` with `zerolog.ConsoleWriter` for console output and `zerolog.JSONEncoder` for file output.
- **Metrics:** Use `prometheus` for metrics.
- **Tracing:** Use `opentelemetry` for tracing.
- **Security:** Use `OWASP` for security.
- **Testing:** Use `go test` for testing. Unit tests must be in the same package as the code they test. Test Driven Development is a must, write tests first and code against them.
- **Code Style:** Use `gofmt` for formatting.
- **Code Organization:** Use `go mod` for dependency management.

## 🚨 CRITICAL: Time & Simulation Rules
1. **No `time.Now()`:** All game logic must use `WorldTime`.
   - Each World has a `TimeMultiplier` (stored in Redis).
   - `VirtualTime = PreviousTime + (RealDelta * Multiplier)`.
2. **Lazy Simulation (The "Catch-Up" Pattern):**
   - Do NOT run a tick loop for empty worlds.
   - When a player enters a Dormant Zone, calculate `DeltaTime` and apply bulk updates (e.g., `GrowPlants(hoursPassed)`).
   - **Time Travel:** If `Multiplier > 1.0`, do NOT increase NATS message rate. Increase the `Delta` per tick.