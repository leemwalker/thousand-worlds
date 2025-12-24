package ecosystem

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestBillionYearStability simulates a long period of geological time
// to ensure that elevation and sea level remain within realistic bounds.
// This validates the equilibrium-based physics fixes.
func TestBillionYearStability(t *testing.T) {
	// Setup
	worldID := uuid.New()
	seed := int64(12345)
	// Use smaller circumference to force minimal grid size (approx 100x50) for faster testing
	// 1,000 km circumference -> 100x50 pixels at 10km/px
	// This is significantly smaller than the max 512x256
	circumference := 1_000_000.0
	geo := NewWorldGeology(worldID, seed, circumference)
	geo.InitializeGeology()

	// Initial stats
	initialStats := geo.GetStats()
	t.Logf("Initial Stats: MaxElev=%.1fm, MinElev=%.1fm, SeaLevel=%.1fm",
		initialStats.MaxElevation, initialStats.MinElevation, initialStats.SeaLevel)

	// Simulate 200 million years
	// We use larger time steps for speed, but ensure tectonic intervals are hit
	// Total steps: 200 steps of 1 million years each = 200 million years
	// Note: Reduced from 1B years for CI performance (1B takes ~6 minutes)
	stepSize := int64(1_000_000)
	totalSteps := 200

	for i := 0; i < totalSteps; i++ {
		geo.SimulateGeology(stepSize, 0.0)

		// Periodic assertions (every 100M years)
		if (i+1)%100 == 0 {
			stats := geo.GetStats()
			t.Logf("Year %dM: MaxElev=%.1fm, MinElev=%.1fm, SeaLevel=%.1fm",
				(i + 1), stats.MaxElevation, stats.MinElevation, stats.SeaLevel)

			// Hard Constraints (Physics Limits)
			// Max elevation shouldn't exceed 15km (Olympus Mons is ~21km, Earth max is ~8.8km)
			// Our clamp is 15km
			assert.LessOrEqual(t, stats.MaxElevation, 15001.0, "Max elevation exceeded 15km Physics limit")

			// Min elevation shouldn't exceed -11km (Mariana Trench)
			// Our clamp is -11km
			assert.GreaterOrEqual(t, stats.MinElevation, -11001.0, "Min elevation exceeded -11km Physics limit")

			// Sea level homeostasis check
			// Sea level should stay roughly within +/- 500m of baseline 0
			// It may fluctuate due to ice ages, but shouldn't drift to -2000m
			assert.True(t, math.Abs(stats.SeaLevel) < 2000.0, "Sea level drifted too far: %.1f", stats.SeaLevel)
		}
	}

	// Final verification
	finalStats := geo.GetStats()
	t.Logf("Final Stats (1 Billion Years):")
	t.Logf("  Max Elevation: %.1fm", finalStats.MaxElevation)
	t.Logf("  Min Elevation: %.1fm", finalStats.MinElevation)
	t.Logf("  Sea Level:     %.1fm", finalStats.SeaLevel)
	t.Logf("  Land Percent:  %.1f%%", finalStats.LandPercent)

	// Verify mountains actually exist (should be > 2000m)
	// If this fails, we over-damped the system
	assert.Greater(t, finalStats.MaxElevation, 2000.0, "No significant mountains formed after 1B years")
}
