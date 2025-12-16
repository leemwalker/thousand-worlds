// Package sapience provides detection and tracking of sapient species emergence.
// This implements both standard (natural evolution) and magic-assisted sapience thresholds.
package sapience

import (
	"math"

	"github.com/google/uuid"
)

// SapienceLevel represents the level of sapience/sentience
type SapienceLevel string

const (
	SapienceNone         SapienceLevel = "none"          // No significant cognition
	SapienceProtoSapient SapienceLevel = "proto_sapient" // Early tool use, basic communication
	SapienceSapient      SapienceLevel = "sapient"       // Full sapience - language, culture
	SapienceAdvanced     SapienceLevel = "advanced"      // Advanced technology/magic
)

// SapienceThresholds define the trait requirements for sapience
type SapienceThresholds struct {
	// Standard (non-magic) sapience thresholds
	StandardIntelligence  float64 `json:"standard_intelligence"`  // Minimum intelligence (0-10)
	StandardSocial        float64 `json:"standard_social"`        // Minimum social (0-10)
	StandardToolUse       float64 `json:"standard_tool_use"`      // Minimum tool use (0-10)
	StandardCommunication float64 `json:"standard_communication"` // Minimum communication (0-10)

	// Magic-assisted sapience thresholds (lower requirements)
	MagicIntelligence  float64 `json:"magic_intelligence"`   // Lower intelligence threshold
	MagicMagicAffinity float64 `json:"magic_magic_affinity"` // Minimum magic affinity (0-10)
	MagicSocial        float64 `json:"magic_social"`         // Lower social threshold

	// Proto-sapience thresholds (early signs)
	ProtoIntelligence float64 `json:"proto_intelligence"`
	ProtoToolUse      float64 `json:"proto_tool_use"`
}

// DefaultThresholds returns the standard sapience thresholds
func DefaultThresholds() SapienceThresholds {
	return SapienceThresholds{
		// Standard sapience: high intelligence, social, tool use, communication
		StandardIntelligence:  7.5,
		StandardSocial:        6.0,
		StandardToolUse:       5.0,
		StandardCommunication: 6.0,

		// Magic-assisted: lower cognitive requirements, but needs magic affinity
		MagicIntelligence:  5.0,
		MagicMagicAffinity: 7.0,
		MagicSocial:        5.0,

		// Proto-sapience: early signs
		ProtoIntelligence: 5.0,
		ProtoToolUse:      3.0,
	}
}

// SapienceCandidate represents a species that might achieve sapience
type SapienceCandidate struct {
	SpeciesID       uuid.UUID     `json:"species_id"`
	SpeciesName     string        `json:"species_name"`
	Level           SapienceLevel `json:"level"`
	Score           float64       `json:"score"` // 0.0-1.0, how close to sapience
	IsMagicAssisted bool          `json:"is_magic_assisted"`
	YearDetected    int64         `json:"year_detected"`

	// Trait values
	Intelligence  float64 `json:"intelligence"`
	Social        float64 `json:"social"`
	ToolUse       float64 `json:"tool_use"`
	Communication float64 `json:"communication"`
	MagicAffinity float64 `json:"magic_affinity"`

	// Additional factors
	PopulationSize  int64 `json:"population_size"`
	LongestLineage  int   `json:"longest_lineage"`  // Generations of sustained intelligence
	CulturalMarkers int   `json:"cultural_markers"` // # of advanced behaviors observed
}

// SapienceDetector monitors species for sapience emergence
type SapienceDetector struct {
	WorldID           uuid.UUID                        `json:"world_id"`
	Thresholds        SapienceThresholds               `json:"thresholds"`
	Candidates        map[uuid.UUID]*SapienceCandidate `json:"candidates"`
	SapientSpecies    []uuid.UUID                      `json:"sapient_species"`
	MagicEnabled      bool                             `json:"magic_enabled"` // World has magic
	CurrentYear       int64                            `json:"current_year"`
	FirstSapienceYear int64                            `json:"first_sapience_year"`
}

// NewSapienceDetector creates a new sapience detector
func NewSapienceDetector(worldID uuid.UUID, magicEnabled bool) *SapienceDetector {
	return &SapienceDetector{
		WorldID:        worldID,
		Thresholds:     DefaultThresholds(),
		Candidates:     make(map[uuid.UUID]*SapienceCandidate),
		SapientSpecies: make([]uuid.UUID, 0),
		MagicEnabled:   magicEnabled,
	}
}

// SpeciesTraits contains traits needed for sapience evaluation
type SpeciesTraits struct {
	Intelligence  float64
	Social        float64
	ToolUse       float64
	Communication float64
	MagicAffinity float64
	Population    int64
	Generation    int64
}

// Evaluate checks a species for sapience potential and updates candidates
func (sd *SapienceDetector) Evaluate(
	speciesID uuid.UUID,
	speciesName string,
	traits SpeciesTraits,
	year int64,
) *SapienceCandidate {
	sd.CurrentYear = year

	candidate := &SapienceCandidate{
		SpeciesID:      speciesID,
		SpeciesName:    speciesName,
		Intelligence:   traits.Intelligence,
		Social:         traits.Social,
		ToolUse:        traits.ToolUse,
		Communication:  traits.Communication,
		MagicAffinity:  traits.MagicAffinity,
		PopulationSize: traits.Population,
		YearDetected:   year,
	}

	// Calculate sapience score and level
	standardScore := sd.calculateStandardScore(traits)
	magicScore := 0.0
	if sd.MagicEnabled {
		magicScore = sd.calculateMagicScore(traits)
	}

	// Use higher of the two scores
	if magicScore > standardScore {
		candidate.Score = magicScore
		candidate.IsMagicAssisted = true
	} else {
		candidate.Score = standardScore
		candidate.IsMagicAssisted = false
	}

	// Determine level
	candidate.Level = sd.determineSapienceLevel(candidate)

	// Track candidates and sapient species
	if candidate.Level == SapienceProtoSapient || candidate.Level == SapienceSapient || candidate.Level == SapienceAdvanced {
		sd.Candidates[speciesID] = candidate

		if candidate.Level == SapienceSapient || candidate.Level == SapienceAdvanced {
			if !sd.isSapient(speciesID) {
				sd.SapientSpecies = append(sd.SapientSpecies, speciesID)
				if sd.FirstSapienceYear == 0 {
					sd.FirstSapienceYear = year
				}
			}
		}
	}

	return candidate
}

// calculateStandardScore calculates sapience score without magic
func (sd *SapienceDetector) calculateStandardScore(traits SpeciesTraits) float64 {
	th := sd.Thresholds

	// Each trait contributes to the score based on how close to threshold
	intScore := traits.Intelligence / th.StandardIntelligence
	socScore := traits.Social / th.StandardSocial
	toolScore := traits.ToolUse / th.StandardToolUse
	commScore := traits.Communication / th.StandardCommunication

	// Weighted average - intelligence most important
	score := (intScore*0.4 + socScore*0.2 + toolScore*0.2 + commScore*0.2)

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateMagicScore calculates sapience score with magic assistance
func (sd *SapienceDetector) calculateMagicScore(traits SpeciesTraits) float64 {
	th := sd.Thresholds

	// Need sufficient magic affinity
	if traits.MagicAffinity < th.MagicMagicAffinity*0.5 {
		return 0 // Not enough magic potential
	}

	intScore := traits.Intelligence / th.MagicIntelligence
	socScore := traits.Social / th.MagicSocial
	magicScore := traits.MagicAffinity / th.MagicMagicAffinity

	// Magic affinity can compensate for lower intelligence
	score := (intScore*0.3 + socScore*0.2 + magicScore*0.5)

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// determineSapienceLevel determines the sapience level from score and traits
func (sd *SapienceDetector) determineSapienceLevel(candidate *SapienceCandidate) SapienceLevel {
	th := sd.Thresholds

	// Check for full sapience
	if candidate.IsMagicAssisted {
		if candidate.Intelligence >= th.MagicIntelligence &&
			candidate.Social >= th.MagicSocial &&
			candidate.MagicAffinity >= th.MagicMagicAffinity {
			return SapienceSapient
		}
	} else {
		if candidate.Intelligence >= th.StandardIntelligence &&
			candidate.Social >= th.StandardSocial &&
			candidate.ToolUse >= th.StandardToolUse &&
			candidate.Communication >= th.StandardCommunication {
			return SapienceSapient
		}
	}

	// Check for proto-sapience
	if candidate.Intelligence >= th.ProtoIntelligence &&
		candidate.ToolUse >= th.ProtoToolUse {
		return SapienceProtoSapient
	}

	return SapienceNone
}

// isSapient checks if a species is already marked as sapient
func (sd *SapienceDetector) isSapient(speciesID uuid.UUID) bool {
	for _, id := range sd.SapientSpecies {
		if id == speciesID {
			return true
		}
	}
	return false
}

// GetCandidates returns all proto-sapient and sapient candidates
func (sd *SapienceDetector) GetCandidates() []*SapienceCandidate {
	candidates := make([]*SapienceCandidate, 0, len(sd.Candidates))
	for _, c := range sd.Candidates {
		candidates = append(candidates, c)
	}
	return candidates
}

// GetSapientCount returns the number of sapient species
func (sd *SapienceDetector) GetSapientCount() int {
	return len(sd.SapientSpecies)
}

// HasAnySapience returns true if any species has achieved sapience
func (sd *SapienceDetector) HasAnySapience() bool {
	return len(sd.SapientSpecies) > 0
}

// GetTopCandidate returns the species most likely to achieve sapience next
func (sd *SapienceDetector) GetTopCandidate() *SapienceCandidate {
	var top *SapienceCandidate
	for _, c := range sd.Candidates {
		if c.Level != SapienceSapient && c.Level != SapienceAdvanced {
			if top == nil || c.Score > top.Score {
				top = c
			}
		}
	}
	return top
}

// CalculateSapienceProgress returns how close the world is to first sapience (0.0-1.0)
func (sd *SapienceDetector) CalculateSapienceProgress() float64 {
	if sd.HasAnySapience() {
		return 1.0
	}

	maxScore := 0.0
	for _, c := range sd.Candidates {
		if c.Score > maxScore {
			maxScore = c.Score
		}
	}

	return maxScore
}

// PredictSapienceYear estimates when sapience might emerge based on current trends
func (sd *SapienceDetector) PredictSapienceYear(intelligenceGrowthPerMY float64) int64 {
	if sd.HasAnySapience() {
		return sd.FirstSapienceYear
	}

	top := sd.GetTopCandidate()
	if top == nil {
		return -1 // No candidates
	}

	// Calculate years needed to reach sapience threshold
	threshold := sd.Thresholds.StandardIntelligence
	if sd.MagicEnabled && top.MagicAffinity > 3.0 {
		threshold = sd.Thresholds.MagicIntelligence
	}

	if top.Intelligence >= threshold {
		return sd.CurrentYear // Already there, just need other traits
	}

	yearsNeeded := int64(math.Ceil((threshold - top.Intelligence) / intelligenceGrowthPerMY * 1000000))
	return sd.CurrentYear + yearsNeeded
}
