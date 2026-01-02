package ecosystem

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRainfallUpdating(t *testing.T) {
	// Initialize geology
	geo := NewWorldGeology(uuid.New(), 12345, 40000000) // Earth circumference
	geo.InitializeGeology(0)

	// Capture initial rainfall (copy it)
	initialRainfall := make([]float64, len(geo.Rainfall))
	copy(initialRainfall, geo.Rainfall)

	// Simulate for 50 Million years (enough for significant drift)
	// We pass a large dt to ensure things move
	dt := int64(10_000_000)
	for i := 0; i < 5; i++ {
		geo.SimulateGeology(dt, 0.0)
	}

	// Calculate difference
	diffCount := 0
	for i := range geo.Rainfall {
		if geo.Rainfall[i] != initialRainfall[i] {
			diffCount++
		}
	}

	// Assert that rainfall HAS changed
	// This should FAIL currently, confirming the bug
	t.Logf("Rainfall changed in %d cells", diffCount)
	assert.True(t, diffCount > 0, "Rainfall map should change as continents drift")
}
