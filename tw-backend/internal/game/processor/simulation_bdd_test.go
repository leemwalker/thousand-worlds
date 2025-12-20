package processor_test

import (
	"testing"

	"tw-backend/internal/game/processor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: World Simulate Command
// =============================================================================
// These tests verify simulation command parsing and configuration.
// The full simulation integration is tested in runner_bdd_test.go.

// -----------------------------------------------------------------------------
// Scenario: Basic Argument Parsing
// -----------------------------------------------------------------------------
// Given: A simple simulation command string
// When: ParseSimulationArgs is called
// Then: Config should be populated with years
func TestBDD_WorldSimulate_BasicParsing(t *testing.T) {
	config := processor.ParseSimulationArgs("1000000")

	require.NotNil(t, config, "ParseSimulationArgs should return a config, not nil")
	assert.Equal(t, int64(1000000), config.Years, "Years should be parsed correctly")
	assert.True(t, config.SimulateGeology, "Geology should be enabled by default")
	assert.True(t, config.SimulateLife, "Life should be enabled by default")
	assert.True(t, config.SimulateDiseases, "Diseases should be enabled by default")
}

// -----------------------------------------------------------------------------
// Scenario: Simulation Flags (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various command strings with flags
// When: ParseSimulationArgs is called
// Then: The config should match expected state
func TestBDD_WorldSimulate_Flags(t *testing.T) {
	scenarios := []struct {
		name             string
		command          string
		expectGeology    bool
		expectLife       bool
		expectDiseases   bool
		expectWaterLevel string
	}{
		{"Default flags", "100", true, true, true, ""},
		{"Only geology", "100 --only-geology", true, false, false, ""},
		{"Only life", "100 --only-life", false, true, true, ""},
		{"No diseases", "100 --no-diseases", true, true, false, ""},
		{"Water level high", "100 --water-level high", true, true, true, "high"},
		{"Water level 90%", "100 --water-level 90%", true, true, true, "90%"},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := processor.ParseSimulationArgs(sc.command)

			require.NotNil(t, config, "ParseSimulationArgs should return a config")
			assert.Equal(t, sc.expectGeology, config.SimulateGeology, "SimulateGeology mismatch")
			assert.Equal(t, sc.expectLife, config.SimulateLife, "SimulateLife mismatch")
			assert.Equal(t, sc.expectDiseases, config.SimulateDiseases, "SimulateDiseases mismatch")
			assert.Equal(t, sc.expectWaterLevel, config.WaterLevel, "WaterLevel mismatch")
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Epoch Labeling
// -----------------------------------------------------------------------------
// Given: Command with --epoch flag
// When: ParseSimulationArgs is called
// Then: Epoch should be captured in config
func TestBDD_WorldSimulate_EpochLabel(t *testing.T) {
	config := processor.ParseSimulationArgs("100000000 --epoch Jurassic")

	require.NotNil(t, config, "ParseSimulationArgs should return a config")
	assert.Equal(t, "Jurassic", config.Epoch, "Epoch should be parsed")
	assert.Equal(t, int64(100000000), config.Years, "Years should be parsed")
}

// -----------------------------------------------------------------------------
// Scenario: Goal Flag
// -----------------------------------------------------------------------------
// Given: Command with --goal flag
// When: ParseSimulationArgs is called
// Then: Goal should be captured in config
func TestBDD_WorldSimulate_GoalFlag(t *testing.T) {
	config := processor.ParseSimulationArgs("1000000 --goal sapience")

	require.NotNil(t, config, "ParseSimulationArgs should return a config")
	assert.Equal(t, "sapience", config.Goal, "Goal should be parsed")
}

// -----------------------------------------------------------------------------
// Scenario: Combined Flags
// -----------------------------------------------------------------------------
// Given: Command with multiple flags
// When: ParseSimulationArgs is called
// Then: All flags should be correctly parsed
func TestBDD_WorldSimulate_CombinedFlags(t *testing.T) {
	config := processor.ParseSimulationArgs("1000000 --only-geology --epoch Hadean")

	require.NotNil(t, config, "ParseSimulationArgs should return a config")
	assert.True(t, config.SimulateGeology, "Geology should be enabled")
	assert.False(t, config.SimulateLife, "Life should be disabled with --only-geology")
	assert.Equal(t, "Hadean", config.Epoch, "Epoch should be parsed")
}

// -----------------------------------------------------------------------------
// Scenario: Input Validation - Negative Years
// -----------------------------------------------------------------------------
// Given: Invalid negative year value
// When: ParseSimulationArgs is called
// Then: Should return nil (invalid input)
func TestBDD_WorldSimulate_InvalidNegativeYears(t *testing.T) {
	config := processor.ParseSimulationArgs("-100")

	assert.Nil(t, config, "Negative years should return nil config")
}

// -----------------------------------------------------------------------------
// Scenario: Input Validation - Zero Years
// -----------------------------------------------------------------------------
// Given: Zero year value
// When: ParseSimulationArgs is called
// Then: Should return nil (invalid input)
func TestBDD_WorldSimulate_InvalidZeroYears(t *testing.T) {
	config := processor.ParseSimulationArgs("0")

	assert.Nil(t, config, "Zero years should return nil config")
}

// -----------------------------------------------------------------------------
// Scenario: Input Validation - Non-Numeric Years
// -----------------------------------------------------------------------------
// Given: Non-numeric year value
// When: ParseSimulationArgs is called
// Then: Should return nil (invalid input)
func TestBDD_WorldSimulate_InvalidNonNumeric(t *testing.T) {
	config := processor.ParseSimulationArgs("abc")

	assert.Nil(t, config, "Non-numeric years should return nil config")
}

// -----------------------------------------------------------------------------
// Scenario: Input Validation - Excessive Years
// -----------------------------------------------------------------------------
// Given: Years exceeding max simulation cap (e.g., > 10 billion)
// When: ParseSimulationArgs is called
// Then: Should return nil or cap the value
func TestBDD_WorldSimulate_ExcessiveYears(t *testing.T) {
	config := processor.ParseSimulationArgs("100000000000") // 100 billion

	// Either nil or capped to max allowed
	if config != nil {
		assert.LessOrEqual(t, config.Years, int64(10_000_000_000),
			"Years should be capped at 10 billion max")
	}
}

// -----------------------------------------------------------------------------
// Scenario: Empty Input
// -----------------------------------------------------------------------------
// Given: Empty string input
// When: ParseSimulationArgs is called
// Then: Should return nil
func TestBDD_WorldSimulate_EmptyInput(t *testing.T) {
	config := processor.ParseSimulationArgs("")

	assert.Nil(t, config, "Empty input should return nil config")
}

// -----------------------------------------------------------------------------
// Scenario: Whitespace Only Input
// -----------------------------------------------------------------------------
// Given: Whitespace-only input
// When: ParseSimulationArgs is called
// Then: Should return nil
func TestBDD_WorldSimulate_WhitespaceInput(t *testing.T) {
	config := processor.ParseSimulationArgs("   ")

	assert.Nil(t, config, "Whitespace-only input should return nil config")
}
