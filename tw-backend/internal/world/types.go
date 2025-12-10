package world

import (
	"time"

	"github.com/google/uuid"
)

// WorldStatus represents the current state of a world
type WorldStatus string

const (
	StatusRunning WorldStatus = "running"
	StatusPaused  WorldStatus = "paused"
	StatusStopped WorldStatus = "stopped"
)

// WorldState tracks the current state of a world
type WorldState struct {
	ID             uuid.UUID
	Name           string
	Status         WorldStatus
	TickCount      int64
	GameTime       time.Duration // Total in-game time elapsed
	DilationFactor float64       // Time speed multiplier (1.0 = real-time)
	CreatedAt      time.Time
	LastTickAt     time.Time
	PausedAt       time.Time // When the world was paused (zero if not paused)
}

// Event payloads for event sourcing

// WorldCreatedPayload is emitted when a world is first created
type WorldCreatedPayload struct {
	WorldID        string    `json:"world_id"`
	Name           string    `json:"name"`
	DilationFactor float64   `json:"dilation_factor"`
	CreatedAt      time.Time `json:"created_at"`
}

// WorldTickedPayload is emitted on each world tick
type WorldTickedPayload struct {
	WorldID            string `json:"world_id"`
	TickCount          int64  `json:"tick_count"`
	GameTimeNs         int64  `json:"game_time_ns"`
	RealTickDurationMs int64  `json:"real_tick_duration_ms"`
}

// WorldPausedPayload is emitted when a world is paused
type WorldPausedPayload struct {
	WorldID   string    `json:"world_id"`
	PausedAt  time.Time `json:"paused_at"`
	TickCount int64     `json:"tick_count"`
}

// WorldResumedPayload is emitted when a world is resumed
type WorldResumedPayload struct {
	WorldID           string    `json:"world_id"`
	ResumedAt         time.Time `json:"resumed_at"`
	NewDilationFactor float64   `json:"new_dilation_factor,omitempty"`
}
