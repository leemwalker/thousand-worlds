package eventstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresReplayEngine_ReplayEvents(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresEventStore(pool)
	replay := NewPostgresReplayEngine(store)
	ctx := context.Background()

	// Seed events
	aggID := "replay-agg-1"
	events := []Event{
		{ID: "123e4567-e89b-12d3-a456-426614174010", EventType: "T1", AggregateID: aggID, AggregateType: "A", Version: 1, Timestamp: time.Now().UTC(), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174011", EventType: "T2", AggregateID: aggID, AggregateType: "A", Version: 2, Timestamp: time.Now().UTC(), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174012", EventType: "T3", AggregateID: aggID, AggregateType: "A", Version: 3, Timestamp: time.Now().UTC(), Payload: json.RawMessage(`{}`)},
	}

	for _, e := range events {
		err := store.AppendEvent(ctx, e)
		require.NoError(t, err)
	}

	t.Run("replays range of events", func(t *testing.T) {
		got, err := replay.ReplayEvents(ctx, aggID, 1, 2)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, int64(1), got[0].Version)
		assert.Equal(t, int64(2), got[1].Version)
	})
}

func TestPostgresReplayEngine_RewindToTimestamp(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresEventStore(pool)
	replay := NewPostgresReplayEngine(store)
	ctx := context.Background()

	aggID := "replay-agg-2"
	baseTime := time.Now().UTC().Add(-1 * time.Hour)

	events := []Event{
		{ID: "123e4567-e89b-12d3-a456-426614174020", EventType: "T1", AggregateID: aggID, AggregateType: "A", Version: 1, Timestamp: baseTime, Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174021", EventType: "T2", AggregateID: aggID, AggregateType: "A", Version: 2, Timestamp: baseTime.Add(10 * time.Minute), Payload: json.RawMessage(`{}`)},
		{ID: "123e4567-e89b-12d3-a456-426614174022", EventType: "T3", AggregateID: aggID, AggregateType: "A", Version: 3, Timestamp: baseTime.Add(20 * time.Minute), Payload: json.RawMessage(`{}`)},
	}

	for _, e := range events {
		err := store.AppendEvent(ctx, e)
		require.NoError(t, err)
	}

	t.Run("rewinds to point in time", func(t *testing.T) {
		// Rewind to 15 minutes after baseTime (should include V1 and V2, but not V3)
		targetTime := baseTime.Add(15 * time.Minute)
		got, err := replay.RewindToTimestamp(ctx, aggID, targetTime)
		assert.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, int64(1), got[0].Version)
		assert.Equal(t, int64(2), got[1].Version)
	})
}
