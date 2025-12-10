package world

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/internal/eventstore"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"

	"github.com/stretchr/testify/mock"
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)

	nonExistentID := uuid.New()

	err := tm.StopTicker(nonExistentID)
	assert.Error(t, err, "Should return error when stopping non-existent ticker")
}

func TestTickerManager_GetTickerStatus(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)
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
	tm := NewTickerManager(registry, eventStore, nil, nil, nil)

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

// MockWeatherRepository
type MockWeatherRepository struct {
	mock.Mock
}

func (m *MockWeatherRepository) SaveWeatherState(ctx context.Context, state *weather.WeatherState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockWeatherRepository) GetWeatherState(ctx context.Context, cellID uuid.UUID, timestamp int64) (*weather.WeatherState, error) {
	args := m.Called(ctx, cellID, timestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather.WeatherState), args.Error(1)
}

func (m *MockWeatherRepository) GetWeatherHistory(ctx context.Context, cellID uuid.UUID, limit int) ([]*weather.WeatherState, error) {
	args := m.Called(ctx, cellID, limit)
	return args.Get(0).([]*weather.WeatherState), args.Error(1)
}

func (m *MockWeatherRepository) GetAnnualPrecipitation(ctx context.Context, cellID uuid.UUID, year int) (float64, error) {
	args := m.Called(ctx, cellID, year)
	return args.Get(0).(float64), args.Error(1)
}

// MockAreaBroadcaster
type MockAreaBroadcaster struct {
	mock.Mock
}

func (m *MockAreaBroadcaster) BroadcastToArea(center spatial.Position, radius float64, msgType string, data interface{}) {
	m.Called(center, radius, msgType, data)
}

func TestTickerManager_WeatherIntegration(t *testing.T) {
	registry := NewRegistry()
	eventStore := &MockEventStore{}
	weatherRepo := &MockWeatherRepository{}
	weatherService := weather.NewService(weatherRepo)
	broadcaster := &MockAreaBroadcaster{}

	tm := NewTickerManager(registry, eventStore, nil, weatherService, broadcaster)
	defer tm.StopAll()

	worldID := uuid.New()

	// Prepare minimal geography for weather service
	cellID := uuid.New()
	cells := []*weather.GeographyCell{
		{CellID: cellID, Location: geography.Point{X: 0, Y: 0}, Elevation: 100, Temperature: 20},
	}
	weatherStates := []*weather.WeatherState{
		{CellID: cellID, State: weather.WeatherClear, Temperature: 20, Timestamp: time.Now()},
	}

	// Initialize weather service cache
	weatherService.InitializeWorldWeather(context.Background(), worldID, weatherStates, cells)

	// Expect SaveWeatherState to be called eventually
	weatherRepo.On("SaveWeatherState", mock.Anything, mock.Anything).Return(nil)

	// Spawn ticker with high dilation to trigger update quickly
	// Default weather update interval is 30 game minutes.
	// At 100x dilation, 30 game minutes = 1800s game time = 18s real time.
	// Too slow.
	// We need 30 game minutes.
	// If we want it in < 1s.
	// 30 min = 1800s.
	// 1800s / 0.5s = 3600 dilation.
	dilationFactor := 3600.0

	err := tm.SpawnTicker(worldID, "Weather World", dilationFactor)
	require.NoError(t, err)

	// Wait for a few ticks to cover 30 game minutes
	time.Sleep(600 * time.Millisecond) // Should be enough for > 30 min game time

	// Assert weather updated
	// Since we mock SaveWeatherState, we can check if it was called
	// It might be called multiple times depending on how much game time passes
	// But at least once.

	// We use Eventually because it happens in goroutine
	assert.Eventually(t, func() bool {
		return len(weatherRepo.Calls) > 0
	}, 2*time.Second, 100*time.Millisecond, "Should have triggered weather update")

	tm.StopTicker(worldID)
}
