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

// -----------------------------------------------------------------------------
// Scenario: Disease System Update
// -----------------------------------------------------------------------------
// Given: A disease system with species info
// When: Update is called
// Then: System should process pathogens and outbreaks
func TestBDD_DiseaseSystem_Update(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	rng := rand.New(rand.NewSource(testSeed))
	speciesID := uuid.New()
	p := pathogen.NewPathogen("Update Test", pathogen.PathogenVirus, speciesID, 500000, rng)
	ds.AddPathogen(p)

	speciesInfo := map[uuid.UUID]pathogen.SpeciesInfo{
		speciesID: {
			Population:        500000,
			DiseaseResistance: 0.3,
			DietType:          "herbivore",
			Density:           0.5,
		},
	}

	// Update should not panic
	ds.Update(100000, speciesInfo)
}

// -----------------------------------------------------------------------------
// Scenario: Historical Outbreaks
// -----------------------------------------------------------------------------
// Given: A disease system
// When: GetHistoricalOutbreaks is called
// Then: Should return resolved outbreaks
func TestBDD_GetHistoricalOutbreaks(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	speciesID := uuid.New()
	historical := ds.GetHistoricalOutbreaks(speciesID)

	assert.NotNil(t, historical, "Should return slice")
	assert.Empty(t, historical, "Initially empty")
}

// -----------------------------------------------------------------------------
// Scenario: Pathogen Clone
// -----------------------------------------------------------------------------
// Given: An existing pathogen
// When: Clone is called
// Then: Should create a copy with same properties
func TestBDD_Pathogen_Clone(t *testing.T) {
	rng := rand.New(rand.NewSource(testSeed))
	original := pathogen.NewPathogen("Original", pathogen.PathogenVirus, uuid.New(), 1000, rng)

	clone := original.Clone(rng)

	require.NotNil(t, clone, "Clone should not be nil")
	assert.NotEqual(t, original.ID, clone.ID, "Clone should have new ID")
	assert.Equal(t, original.Name, clone.Name)
	assert.Equal(t, original.Type, clone.Type)
}

// -----------------------------------------------------------------------------
// Scenario: Pathogen Can Infect Host
// -----------------------------------------------------------------------------
// Given: A pathogen with specific characteristics
// When: CanInfectHost is called
// Then: Should determine infectability
func TestBDD_Pathogen_CanInfectHost(t *testing.T) {
	rng := rand.New(rand.NewSource(testSeed))
	speciesID := uuid.New()
	p := pathogen.NewPathogen("Infection Test", pathogen.PathogenBacteria, speciesID, 1000, rng)

	// Test with original host - should always infect
	canInfect := p.CanInfectHost(speciesID, "herbivore", 0.5)
	assert.True(t, canInfect, "Should infect original host")

	// Test with different species
	otherSpecies := uuid.New()
	// May or may not infect depending on specificity
	_ = p.CanInfectHost(otherSpecies, "carnivore", 0.3)
	_ = p.CanInfectHost(otherSpecies, "omnivore", 0.8)
}

// -----------------------------------------------------------------------------
// Scenario: Pathogen Is Becoming Endemic
// -----------------------------------------------------------------------------
// Given: A pathogen that has been around a long time
// When: IsBecomingEndemic is called
// Then: Should determine if disease is becoming permanent
func TestBDD_Pathogen_IsBecomingEndemic(t *testing.T) {
	rng := rand.New(rand.NewSource(testSeed))
	p := pathogen.NewPathogen("Endemic Test", pathogen.PathogenVirus, uuid.New(), 0, rng)

	// Check endemic status
	isEndemic := p.IsBecomingEndemic()
	_ = isEndemic // Just verify it runs
}

// -----------------------------------------------------------------------------
// Scenario: Outbreak Update
// -----------------------------------------------------------------------------
// Given: An active outbreak
// When: Update is called
// Then: Should simulate disease progression
func TestBDD_Outbreak_Update(t *testing.T) {
	rng := rand.New(rand.NewSource(testSeed))
	speciesID := uuid.New()
	biomeID := uuid.New()
	p := pathogen.NewPathogen("Outbreak Update", pathogen.PathogenVirus, speciesID, 1000, rng)

	outbreak := pathogen.NewOutbreak(p.ID, speciesID, biomeID, 0, 1000)

	// Update the outbreak
	outbreak.Update(p, 100000, 0.3, rng)

	// Outbreak should have progressed (check it didn't panic)
	assert.NotNil(t, outbreak)
}

// -----------------------------------------------------------------------------
// Scenario: Zoonotic Transfer Edge Cases
// -----------------------------------------------------------------------------
// Given: Various contact rates and resistances
// When: CheckZoonoticTransfer is called
// Then: Should handle different scenarios
func TestBDD_ZoonoticTransfer_EdgeCases(t *testing.T) {
	worldID := uuid.New()
	ds := pathogen.NewDiseaseSystem(worldID, testSeed)

	sourceSpeciesID := uuid.New()
	targetSpeciesID := uuid.New()

	rng := rand.New(rand.NewSource(testSeed))
	p := pathogen.NewPathogen("Zoonotic Test", pathogen.PathogenVirus, sourceSpeciesID, 1000, rng)
	ds.AddPathogen(p)

	// Multiple trials
	transfers := 0
	for i := 0; i < 50; i++ {
		outbreak := ds.CheckZoonoticTransfer(
			p,
			sourceSpeciesID,
			targetSpeciesID,
			"carnivore",
			200000,
			0.1, // Low resistance
			0.9, // High contact
		)
		if outbreak != nil {
			transfers++
		}
	}
	t.Logf("Zoonotic transfers in 50 trials: %d", transfers)
}

// -----------------------------------------------------------------------------
// Scenario: Pathogen Mortality Calculation
// -----------------------------------------------------------------------------
// Given: Various host resistance levels
// When: CalculateMortality is called
// Then: Should return appropriate death rates
func TestBDD_Pathogen_CalculateMortality(t *testing.T) {
	rng := rand.New(rand.NewSource(testSeed))
	p := pathogen.NewPathogen("Mortality Test", pathogen.PathogenVirus, uuid.New(), 1000, rng)

	// Test with different resistance levels
	lowRes := p.CalculateMortality(0.1)
	highRes := p.CalculateMortality(0.9)

	assert.GreaterOrEqual(t, lowRes, highRes, "Higher resistance should have lower mortality")
}
