package geography_test

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed seed for deterministic test results across runs.
// All BDD tests should use this seed when generating world data.
const testSeed int64 = 42

// =============================================================================
// BDD Tests: Tectonics (Spherical API)
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Plate Generation - Basic Properties
// -----------------------------------------------------------------------------
// Given: A request to generate N tectonic plates
// When: GeneratePlates is called
// Then: The correct number of plates should be created
//
//	AND Each plate should have a valid centroid, velocity, and type
func TestBDD_Tectonics_PlateGeneration(t *testing.T) {
	scenarios := []struct {
		name       string
		plateCount int
		resolution int
	}{
		{"Small world - 3 plates", 3, 16},
		{"Medium world - 5 plates", 5, 24},
		{"Large world - 8 plates", 8, 32},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			topology := spatial.NewCubeSphereTopology(sc.resolution)
			plates := geography.GeneratePlates(sc.plateCount, topology, testSeed)

			require.Len(t, plates, sc.plateCount, "Should generate exact plate count")

			for i, plate := range plates {
				// Verify centroid is valid (on a face)
				assert.GreaterOrEqual(t, plate.Centroid.Face, 0,
					"Plate %d centroid Face should be >= 0", i)
				assert.Less(t, plate.Centroid.Face, 6,
					"Plate %d centroid Face should be < 6", i)
				assert.GreaterOrEqual(t, plate.Centroid.X, 0,
					"Plate %d centroid X should be >= 0", i)
				assert.Less(t, plate.Centroid.X, sc.resolution,
					"Plate %d centroid X should be < resolution", i)

				// Verify velocity magnitude is reasonable (normalized ~1)
				vel := plate.Velocity
				mag := vel.X*vel.X + vel.Y*vel.Y + vel.Z*vel.Z
				assert.InDelta(t, 1.0, mag, 0.01,
					"Plate %d velocity vector should be normalized", i)

				// Verify plate type is set
				assert.True(t, plate.Type == geography.PlateContinental ||
					plate.Type == geography.PlateOceanic,
					"Plate %d should have valid type", i)

				// Verify region is populated
				assert.Greater(t, len(plate.Region), 0,
					"Plate %d should have cells in its region", i)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Continental/Oceanic Plate Distribution
// -----------------------------------------------------------------------------
// Given: Plate generation with default parameters
// When: GeneratePlates is called
// Then: Approximately 30% should be continental, 70% oceanic
func TestBDD_Tectonics_PlateTypeDistribution(t *testing.T) {
	resolution := 24
	topology := spatial.NewCubeSphereTopology(resolution)
	plates := geography.GeneratePlates(10, topology, testSeed)

	continentalCount := 0
	oceanicCount := 0
	for _, plate := range plates {
		if plate.Type == geography.PlateContinental {
			continentalCount++
		} else {
			oceanicCount++
		}
	}

	// With 10 plates: expect ~3 continental, ~7 oceanic (30/70 split)
	assert.GreaterOrEqual(t, continentalCount, 2,
		"Should have at least 2 continental plates")
	assert.LessOrEqual(t, continentalCount, 4,
		"Should have at most 4 continental plates")
	assert.GreaterOrEqual(t, oceanicCount, 6,
		"Should have at least 6 oceanic plates")
}

// -----------------------------------------------------------------------------
// Scenario: Tectonic Simulation - Heightmap Generation
// -----------------------------------------------------------------------------
// Given: A set of tectonic plates
// When: SimulateTectonics is called
// Then: A heightmap should be generated
//
//	AND Boundary zones should show elevation changes
func TestBDD_Tectonics_HeightmapGeneration(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)

	plates := geography.GeneratePlates(5, topology, testSeed)
	heightmap := geography.NewSphereHeightmap(topology)

	// Initialize to zero
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				heightmap.Set(spatial.Coordinate{Face: face, X: x, Y: y}, 0)
			}
		}
	}

	result := geography.SimulateTectonics(plates, heightmap, topology)

	require.NotNil(t, result, "Heightmap should be generated")

	// Check that some elevation variation exists (not all zeros)
	hasVariation := false
	for face := 0; face < 6 && !hasVariation; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				if result.Get(spatial.Coordinate{Face: face, X: x, Y: y}) != 0 {
					hasVariation = true
					break
				}
			}
		}
	}
	assert.True(t, hasVariation, "Heightmap should have some elevation variation from plate boundaries")
}

// -----------------------------------------------------------------------------
// Scenario: Hadean Eon - No Stable Crust
// -----------------------------------------------------------------------------
// Given: A newly formed planet (age < 500 million years)
// When: Tectonic simulation begins
// Then: Surface should be molten with no stable continental crust
//
//	AND plate count should be 0 or undefined
//	AND heightmap should show uniform low elevation
func TestBDD_Hadean_NoStableCrust(t *testing.T) {
	// Pseudocode became implementation:
	plateCount, surface := geography.SimulateGeologicalAge(geography.AgeHadean)

	assert.Equal(t, 0, plateCount, "Hadean should have 0 stable plates")
	assert.Equal(t, "molten", surface, "Hadean surface should be molten")
}

// -----------------------------------------------------------------------------
// Scenario: Archean Eon - First Cratons Form
// -----------------------------------------------------------------------------
// Given: A planet aged 500M-2.5B years
// When: Tectonic simulation runs
// Then: Small stable continental nuclei (cratons) should form
//
//	AND ocean coverage should be > 90%
//	AND plate count should be 1-3 proto-plates
func TestBDD_Archean_FirstCratons(t *testing.T) {
	// Pseudocode became implementation:
	plateCount, surface := geography.SimulateGeologicalAge(geography.AgeArchean)

	assert.GreaterOrEqual(t, plateCount, 1, "Archean should have at least 1 craton")
	assert.Equal(t, "cratons", surface, "Archean surface should feature cratons")
}

// -----------------------------------------------------------------------------
// Scenario: Pangaea Assembly - Supercontinent Formation
// -----------------------------------------------------------------------------
// Given: Multiple continental plates
// When: Plates converge over 200M+ years
// Then: A single supercontinent should form
//
//	AND continental fragmentation should approach 0.0
//	AND interior deserts should expand (> 50% of land)
func TestBDD_Pangaea_SupercontinentFormation(t *testing.T) {
	// Pseudocode became implementation:
	desertPct, speciationRate := geography.CalculateSupercontinentEffects(0.9) // High Pangaea Index

	assert.Greater(t, desertPct, 0.5, "Supercontinent should have extensive deserts")
	assert.Less(t, speciationRate, 0.5, "Supercontinent should have reduced speciation rate")
}

// -----------------------------------------------------------------------------
// Scenario: Atlantic Opening - Continental Rift
// -----------------------------------------------------------------------------
// Given: A supercontinent configuration
// When: Divergent boundary forms between two plate sections
// Then: New ocean basin should form between separating continents
//
//	AND volcanic activity should increase at rift zone
//	AND sea level should rise (new mid-ocean ridge displaces water)
func TestBDD_Atlantic_ContinentalRift(t *testing.T) {
	// Pseudocode became implementation:
	hasRift, volcanicActivity := geography.SimulateContinentalRift(true) // Divergent

	assert.True(t, hasRift, "Divergent boundary should form rift")
	assert.Greater(t, volcanicActivity, 0.5, "Rifting should have high volcanic activity")
}

// -----------------------------------------------------------------------------
// Scenario: Continental Fragmentation - Speciation Rate
// -----------------------------------------------------------------------------
// Given: High continental fragmentation (> 0.7)
// When: Species evolution is simulated
// Then: Speciation rate should be ~2x baseline
//
//	AND genetic drift rate should increase
//	AND large animals should face size penalty (island dwarfism)
func TestBDD_Fragmentation_SpeciationRate(t *testing.T) {
	// Pseudocode became implementation:
	specMult, sizeMult := geography.CalculateFragmentationEffects(0.8) // High fragmentation

	assert.GreaterOrEqual(t, specMult, 1.8, "High fragmentation should boost speciation")
	assert.Less(t, sizeMult, 0.9, "High fragmentation should penalize large animal size")
}

// -----------------------------------------------------------------------------
// Scenario: Earth-Realistic Crust Layers
// -----------------------------------------------------------------------------
// Given: A continental region
// When: Underground column is initialized
// Then: Crust should be ~35km thick (continental)
//
//	AND Moho discontinuity should be present
//	AND Layer sequence: sedimentary → granite → basalt → mantle
func TestBDD_ContinentalCrust_Layers(t *testing.T) {
	// Pseudocode became implementation:
	crust := geography.GetCrustLayers(geography.PlateContinental)

	assert.Equal(t, "sedimentary", crust.Layers[0])
	assert.Equal(t, "granite", crust.Layers[1])
	assert.GreaterOrEqual(t, crust.Thickness, 35000.0, "Continental crust should be > 35km thick")
}

// -----------------------------------------------------------------------------
// Scenario: Oceanic Crust - Thin Layer
// -----------------------------------------------------------------------------
// Given: An oceanic region
// When: Underground column is initialized
// Then: Crust should be ~7km thick (oceanic)
//
//	AND Layer sequence: sediment → basalt → gabbro → mantle
//	AND High cave potential in limestone zones
func TestBDD_OceanicCrust_Layers(t *testing.T) {
	// Pseudocode became implementation:
	crust := geography.GetCrustLayers(geography.PlateOceanic)

	assert.GreaterOrEqual(t, crust.Thickness, 7000.0, "Oceanic crust should be >= 7km thick")
	assert.Equal(t, "basalt", crust.Layers[1], "Oceanic crust should have basalt")
}

// -----------------------------------------------------------------------------
// Scenario: Seismic Profile by Boundary (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Different tectonic boundary interactions
// When: Stress releases (Earthquake)
// Then: The focal depth and max magnitude should match geological physics
func TestBDD_Tectonics_SeismicProfiles(t *testing.T) {
	scenarios := []struct {
		boundaryType  string
		expectedDepth string
		minMagnitude  float64
	}{
		{"divergent_ridge", "Shallow", 6.0},
		{"transform_fault", "Shallow", 7.5},
		{"subduction_zone", "Deep", 9.0},
		{"continental_collision", "Intermediate", 8.0},
	}

	for _, sc := range scenarios {
		event := geography.CalculateSeismicActivity(geography.BoundaryType(sc.boundaryType))

		assert.Equal(t, sc.expectedDepth, event.Depth, "Depth mismatch for %s", sc.boundaryType)
		assert.GreaterOrEqual(t, event.Magnitude, sc.minMagnitude, "Magnitude mismatch for %s", sc.boundaryType)
	}
}

// -----------------------------------------------------------------------------
// Scenario: Tsunami Generation - Underwater Megathrust
// -----------------------------------------------------------------------------
// Given: A convergent boundary (subduction zone) in an oceanic region
// When: A high-magnitude earthquake (M > 7.5) occurs
// Then: A tsunami wave should be generated
//
//	AND wave height should be proportional to vertical displacement
//	AND coastal regions within range should be flagged for impact
func TestBDD_Tsunami_Generation(t *testing.T) {
	event := geography.SeismicEvent{
		Magnitude:    8.5,
		BoundaryType: geography.BoundaryConvergent,
	}
	waterDepth := 4000.0 // Deep ocean

	tsunami := geography.GenerateTsunami(event, waterDepth)

	require.NotNil(t, tsunami, "Tsunami should be generated")
	assert.Greater(t, tsunami.InitialWaveHeight, 2.0, "Megathrust should create large wave")
	assert.Greater(t, tsunami.TravelVelocity, 500.0, "Tsunami should travel fast in deep water")
}

// -----------------------------------------------------------------------------
// Scenario: Isostatic Adjustment (Glacial Rebound)
// -----------------------------------------------------------------------------
// Given: A continent covered by a massive ice sheet (load)
// When: The Ice Age ends and ice melts
// Then: The crust elevation should slowly rise over ticks (Rebound)
//
//	AND Relative sea level should drop in that region
func TestBDD_Tectonics_Isostasy(t *testing.T) {
	// Pseudocode became implementation:
	currentElevation := 100.0
	iceLoad := 0.0 // Melting complete

	newElev, state := geography.SimulateIsostasy(currentElevation, iceLoad)

	assert.Greater(t, newElev, currentElevation, "Rebound should increase elevation")
	assert.Equal(t, "uplifting", state)
}

// -----------------------------------------------------------------------------
// Scenario: Exotic Terrane Accretion
// -----------------------------------------------------------------------------
// Given: A continental plate moving West and an island arc moving East
// When: They collide (subduction consumes the ocean between)
// Then: The island arc should not subduct but "suture" onto the continent
//
//	AND The continent's edge should gain a new geological province (Terrane)
func TestBDD_Tectonics_TerraneAccretion(t *testing.T) {
	// Pseudocode became implementation:
	accretionParams := struct{ ContinentMass, ArcMass float64 }{1000.0, 10.0}

	occured := geography.SimulateTerraneAccretion(accretionParams.ContinentMass, accretionParams.ArcMass)

	assert.True(t, occured, "Small arc should accrete onto large continent")
}

// -----------------------------------------------------------------------------
// Scenario: Crustal Extension (Horst and Graben)
// -----------------------------------------------------------------------------
// Given: A continental region undergoing tensile stress (pulling apart)
// When: The crust thins and faults
// Then: Parallel mountain ranges and valleys should form
//
//	AND Valleys should drop in elevation (Subsidence)
func TestBDD_Tectonics_Extension(t *testing.T) {
	// Pseudocode became implementation:
	stress := 0.8 // High tension

	topo, thicknessMult := geography.SimulateExtension(stress)

	assert.Equal(t, "alternating_ridge_valley", topo)
	assert.Less(t, thicknessMult, 1.0, "Extension should thin the crust")
}

// -----------------------------------------------------------------------------
// Scenario: The Wilson Cycle Integration
// -----------------------------------------------------------------------------
// Given: A stable supercontinent
// When: Simulation runs for 500M years
// Then: It should Rift (Open Ocean) -> Spread -> Subduct -> Collide
//
//	AND The final state should be a new mountain belt (suture zone)
func TestBDD_Tectonics_WilsonCycle(t *testing.T) {
	// Pseudocode became implementation:
	phase := geography.SimulateWilsonCycle(50_000_000) // Early check
	assert.Equal(t, "Rifting", phase)

	phaseMidd := geography.SimulateWilsonCycle(150_000_000)
	assert.Equal(t, "OceanFloorSpreading", phaseMidd)
}

// -----------------------------------------------------------------------------
// Scenario: Orogenic Collapse (Mountain Limits)
// -----------------------------------------------------------------------------
// Given: A mountain range exceeding maximum crustal support height (> 8800m)
// When: Gravity acts over geological time
// Then: The range should spread laterally and lower in height
//
//	AND "Normal faults" should appear at the high peaks (gravitational collapse)
func TestBDD_Tectonics_MountainLimits(t *testing.T) {
	// Pseudocode became implementation:
	elev := 10000.0 // Unrealistic

	newElev, state := geography.SimulateMountainCollapse(elev)

	assert.Less(t, newElev, elev, "Mountains should collapse under gravity")
	assert.Equal(t, "extensional", state)
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Plate Generation
// -----------------------------------------------------------------------------
// Given: Same seed value
// When: GeneratePlates is called twice
// Then: Results should be identical
func TestBDD_Tectonics_Determinism(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)

	plates1 := geography.GeneratePlates(5, topology, testSeed)
	plates2 := geography.GeneratePlates(5, topology, testSeed)

	require.Len(t, plates1, len(plates2))
	for i := range plates1 {
		assert.Equal(t, plates1[i].Centroid, plates2[i].Centroid,
			"Plate %d centroids should be identical", i)
		assert.Equal(t, plates1[i].Velocity, plates2[i].Velocity,
			"Plate %d velocity vectors should be identical", i)
		assert.Equal(t, plates1[i].Type, plates2[i].Type,
			"Plate %d types should be identical", i)
	}
}

// -----------------------------------------------------------------------------
// Scenario: Full Cell Coverage (Spherical)
// -----------------------------------------------------------------------------
// Given: Generated plates on a cube-sphere
// When: All cells are checked
// Then: Every cell should belong to exactly one plate
func TestBDD_Tectonics_FullCellCoverage(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)

	plates := geography.GeneratePlates(5, topology, testSeed)

	// Build map of all assigned cells
	assigned := make(map[spatial.Coordinate]int) // cell -> plate count
	for _, plate := range plates {
		for coord := range plate.Region {
			assigned[coord]++
		}
	}

	// Verify all cells are assigned exactly once
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				count, exists := assigned[coord]
				assert.True(t, exists, "Cell %v should be assigned to a plate", coord)
				assert.Equal(t, 1, count, "Cell %v should be assigned to exactly one plate", coord)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Scenario: Cross-Face Plate Regions
// -----------------------------------------------------------------------------
// Given: Few plates on a cube-sphere
// When: Plates are generated
// Then: At least one plate should span multiple faces
func TestBDD_Tectonics_CrossFacePlates(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)

	// Use few plates so they must span faces
	plates := geography.GeneratePlates(3, topology, testSeed)

	crossFaceFound := false
	for _, plate := range plates {
		facesPresent := make(map[int]bool)
		for coord := range plate.Region {
			facesPresent[coord.Face] = true
		}
		if len(facesPresent) > 1 {
			crossFaceFound = true
			break
		}
	}

	assert.True(t, crossFaceFound, "At least one plate should span multiple cube-sphere faces")
}
