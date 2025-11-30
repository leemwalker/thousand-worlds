package geography

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePlates(t *testing.T) {
	width, height := 100, 100
	count := 10
	seed := int64(12345)

	plates := GeneratePlates(count, width, height, seed)

	assert.Equal(t, count, len(plates))

	continentalCount := 0
	oceanicCount := 0

	for _, p := range plates {
		if p.Type == PlateContinental {
			continentalCount++
		} else {
			oceanicCount++
		}
		assert.NotEqual(t, 0, p.PlateID)
		assert.True(t, p.Thickness > 0)
	}

	// Check ratio (approx 30% continental)
	assert.Equal(t, 3, continentalCount)
	assert.Equal(t, 7, oceanicCount)
}

func TestSimulateTectonics(t *testing.T) {
	width, height := 50, 50
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, width, height, seed)
	modifiers := SimulateTectonics(plates, width, height)

	assert.NotNil(t, modifiers)
	assert.Equal(t, width, modifiers.Width)
	assert.Equal(t, height, modifiers.Height)

	// Check that we have some non-zero modifiers (boundaries)
	hasChanges := false
	for _, val := range modifiers.Elevations {
		if val != 0 {
			hasChanges = true
			break
		}
	}
	assert.True(t, hasChanges, "Tectonic simulation should produce elevation changes")
}
