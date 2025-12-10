package geography

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHeightmap(t *testing.T) {
	width, height := 50, 50
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, width, height, seed)
	hm := GenerateHeightmap(10, 10, plates, 0, 1.0, 1.0)

	assert.NotNil(t, hm)
	assert.Equal(t, width, hm.Width)
	assert.Equal(t, height, hm.Height)

	// Check elevation ranges
	// Oceanic plates should be deep negative
	// Continental should be positive
	// Tectonic interactions should create extremes

	hasOcean := false
	hasLand := false
	hasMountains := false
	hasTrenches := false

	for _, val := range hm.Elevations {
		if val < -3000 {
			hasOcean = true
		}
		if val > 0 {
			hasLand = true
		}
		if val > 2000 {
			hasMountains = true
		}
		if val < -6000 {
			hasTrenches = true
		}
	}

	assert.True(t, hasOcean, "Should have deep ocean")
	assert.True(t, hasLand, "Should have land")
	// Mountains and trenches depend on random plate movement, but with 5 plates and seed 12345,
	// we expect some interaction.
	assert.True(t, hasMountains || hasTrenches, "Should have some tectonic features")

	assert.True(t, hm.MinElev < hm.MaxElev)
}
