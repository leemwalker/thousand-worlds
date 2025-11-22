package player

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypePlayerMoved    = "PlayerMoved"
	EventTypeStaminaChanged = "StaminaChanged"
)

// PlayerMovedEvent is emitted when a player moves
type PlayerMovedEvent struct {
	CharacterID  uuid.UUID    `json:"character_id"`
	FromX        float64      `json:"from_x"`
	FromY        float64      `json:"from_y"`
	FromZ        float64      `json:"from_z"`
	ToX          float64      `json:"to_x"`
	ToY          float64      `json:"to_y"`
	ToZ          float64      `json:"to_z"`
	MovementType MovementType `json:"movement_type"`
	StaminaCost  int          `json:"stamina_cost"`
	Timestamp    time.Time    `json:"timestamp"`
}

// StaminaChangedEvent is emitted when stamina changes (regen or drain)
type StaminaChangedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	OldValue    int       `json:"old_value"`
	NewValue    int       `json:"new_value"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}
