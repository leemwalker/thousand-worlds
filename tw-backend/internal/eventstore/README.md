# Event Store Package

Event sourcing implementation for Thousand Worlds. Stores immutable events as the source of truth for game state.

## Core Concepts

**Event Sourcing**: Instead of storing current state, we store the sequence of events that led to that state. This enables:
- Complete audit trail
- Time-travel debugging (replay to any point)
- Event-driven architecture
- Rebuilding read models from events

## Architecture

```
eventstore/
├── types.go        # Event, EventType, AggregateType, Command interface
├── store.go        # EventStore interface + PostgresEventStore
├── projections.go  # Read model building from events
├── replay.go       # Event replay for state reconstruction
└── versioning.go   # Event schema versioning
```

---

## Types

### Event
```go
type Event struct {
    ID            string          // Unique event identifier
    EventType     EventType       // e.g., "PlayerMoved", "ItemPickedUp"
    AggregateID   string          // Entity this event belongs to
    AggregateType AggregateType   // e.g., "Player", "World", "NPC"
    Version       int64           // Monotonic version for ordering
    Timestamp     time.Time       // When event occurred
    Payload       json.RawMessage // Event-specific data
    Metadata      map[string]any  // Context (user, session, etc.)
}
```

### Command Interface
```go
type Command interface {
    AggregateID() string
    CommandType() string
}
```

---

## EventStore Interface

```go
type EventStore interface {
    AppendEvent(ctx, event Event) error
    GetEventsByAggregate(ctx, aggregateID string, fromVersion int64) ([]Event, error)
    GetEventsByType(ctx, eventType, fromTimestamp, toTimestamp) ([]Event, error)
    GetAllEvents(ctx, fromTimestamp, limit int) ([]Event, error)
}
```

**Implementations**:
- `PostgresEventStore` - Production implementation

---

## Features

### Projections (`projections.go`)
Build read-optimized views from event streams:
```go
projector := NewProjector(eventStore)
projector.Project("player-positions", handlePlayerMoved)
```

### Replay (`replay.go`)
Reconstruct state by replaying events:
```go
replayer := NewReplayer(eventStore)
state := replayer.ReplayTo(aggregateID, timestamp)
```

### Versioning (`versioning.go`)
Handle event schema evolution:
- Upcasters for old event formats
- Version metadata on events

---

## Usage Example

```go
event := Event{
    ID:            uuid.New().String(),
    EventType:     "PlayerMoved",
    AggregateID:   playerID,
    AggregateType: "Player",
    Version:       nextVersion,
    Timestamp:     time.Now(),
    Payload:       json.RawMessage(`{"x": 100, "y": 200}`),
}

err := store.AppendEvent(ctx, event)
```

## Testing

```bash
go test ./internal/eventstore/...
```
