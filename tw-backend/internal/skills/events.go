package skills

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeSkillIncreased = "SkillIncreased"
	EventTypeSkillUsed      = "SkillUsed"
	EventTypeSkillLeveledUp = "SkillLeveledUp"
)

type SkillIncreasedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	SkillName   string    `json:"skill_name"`
	OldValue    int       `json:"old_value"`
	NewValue    int       `json:"new_value"`
	XPGained    float64   `json:"xp_gained"`
	Timestamp   time.Time `json:"timestamp"`
}

type SkillUsedEvent struct {
	CharacterID             uuid.UUID `json:"character_id"`
	SkillName               string    `json:"skill_name"`
	Context                 string    `json:"context"`
	XPGained                float64   `json:"xp_gained"`
	DiminishingReturnFactor float64   `json:"diminishing_return_factor"`
	Timestamp               time.Time `json:"timestamp"`
}

type SkillLeveledUpEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	SkillName   string    `json:"skill_name"`
	NewLevel    int       `json:"new_level"`
	Timestamp   time.Time `json:"timestamp"`
}
