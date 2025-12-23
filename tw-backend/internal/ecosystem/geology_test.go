package ecosystem

import (
	"testing"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWorldGeology_Lifecycle(t *testing.T) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0 // 40,000 km

	// 1. Creation
	geo := NewWorldGeology(worldID, seed, circumference)
	assert.NotNil(t, geo)
	assert.False(t, geo.IsInitialized())

	// 2. Initialization
	geo.InitializeGeology()
	assert.True(t, geo.IsInitialized())
	assert.NotNil(t, geo.Heightmap)
	assert.NotEmpty(t, geo.Plates)
	assert.NotEmpty(t, geo.Hotspots)
	// Rivers might be empty if random generation yields none, but unlikely with seed
	// Biomes should be populated
	assert.NotEmpty(t, geo.Biomes)

	initialStats := geo.GetStats()
	assert.Greater(t, initialStats.PlateCount, 0)
	assert.Greater(t, initialStats.HotspotCount, 0)

	// 3. Simulation
	// Simulate 100,000 years (enough for tectonic shift and hotspot activity)
	geo.SimulateGeology(100000, 0.0)

	simStats := geo.GetStats()
	assert.Equal(t, int64(100_000), simStats.YearsSimulated)
	// Check that rivers/biomes were regenerated (not nil)
	assert.NotNil(t, geo.Rivers)
	assert.NotNil(t, geo.Biomes)

	// 4. Events
	event := GeologicalEvent{
		Type:      EventAsteroidImpact,
		Severity:  1.0,
		StartTick: 0,
	}
	geo.ApplyEvent(event)
	// Just check it doesn't crash and stats might change
	eventStats := geo.GetStats()
	assert.NotEqual(t, simStats.AverageElevation, eventStats.AverageElevation, "Impact should change elevation stats")
}

func TestApplyHotspotActivity(t *testing.T) {
	// Setup localized test
	worldID := uuid.New()
	geo := NewWorldGeology(worldID, 123, 10_000_000)
	geo.InitializeGeology()

	// Force a hotspot at center
	cx, cy := float64(geo.Heightmap.Width/2), float64(geo.Heightmap.Height/2)
	geo.Hotspots = []geography.Point{{X: cx, Y: cy}}

	// Get elevation before
	elevBefore := geo.Heightmap.Get(int(cx), int(cy))

	// Simulate million years of activity
	geo.applyHotspotActivity(1_000_000)

	// Get elevation after
	elevAfter := geo.Heightmap.Get(int(cx), int(cy))

	assert.Greater(t, elevAfter, elevBefore, "Hotspot activity should build mountains")
}

func TestPlateMovement_Realism(t *testing.T) {
	// Setup world for testing plate movement
	worldID := uuid.New()
	geo := NewWorldGeology(worldID, 42, 40_000_000) // Earth-size
	geo.InitializeGeology()

	// Record initial plate positions (centroids)
	initialCentroids := make([]struct{ Face, X, Y int }, len(geo.Plates))
	for i, plate := range geo.Plates {
		initialCentroids[i].Face = plate.Centroid.Face
		initialCentroids[i].X = plate.Centroid.X
		initialCentroids[i].Y = plate.Centroid.Y
	}

	// Record initial heightmap state
	initialElevSum := 0.0
	for i := range geo.Heightmap.Elevations {
		initialElevSum += geo.Heightmap.Elevations[i]
	}

	// Simulate 1 million years of plate movement
	geo.advancePlates(1_000_000)

	// Verify plates have moved
	movedCount := 0
	for i, plate := range geo.Plates {
		if plate.Centroid.Face != initialCentroids[i].Face ||
			plate.Centroid.X != initialCentroids[i].X ||
			plate.Centroid.Y != initialCentroids[i].Y {
			movedCount++
		}
	}

	// At least some plates should have moved over 1M years
	// Movement rate is 2cm/year = 20km over 1M years
	// On a 40,000km circumference, this is 20/40000 = 0.0005 of circumference
	// With resolution typically 100-500, plates should move at least 1 cell
	assert.Greater(t, movedCount, 0, "At least one plate should have moved over 1M years")

	// Verify plate age was updated
	for _, plate := range geo.Plates {
		assert.GreaterOrEqual(t, plate.Age, 1.0, "Plate age should increase by 1M years")
	}

	// Now apply continental drift event and check heightmap changes
	if geo.SphereHeightmap != nil {
		// Record pre-drift sphere heightmap state
		preDriftSum := 0.0
		resolution := geo.Topology.Resolution()
		for face := 0; face < 6; face++ {
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					preDriftSum += geo.SphereHeightmap.Get(coord)
				}
			}
		}

		// Apply continental drift event
		geo.applyContinentalDrift(0.5)

		// Verify SphereHeightmap has changed (tectonic activity should modify elevations)
		postDriftSum := 0.0
		for face := 0; face < 6; face++ {
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					postDriftSum += geo.SphereHeightmap.Get(coord)
				}
			}
		}

		assert.NotEqual(t, preDriftSum, postDriftSum, "SphereHeightmap should change after continental drift")
	}
}
