package population_test

import (
	"testing"

	"tw-backend/internal/ecosystem/population"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed seed for deterministic test results.
const testSeed int64 = 42

// =============================================================================
// BDD Tests: Population Simulator
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Population Simulator Creation
// -----------------------------------------------------------------------------
// Given: A world ID and seed
// When: NewPopulationSimulator is called
// Then: Simulator should be created with empty biomes
func TestBDD_PopulationSimulator_Creation(t *testing.T) {
	worldID := uuid.New()

	sim := population.NewPopulationSimulator(worldID, testSeed)

	require.NotNil(t, sim, "Simulator should be created")
	assert.NotNil(t, sim.Biomes, "Biomes map should be initialized")
	assert.Equal(t, int64(0), sim.CurrentYear, "Should start at year 0")
}

// -----------------------------------------------------------------------------
// Scenario: Metabolic Rate Calculation (Kleiber's Law)
// -----------------------------------------------------------------------------
// Given: Animals of different sizes
// When: CalculateMetabolicRate is called
// Then: Larger animals should have lower per-kg metabolic rate (more efficient)
func TestBDD_MetabolicRate_KleiberLaw(t *testing.T) {
	scenarios := []struct {
		name string
		size float64
	}{
		{"Mouse (size 0.1)", 0.1},
		{"Rabbit (size 2)", 2.0},
		{"Human (size 5)", 5.0},
		{"Elephant (size 50)", 50.0},
	}

	prevRate := 0.0
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			rate := population.CalculateMetabolicRate(sc.size)
			assert.Greater(t, rate, 0.0, "Metabolic rate should be positive")
			if prevRate > 0 {
				// Larger animals have lower per-kg rate (Kleiber's Law)
				// But total rate still increases with size
				assert.Greater(t, rate, prevRate, "Total metabolic rate should increase with size")
			}
			prevRate = rate
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Reproduction Modifier (r/K Selection)
// -----------------------------------------------------------------------------
// Given: Animals of different sizes
// When: CalculateReproductionModifier is called
// Then: Smaller animals should have higher reproduction rate
func TestBDD_ReproductionModifier_rKSelection(t *testing.T) {
	smallRate := population.CalculateReproductionModifier(0.1)
	largeRate := population.CalculateReproductionModifier(50.0)

	assert.Greater(t, smallRate, largeRate, "Small animals should reproduce faster (r-strategy)")
}

// -----------------------------------------------------------------------------
// Scenario: Juvenile Survival
// -----------------------------------------------------------------------------
// Given: Species with different traits
// When: CalculateJuvenileSurvival is called
// Then: K-strategists (larger, smarter) should have higher juvenile survival
func TestBDD_JuvenileSurvival(t *testing.T) {
	rStrategyTraits := population.EvolvableTraits{
		Size:         0.5,
		Intelligence: 10,
	}
	kStrategyTraits := population.EvolvableTraits{
		Size:         30.0,
		Intelligence: 80,
	}

	rSurvival := population.CalculateJuvenileSurvival(rStrategyTraits)
	kSurvival := population.CalculateJuvenileSurvival(kStrategyTraits)

	assert.GreaterOrEqual(t, kSurvival, rSurvival, "K-strategists should have higher juvenile survival")
}

// -----------------------------------------------------------------------------
// Scenario: Maturation Rate
// -----------------------------------------------------------------------------
// Given: Different maturity ages
// When: CalculateMaturationRate is called
// Then: Lower maturity age should give higher maturation rate
func TestBDD_MaturationRate(t *testing.T) {
	fastMaturation := population.CalculateMaturationRate(1.0)  // 1 year to maturity
	slowMaturation := population.CalculateMaturationRate(20.0) // 20 years to maturity

	assert.Greater(t, fastMaturation, slowMaturation, "Fast maturing species should have higher rate")
}

// -----------------------------------------------------------------------------
// Scenario: Continental Configuration Update
// -----------------------------------------------------------------------------
// Given: A population simulator
// When: UpdateContinentalConfiguration is called with drift event
// Then: Fragmentation should change
func TestBDD_ContinentalConfiguration_Update(t *testing.T) {
	worldID := uuid.New()
	sim := population.NewPopulationSimulator(worldID, testSeed)
	sim.ContinentalFragmentation = 0.5

	// Major drift event
	newFrag := sim.UpdateContinentalConfiguration(true, 0.8)

	assert.NotEqual(t, 0.5, newFrag, "Fragmentation should change with drift event")
}

// -----------------------------------------------------------------------------
// Scenario: Oxygen Level Effects
// -----------------------------------------------------------------------------
// Given: Different oxygen levels
// When: CalculateOxygenSizeModifier is called
// Then: High O2 should allow larger organisms
func TestBDD_OxygenEffects(t *testing.T) {
	lowO2Modifier := population.CalculateOxygenSizeModifier(0.15)    // 15%
	normalO2Modifier := population.CalculateOxygenSizeModifier(0.21) // 21%
	highO2Modifier := population.CalculateOxygenSizeModifier(0.35)   // 35% (Carboniferous)

	assert.Less(t, lowO2Modifier, normalO2Modifier, "Low O2 should limit size")
	assert.Greater(t, highO2Modifier, normalO2Modifier, "High O2 should allow larger sizes")
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Results
// -----------------------------------------------------------------------------
// Given: Same seed value
// When: PopulationSimulator is created twice
// Then: Initial state should be identical
func TestBDD_PopulationSimulator_Determinism(t *testing.T) {
	worldID := uuid.New()

	sim1 := population.NewPopulationSimulator(worldID, testSeed)
	sim2 := population.NewPopulationSimulator(worldID, testSeed)

	assert.Equal(t, sim1.ContinentalFragmentation, sim2.ContinentalFragmentation,
		"Initial fragmentation should be identical")
	assert.Equal(t, sim1.OxygenLevel, sim2.OxygenLevel,
		"Initial oxygen level should be identical")
}
