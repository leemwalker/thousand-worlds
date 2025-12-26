package geography

import (
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePlates(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	count := 10
	seed := int64(12345)

	plates := GeneratePlates(count, topology, seed)

	assert.Equal(t, count, len(plates))

	continentalCount := 0
	oceanicCount := 0

	for _, p := range plates {
		if p.Type == PlateContinental {
			continentalCount++
		} else {
			oceanicCount++
		}
		assert.NotEqual(t, 0, p.ID)
		assert.True(t, p.Thickness > 0)
	}

	// Check ratio (approx 30% continental)
	assert.Equal(t, 3, continentalCount)
	assert.Equal(t, 7, oceanicCount)
}

func TestSimulateTectonics(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, topology, seed)
	hm := NewSphereHeightmap(topology)

	// Initialize with zeros
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				hm.Set(spatial.Coordinate{Face: face, X: x, Y: y}, 0)
			}
		}
	}

	result := SimulateTectonics(plates, hm, topology, 1.0)

	assert.NotNil(t, result)

	// Check that we have some non-zero modifiers (boundaries)
	hasChanges := false
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				val := result.Get(spatial.Coordinate{Face: face, X: x, Y: y})
				if val != 0 {
					hasChanges = true
					break
				}
			}
			if hasChanges {
				break
			}
		}
		if hasChanges {
			break
		}
	}
	assert.True(t, hasChanges, "Tectonic simulation should produce elevation changes")
}
