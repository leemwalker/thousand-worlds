package ecosystem

import (
	"testing"
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
	geo.SimulateGeology(100_000)

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
