package geography

import (
	"testing"

	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRivers(t *testing.T) {
	// Use spherical topology but convert to flat heightmap for river generation
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, topology, seed)
	sphereHm := NewSphereHeightmap(topology)
	sphereHm = GenerateHeightmap(plates, sphereHm, topology, seed, 1.0, 1.0)

	// Convert to flat heightmap for river generation (legacy)
	width, height := 50, 50
	hm := NewHeightmap(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			face := (x / resolution) % 6
			fx := x % resolution
			fy := y % resolution
			if fx >= resolution {
				fx = resolution - 1
			}
			if fy >= resolution {
				fy = resolution - 1
			}
			coord := spatial.Coordinate{Face: face, X: fx, Y: fy}
			hm.Set(x, y, sphereHm.Get(coord))
		}
	}

	// Ensure some land exists
	seaLevel := AssignOceanLand(hm, 0.3) // 30% land

	rivers := GenerateRivers(hm, seaLevel, seed)

	if len(rivers) > 0 {
		for _, river := range rivers {
			assert.True(t, len(river) > 1)

			// Check valid points
			for i := 0; i < len(river)-1; i++ {
				p1 := river[i]
				assert.True(t, p1.X >= 0 && p1.X < float64(width))
				assert.True(t, p1.Y >= 0 && p1.Y < float64(height))
			}
		}
	}
}
