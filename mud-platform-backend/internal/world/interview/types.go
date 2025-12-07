package interview

import (
	"time"

	"github.com/google/uuid"
)

// Category represents the section of world building
type Category string

const (
	CategoryTheme     Category = "Theme"
	CategoryTechLevel Category = "Tech Level"
	CategoryGeography Category = "Geography"
	CategoryCulture   Category = "Culture"
)

// Topic represents a specific question area within a category
type Topic struct {
	Category    Category
	Name        string
	Description string // Internal description for the LLM
}

// InterviewState tracks the progress of the interview
type InterviewState struct {
	CurrentCategory   Category
	CurrentTopicIndex int
	Answers           map[string]string // Topic Name -> Player Answer
	IsComplete        bool
}

// InterviewSession represents an active interview
type InterviewSession struct {
	ID        uuid.UUID
	PlayerID  uuid.UUID
	State     InterviewState
	History   []ConversationTurn // For context
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ConversationTurn holds a single exchange
type ConversationTurn struct {
	Question string
	Answer   string
}

// WorldConfiguration represents the structured world parameters extracted from interview
type WorldConfiguration struct {
	ID          uuid.UUID
	InterviewID uuid.UUID
	WorldID     *uuid.UUID // Optional, set when world is generated
	CreatedBy   uuid.UUID

	// Identity
	WorldName string

	// Theme
	Theme          string
	Tone           string
	Inspirations   []string
	UniqueAspect   string
	MajorConflicts []string

	// Tech Level
	TechLevel    string // "stone_age", "medieval", "industrial", "modern", "futuristic", "mixed"
	MagicLevel   string // "none", "rare", "common", "dominant"
	AdvancedTech string
	MagicImpact  string

	// Geography
	PlanetSize          string
	ClimateRange        string
	LandWaterRatio      string
	UniqueFeatures      []string
	ExtremeEnvironments []string

	// Culture
	SentientSpecies    []string
	PoliticalStructure string
	CulturalValues     []string
	EconomicSystem     string
	Religions          []string
	Taboos             []string

	// Generation Parameters (derived)
	BiomeWeights           map[string]float64
	ResourceDistribution   map[string]float64
	SpeciesStartAttributes map[string]interface{}

	CreatedAt time.Time
}

// Status represents the state of an interview
type Status string

const (
	StatusNotStarted Status = "not_started"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

// Interview represents a world creation interview
type Interview struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Status               Status
	CurrentQuestionIndex int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Answer represents a user's answer to an interview question
type Answer struct {
	ID            uuid.UUID
	InterviewID   uuid.UUID
	QuestionIndex int
	AnswerText    string
	CreatedAt     time.Time
}
