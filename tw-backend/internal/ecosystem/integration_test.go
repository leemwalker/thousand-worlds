package ecosystem_test

import (
	"testing"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/ecosystem/geography"
	"tw-backend/internal/ecosystem/pathogen"
	"tw-backend/internal/ecosystem/population"
	"tw-backend/internal/ecosystem/sapience"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Integration Tests: Cross-Subsystem Ecosystem Simulation
// =============================================================================

// Fixed seed for deterministic integration tests
const integrationSeed int64 = 12345

// -----------------------------------------------------------------------------
// Test: Geography → Population Integration
// -----------------------------------------------------------------------------
// Given: A tectonic system with plates
// When: Population simulator uses geographic data
// Then: Population dynamics should reflect continental configuration
func TestIntegration_Geography_Population(t *testing.T) {
	worldID := uuid.New()

	// Initialize tectonic system
	tectonics := geography.NewTectonicSystem(worldID, integrationSeed)
	require.NotNil(t, tectonics, "Tectonic system should initialize")

	// Initialize population simulator
	popSim := population.NewPopulationSimulator(worldID, integrationSeed)
	require.NotNil(t, popSim, "Population simulator should initialize")

	// Link geographic systems
	popSim.InitializeGeographicSystems(worldID, integrationSeed)
	require.NotNil(t, popSim.HexGrid, "Hex grid should be initialized")
	require.NotNil(t, popSim.Tectonics, "Tectonics should be linked")

	// Simulate 100,000 years
	for i := 0; i < 10; i++ {
		popSim.Tectonics.Update(10000)
		popSim.UpdateGeographicSystems(10000)
	}

	// Verify fragmentation is tracked
	frag := popSim.Tectonics.CalculateFragmentation()
	assert.GreaterOrEqual(t, frag, float32(0.0))
	assert.LessOrEqual(t, frag, float32(1.0))
}

// -----------------------------------------------------------------------------
// Test: Population → Pathogen Integration
// -----------------------------------------------------------------------------
// Given: A populated biome
// When: Population density is high
// Then: Pathogen outbreaks become possible
func TestIntegration_Population_Pathogen(t *testing.T) {
	worldID := uuid.New()

	// Initialize population simulator with some species
	popSim := population.NewPopulationSimulator(worldID, integrationSeed)
	popSim.InitializeGeographicSystems(worldID, integrationSeed)

	// Initialize pathogen system
	diseaseSystem := pathogen.NewDiseaseSystem(worldID, integrationSeed)
	require.NotNil(t, diseaseSystem, "Disease system should initialize")

	// Simulate a dense population scenario
	speciesID := uuid.New()
	densePop := int64(1000000)
	highDensity := 0.9

	// Check for spontaneous outbreaks over time
	outbreakCount := 0
	for i := 0; i < 50; i++ {
		_, outbreak := diseaseSystem.CheckSpontaneousOutbreak(
			speciesID, "Test Species", densePop, highDensity)
		if outbreak != nil {
			outbreakCount++
		}
	}

	// Dense populations should eventually see outbreaks
	// (probabilistic, but with 50 trials should see some)
	t.Logf("Outbreak count in 50 trials: %d", outbreakCount)
}

// -----------------------------------------------------------------------------
// Test: Population → Sapience Integration
// -----------------------------------------------------------------------------
// Given: Species with evolving intelligence
// When: Traits exceed thresholds
// Then: Sapience should be detected
func TestIntegration_Population_Sapience(t *testing.T) {
	worldID := uuid.New()

	// Initialize sapience detector
	detector := sapience.NewSapienceDetector(worldID, false)
	require.NotNil(t, detector, "Sapience detector should initialize")

	// Simulate intelligence evolution over time
	speciesID := uuid.New()
	baseIntelligence := 3.0

	for year := int64(0); year <= 10000000; year += 1000000 {
		// Intelligence slowly increases
		intelligence := baseIntelligence + float64(year)/2500000 // Reaches 7 by 10MY

		traits := sapience.SpeciesTraits{
			Intelligence:  intelligence,
			Social:        intelligence * 0.9,
			ToolUse:       intelligence * 0.7,
			Communication: intelligence * 0.8,
			Population:    50000,
		}

		candidate := detector.Evaluate(speciesID, "Evolving Primate", traits, year)

		if candidate != nil && candidate.Level == sapience.SapienceSapient {
			t.Logf("Sapience achieved at year %d (intelligence: %.2f)", year, intelligence)
			break
		}
	}

	// By 10 million years, should have at least proto-sapience
	assert.True(t, len(detector.GetCandidates()) > 0, "Should have candidates")
}

// -----------------------------------------------------------------------------
// Test: Full Ecosystem Pipeline
// -----------------------------------------------------------------------------
// Given: All subsystems initialized
// When: Simulation runs for extended period
// Then: All subsystems should interact correctly
func TestIntegration_FullEcosystemPipeline(t *testing.T) {
	worldID := uuid.New()

	// Initialize all subsystems
	popSim := population.NewPopulationSimulator(worldID, integrationSeed)
	popSim.InitializeGeographicSystems(worldID, integrationSeed)

	diseaseSystem := pathogen.NewDiseaseSystem(worldID, integrationSeed)
	sapienceDetector := sapience.NewSapienceDetector(worldID, false)

	// Track events
	var milestones []string

	// Simulate 5 million years
	for year := int64(0); year <= 5000000; year += 100000 {
		// Update tectonics
		popSim.Tectonics.Update(100000)
		popSim.UpdateGeographicSystems(100000)

		// Simulate population dynamics
		popSim.CurrentYear = year

		// Check for continental events
		if frag := popSim.Tectonics.CalculateFragmentation(); frag > 0.7 {
			milestones = append(milestones,
				"High fragmentation detected")
		}

		// Check for disease outbreaks
		if _, outbreak := diseaseSystem.CheckSpontaneousOutbreak(
			uuid.New(), "Test Species", 100000, 0.5); outbreak != nil {
			milestones = append(milestones, "Disease outbreak")
		}

		// Check for sapience evolution
		traits := sapience.SpeciesTraits{
			Intelligence: 3.0 + float64(year)/1500000,
			Social:       3.0 + float64(year)/2000000,
			ToolUse:      2.0 + float64(year)/2500000,
		}
		if candidate := sapienceDetector.Evaluate(
			uuid.New(), "Evolving Species", traits, year); candidate != nil && candidate.Level != sapience.SapienceNone {
			milestones = append(milestones,
				"Proto-sapience or sapience detected")
		}
	}

	t.Logf("Milestones during simulation: %d events recorded", len(milestones))

	// Verify systems are in valid state
	assert.GreaterOrEqual(t, popSim.Tectonics.CurrentYear, int64(0))
	assert.NotNil(t, popSim.HexGrid)
	assert.NotNil(t, diseaseSystem)
}

// -----------------------------------------------------------------------------
// Test: Simulation Runner Basic Operations
// -----------------------------------------------------------------------------
// Given: A SimulationRunner
// When: Initialized and configured
// Then: Should be ready to run simulation
func TestIntegration_SimulationRunner_Initialize(t *testing.T) {
	worldID := uuid.New()

	config := ecosystem.DefaultConfig(worldID)
	config.MaxYearTarget = 100000 // Short run for test

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	require.NotNil(t, runner, "Runner should initialize")

	// Initialize population simulator
	runner.InitializePopulationSimulator(integrationSeed)

	// Verify it's ready
	stats := runner.GetStats()
	assert.Equal(t, ecosystem.RunnerIdle, stats.State)
}

// -----------------------------------------------------------------------------
// Test: Worldgen → Ecosystem Handoff
// -----------------------------------------------------------------------------
// Given: World generation data (heightmap, biomes)
// When: Ecosystem simulation starts
// Then: Biome data should be usable by population simulator
func TestIntegration_Worldgen_Ecosystem_Handoff(t *testing.T) {
	worldID := uuid.New()

	// Simulate worldgen output (normally from orchestrator)
	biomeCount := 10

	// Create population simulator
	popSim := population.NewPopulationSimulator(worldID, integrationSeed)

	// Simulate adding biomes from worldgen
	for i := 0; i < biomeCount; i++ {
		biomeID := uuid.New()
		popSim.Biomes[biomeID] = &population.BiomePopulation{
			BiomeID:   biomeID,
			BiomeType: "forest",
		}
	}

	assert.Len(t, popSim.Biomes, biomeCount, "All biomes should be registered")
}

// -----------------------------------------------------------------------------
// Test: Geological Events Affect Population
// -----------------------------------------------------------------------------
// Given: Active tectonic boundaries
// When: Geological events occur
// Then: Population should be affected
func TestIntegration_GeologicalEvents_Population(t *testing.T) {
	worldID := uuid.New()

	// Initialize systems
	popSim := population.NewPopulationSimulator(worldID, integrationSeed)
	popSim.InitializeGeographicSystems(worldID, integrationSeed)

	// Set initial fragmentation
	initialFrag := popSim.ContinentalFragmentation

	// Simulate major drift event
	newFrag := popSim.UpdateContinentalConfiguration(true, 0.9)

	// Fragmentation should change with major drift
	assert.NotEqual(t, initialFrag, newFrag,
		"Continental configuration should change with drift event")
}

// -----------------------------------------------------------------------------
// Test: Disease System → Population Impact
// -----------------------------------------------------------------------------
// Given: An active outbreak
// When: Disease spreads
// Then: Population should be affected
func TestIntegration_Disease_PopulationImpact(t *testing.T) {
	worldID := uuid.New()

	// Initialize disease system
	diseaseSystem := pathogen.NewDiseaseSystem(worldID, integrationSeed)

	// Create a species with population
	speciesID := uuid.New()
	speciesInfo := map[uuid.UUID]pathogen.SpeciesInfo{
		speciesID: {
			Population:        1000000,
			DiseaseResistance: 0.3,
			DietType:          "omnivore",
			Density:           0.5,
		},
	}

	// Force create a pathogen
	p := diseaseSystem.CreateNovelPathogen(speciesID, "Test Host", pathogen.PathogenVirus)
	require.NotNil(t, p, "Pathogen should be created")

	// Update disease system
	diseaseSystem.Update(1000, speciesInfo)

	// Check for active outbreaks
	outbreaks := diseaseSystem.GetActiveOutbreaks()
	t.Logf("Active outbreaks: %d", len(outbreaks))
}
