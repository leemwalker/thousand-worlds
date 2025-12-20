package geography_test

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed seed for deterministic test results across runs.
// All BDD tests should use this seed when generating world data.
const testSeed int64 = 42

// =============================================================================
// BDD Tests: Tectonics
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Plate Generation - Basic Properties
// -----------------------------------------------------------------------------
// Given: A request to generate N tectonic plates
// When: GeneratePlates is called
// Then: The correct number of plates should be created
//
//	AND Each plate should have a valid centroid, movement vector, and type
func TestBDD_Tectonics_PlateGeneration(t *testing.T) {
	scenarios := []struct {
		name       string
		plateCount int
		width      int
		height     int
	}{
		{"Small world - 3 plates", 3, 100, 100},
		{"Medium world - 5 plates", 5, 200, 200},
		{"Large world - 8 plates", 8, 500, 500},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			plates := geography.GeneratePlates(sc.plateCount, sc.width, sc.height, testSeed)

			require.Len(t, plates, sc.plateCount, "Should generate exact plate count")

			for i, plate := range plates {
				// Verify centroid is within world bounds
				assert.GreaterOrEqual(t, plate.Centroid.X, 0.0,
					"Plate %d centroid X should be >= 0", i)
				assert.Less(t, plate.Centroid.X, float64(sc.width),
					"Plate %d centroid X should be < width", i)
				assert.GreaterOrEqual(t, plate.Centroid.Y, 0.0,
					"Plate %d centroid Y should be >= 0", i)
				assert.Less(t, plate.Centroid.Y, float64(sc.height),
					"Plate %d centroid Y should be < height", i)

				// Verify movement vector is normalized (magnitude ~1)
				mag := plate.MovementVector.X*plate.MovementVector.X +
					plate.MovementVector.Y*plate.MovementVector.Y
				assert.InDelta(t, 1.0, mag, 0.01,
					"Plate %d movement vector should be normalized", i)

				// Verify plate type is set
				assert.True(t, plate.Type == geography.PlateContinental ||
					plate.Type == geography.PlateOceanic,
					"Plate %d should have valid type", i)
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
	// Generate many plates to verify distribution
	plates := geography.GeneratePlates(10, 200, 200, testSeed)

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
//	AND Heightmap dimensions should match input
//	AND Boundary zones should show elevation changes
func TestBDD_Tectonics_HeightmapGeneration(t *testing.T) {
	width, height := 100, 100
	plates := geography.GeneratePlates(5, width, height, testSeed)

	heightmap := geography.SimulateTectonics(plates, width, height)

	require.NotNil(t, heightmap, "Heightmap should be generated")
	assert.Equal(t, width, heightmap.Width, "Heightmap width should match")
	assert.Equal(t, height, heightmap.Height, "Heightmap height should match")

	// Check that some elevation variation exists (not all zeros)
	hasVariation := false
	for y := 0; y < height && !hasVariation; y++ {
		for x := 0; x < width; x++ {
			if heightmap.Get(x, y) != 0 {
				hasVariation = true
				break
			}
		}
	}
	assert.True(t, hasVariation, "Heightmap should have some elevation variation from plate boundaries")
}

// -----------------------------------------------------------------------------
// Scenario: Plate Boundary Types - Convergent Mountains
// -----------------------------------------------------------------------------
// Given: Two continental plates moving toward each other
// When: Tectonic simulation runs
// Then: Convergent boundary should create high elevation (mountains)
func TestBDD_Tectonics_ConvergentMountains(t *testing.T) {
	width, height := 100, 100

	// Create two continental plates moving toward each other
	plates := []geography.TectonicPlate{
		{
			Centroid:       geography.Point{X: 25, Y: 50},
			MovementVector: geography.Vector{X: 1, Y: 0}, // Moving right
			Type:           geography.PlateContinental,
			Thickness:      35,
		},
		{
			Centroid:       geography.Point{X: 75, Y: 50},
			MovementVector: geography.Vector{X: -1, Y: 0}, // Moving left
			Type:           geography.PlateContinental,
			Thickness:      35,
		},
	}

	heightmap := geography.SimulateTectonics(plates, width, height)

	// Find max elevation (should be at boundary around x=50)
	maxElev := 0.0
	maxX := 0
	for x := 40; x < 60; x++ {
		elev := heightmap.Get(x, 50)
		if elev > maxElev {
			maxElev = elev
			maxX = x
		}
	}

	assert.Greater(t, maxElev, 3000.0,
		"Convergent continental boundary should create mountains >3000m elevation")
	assert.Greater(t, maxX, 40, "Mountain peak should be near boundary")
	assert.Less(t, maxX, 60, "Mountain peak should be near boundary")
}

// -----------------------------------------------------------------------------
// Scenario: Plate Boundary Types - Oceanic Trench
// -----------------------------------------------------------------------------
// Given: Two oceanic plates converging
// When: Tectonic simulation runs
// Then: Should create deep trench (negative elevation)
func TestBDD_Tectonics_OceanicTrench(t *testing.T) {
	width, height := 100, 100

	plates := []geography.TectonicPlate{
		{
			Centroid:       geography.Point{X: 25, Y: 50},
			MovementVector: geography.Vector{X: 1, Y: 0},
			Type:           geography.PlateOceanic,
			Thickness:      7,
		},
		{
			Centroid:       geography.Point{X: 75, Y: 50},
			MovementVector: geography.Vector{X: -1, Y: 0},
			Type:           geography.PlateOceanic,
			Thickness:      7,
		},
	}

	heightmap := geography.SimulateTectonics(plates, width, height)

	// Find min elevation at boundary (trench)
	minElev := 0.0
	for x := 40; x < 60; x++ {
		elev := heightmap.Get(x, 50)
		if elev < minElev {
			minElev = elev
		}
	}

	assert.Less(t, minElev, -4000.0,
		"Convergent oceanic boundary should create trench <-4000m")
}

// -----------------------------------------------------------------------------
// Scenario: Divergent Boundary - Mid-Ocean Ridge
// -----------------------------------------------------------------------------
// Given: Two oceanic plates moving apart
// When: Tectonic simulation runs
// Then: Should create mid-ocean ridge (elevated seafloor)
func TestBDD_Tectonics_DivergentRidge(t *testing.T) {
	width, height := 100, 100

	plates := []geography.TectonicPlate{
		{
			Centroid:       geography.Point{X: 25, Y: 50},
			MovementVector: geography.Vector{X: -1, Y: 0}, // Moving away
			Type:           geography.PlateOceanic,
			Thickness:      7,
		},
		{
			Centroid:       geography.Point{X: 75, Y: 50},
			MovementVector: geography.Vector{X: 1, Y: 0}, // Moving away
			Type:           geography.PlateOceanic,
			Thickness:      7,
		},
	}

	heightmap := geography.SimulateTectonics(plates, width, height)

	// Find elevation at boundary (ridge)
	maxElev := 0.0
	for x := 40; x < 60; x++ {
		elev := heightmap.Get(x, 50)
		if elev > maxElev {
			maxElev = elev
		}
	}

	assert.Greater(t, maxElev, 200.0,
		"Divergent oceanic boundary should create mid-ocean ridge with positive elevation")
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
	assert.Fail(t, "BDD RED: Hadean era simulation not yet implemented - requires GeologicalAge parameter")
	// Pseudocode:
	// params := GenerationParams{GeologicalAge: "hadean", PlateCount: 0}
	// world := GenerateWorld(params)
	// assert world.Plates == nil || len(world.Plates) == 0
	// assert world.Heightmap.MaxElev < 100 // No mountains yet
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
	assert.Fail(t, "BDD RED: Archean craton formation not yet implemented")
	// Pseudocode:
	// plates := GeneratePlates(3, width, height, seed)
	// assert len(plates) >= 1
	// assert plates[0].Type == "proto-continental"
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
	assert.Fail(t, "BDD RED: Supercontinent assembly not yet implemented")
	// Pseudocode:
	// config := ContinentalConfiguration{PangaeaIndex: 0.9}
	// effects := config.calculateClimaticEffects()
	// assert effects.InteriorDesertPercent > 0.5
	// assert effects.SpeciationRate < 0.5 // Reduced due to land connectivity
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
	assert.Fail(t, "BDD RED: Continental rifting not yet implemented")
	// Pseudocode:
	// plates := []TectonicPlate{continental1, continental2}
	// boundaries := SimulateTectonics(plates, width, height)
	// divergent := findDivergentBoundaries(boundaries)
	// assert len(divergent) > 0
	// assert divergent[0].VolcanicActivity > 0.5
}

// -----------------------------------------------------------------------------
// Scenario: Himalaya Formation - Continental Collision
// -----------------------------------------------------------------------------
// Given: Two continental plates (India and Eurasia analogs)
// When: Plates collide at convergent boundary
// Then: Mountain range should form at collision zone
//
//	AND elevation should exceed 5000m at peaks
//	AND no subduction (both continental)
func TestBDD_Himalaya_ContinentalCollision(t *testing.T) {
	// This partially tests the existing implementation
	width, height := 200, 200

	// Simulate India-Eurasia style collision
	plates := []geography.TectonicPlate{
		{
			Centroid:       geography.Point{X: 50, Y: 100},
			MovementVector: geography.Vector{X: 1, Y: 0},
			Type:           geography.PlateContinental,
			Thickness:      40,
		},
		{
			Centroid:       geography.Point{X: 150, Y: 100},
			MovementVector: geography.Vector{X: 0, Y: 0}, // Stationary
			Type:           geography.PlateContinental,
			Thickness:      40,
		},
	}

	heightmap := geography.SimulateTectonics(plates, width, height)

	// Find max elevation at collision zone
	maxElev := 0.0
	for x := 80; x < 120; x++ {
		elev := heightmap.Get(x, 100)
		if elev > maxElev {
			maxElev = elev
		}
	}

	assert.Greater(t, maxElev, 5000.0,
		"Continental collision should create Himalaya-scale mountains (>5000m) - EXPECTED TO FAIL: current impl uses 6000m factor")
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
	assert.Fail(t, "BDD RED: Fragmentation → speciation effects not yet implemented")
	// Pseudocode:
	// config := ContinentalConfiguration{FragmentationIndex: 0.8}
	// effects := ApplyContinentalEffects(config, population)
	// assert effects.SpeciationRateMultiplier >= 1.8
	// assert effects.LargeAnimalSizeMultiplier < 0.9
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
	assert.Fail(t, "BDD RED: Continental crust layer generation not yet implemented in tectonics")
	// This belongs in underground module
	// Pseudocode:
	// col := WorldColumn{Composition: "continental"}
	// generateStrata(col, surfaceElevation)
	// assert col.Strata[0].Material == "sedimentary"
	// assert col.Strata[1].Material == "granite"
	// assert col.Strata[2].Material == "basalt"
	// assert col.TotalThickness() >= 35000 // 35km
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
	assert.Fail(t, "BDD RED: Oceanic crust layer generation not yet implemented in tectonics")
	// Pseudocode:
	// col := WorldColumn{Composition: "oceanic"}
	// generateStrata(col, surfaceElevation)
	// assert col.TotalThickness() >= 7000 // 7km
	// assert col.CavePotential() > 0.7
}

// -----------------------------------------------------------------------------
// Scenario: Seismic Profile by Boundary (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Different tectonic boundary interactions
// When: Stress releases (Earthquake)
// Then: The focal depth and max magnitude should match geological physics
func TestBDD_Tectonics_SeismicProfiles(t *testing.T) {
	assert.Fail(t, "BDD RED: Seismic physics not yet implemented")

	scenarios := []struct {
		boundaryType  string
		expectedDepth string // "Shallow", "Intermediate", "Deep"
		maxMagnitude  float64
	}{
		{"divergent_ridge", "Shallow", 6.5},            // Thin crust, hot rock
		{"transform_fault", "Shallow", 8.0},            // San Andreas style
		{"subduction_zone", "Deep", 9.5},               // Megathrust + Benioff zone
		{"continental_collision", "Intermediate", 8.5}, // Himalayas
	}
	_ = scenarios // For BDD stub - will be used when implemented
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
	assert.Fail(t, "BDD RED: Tsunami generation from seismic events not yet implemented")
	// Pseudocode:
	// boundary := Boundary{Type: BoundaryConvergent, Submerged: true}
	// event := boundary.TriggerEarthquake(8.5) // Megathrust event
	// tsunami := GenerateTsunami(event, waterDepth)
	//
	// assert tsunami.InitialWaveHeight > 2.0
	// assert tsunami.TravelVelocity > 500 // km/h in deep ocean
	// assert len(tsunami.AffectedCoasts) > 0
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
	assert.Fail(t, "BDD RED: Crustal buoyancy not yet implemented")
	// Pseudocode:
	// tile := Tile{Elevation: 100, IceLoad: 5000} // Compressed
	// tile.IceLoad = 0
	//
	// sim.Run(years: 5000)
	//
	// assert tile.Elevation > 100 // Rebounded
	// assert tile.IsostaticState == "uplifting"
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
	assert.Fail(t, "BDD RED: Accretionary wedges not yet implemented")
	// Pseudocode:
	// continent := Plate{Mass: Huge}
	// islandArc := Plate{Mass: Small, Composition: "volcanic"}
	//
	// SimulateCollision(continent, islandArc)
	//
	// assert continent.ContainsTerrane(islandArc.ID)
	// assert continent.WesternEdge.Geology != continent.Core.Geology
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
	assert.Fail(t, "BDD RED: Normal faulting not yet implemented")
	// Pseudocode:
	// region := Region{Stress: Tension}
	// SimulateTectonics(region)
	//
	// assert region.Topography == "alternating_ridge_valley"
	// assert region.CrustThickness < originalThickness
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
	assert.Fail(t, "BDD RED: Long-term tectonic loop not yet implemented")
	// Pseudocode:
	// history := RunLongSimulation(500_000_000)
	// assert history.HasEvent("Rifting")
	// assert history.HasEvent("OceanFloorSpreading")
	// assert history.HasEvent("Orogeny") // Mountain building
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
	assert.Fail(t, "BDD RED: Gravitational potential limits not yet implemented")
	// Pseudocode:
	// peak := Mountain{Height: 10000} // Unrealistic on Earth
	// sim.TickGeology()
	// assert peak.Height < 9000
	// assert peak.Faulting == "extensional" // Spreading out
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Plate Generation
// -----------------------------------------------------------------------------
// Given: Same seed value
// When: GeneratePlates is called twice
// Then: Results should be identical
func TestBDD_Tectonics_Determinism(t *testing.T) {
	plates1 := geography.GeneratePlates(5, 100, 100, testSeed)
	plates2 := geography.GeneratePlates(5, 100, 100, testSeed)

	require.Len(t, plates1, len(plates2))
	for i := range plates1 {
		assert.Equal(t, plates1[i].Centroid, plates2[i].Centroid,
			"Plate %d centroids should be identical", i)
		assert.Equal(t, plates1[i].MovementVector, plates2[i].MovementVector,
			"Plate %d movement vectors should be identical", i)
		assert.Equal(t, plates1[i].Type, plates2[i].Type,
			"Plate %d types should be identical", i)
	}
}
