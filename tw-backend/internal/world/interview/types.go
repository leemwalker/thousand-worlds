package interview

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Category represents the section of world building
type Category string

const (
	CategoryTheme     Category = "Theme"
	CategoryTechLevel Category = "Tech Level"
	CategoryGeography Category = "Geography"
	CategoryGeology   Category = "Geology"
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

	// Geological Age
	GeologicalAge string // "young", "mature", "old"

	// Tech Level
	TechLevel    string // "stone_age", "medieval", "industrial", "modern", "futur", "mixed"
	MagicLevel   string // "none", "rare", "common", "dominant"
	AdvancedTech string
	MagicImpact  string

	// Geography
	PlanetSize        string
	ClimateRange      string
	LandWaterRatio    string
	UniqueFeatures    []string
	NaturalSatellites string // "none", "one", "many", "random"

	ExtremeEnvironments []string
	WaterLevel          string // "high", "low", "50%", etc.

	// Simulation Flags
	SimulateGeology bool
	SimulateLife    bool
	DisableDiseases bool

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

// Getter methods to implement orchestrator.WorldConfig interface
func (w *WorldConfiguration) GetPlanetSize() string        { return w.PlanetSize }
func (w *WorldConfiguration) GetLandWaterRatio() string    { return w.LandWaterRatio }
func (w *WorldConfiguration) GetClimateRange() string      { return w.ClimateRange }
func (w *WorldConfiguration) GetTechLevel() string         { return w.TechLevel }
func (w *WorldConfiguration) GetMagicLevel() string        { return w.MagicLevel }
func (w *WorldConfiguration) GetGeologicalAge() string     { return w.GeologicalAge }
func (w *WorldConfiguration) GetSentientSpecies() []string { return w.SentientSpecies }
func (w *WorldConfiguration) GetResourceDistribution() map[string]float64 {
	return w.ResourceDistribution
}
func (w *WorldConfiguration) GetSimulationFlags() map[string]bool {
	return map[string]bool{
		"simulate_geology": w.SimulateGeology,
		"simulate_life":    w.SimulateLife,
		"disable_diseases": w.DisableDiseases,
	}
}

func (w *WorldConfiguration) GetSeaLevel() *float64 {
	// Parse WaterLevel string if needed, or if we store it as float eventually
	// For now, let's parse the string "high" -> 0.8, "low" -> 0.2, etc.
	// Default nil if not set or standard
	if w.WaterLevel == "" {
		return nil
	}

	level := strings.ToLower(w.WaterLevel)
	var val float64
	if strings.Contains(level, "high") || strings.Contains(level, "flood") {
		val = 0.8
	} else if strings.Contains(level, "low") || strings.Contains(level, "dry") {
		val = 0.2
	} else if strings.Contains(level, "%") {
		// Try parsing percentage
		var percent float64
		if _, err := fmt.Sscanf(level, "%f%%", &percent); err == nil {
			val = percent / 100.0
		} else {
			return nil
		}
	} else {
		return nil
	}
	return &val
}

// GetSeed returns nil (random seed) since interviews don't specify seeds
func (w *WorldConfiguration) GetSeed() *int64 {
	return nil
}

// GetNaturalSatellites returns the natural satellites configuration
// Returns "none", "one", "many", "random", or a specific number
func (w *WorldConfiguration) GetNaturalSatellites() string {
	if w.NaturalSatellites == "" {
		return "random" // Default to random if not specified
	}
	return w.NaturalSatellites
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
