// Package ecosystem provides the background simulation runner and related systems.
package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FailedEvent represents a simulation event that failed to process
type FailedEvent struct {
	ID          uuid.UUID              `json:"id"`
	WorldID     uuid.UUID              `json:"world_id"`
	Year        int64                  `json:"year"`
	EventType   string                 `json:"event_type"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Error       string                 `json:"error"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	OccurredAt  time.Time              `json:"occurred_at"`
	RetryCount  int                    `json:"retry_count"`
	Recoverable bool                   `json:"recoverable"`
}

// DeadLetterQueue captures failed simulation events for later analysis and retry
type DeadLetterQueue struct {
	mu      sync.Mutex
	events  []FailedEvent
	maxSize int
	logPath string
	logFile *os.File
}

// DeadLetterConfig holds configuration for the dead letter queue
type DeadLetterConfig struct {
	MaxSize int    // Maximum events to keep in memory (0 = unlimited)
	LogPath string // Path to write failed events (empty = no file logging)
}

// DefaultDeadLetterConfig returns sensible defaults
func DefaultDeadLetterConfig() DeadLetterConfig {
	return DeadLetterConfig{
		MaxSize: 1000,
		LogPath: "",
	}
}

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(config DeadLetterConfig) *DeadLetterQueue {
	dlq := &DeadLetterQueue{
		events:  make([]FailedEvent, 0),
		maxSize: config.MaxSize,
		logPath: config.LogPath,
	}

	// Open log file if path specified
	if config.LogPath != "" {
		f, err := os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("[DLQ] Warning: failed to open log file %s: %v\n", config.LogPath, err)
		} else {
			dlq.logFile = f
		}
	}

	return dlq
}

// RecordFailure adds a failed event to the queue
func (dlq *DeadLetterQueue) RecordFailure(worldID uuid.UUID, year int64, eventType string, err error, payload map[string]interface{}, recoverable bool) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	event := FailedEvent{
		ID:          uuid.New(),
		WorldID:     worldID,
		Year:        year,
		EventType:   eventType,
		Payload:     payload,
		Error:       err.Error(),
		OccurredAt:  time.Now(),
		RetryCount:  0,
		Recoverable: recoverable,
	}

	dlq.events = append(dlq.events, event)

	// Trim if over max size
	if dlq.maxSize > 0 && len(dlq.events) > dlq.maxSize {
		dlq.events = dlq.events[len(dlq.events)-dlq.maxSize:]
	}

	// Log to file
	if dlq.logFile != nil {
		if data, marshalErr := json.Marshal(event); marshalErr == nil {
			dlq.logFile.WriteString(string(data) + "\n")
		}
	}
}

// RecordPanic captures a panic recovery as a failed event
func (dlq *DeadLetterQueue) RecordPanic(worldID uuid.UUID, year int64, eventType string, recovered interface{}, stackTrace string) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	event := FailedEvent{
		ID:          uuid.New(),
		WorldID:     worldID,
		Year:        year,
		EventType:   eventType,
		Error:       fmt.Sprintf("panic: %v", recovered),
		StackTrace:  stackTrace,
		OccurredAt:  time.Now(),
		RetryCount:  0,
		Recoverable: false,
	}

	dlq.events = append(dlq.events, event)

	// Trim if over max size
	if dlq.maxSize > 0 && len(dlq.events) > dlq.maxSize {
		dlq.events = dlq.events[len(dlq.events)-dlq.maxSize:]
	}

	// Log to file (panics are always logged)
	if dlq.logFile != nil {
		if data, marshalErr := json.Marshal(event); marshalErr == nil {
			dlq.logFile.WriteString(string(data) + "\n")
		}
	}

	// Also log to stderr for visibility
	fmt.Fprintf(os.Stderr, "[DLQ] PANIC recorded: world=%s year=%d type=%s error=%v\n",
		worldID, year, eventType, recovered)
}

// GetFailedEvents returns all failed events, optionally filtered by world
func (dlq *DeadLetterQueue) GetFailedEvents(worldID *uuid.UUID, limit int) []FailedEvent {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	var result []FailedEvent

	for i := len(dlq.events) - 1; i >= 0 && (limit == 0 || len(result) < limit); i-- {
		event := dlq.events[i]
		if worldID == nil || event.WorldID == *worldID {
			result = append(result, event)
		}
	}

	return result
}

// GetRecoverableEvents returns events that can be retried
func (dlq *DeadLetterQueue) GetRecoverableEvents(worldID uuid.UUID) []FailedEvent {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	var result []FailedEvent
	for _, event := range dlq.events {
		if event.WorldID == worldID && event.Recoverable && event.RetryCount < 3 {
			result = append(result, event)
		}
	}

	return result
}

// MarkRetried increments the retry count for an event
func (dlq *DeadLetterQueue) MarkRetried(eventID uuid.UUID) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	for i := range dlq.events {
		if dlq.events[i].ID == eventID {
			dlq.events[i].RetryCount++
			break
		}
	}
}

// RemoveEvent removes a successfully retried event
func (dlq *DeadLetterQueue) RemoveEvent(eventID uuid.UUID) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	for i, event := range dlq.events {
		if event.ID == eventID {
			dlq.events = append(dlq.events[:i], dlq.events[i+1:]...)
			break
		}
	}
}

// Count returns the number of events in the queue
func (dlq *DeadLetterQueue) Count() int {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	return len(dlq.events)
}

// CountByWorld returns the number of failed events for a specific world
func (dlq *DeadLetterQueue) CountByWorld(worldID uuid.UUID) int {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	count := 0
	for _, event := range dlq.events {
		if event.WorldID == worldID {
			count++
		}
	}
	return count
}

// Close closes the log file if open
func (dlq *DeadLetterQueue) Close() error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if dlq.logFile != nil {
		return dlq.logFile.Close()
	}
	return nil
}

// Clear removes all events from the queue
func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.events = make([]FailedEvent, 0)
}
