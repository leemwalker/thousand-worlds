package character

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeCharacterCreatedViaGeneration  = "CharacterCreatedViaGeneration"
	EventTypeCharacterCreatedViaInhabitance = "CharacterCreatedViaInhabitance"
	EventTypeAttributeModified              = "AttributeModified"
)

// CharacterCreatedViaGenerationEvent is emitted when a new character is generated
type CharacterCreatedViaGenerationEvent struct {
	CharacterID     uuid.UUID      `json:"character_id"`
	PlayerID        uuid.UUID      `json:"player_id"`
	Name            string         `json:"name"`
	Species         string         `json:"species"`
	BaseAttributes  Attributes     `json:"base_attributes"`
	Variance        Attributes     `json:"variance"`
	PointBuyChoices map[string]int `json:"point_buy_choices"`
	FinalAttributes Attributes     `json:"final_attributes"`
	Timestamp       time.Time      `json:"timestamp"`
}

// CharacterCreatedViaInhabitanceEvent is emitted when a player inhabits an NPC
type CharacterCreatedViaInhabitanceEvent struct {
	CharacterID      uuid.UUID          `json:"character_id"`
	PlayerID         uuid.UUID          `json:"player_id"`
	NPCID            uuid.UUID          `json:"npc_id"`
	BaselineSnapshot BehavioralBaseline `json:"baseline_snapshot"`
	Timestamp        time.Time          `json:"timestamp"`
}

// BehavioralBaseline represents the snapshot of an NPC's personality at inhabitation
type BehavioralBaseline struct {
	Aggression   float64 `json:"aggression"`
	Generosity   float64 `json:"generosity"`
	Honesty      float64 `json:"honesty"`
	Sociability  float64 `json:"sociability"`
	Recklessness float64 `json:"recklessness"`
	Loyalty      float64 `json:"loyalty"`
}

// AttributeModifiedEvent is emitted when an attribute changes
type AttributeModifiedEvent struct {
	CharacterID uuid.UUID `json:"character_id"`
	Attribute   string    `json:"attribute"`
	OldValue    int       `json:"old_value"`
	NewValue    int       `json:"new_value"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}
