package world

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNATSPublisher for testing
type MockNATSPublisher struct {
	mu        sync.Mutex
	published []PublishedMessage
}

type PublishedMessage struct {
	Subject string
	Data    []byte
	Time    time.Time
}

func (m *MockNATSPublisher) Publish(subject string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, PublishedMessage{
		Subject: subject,
		Data:    data,
		Time:    time.Now(),
	})
	return nil
}

func (m *MockNATSPublisher) GetPublished() []PublishedMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]PublishedMessage{}, m.published...)
}

func (m *MockNATSPublisher) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = nil
}

func TestNATSBroadcast_TicksPublished(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Wait for a few ticks
	time.Sleep(350 * time.Millisecond)

	published := natsPublisher.GetPublished()
	assert.Greater(t, len(published), 2, "Should have published multiple tick events")
}

func TestNATSBroadcast_CorrectSubject(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()
	expectedSubject := "world.tick." + worldID.String()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	published := natsPublisher.GetPublished()
	require.Greater(t, len(published), 0, "Should have published at least one tick")

	// All messages should have correct subject
	for _, msg := range published {
		assert.Equal(t, expectedSubject, msg.Subject)
	}
}

func TestNATSBroadcast_PayloadStructure(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()
	dilationFactor := 5.0

	err := tm.SpawnTicker(worldID, "Test World", dilationFactor)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	published := natsPublisher.GetPublished()
	require.Greater(t, len(published), 0, "Should have published at least one tick")

	// Parse first message
	var broadcast TickBroadcast
	err = json.Unmarshal(published[0].Data, &broadcast)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, worldID.String(), broadcast.WorldID)
	assert.Greater(t, broadcast.TickNumber, int64(0))
	assert.Greater(t, broadcast.GameTimeMs, int64(0))
	assert.Equal(t, int64(100), broadcast.RealTimeMs, "Real time should be 100ms tick interval")
	assert.Equal(t, dilationFactor, broadcast.DilationFactor)
}

func TestNATSBroadcast_Frequency(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Run for 1 second
	time.Sleep(1 * time.Second)

	published := natsPublisher.GetPublished()

	// At 10 Hz (100ms interval), should have ~10 ticks in 1 second
	// Allow some variance for timing
	assert.GreaterOrEqual(t, len(published), 8, "Should have at least 8 ticks in 1 second")
	assert.LessOrEqual(t, len(published), 12, "Should have at most 12 ticks in 1 second")
}

func TestNATSBroadcast_MultipleWorlds(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	world1 := uuid.New()
	world2 := uuid.New()

	err := tm.SpawnTicker(world1, "World 1", 1.0)
	require.NoError(t, err)

	err = tm.SpawnTicker(world2, "World 2", 10.0)
	require.NoError(t, err)

	time.Sleep(250 * time.Millisecond)

	published := natsPublisher.GetPublished()

	// Count messages per world
	world1Count := 0
	world2Count := 0
	subject1 := "world.tick." + world1.String()
	subject2 := "world.tick." + world2.String()

	for _, msg := range published {
		if msg.Subject == subject1 {
			world1Count++
		} else if msg.Subject == subject2 {
			world2Count++
		}
	}

	assert.Greater(t, world1Count, 0, "World 1 should have published ticks")
	assert.Greater(t, world2Count, 0, "World 2 should have published ticks")
}

func TestNATSBroadcast_GameTimeProgression(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()
	dilationFactor := 10.0

	err := tm.SpawnTicker(worldID, "Fast World", dilationFactor)
	require.NoError(t, err)

	time.Sleep(350 * time.Millisecond)

	published := natsPublisher.GetPublished()
	require.GreaterOrEqual(t, len(published), 2, "Should have at least 2 ticks")

	// Parse first and last tick
	var firstTick, lastTick TickBroadcast
	json.Unmarshal(published[0].Data, &firstTick)
	json.Unmarshal(published[len(published)-1].Data, &lastTick)

	// Game time should advance by realTime * dilation
	gameTimeDelta := lastTick.GameTimeMs - firstTick.GameTimeMs
	tickCount := lastTick.TickNumber - firstTick.TickNumber
	expectedGameTimeDelta := tickCount * 100 * int64(dilationFactor)

	assert.Equal(t, expectedGameTimeDelta, gameTimeDelta)
}

func TestNATSBroadcast_NilPublisher(t *testing.T) {
	// Test that ticker works without NATS publisher (backward compatibility)
	registry := NewRegistry()
	tm := NewTickerManager(registry, nil, nil) // nil NATS publisher
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	// Should not crash, ticker should still work
	running, tickCount, _ := tm.GetTickerStatus(worldID)
	assert.True(t, running)
	assert.Greater(t, tickCount, int64(0))
}
