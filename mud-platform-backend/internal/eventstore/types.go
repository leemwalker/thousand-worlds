package eventstore

import (
	"encoding/json"
	"time"
)

// EventType represents the type of an event.
type EventType string

// AggregateType represents the type of an aggregate.
type AggregateType string

// Event represents a fact that has happened in the system.
type Event struct {
	ID            string          `json:"id"`
	EventType     EventType       `json:"event_type"`
	AggregateID   string          `json:"aggregate_id"`
	AggregateType AggregateType   `json:"aggregate_type"`
	Version       int64           `json:"version"`
	Timestamp     time.Time       `json:"timestamp"`
	Payload       json.RawMessage `json:"payload"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

// Command represents a request to perform an action.
type Command interface {
	AggregateID() string
	CommandType() string
}
