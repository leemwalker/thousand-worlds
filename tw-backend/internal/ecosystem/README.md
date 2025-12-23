# Ecosystem Package

Wildlife simulation including geology, evolution, creature spawning, and pathfinding.

## Architecture

```
ecosystem/
├── geology.go       # Geological events and terrain evolution
├── evolution.go     # Species adaptation over time
├── simulation/      # Core world simulation engine and turn orchestration
├── spawner.go       # Creature population management
├── pathfinding.go   # A* pathfinding for entities
├── events.go        # Ecosystem event types
├── service.go       # Main ecosystem service
├── fitness.go       # Creature fitness calculations
└── state/           # World state management
```

---

## Components

### Geology (`geology.go`)
Simulates geological processes affecting terrain:

| Event Type | Effect |
|------------|--------|
| Continental Drift | Plate movement, mountain formation |
| Volcanic Winter | Temperature drop, ash coverage |
| Asteroid Impact | Crater creation, mass extinction |
| Ice Age | Glaciation, sea level drop |
| Flood Basalt | Lava flows, gas emissions |

```go
geology := NewWorldGeology(worldID, seed)
geology.Tick(deltaTime) // Advance simulation
events := geology.GetPendingEvents()
```

---

### Evolution (`evolution.go`)
Species adaptation based on environmental pressures:
- Trait inheritance from parent creatures
- Mutation based on environmental stress
- Fitness-based survival
- Population dynamics

---

### Spawner (`spawner.go`)
Manages creature populations:
- Biome-appropriate creature selection
- Population density limits
- Spawn point management
- Despawn for low-population areas

---

### Pathfinding (`pathfinding.go`)
A* pathfinding for NPCs and creatures:
- Terrain cost consideration
- Obstacle avoidance
- Path caching for performance

```go
path := FindPath(start, goal, worldGrid)
```

---

### Service (`service.go`)
Main ecosystem service that coordinates all subsystems:

```go
service := ecosystem.NewService(seed)

// Called every game tick
service.Tick()

// Get creatures in area
creatures := service.GetCreaturesInRadius(center, radius)
```

---

## State Management

The `state/` subdirectory manages:
- World state persistence
- Creature state tracking
- Population statistics

---

## Integration Points

- **Worldgen**: Uses biome data from `internal/worldgen`
- **NPC System**: Creature behaviors from `internal/npc`
- **Spatial**: Location queries via `internal/spatial`
- **Game Loop**: Ticked by game processor

## Testing

```bash
go test ./internal/ecosystem/...
```
