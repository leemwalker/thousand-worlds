package ecosystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeadLetterQueue_RecordFailure(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	worldID := uuid.New()
	err := errors.New("simulation error")

	dlq.RecordFailure(worldID, 1000, "tick", err, nil, true)

	assert.Equal(t, 1, dlq.Count())
	events := dlq.GetFailedEvents(nil, 0)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, worldID, event.WorldID)
	assert.Equal(t, int64(1000), event.Year)
	assert.Equal(t, "tick", event.EventType)
	assert.Equal(t, "simulation error", event.Error)
	assert.True(t, event.Recoverable)
}

func TestDeadLetterQueue_RecordPanic(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	worldID := uuid.New()

	dlq.RecordPanic(worldID, 5000, "evolution", "nil pointer dereference", "goroutine 1 [running]: ...")

	events := dlq.GetFailedEvents(nil, 0)
	require.Len(t, events, 1)

	event := events[0]
	assert.Contains(t, event.Error, "panic: nil pointer dereference")
	assert.Contains(t, event.StackTrace, "goroutine")
	assert.False(t, event.Recoverable)
}

func TestDeadLetterQueue_MaxSize(t *testing.T) {
	config := DeadLetterConfig{MaxSize: 5}
	dlq := NewDeadLetterQueue(config)
	defer dlq.Close()

	worldID := uuid.New()
	err := errors.New("test error")

	// Add 10 events
	for i := 0; i < 10; i++ {
		dlq.RecordFailure(worldID, int64(i), "tick", err, nil, true)
	}

	// Should only keep last 5
	assert.Equal(t, 5, dlq.Count())

	events := dlq.GetFailedEvents(nil, 0)
	// Most recent first
	assert.Equal(t, int64(9), events[0].Year)
	assert.Equal(t, int64(5), events[4].Year)
}

func TestDeadLetterQueue_FilterByWorld(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	world1 := uuid.New()
	world2 := uuid.New()
	err := errors.New("test error")

	dlq.RecordFailure(world1, 1000, "tick", err, nil, true)
	dlq.RecordFailure(world2, 2000, "tick", err, nil, true)
	dlq.RecordFailure(world1, 3000, "tick", err, nil, true)

	// Filter by world1
	events := dlq.GetFailedEvents(&world1, 0)
	assert.Len(t, events, 2)

	// Count by world
	assert.Equal(t, 2, dlq.CountByWorld(world1))
	assert.Equal(t, 1, dlq.CountByWorld(world2))
}

func TestDeadLetterQueue_RecoverableEvents(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	worldID := uuid.New()

	// Mix of recoverable and non-recoverable
	dlq.RecordFailure(worldID, 1000, "tick", errors.New("err1"), nil, true)
	dlq.RecordFailure(worldID, 2000, "tick", errors.New("err2"), nil, false)
	dlq.RecordFailure(worldID, 3000, "tick", errors.New("err3"), nil, true)

	recoverable := dlq.GetRecoverableEvents(worldID)
	assert.Len(t, recoverable, 2)
}

func TestDeadLetterQueue_RetryTracking(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	worldID := uuid.New()
	dlq.RecordFailure(worldID, 1000, "tick", errors.New("err"), nil, true)

	events := dlq.GetFailedEvents(nil, 0)
	eventID := events[0].ID

	// Mark as retried
	dlq.MarkRetried(eventID)
	dlq.MarkRetried(eventID)
	dlq.MarkRetried(eventID)

	// Should no longer be in recoverable (retry count >= 3)
	recoverable := dlq.GetRecoverableEvents(worldID)
	assert.Len(t, recoverable, 0)
}

func TestDeadLetterQueue_RemoveEvent(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	worldID := uuid.New()
	dlq.RecordFailure(worldID, 1000, "tick", errors.New("err1"), nil, true)
	dlq.RecordFailure(worldID, 2000, "tick", errors.New("err2"), nil, true)

	events := dlq.GetFailedEvents(nil, 0)
	require.Len(t, events, 2)

	// Remove first event
	dlq.RemoveEvent(events[0].ID)
	assert.Equal(t, 1, dlq.Count())
}

func TestDeadLetterQueue_FileLogging(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "dlq.log")

	config := DeadLetterConfig{
		MaxSize: 100,
		LogPath: logPath,
	}
	dlq := NewDeadLetterQueue(config)

	worldID := uuid.New()
	dlq.RecordFailure(worldID, 1000, "tick", errors.New("test error"), nil, true)
	dlq.RecordPanic(worldID, 2000, "evolution", "panic!", "stack trace")

	dlq.Close()

	// Read log file
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "test error")
	assert.Contains(t, content, "panic!")
}

func TestDeadLetterQueue_Clear(t *testing.T) {
	dlq := NewDeadLetterQueue(DefaultDeadLetterConfig())
	defer dlq.Close()

	worldID := uuid.New()
	dlq.RecordFailure(worldID, 1000, "tick", errors.New("err"), nil, true)
	assert.Equal(t, 1, dlq.Count())

	dlq.Clear()
	assert.Equal(t, 0, dlq.Count())
}
