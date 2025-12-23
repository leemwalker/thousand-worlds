package geography

import (
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHeightmap(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, topology, seed)
	hm := NewSphereHeightmap(topology)
	hm = GenerateHeightmap(plates, hm, topology, seed, 1.0, 1.0)

	assert.NotNil(t, hm)
	assert.Equal(t, resolution, hm.Resolution())

	// Check elevation ranges
	// Oceanic plates should be deep negative
	// Continental should be positive
	// Tectonic interactions should create extremes

	hasOcean := false
	hasLand := false
	hasMountains := false
	hasTrenches := false

	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				val := hm.Get(spatial.Coordinate{Face: face, X: x, Y: y})
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
		}
	}

	assert.True(t, hasOcean, "Should have deep ocean")
	assert.True(t, hasLand, "Should have land")
	// Mountains and trenches depend on random plate movement, but with 5 plates and seed 12345,
	// we expect some interaction.
	assert.True(t, hasMountains || hasTrenches, "Should have some tectonic features")

	hm.UpdateMinMax()
	assert.True(t, hm.MinElev < hm.MaxElev)
}
