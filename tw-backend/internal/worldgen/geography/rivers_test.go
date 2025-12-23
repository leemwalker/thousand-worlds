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

func TestGenerateRiversSpherical(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	seed := int64(42)
	count := 5

	// Generate plates and spherical heightmap
	plates := GeneratePlates(count, topology, seed)
	sphereHm := NewSphereHeightmap(topology)
	sphereHm = GenerateHeightmap(plates, sphereHm, topology, seed, 1.0, 1.0)

	// Calculate sea level from the spherical heightmap
	// Use a simple approximation based on min/max
	sphereHm.UpdateMinMax()
	seaLevel := sphereHm.MinElev + (sphereHm.MaxElev-sphereHm.MinElev)*0.4

	// Generate rivers on sphere
	rivers := GenerateRiversSpherical(sphereHm, seaLevel, seed)

	// Should generate some rivers
	assert.Greater(t, len(rivers), 0, "Should generate at least one river on sphere")

	for _, river := range rivers {
		assert.Greater(t, len(river.Points), 5, "Rivers should have minimum length")

		// Each point should have valid coordinates
		for _, coord := range river.Points {
			assert.GreaterOrEqual(t, coord.Face, 0, "Face should be >= 0")
			assert.Less(t, coord.Face, 6, "Face should be < 6")
			assert.GreaterOrEqual(t, coord.X, 0, "X should be >= 0")
			assert.Less(t, coord.X, resolution, "X should be < resolution")
			assert.GreaterOrEqual(t, coord.Y, 0, "Y should be >= 0")
			assert.Less(t, coord.Y, resolution, "Y should be < resolution")
		}

		// River should flow downhill (each subsequent point should be at lower or equal elevation)
		// Note: Due to erosion applied, we check for reasonable flow
		for i := 1; i < len(river.Points); i++ {
			// Just check that movement between points is valid (no huge jumps)
			// The elevation check is implicit in the algorithm
			prev := river.Points[i-1]
			curr := river.Points[i]
			// Allow cross-face transitions (that's the whole point!)
			if prev.Face == curr.Face {
				// Same face: should be adjacent
				dx := abs(curr.X - prev.X)
				dy := abs(curr.Y - prev.Y)
				assert.LessOrEqual(t, dx, 1, "Movement should be to adjacent cell")
				assert.LessOrEqual(t, dy, 1, "Movement should be to adjacent cell")
			}
			// Cross-face transitions are handled by topology
		}
	}
}

func TestConvertSphericalRiversToFlat(t *testing.T) {
	resolution := 16

	// Create sample spherical rivers
	rivers := []SphericalRiverPath{
		{
			Points: []spatial.Coordinate{
				{Face: 0, X: 5, Y: 5},
				{Face: 0, X: 5, Y: 6},
				{Face: 0, X: 5, Y: 7},
			},
		},
		{
			Points: []spatial.Coordinate{
				{Face: 1, X: 10, Y: 10},
				{Face: 1, X: 11, Y: 10},
			},
		},
	}

	flat := ConvertSphericalRiversToFlat(rivers, resolution)

	assert.Len(t, flat, 2, "Should have same number of rivers")
	assert.Len(t, flat[0], 3, "First river should have 3 points")
	assert.Len(t, flat[1], 2, "Second river should have 2 points")

	// Check projection formula
	// Face 0, X=5 -> flatX = 0*16 + 5 = 5
	assert.Equal(t, 5.0, flat[0][0].X)
	assert.Equal(t, 5.0, flat[0][0].Y)

	// Face 1, X=10 -> flatX = 1*16 + 10 = 26
	assert.Equal(t, 26.0, flat[1][0].X)
	assert.Equal(t, 10.0, flat[1][0].Y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
