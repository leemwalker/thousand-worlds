// Package eventstore implements event sourcing for the Thousand Worlds game.
//
// Event sourcing stores immutable events as the source of truth rather than
// current state. This enables:
//   - Complete audit trail of all changes
//   - Time-travel debugging (replay to any point)
//   - Event-driven architecture
//   - Rebuilding read models from events
//
// # Core Types
//
//   - Event: Immutable fact that occurred (ID, type, aggregate, payload)
//   - EventStore: Interface for storing and retrieving events
//   - PostgresEventStore: Production implementation using PostgreSQL
//
// # Usage
//
//	store := eventstore.NewPostgresEventStore(pool)
//
//	// Append an event
//	event := Event{
//	    ID:            uuid.New().String(),
//	    EventType:     "PlayerMoved",
//	    AggregateID:   playerID,
//	    AggregateType: "Player",
//	    Timestamp:     time.Now(),
//	    Payload:       json.RawMessage(`{"x": 100, "y": 200}`),
//	}
//	store.AppendEvent(ctx, event)
//
//	// Retrieve events for an aggregate
//	events, _ := store.GetEventsByAggregate(ctx, playerID, 0)
//
// # Features
//
//   - Projections: Build read-optimized views from event streams
//   - Replay: Reconstruct state by replaying events
//   - Versioning: Handle event schema evolution with upcasters
package eventstore
