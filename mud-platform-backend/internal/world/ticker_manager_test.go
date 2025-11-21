package world

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/eventstore"
)

// MockEventStore for testing
type MockEventStore struct {
	events []eventstore.Event
}

func (m *MockEventStore) AppendEvent(ctx context.Context, event eventstore.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventStore) GetEventsByAggregate(ctx context.Context, aggregateID string, fromVersion int64) ([]eventstore.Event, error) {
	return nil, nil
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, eventType eventstore.EventType, fromTimestamp, toTimestamp time.Time) ([]eventstore.Event, error) {
	return nil, nil
}

func (m *MockEventStore) GetAllEvents(ctx context.Context, fromTimestamp time.Time, limit int) ([]eventstore.Event, error) {
	return nil, nil
}

func TestTickerManager_SpawnTicker(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err, "Should spawn ticker successfully")

	// Check ticker status
	running, tickCount, gameTime := tm.GetTickerStatus(worldID)
	assert.True(t, running, "Ticker should be running")
	assert.GreaterOrEqual(t, tickCount, int64(0), "Tick count should be initialized")
	assert.GreaterOrEqual(t, gameTime, time.Duration(0), "Game time should be initialized")
}

func TestTickerManager_SpawnTicker_AlreadyRunning(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Try to spawn again
	err = tm.SpawnTicker(worldID, "Test World", 1.0)
	assert.Error(t, err, "Should return error when spawning ticker that's already running")
}

func TestTickerManager_StopTicker(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Stop the ticker
	err = tm.StopTicker(worldID)
	require.NoError(t, err, "Should stop ticker successfully")

	// Verify it's stopped
	running, _, _ := tm.GetTickerStatus(worldID)
	assert.False(t, running, "Ticker should be stopped")
}

func TestTickerManager_StopTicker_NotRunning(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)

	nonExistentID := uuid.New()

	err := tm.StopTicker(nonExistentID)
	assert.Error(t, err, "Should return error when stopping non-existent ticker")
}

func TestTickerManager_GetTickerStatus(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	// Before spawning
	running, tickCount, gameTime := tm.GetTickerStatus(worldID)
	assert.False(t, running, "Ticker should not be running initially")
	assert.Equal(t, int64(0), tickCount)
	assert.Equal(t, time.Duration(0), gameTime)

	// After spawning
	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	running, _, _ = tm.GetTickerStatus(worldID)
	assert.True(t, running, "Ticker should be running after spawn")
}

func TestTickerManager_TickerEmitsEvents(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Wait for at least one tick
	time.Sleep(150 * time.Millisecond)

	// Stop to check events
	err = tm.StopTicker(worldID)
	require.NoError(t, err)

	// Verify events were emitted
	assert.Greater(t, len(eventStore.events), 0, "Should have emitted at least WorldCreated event")

	// First event should be WorldCreated
	assert.Equal(t, eventstore.EventType("WorldCreated"), eventStore.events[0].EventType)

	// If we had ticks, should have WorldTicked events
	if len(eventStore.events) > 1 {
		hasTickedEvent := false
		for _, evt := range eventStore.events {
			if evt.EventType == eventstore.EventType("WorldTicked") {
				hasTickedEvent = true
				break
			}
		}
		assert.True(t, hasTickedEvent, "Should have emitted WorldTicked events")
	}
}

func TestTickerManager_TickProgression(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Get initial tick count
	_, initialTick, _ := tm.GetTickerStatus(worldID)

	// Wait for several ticks (tick interval is 100ms)
	time.Sleep(350 * time.Millisecond)

	// Get final tick count
	_, finalTick, finalGameTime := tm.GetTickerStatus(worldID)

	assert.Greater(t, finalTick, initialTick, "Tick count should increase over time")
	assert.Greater(t, finalGameTime, time.Duration(0), "Game time should increase over time")

	tm.StopTicker(worldID)
}

func TestTickerManager_DilationFactor(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)
	defer tm.StopAll()

	worldID := uuid.New()
	dilationFactor := 10.0 // 10x time speed

	err := tm.SpawnTicker(worldID, "Fast World", dilationFactor)
	require.NoError(t, err)

	// Wait for several ticks
	time.Sleep(250 * time.Millisecond)

	_, _, gameTime := tm.GetTickerStatus(worldID)

	// With 100ms tick interval and 10x dilation, after 250ms:
	// We should have ~2 ticks = 200ms real time = 2000ms game time
	// Allow some tolerance for timing
	assert.Greater(t, gameTime, 1*time.Second, "Game time should be dilated by factor")

	tm.StopTicker(worldID)
}

func TestTickerManager_StopAll(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil)

	// Spawn multiple tickers
	world1 := uuid.New()
	world2 := uuid.New()
	world3 := uuid.New()

	tm.SpawnTicker(world1, "World 1", 1.0)
	tm.SpawnTicker(world2, "World 2", 1.0)
	tm.SpawnTicker(world3, "World 3", 1.0)

	// Verify all running
	running1, _, _ := tm.GetTickerStatus(world1)
	running2, _, _ := tm.GetTickerStatus(world2)
	running3, _, _ := tm.GetTickerStatus(world3)

	assert.True(t, running1 && running2 && running3, "All tickers should be running")

	// Stop all
	tm.StopAll()

	// Verify all stopped
	running1, _, _ = tm.GetTickerStatus(world1)
	running2, _, _ = tm.GetTickerStatus(world2)
	running3, _, _ = tm.GetTickerStatus(world3)

	assert.False(t, running1 || running2 || running3, "All tickers should be stopped")
}
