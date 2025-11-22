package interaction

import (
	"time"

	"github.com/google/uuid"
)

// Conversation Stages
const (
	StageGreeting = "greeting"
	StageTopic    = "topic"
	StageResponse = "response"
	StageEnd      = "end"
)

// Conversation Outcomes
const (
	OutcomePositive = "positive"
	OutcomeNeutral  = "neutral"
	OutcomeNegative = "negative"
)

// DialogueTurn represents a single line of dialogue
type DialogueTurn struct {
	SpeakerID uuid.UUID `json:"speaker_id"`
	Text      string    `json:"text"`
	Emotion   string    `json:"emotion"`
}

// Conversation represents an active or completed interaction
type Conversation struct {
	ID           uuid.UUID      `json:"id"`
	InitiatorID  uuid.UUID      `json:"initiator_id"`
	ResponderID  uuid.UUID      `json:"responder_id"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Dialogue     []DialogueTurn `json:"dialogue"`
	Topic        string         `json:"topic"`
	Outcome      string         `json:"outcome"`
	CurrentStage string         `json:"current_stage"`
}

// InteractionContext holds data for initiation and flow
type InteractionContext struct {
	InitiatorID        uuid.UUID
	TargetID           uuid.UUID
	Distance           float64
	InitiatorAffection int // Affection towards target
	TargetAffection    int // Affection towards initiator
	SharedLocation     bool
	TimeIdle           float64 // Minutes
}
