package evolution_test

import (
	"testing"

	"tw-backend/internal/worldgen/evolution"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: Evolution
// =============================================================================
// These tests verify evolutionary mechanics including mass extinctions,
// O2-size relationships, and diversification events.

// -----------------------------------------------------------------------------
// Scenario: Cambrian Explosion
// -----------------------------------------------------------------------------
// Given: Stable ocean environment with O2 > 10%
// When: 50 million years of evolution simulated
// Then: Species count should increase 10x or more
//
//	AND Body plans should diversify dramatically
//	AND Predation should emerge
func TestBDD_CambrianExplosion_RapidDiversification(t *testing.T) {
	// Create initial species (simple organisms)
	initialSpecies := []*evolution.Species{
		{SpeciesID: uuid.New(), Name: "Primordial Alga", Type: evolution.SpeciesFlora, Population: 1000},
		{SpeciesID: uuid.New(), Name: "Early Worm", Type: evolution.SpeciesHerbivore, Population: 500},
	}
	initialCount := len(initialSpecies)

	// Simulate Cambrian explosion with O2 at 12%
	newSpecies := evolution.SimulateCambrianExplosion(initialSpecies, 0.12, 50_000_000)

	// Should have 10x species increase
	expectedMinNew := initialCount * 10
	assert.GreaterOrEqual(t, newSpecies, expectedMinNew,
		"Cambrian explosion should create at least 10x species (expected %d, got %d)", expectedMinNew, newSpecies)
}

// -----------------------------------------------------------------------------
// Scenario: K-Pg Extinction Event
// -----------------------------------------------------------------------------
// Given: Diverse ecosystem with large dinosaurs
// When: Asteroid impact event (severity 0.9)
// Then: 75% of species should go extinct
//
//	AND Large animals (size > 5) should suffer most
//	AND Small mammals should survive
func TestBDD_KPgExtinction_MassExtinction(t *testing.T) {
	// Create diverse ecosystem
	species := []*evolution.Species{
		{SpeciesID: uuid.New(), Name: "T-Rex", Type: evolution.SpeciesCarnivore, Size: 8.0, Population: 100},
		{SpeciesID: uuid.New(), Name: "Brontosaurus", Type: evolution.SpeciesHerbivore, Size: 30.0, Population: 200},
		{SpeciesID: uuid.New(), Name: "Small Mammal", Type: evolution.SpeciesHerbivore, Size: 0.5, Population: 10000},
		{SpeciesID: uuid.New(), Name: "Fern", Type: evolution.SpeciesFlora, Size: 0.3, Population: 50000},
	}
	beforeCount := len(species)

	// Apply K-Pg extinction event
	event := evolution.ExtinctionEvent{Type: "asteroid", Severity: 0.9}
	extinctCount := evolution.ApplyExtinctionEvent(species, event)

	// 75%+ should go extinct
	expectedMinExtinct := int(float64(beforeCount) * 0.75)
	assert.GreaterOrEqual(t, extinctCount, expectedMinExtinct,
		"Asteroid impact should kill 75%+ of species (expected %d, got %d)", expectedMinExtinct, extinctCount)
}

// -----------------------------------------------------------------------------
// Scenario: Adaptive Radiation (Darwin's Finches)
// -----------------------------------------------------------------------------
// Given: Single ancestral species colonizes island archipelago
// When: Islands have different food sources
// Then: Multiple descendant species should evolve
//
//	AND Beak shapes should diverge based on food type
func TestBDD_AdaptiveRadiation_DarwinFinches(t *testing.T) {
	// Create ancestral finch
	ancestor := &evolution.Species{
		SpeciesID:  uuid.New(),
		Name:       "Ancestral Finch",
		Type:       evolution.SpeciesOmnivore,
		Population: 1000,
	}

	// Islands with different food sources would cause speciation
	niches := []string{"seed-eating", "insect-eating", "nectar-feeding", "cactus-feeding"}

	// Simulate adaptive radiation
	newSpeciesCount := evolution.SimulateAdaptiveRadiation(ancestor, niches, 0.8, 42)

	assert.Greater(t, newSpeciesCount, 0, "Should produce at least one new species")
	assert.GreaterOrEqual(t, newSpeciesCount, len(niches)/2, "Should fill multiple niches")
}

// -----------------------------------------------------------------------------
// Scenario: Convergent Evolution
// -----------------------------------------------------------------------------
// Given: Unrelated species in similar environments
// When: Evolution proceeds for millions of years
// Then: Similar traits should evolve independently
//
//	AND Eyes, wings, echolocation should appear multiple times
func TestBDD_ConvergentEvolution_SimilarTraits(t *testing.T) {
	// Create unrelated species with similar traits (high speed)
	species := []*evolution.Species{
		{SpeciesID: uuid.New(), Name: "Cheetah", Type: evolution.SpeciesCarnivore, Speed: 15.0},
		{SpeciesID: uuid.New(), Name: "Pronghorn", Type: evolution.SpeciesHerbivore, Speed: 12.0},
		{SpeciesID: uuid.New(), Name: "Slow Turtle", Type: evolution.SpeciesHerbivore, Speed: 0.5},
	}

	// Detect convergent traits
	traits := evolution.DetectConvergentEvolution(species, "grassland")

	assert.Greater(t, len(traits), 0, "Should detect convergent traits")
	assert.Equal(t, "High Speed", traits[0].TraitName, "Should detect speed convergence")
}

// -----------------------------------------------------------------------------
// Scenario: Island Dwarfism
// -----------------------------------------------------------------------------
// Given: Large continental species
// When: Isolated on small island with limited resources
// Then: Descendants should decrease in size over time
//
//	AND Size reduction should be proportional to resource scarcity
func TestBDD_IslandDwarfism_SizeReduction(t *testing.T) {
	elephant := &evolution.Species{
		SpeciesID:  uuid.New(),
		Name:       "Continental Elephant",
		Type:       evolution.SpeciesHerbivore,
		Size:       8.0,
		Population: 100,
	}

	// Configure small island with limited resources
	config := evolution.IsolationConfig{
		IslandArea:      100,       // Small island
		ResourceDensity: 0.3,       // Low resources
		Years:           1_000_000, // 1 million years
	}

	// Simulate isolation
	sizeMultiplier := evolution.SimulateIsolation(elephant, config)

	// Should experience dwarfism (multiplier < 1)
	assert.Less(t, sizeMultiplier, 1.0, "Large animals should shrink on islands")
	assert.Greater(t, sizeMultiplier, 0.0, "Should still have positive size")
}

// -----------------------------------------------------------------------------
// Scenario: Giant Insects at High O2 (Carboniferous)
// -----------------------------------------------------------------------------
// Given: Atmospheric O2 at 35%
// When: Arthropod species evolve
// Then: Maximum arthropod size should allow giants (2m dragonflies)
//
//	AND O2 cycle should feed back from flora biomass
func TestBDD_GiantInsects_HighOxygen(t *testing.T) {
	// High O2 allows giant arthropods
	maxSize, multiplier := evolution.CalculateO2Effects(0.35, 0.0)

	assert.GreaterOrEqual(t, maxSize, 3.0,
		"O2 at 35%% should allow giant arthropods (size >= 3.0), got %.2f", maxSize)
	assert.Greater(t, multiplier, 1.0,
		"O2 at 35%% should increase size multiplier, got %.2f", multiplier)
}

// -----------------------------------------------------------------------------
// Scenario: O2 Cycle - Flora Production
// -----------------------------------------------------------------------------
// Given: High flora biomass (forests)
// When: Photosynthesis occurs
// Then: Atmospheric O2 should increase
//
//	AND O2 level should be bounded (10%-40%)
func TestBDD_O2Cycle_FloraProduction(t *testing.T) {
	// Pseudocode became implementation:
	currentO2 := 0.10
	floraBiomass := 100.0

	newO2 := evolution.UpdateAtmosphere(currentO2, floraBiomass)

	assert.Greater(t, newO2, currentO2, "O2 should increase with flora biomass")
	assert.LessOrEqual(t, newO2, 0.40, "O2 should be capped at 40%")
}

// -----------------------------------------------------------------------------
// Scenario: Solar Evolution - Biosphere Stress
// -----------------------------------------------------------------------------
// Given: Simulation 800 million years in the future
// When: Solar luminosity increases by 8%
// Then: Global temperature should increase
//
//	AND Biosphere should experience stress (fitness reduction)
//	AND Desert areas should expand
func TestBDD_SolarEvolution_BiosphereStress(t *testing.T) {
	// Pseudocode became implementation:
	luminosity := evolution.CalculateSolarLuminosity(0.8) // 800M years

	// Linear: 1.0 + 0.08 * 0.8 = 1.064
	assert.Greater(t, luminosity, 1.05, "Solar luminosity should increase over time")
	assert.Less(t, luminosity, 1.1, "Solar luminosity should be reasonable")
}

// -----------------------------------------------------------------------------
// Scenario: Genetic Mutation Mechanics (Table-Driven)
// -----------------------------------------------------------------------------
// Given: A base genetic code
// When: Specific mutation operators are applied
// Then: The resulting code should reflect the specific change type
func TestBDD_Genetics_MutationTypes(t *testing.T) {
	scenarios := []struct {
		name         string
		operator     string // Point, Insertion, Deletion, Duplication
		expectLength int    // 0 = same, 1 = longer, -1 = shorter
	}{
		{"Point Mutation", "point", 0},    // ATCG -> ACCG (Same length)
		{"Insertion", "insertion", 1},     // ATCG -> ATCCG (Longer)
		{"Deletion", "deletion", -1},      // ATCG -> ACG (Shorter)
		{"Duplication", "duplication", 1}, // Gene doubling
	}

	baseGenome := "ATCGATCGATCG" // 12 base pairs

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			mutated := evolution.ApplyMutationOperator(baseGenome, evolution.MutationOperator(sc.operator), 5, 42)

			switch sc.expectLength {
			case 0:
				assert.Equal(t, len(baseGenome), len(mutated), "Point mutation should preserve length")
				assert.NotEqual(t, baseGenome, mutated, "Point mutation should change genome")
			case 1:
				assert.Greater(t, len(mutated), len(baseGenome), "Insertion/Duplication should increase length")
			case -1:
				assert.Less(t, len(mutated), len(baseGenome), "Deletion should decrease length")
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Co-Evolutionary Arms Race (Predator vs Prey)
// -----------------------------------------------------------------------------
// Given: A predator species and a prey species
// When: Prey evolves "Speed" trait to escape
// Then: Predators should subsequently evolve "Speed" or "Ambush" to compensate
//
//	AND Both populations should oscillate rather than one wiping out the other
func TestBDD_CoEvolution_RedQueen(t *testing.T) {
	predator := &evolution.Species{
		SpeciesID:  uuid.New(),
		Name:       "Fast Predator",
		Type:       evolution.SpeciesCarnivore,
		Speed:      11.0,
		Population: 50,
	}
	prey := &evolution.Species{
		SpeciesID:  uuid.New(),
		Name:       "Slower Prey",
		Type:       evolution.SpeciesHerbivore,
		Speed:      10.0,
		Population: 500,
	}

	// Simulate co-evolution over 100 generations
	result := evolution.SimulateCoEvolution(predator, prey, 100, 42)

	// Prey should speed up to escape predators
	assert.Greater(t, result.PreySpeedChange, 0.0, "Prey should evolve faster speed")
	// Arms race should produce changes in both
	assert.NotEqual(t, 0.0, result.PredatorSpeedChange+result.PreySpeedChange, "Arms race should cause changes")
}

// -----------------------------------------------------------------------------
// Scenario: Genetic Drift in Small Populations
// -----------------------------------------------------------------------------
// Given: A small population (N=10) with 50% Red and 50% Blue alleles
// When: No selective pressure is applied (neutral traits)
// Then: Allele frequencies should fluctuate randomly
//
//	AND Eventually one allele should fixate (reach 100%) purely by chance
func TestBDD_Evolution_GeneticDrift(t *testing.T) {
	// Start with 50% Red, 50% Blue alleles
	alleles := []evolution.AlleleFrequency{
		{Allele: "Red", Frequency: 0.5},
		{Allele: "Blue", Frequency: 0.5},
	}

	// Small population (N=10) over 50 generations
	result := evolution.SimulateGeneticDrift(alleles, 10, 50, 42)

	require.Len(t, result, 2, "Should still have two alleles")

	// Frequencies should have changed from initial 50/50
	changed := result[0].Frequency != 0.5 || result[1].Frequency != 0.5
	assert.True(t, changed, "Allele frequencies should fluctuate due to drift")
}

// -----------------------------------------------------------------------------
// Scenario: Trophic Cascade / Over-Grazing
// -----------------------------------------------------------------------------
// Given: Herbivores become extremely efficient (high metabolism, high reproduction)
// When: Plant biomass is depleted below recovery threshold
// Then: Herbivore population should crash (starvation)
//
//	AND Plant biomass should eventually recover (predator-prey cycles)
func TestBDD_Ecosystem_TrophicCollapse(t *testing.T) {
	// Pseudocode became implementation:
	herbivores := []*evolution.Species{
		{SpeciesID: uuid.New(), Population: 50000},
		{SpeciesID: uuid.New(), Population: 30000},
	}

	// Low biomass, high herbivore pop -> starvation
	result := evolution.SimulateTrophicDynamics(herbivores, 100.0, 50.0, 30.0)

	assert.True(t, result.StarvationOccurred, "Starvation should occur when herbivores exceed capacity")
	assert.Less(t, result.HerbivorePop, 80000, "Population should crash")
}

// -----------------------------------------------------------------------------
// Scenario: Emergence of Sapience
// -----------------------------------------------------------------------------
// Given: A social species with high intelligence and manipulative appendages
// When: Environmental pressure requires complex problem solving (e.g., Ice Age)
// Then: "Tool Use" trait should appear
//
//	AND Sapience flag should eventually trigger
func TestBDD_Evolution_SapienceEmergence(t *testing.T) {
	// Pseudocode became implementation:
	isSapient := evolution.CheckSapienceEmergence(0.95, 0.85)

	assert.True(t, isSapient, "High intelligence + social should trigger sapience")

	notSapient := evolution.CheckSapienceEmergence(0.7, 0.9)
	assert.False(t, notSapient, "Low intelligence should not trigger sapience")
}

// -----------------------------------------------------------------------------
// Scenario: Square-Cube Law (Biological Limits)
// -----------------------------------------------------------------------------
// Given: A land animal evolving larger size
// When: Mass increases by factor of 8 (2x height)
// Then: Bone strength/Leg cross-section must increase disproportionately
//
//	AND If strength < mass requirements, the species should go extinct or stop growing
func TestBDD_Physics_SquareCubeLaw(t *testing.T) {
	// A titan species too big for normal bones
	titan := &evolution.Species{
		SpeciesID:  uuid.New(),
		Name:       "Impossibly Large Titan",
		Type:       evolution.SpeciesHerbivore,
		Size:       20.0, // Way too big
		Population: 10,
	}

	// Calculate biomechanical fitness
	fitness := evolution.CalculateBiomechanicalFitness(titan)

	// Should have very low fitness due to square-cube law
	assert.Less(t, fitness, 0.5, "Giant animal should have reduced fitness")
	assert.Greater(t, fitness, 0.0, "Should still have positive fitness")

	// Normal-sized animal should have full fitness
	normalAnimal := &evolution.Species{Size: 5.0}
	normalFitness := evolution.CalculateBiomechanicalFitness(normalAnimal)
	assert.Equal(t, 1.0, normalFitness, "Normal-sized animal should have full fitness")
}

// -----------------------------------------------------------------------------
// Helper: Verify CalculateO2Effects returns appropriate values
// -----------------------------------------------------------------------------
func TestBDD_O2Effects_LowO2(t *testing.T) {
	// Current Earth O2 level (21%)
	maxSize, _ := evolution.CalculateO2Effects(0.21, 0.0)

	// At 21% O2, arthropods should be limited to reasonable sizes
	require.NotEqual(t, 0.0, maxSize, "CalculateO2Effects should return non-zero maxSize")
	assert.LessOrEqual(t, maxSize, 1.0,
		"O2 at 21%% should limit arthropods to smaller sizes, got %.2f", maxSize)
}
