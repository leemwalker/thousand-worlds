package eventstore

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test setup
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Assume DB is running via docker-compose on localhost:5432
	// Default credentials from docker-compose.yml
	connString := "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"

	// Allow overriding via env var
	if env := os.Getenv("TEST_DB_URL"); env != "" {
		connString = env
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Ping to verify connection
	err = pool.Ping(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	// Clean up events table before test
	_, err = pool.Exec(ctx, "TRUNCATE TABLE events")
	require.NoError(t, err)

	return pool
}

func TestPostgresEventStore_AppendEvent(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresEventStore(pool)
	ctx := context.Background()

	t.Run("stores event successfully", func(t *testing.T) {
		event := Event{
			ID:            "123e4567-e89b-12d3-a456-426614174000",
			EventType:     "WorldCreated",
			AggregateID:   "world-1",
			AggregateType: "World",
			Version:       1,
			Timestamp:     time.Now().UTC(),
			Payload:       json.RawMessage(`{"name":"Test World"}`),
		}

		err := store.AppendEvent(ctx, event)
		assert.NoError(t, err)

		// Verify it's in the DB
		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM events WHERE id=$1", event.ID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("enforces append-only (duplicate version)", func(t *testing.T) {
		event1 := Event{
			ID:            "123e4567-e89b-12d3-a456-426614174001",
			EventType:     "PlayerMoved",
			AggregateID:   "player-1",
			AggregateType: "Player",
			Version:       1,
			Timestamp:     time.Now().UTC(),
			Payload:       json.RawMessage(`{}`),
		}
		err := store.AppendEvent(ctx, event1)
		require.NoError(t, err)

		// Try to append same version for same aggregate
		event2 := Event{
			ID:            "123e4567-e89b-12d3-a456-426614174002",
			EventType:     "PlayerAttacked",
			AggregateID:   "player-1",
			AggregateType: "Player",
			Version:       1, // Duplicate version!
			Timestamp:     time.Now().UTC(),
			Payload:       json.RawMessage(`{}`),
		}
		err = store.AppendEvent(ctx, event2)
		assert.Error(t, err) // Should fail due to unique constraint
	})
}

func TestPostgresEventStore_GetEventsByAggregate(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresEventStore(pool)
	ctx := context.Background()

	// Seed events
	aggID := "agg-1"
	events := []Event{
		{ID: "123e4567-e89b-12d3-a456-426614174003", EventType: "T1", AggregateID: aggID, AggregateType: "A", Version: 1, Timestamp: time.Now().UTC(), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174004", EventType: "T2", AggregateID: aggID, AggregateType: "A", Version: 2, Timestamp: time.Now().UTC(), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174005", EventType: "T3", AggregateID: aggID, AggregateType: "A", Version: 3, Timestamp: time.Now().UTC(), Payload: json.RawMessage(`{}`)},
	}

	for _, e := range events {
		err := store.AppendEvent(ctx, e)
		require.NoError(t, err)
	}

	t.Run("retrieves all events", func(t *testing.T) {
		got, err := store.GetEventsByAggregate(ctx, aggID, 0)
		assert.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, events[0].ID, got[0].ID)
		assert.Equal(t, events[2].ID, got[2].ID)
	})

	t.Run("retrieves from version", func(t *testing.T) {
		got, err := store.GetEventsByAggregate(ctx, aggID, 2)
		assert.NoError(t, err)
		assert.Len(t, got, 2) // Version 2 and 3
		assert.Equal(t, events[1].ID, got[0].ID)
	})
}

func TestPostgresEventStore_GetEventsByType(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresEventStore(pool)
	ctx := context.Background()

	// Seed events
	baseTime := time.Now().UTC().Truncate(time.Millisecond) // Truncate for DB precision matching
	events := []Event{
		{ID: "123e4567-e89b-12d3-a456-426614174030", EventType: "TypeA", AggregateID: "agg-1", AggregateType: "A", Version: 1, Timestamp: baseTime, Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174031", EventType: "TypeB", AggregateID: "agg-2", AggregateType: "B", Version: 1, Timestamp: baseTime.Add(1 * time.Hour), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174032", EventType: "TypeA", AggregateID: "agg-3", AggregateType: "A", Version: 1, Timestamp: baseTime.Add(2 * time.Hour), Payload: json.RawMessage(`{}`)},
	}

	for _, e := range events {
		err := store.AppendEvent(ctx, e)
		require.NoError(t, err)
	}

	t.Run("retrieves events by type and time range", func(t *testing.T) {
		// Query for TypeA between baseTime and baseTime + 3 hours
		got, err := store.GetEventsByType(ctx, "TypeA", baseTime.Add(-1*time.Minute), baseTime.Add(3*time.Hour))
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, events[0].ID, got[0].ID)
		assert.Equal(t, events[2].ID, got[1].ID)
	})
}

func TestPostgresEventStore_GetAllEvents(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresEventStore(pool)
	ctx := context.Background()

	// Seed events
	baseTime := time.Now().UTC().Truncate(time.Millisecond)
	events := []Event{
		{ID: "123e4567-e89b-12d3-a456-426614174040", EventType: "T1", AggregateID: "agg-1", AggregateType: "A", Version: 1, Timestamp: baseTime, Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174041", EventType: "T2", AggregateID: "agg-2", AggregateType: "B", Version: 1, Timestamp: baseTime.Add(1 * time.Hour), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174042", EventType: "T3", AggregateID: "agg-3", AggregateType: "C", Version: 1, Timestamp: baseTime.Add(2 * time.Hour), Payload: json.RawMessage(`{}`)},
	}

	for _, e := range events {
		err := store.AppendEvent(ctx, e)
		require.NoError(t, err)
	}

	t.Run("retrieves all events with limit", func(t *testing.T) {
		got, err := store.GetAllEvents(ctx, baseTime.Add(-1*time.Minute), 2)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, events[0].ID, got[0].ID)
		assert.Equal(t, events[1].ID, got[1].ID)
	})
}
