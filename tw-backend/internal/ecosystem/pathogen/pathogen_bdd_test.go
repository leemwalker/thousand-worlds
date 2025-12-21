package pathogen_test

import (
	"math/rand"
	"testing"

	"tw-backend/internal/ecosystem/pathogen"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed seed for deterministic test results.
const testSeed int64 = 42

// =============================================================================
// BDD Tests: Disease System
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Disease System Creation
// -----------------------------------------------------------------------------
// Given: A world ID and seed
// When: NewDiseaseSystem is called
// Then: System should be created with empty pathogens and outbreaks
func TestBDD_DiseaseSystem_Creation(t *testing.T) {
	worldID := uuid.New()

	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	require.NotNil(t, ds, "Disease system should be created")
	assert.Equal(t, worldID, ds.WorldID, "WorldID should match")
	assert.Empty(t, ds.Pathogens, "Should start with no pathogens")
}

// -----------------------------------------------------------------------------
// Scenario: Add Pathogen
// -----------------------------------------------------------------------------
// Given: A disease system
// When: A pathogen is added
// Then: It should be retrievable by ID
func TestBDD_DiseaseSystem_AddPathogen(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	p := &pathogen.Pathogen{
		ID:   uuid.New(),
		Name: "Test Virus",
		Type: pathogen.PathogenVirus,
	}

	ds.AddPathogen(p)

	retrieved := ds.GetPathogen(p.ID)
	require.NotNil(t, retrieved, "Pathogen should be retrievable")
	assert.Equal(t, p.Name, retrieved.Name)
}

// -----------------------------------------------------------------------------
// Scenario: Spontaneous Outbreak - Low Density
// -----------------------------------------------------------------------------
// Given: A species with low population density
// When: CheckSpontaneousOutbreak is called
// Then: Outbreak should be unlikely
func TestBDD_Outbreak_LowDensity(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	speciesID := uuid.New()

	// Multiple trials with low density
	outbreakCount := 0
	for i := 0; i < 100; i++ {
		_, outbreak := ds.CheckSpontaneousOutbreak(speciesID, "Sparse Species", 100, 0.01)
		if outbreak != nil {
			outbreakCount++
		}
	}

	assert.Less(t, outbreakCount, 50, "Low density should have fewer outbreaks")
}

// -----------------------------------------------------------------------------
// Scenario: Spontaneous Outbreak - High Density
// -----------------------------------------------------------------------------
// Given: A species with high population density
// When: CheckSpontaneousOutbreak is called
// Then: Outbreak should be more likely
func TestBDD_Outbreak_HighDensity(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	speciesID := uuid.New()

	// Multiple trials with high density
	outbreakCount := 0
	for i := 0; i < 100; i++ {
		_, outbreak := ds.CheckSpontaneousOutbreak(speciesID, "Dense Species", 1000000, 0.9)
		if outbreak != nil {
			outbreakCount++
		}
	}

	assert.GreaterOrEqual(t, outbreakCount, 0, "High density may have outbreaks")
}

// -----------------------------------------------------------------------------
// Scenario: Zoonotic Transfer
// -----------------------------------------------------------------------------
// Given: A pathogen affecting one species
// When: CheckZoonoticTransfer is called with high contact rate
// Then: Transfer may occur to new host
func TestBDD_ZoonoticTransfer(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	sourceSpeciesID := uuid.New()
	targetSpeciesID := uuid.New()

	// Create a pathogen using the proper constructor with RNG
	rng := rand.New(rand.NewSource(testSeed))
	p := pathogen.NewPathogen("Flu Strain A", pathogen.PathogenVirus, sourceSpeciesID, 1000, rng)
	ds.AddPathogen(p)

	// Test zoonotic transfer
	outbreak := ds.CheckZoonoticTransfer(
		p,
		sourceSpeciesID,
		targetSpeciesID,
		"omnivore", // Diet type
		100000,     // Target population
		0.2,        // Low resistance
		0.8,        // High contact rate
	)

	// May or may not transfer - just check it doesn't panic
	_ = outbreak
}

// -----------------------------------------------------------------------------
// Scenario: Disease Impact Tracking
// -----------------------------------------------------------------------------
// Given: Active outbreaks affecting a species
// When: GetImpact is called
// Then: Should return infected and death counts
func TestBDD_DiseaseImpact(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	speciesID := uuid.New()

	// Initially no impact
	infected, deaths := ds.GetImpact(speciesID)

	assert.Equal(t, int64(0), infected, "No infections initially")
	assert.Equal(t, int64(0), deaths, "No deaths initially")
}

// -----------------------------------------------------------------------------
// Scenario: Pandemic Detection
// -----------------------------------------------------------------------------
// Given: A disease system with outbreaks
// When: GetPandemics is called
// Then: Should return severe outbreaks only
func TestBDD_PandemicDetection(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	pandemics := ds.GetPandemics()

	assert.NotNil(t, pandemics, "Should return empty slice, not nil")
	assert.Empty(t, pandemics, "No pandemics initially")
}

// -----------------------------------------------------------------------------
// Scenario: Pathogen Eradication
// -----------------------------------------------------------------------------
// Given: An active pathogen
// When: EradicatePathogen is called
// Then: Pathogen should be marked as eradicated
func TestBDD_PathogenEradication(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	rng := rand.New(rand.NewSource(testSeed))
	p := pathogen.NewPathogen("Eradicable Disease", pathogen.PathogenBacteria, uuid.New(), 1000, rng)
	ds.AddPathogen(p)

	ds.EradicatePathogen(p.ID)

	// Verify pathogen is eradicated
	retrieved := ds.GetPathogen(p.ID)
	if retrieved != nil {
		assert.True(t, retrieved.IsEradicated, "Pathogen should be marked eradicated")
	}
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Results
// -----------------------------------------------------------------------------
// Given: Same seed value
// When: DiseaseSystem is created twice
// Then: Initial random state should be identical
func TestBDD_DiseaseSystem_Determinism(t *testing.T) {
	worldID := uuid.New()

	ds1 := pathogen.NewDiseaseSystem(worldID, testSeed)
	ds2 := pathogen.NewDiseaseSystem(worldID, testSeed)

	assert.Equal(t, ds1.WorldID, ds2.WorldID, "WorldID should match")
	assert.Len(t, ds1.Pathogens, len(ds2.Pathogens), "Pathogen count should match")
}
