package world

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPauseResume_RecordsPauseTime(t *testing.T) {
	registry := NewRegistry()
	tm := NewTickerManager(registry, nil, nil)
	defer tm.StopAll()

	worldID := uuid.New()

	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	beforePause := time.Now()
	err = tm.StopTicker(worldID)
	require.NoError(t, err)
	afterPause := time.Now()

	// Check world state in registry
	world, err := registry.GetWorld(worldID)
	require.NoError(t, err)

	// PausedAt should be set and within the pause window
	assert.False(t, world.PausedAt.IsZero(), "PausedAt should be set")
	assert.True(t, world.PausedAt.After(beforePause) || world.PausedAt.Equal(beforePause))
	assert.True(t, world.PausedAt.Before(afterPause) || world.PausedAt.Equal(afterPause))
}

func TestPauseResume_CatchupEmitted(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()
	dilationFactor := 10.0

	// Start ticker
	err := tm.SpawnTicker(worldID, "Test World", dilationFactor)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	// Pause
	err = tm.StopTicker(worldID)
	require.NoError(t, err)

	ticksBeforePause := len(natsPublisher.GetPublished())
	natsPublisher.Clear()

	// Simulate pause for 1 second real time
	// At 10x dilation, this means 10 seconds of missed game time
	// At 100ms per tick, that's 100 missed ticks
	time.Sleep(1 * time.Second)

	// Resume
	err = tm.SpawnTicker(worldID, "Test World", dilationFactor)
	require.NoError(t, err)

	// Wait for catch-up to complete (at 100x speed, 100 ticks should take ~100ms)
	time.Sleep(500 * time.Millisecond)

	published := natsPublisher.GetPublished()

	// Should have emitted catch-up ticks
	// We expect 10 catch-up ticks (1 second pause / 100ms tick interval)
	// Plus a few more from normal operation after catch-up (3-5 more in 500ms)
	assert.GreaterOrEqual(t, len(published), 10, "Should have emitted catch-up ticks")
	assert.LessOrEqual(t, len(published), 20, "Should not have too many extra ticks")

	t.Logf("Ticks before pause: %d, catch-up ticks: %d", ticksBeforePause, len(published))
}

func TestPauseResume_CatchupSpeed(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()

	// Start and immediately pause
	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	err = tm.StopTicker(worldID)
	require.NoError(t, err)

	natsPublisher.Clear()

	// Pause for 2 seconds (20 missed ticks at 1x dilation)
	time.Sleep(2 * time.Second)

	// Resume and measure catch-up time
	startCatchup := time.Now()
	err = tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Wait a bit for catch-up to start
	time.Sleep(100 * time.Millisecond)

	published := natsPublisher.GetPublished()
	catchupDuration := time.Since(startCatchup)

	t.Logf("Catch-up ticks: %d, duration: %v", len(published), catchupDuration)

	// At 100x speed, 20 ticks should take ~200ms
	// Allow variance, but should be much faster than real-time (which would be 2 seconds)
	assert.Less(t, catchupDuration.Seconds(), 1.0, "Catch-up should be faster than real-time")
}

func TestPauseResume_GameTimeConsistency(t *testing.T) {
	registry := NewRegistry()
	tm := NewTickerManager(registry, nil, nil)
	defer tm.StopAll()

	worldID := uuid.New()
	dilationFactor := 5.0

	// Start ticker
	err := tm.SpawnTicker(worldID, "Test World", dilationFactor)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond) // ~5 ticks, 2.5 seconds game time

	// Get game time before pause
	world1, _ := registry.GetWorld(worldID)
	gameTimeBeforePause := world1.GameTime

	// Pause
	err = tm.StopTicker(worldID)
	require.NoError(t, err)

	// Wait 1 second (5 seconds missed game time at 5x dilation)
	time.Sleep(1 * time.Second)

	// Resume
	err = tm.SpawnTicker(worldID, "Test World", dilationFactor)
	require.NoError(t, err)

	// Wait for catch-up to complete
	time.Sleep(500 * time.Millisecond)

	// Get game time after catch-up
	world2, _ := registry.GetWorld(worldID)
	gameTimeAfterCatchup := world2.GameTime

	// Game time should have advanced by approximately:
	// - 1 second pause * 5x dilation = 5 seconds (from catch-up)
	// - Plus ~500ms * 5x dilation = 2.5 seconds (from normal ticking after catch-up)
	// Total: ~7.5 seconds
	gameTimeDelta := gameTimeAfterCatchup - gameTimeBeforePause
	expectedDelta := 7 * time.Second // Conservative estimate: 5s catch-up + 2s normal

	// Allow 20% variance due to timing
	lowerBound := expectedDelta - (expectedDelta * 2 / 10)
	upperBound := expectedDelta + (expectedDelta * 2 / 10)

	assert.GreaterOrEqual(t, gameTimeDelta, lowerBound)
	assert.LessOrEqual(t, gameTimeDelta, upperBound)

	t.Logf("Game time delta: %v, expected: %v", gameTimeDelta, expectedDelta)
}

func TestPauseResume_MultipleCycles(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()

	// Perform multiple pause/resume cycles
	for i := 0; i < 3; i++ {
		natsPublisher.Clear()

		err := tm.SpawnTicker(worldID, "Test World", 1.0)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		err = tm.StopTicker(worldID)
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)
	}

	// Final resume
	natsPublisher.Clear()
	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	// Should still be working after multiple cycles
	world, err := registry.GetWorld(worldID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, world.Status)
	assert.Greater(t, world.TickCount, int64(0))
}

func TestPauseResume_NoPauseNoticeCatchup(t *testing.T) {
	// Test that initial spawn does not trigger catch-up
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()

	// First spawn should not trigger catch-up
	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	published := natsPublisher.GetPublished()

	// Should have normal ticks (2-3 at 100ms interval)
	// NOT catch-up burst
	assert.LessOrEqual(t, len(published), 5, "Should not have catch-up burst on initial spawn")
}

func TestPauseResume_StopDuringCatchup(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()

	// Start and pause
	err := tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	err = tm.StopTicker(worldID)
	require.NoError(t, err)

	// Long pause to ensure significant catch-up needed
	time.Sleep(5 * time.Second) // 50 ticks worth of catch-up

	// Resume (starts catch-up)
	err = tm.SpawnTicker(worldID, "Test World", 1.0)
	require.NoError(t, err)

	// Stop during catch-up
	time.Sleep(50 * time.Millisecond)
	err = tm.StopTicker(worldID)
	require.NoError(t, err)

	// Should not crash or hang
	world, err := registry.GetWorld(worldID)
	require.NoError(t, err)
	assert.Equal(t, StatusPaused, world.Status)
}
