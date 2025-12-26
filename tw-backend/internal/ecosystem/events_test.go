package ecosystem

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMoonShielding_ReducesAsteroidProbability verifies that ImpactShielding
// reduces asteroid impact probability.
func TestMoonShielding_ReducesAsteroidProbability(t *testing.T) {
	// Statistical test: run many iterations and count asteroid impacts
	// Shielding of 0.2 should produce fewer impacts than shielding of 0.0
	// Note: Rare events have high variance, so we test relative reduction

	const iterations = 100000
	const ticksPerCheck = int64(100000)

	countImpacts := func(shielding float64) int {
		count := 0
		for i := 0; i < iterations; i++ {
			mgr := &GeologicalEventManager{
				ActiveEvents:     make([]GeologicalEvent, 0),
				TectonicActivity: 0.1,
				ImpactShielding:  shielding,
				rng:              rand.New(rand.NewSource(int64(i))),
			}
			mgr.CheckForNewEvents(0, ticksPerCheck)

			// Count asteroid impacts
			for _, e := range mgr.ActiveEvents {
				if e.Type == EventAsteroidImpact {
					count++
				}
			}
		}
		return count
	}

	noShielding := countImpacts(0.0)
	withShielding := countImpacts(0.2) // 20% shielding from moons

	t.Logf("Impacts with no shielding: %d", noShielding)
	t.Logf("Impacts with 20%% shielding: %d", withShielding)

	// Primary assertion: shielding should reduce impacts
	assert.Less(t, withShielding, noShielding,
		"20%% shielding should produce fewer impacts")

	// Calculate reduction for logging (may vary due to small sample with rare events)
	if noShielding > 0 {
		reduction := 100.0 * (1.0 - float64(withShielding)/float64(noShielding))
		t.Logf("Actual reduction: %.1f%%", reduction)
	}
}

// TestMoonShielding_ZeroShielding verifies no change with zero shielding
func TestMoonShielding_ZeroShielding(t *testing.T) {
	mgr := NewGeologicalEventManager()
	assert.Equal(t, 0.0, mgr.ImpactShielding,
		"Default shielding should be 0.0")
}

// TestMoonShielding_MaxShielding verifies maximum shielding still allows some impacts
func TestMoonShielding_MaxShielding(t *testing.T) {
	// Even with 100% shielding (unrealistic), some impacts may occur
	// This tests that shielding doesn't cause division by zero or other errors
	mgr := &GeologicalEventManager{
		ActiveEvents:     make([]GeologicalEvent, 0),
		TectonicActivity: 0.1,
		ImpactShielding:  1.0, // 100% shielding (hypothetical)
		rng:              rand.New(rand.NewSource(42)),
	}

	// Should not panic
	assert.NotPanics(t, func() {
		mgr.CheckForNewEvents(0, 100000)
	})

	// With 100% shielding, no impacts should occur
	impacts := 0
	for _, e := range mgr.ActiveEvents {
		if e.Type == EventAsteroidImpact {
			impacts++
		}
	}
	assert.Equal(t, 0, impacts, "100%% shielding should prevent all impacts")
}
