package sapience_test

import (
	"testing"

	"tw-backend/internal/ecosystem/sapience"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: Sapience Detector
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Sapience Detector Creation
// -----------------------------------------------------------------------------
// Given: A world ID
// When: NewSapienceDetector is called
// Then: Detector should be created with default thresholds
func TestBDD_SapienceDetector_Creation(t *testing.T) {
	worldID := uuid.New()

	detector := sapience.NewSapienceDetector(worldID, false)

	require.NotNil(t, detector, "Detector should be created")
	assert.Equal(t, worldID, detector.WorldID, "WorldID should match")
}

// -----------------------------------------------------------------------------
// Scenario: Low Intelligence Species - No Sapience
// -----------------------------------------------------------------------------
// Given: A species with low intelligence traits
// When: Evaluate is called
// Then: Sapience level should be None
func TestBDD_Sapience_LowIntelligence(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	speciesID := uuid.New()
	traits := sapience.SpeciesTraits{
		Intelligence:  2.0, // Low (scale is 0-10)
		Social:        3.0,
		ToolUse:       1.0,
		Communication: 1.5,
		MagicAffinity: 0,
		Population:    100000,
		Generation:    100,
	}

	candidate := detector.Evaluate(speciesID, "Primitive Fish", traits, 1000000)

	require.NotNil(t, candidate, "Should return candidate")
	assert.Equal(t, sapience.SapienceNone, candidate.Level,
		"Low intelligence species should not be sapient")
}

// -----------------------------------------------------------------------------
// Scenario: Proto-Sapient Species
// -----------------------------------------------------------------------------
// Given: A species with moderate intelligence and tool use
// When: Evaluate is called
// Then: Sapience level should be ProtoSapient
func TestBDD_Sapience_ProtoSapient(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	speciesID := uuid.New()
	traits := sapience.SpeciesTraits{
		Intelligence:  5.5, // Above proto threshold (5.0)
		Social:        5.0,
		ToolUse:       4.0, // Above proto tool use (3.0)
		Communication: 4.0,
		MagicAffinity: 0,
		Population:    50000,
		Generation:    200,
	}

	candidate := detector.Evaluate(speciesID, "Early Hominid", traits, 2000000)

	require.NotNil(t, candidate, "Should create candidate for proto-sapient")
	assert.Equal(t, sapience.SapienceProtoSapient, candidate.Level,
		"Moderate traits should be proto-sapient")
}

// -----------------------------------------------------------------------------
// Scenario: Full Sapience Achievement
// -----------------------------------------------------------------------------
// Given: A species meeting standard sapience thresholds
// When: Evaluate is called
// Then: Sapience level should be Sapient
func TestBDD_Sapience_FullSapience(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	speciesID := uuid.New()
	traits := sapience.SpeciesTraits{
		Intelligence:  8.0, // Above standard (7.5)
		Social:        7.0, // Above standard (6.0)
		ToolUse:       6.0, // Above standard (5.0)
		Communication: 7.0, // Above standard (6.0)
		MagicAffinity: 0,
		Population:    100000,
		Generation:    500,
	}

	candidate := detector.Evaluate(speciesID, "Homo Sapiens", traits, 3000000)

	require.NotNil(t, candidate, "Should create candidate for sapient species")
	assert.Equal(t, sapience.SapienceSapient, candidate.Level,
		"High traits should achieve full sapience")
}

// -----------------------------------------------------------------------------
// Scenario: Magic-Assisted Sapience
// -----------------------------------------------------------------------------
// Given: Magic is enabled and species has high magic affinity
// When: Evaluate is called with lower intelligence thresholds
// Then: Sapience may be achieved through magic path
func TestBDD_Sapience_MagicPath(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, true) // Magic enabled

	speciesID := uuid.New()
	traits := sapience.SpeciesTraits{
		Intelligence:  5.5, // Above magic threshold (5.0)
		Social:        5.5, // Above magic threshold (5.0)
		ToolUse:       3.0,
		Communication: 5.0,
		MagicAffinity: 8.0, // Above magic affinity threshold (7.0)
		Population:    10000,
		Generation:    100,
	}

	candidate := detector.Evaluate(speciesID, "Fey Creature", traits, 1000000)

	require.NotNil(t, candidate, "Should create candidate for magic-sapient species")
	assert.Equal(t, sapience.SapienceSapient, candidate.Level,
		"High magic affinity should achieve sapience via magic path")
	assert.True(t, candidate.IsMagicAssisted, "Should be magic-assisted")
}

// -----------------------------------------------------------------------------
// Scenario: Get Candidates
// -----------------------------------------------------------------------------
// Given: Multiple species have been evaluated
// When: GetCandidates is called
// Then: Should return all proto-sapient and sapient candidates
func TestBDD_Sapience_GetCandidates(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	// Evaluate a proto-sapient species
	detector.Evaluate(uuid.New(), "Hominid A", sapience.SpeciesTraits{
		Intelligence: 5.5, Social: 5.0, ToolUse: 4.0, Communication: 4.0,
	}, 1000000)

	candidates := detector.GetCandidates()

	assert.NotNil(t, candidates, "Should return candidates slice")
	assert.GreaterOrEqual(t, len(candidates), 1, "Should have at least one candidate")
}

// -----------------------------------------------------------------------------
// Scenario: Sapience Progress
// -----------------------------------------------------------------------------
// Given: A world with proto-sapient species
// When: CalculateSapienceProgress is called
// Then: Should return progress toward first full sapience (0-1)
func TestBDD_Sapience_Progress(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	// Initially no progress
	progress := detector.CalculateSapienceProgress()
	assert.GreaterOrEqual(t, progress, 0.0, "Progress should be >= 0")
	assert.LessOrEqual(t, progress, 1.0, "Progress should be <= 1")
}

// -----------------------------------------------------------------------------
// Scenario: Sapient Count
// -----------------------------------------------------------------------------
// Given: A detector with evaluated species
// When: GetSapientCount is called
// Then: Should return count of fully sapient species
func TestBDD_Sapience_SapientCount(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	// Initially no sapient species
	count := detector.GetSapientCount()
	assert.Equal(t, 0, count, "Should start with no sapient species")

	// Add a sapient species
	detector.Evaluate(uuid.New(), "Homo Sapiens", sapience.SpeciesTraits{
		Intelligence: 8.0, Social: 7.0, ToolUse: 6.0, Communication: 7.0, Population: 100000,
	}, 3000000)

	count = detector.GetSapientCount()
	assert.GreaterOrEqual(t, count, 1, "Should have sapient species after evaluation")
}

// -----------------------------------------------------------------------------
// Scenario: Has Any Sapience
// -----------------------------------------------------------------------------
// Given: A detector
// When: HasAnySapience is called
// Then: Should return true only if at least one species is sapient
func TestBDD_Sapience_HasAnySapience(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	// Initially false
	assert.False(t, detector.HasAnySapience(), "Should have no sapience initially")

	// Add a sapient species
	detector.Evaluate(uuid.New(), "Homo Sapiens", sapience.SpeciesTraits{
		Intelligence: 8.0, Social: 7.0, ToolUse: 6.0, Communication: 7.0, Population: 100000,
	}, 3000000)

	assert.True(t, detector.HasAnySapience(), "Should have sapience after evaluation")
}

// -----------------------------------------------------------------------------
// Scenario: Predict Sapience Year
// -----------------------------------------------------------------------------
// Given: A world with proto-sapient species evolving
// When: PredictSapienceYear is called
// Then: Should estimate year when sapience might emerge
func TestBDD_Sapience_PredictYear(t *testing.T) {
	worldID := uuid.New()
	detector := sapience.NewSapienceDetector(worldID, false)

	// With proto-sapient species
	detector.Evaluate(uuid.New(), "Evolving Hominid", sapience.SpeciesTraits{
		Intelligence: 5.5, Social: 5.0, ToolUse: 4.0, Communication: 4.0, Population: 50000,
	}, 2000000)

	predictedYear := detector.PredictSapienceYear(0.001) // Intelligence growth per million years

	assert.Greater(t, predictedYear, int64(0), "Predicted year should be positive")
}
