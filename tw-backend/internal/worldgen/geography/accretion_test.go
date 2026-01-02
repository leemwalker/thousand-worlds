package geography

import (
	"math/rand"
	"testing"
	"tw-backend/internal/spatial"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccretionCreatesRuggedTerrain(t *testing.T) {
	// Setup consistency
	rand.Seed(12345)

	// Create topology
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	hm := NewSphereHeightmap(topology)

	// Create two oceanic plates
	// Plate 1: Cell Owner (Younger, Less Dense) -> Should become Island Arc
	// Plate 2: Neighbor (Older, Denser) -> Should subduct
	p1 := TectonicPlate{
		ID:        uuid.New(),
		Type:      PlateOceanic,
		Age:       10.0, // Young
		Thickness: 6.0,
		Region:    make(map[spatial.Coordinate]struct{}),
	}
	// Calculate density for P1 (Base + Age*0.001)
	p1Density := GetPlateDensity(p1) // 3.0 + 0.01 = 3.01

	p2 := TectonicPlate{
		ID:        uuid.New(),
		Type:      PlateOceanic,
		Age:       100.0, // Old
		Thickness: 6.0,
		Region:    make(map[spatial.Coordinate]struct{}),
	}
	p2Density := GetPlateDensity(p2) // 3.0 + 0.1 = 3.1

	assert.Less(t, p1Density, p2Density, "Plate 1 should be less dense than Plate 2")

	plates := []TectonicPlate{p1, p2}

	// Create a fake boundary cache with one cell
	// The cell belongs to P1, neighbor belongs to P2
	coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}

	// Pre-set elevation to deep ocean
	hm.Set(coord, -4000.0)

	cache := &BoundaryCache{
		Cells: []BoundaryCell{
			{
				Coord:        coord,
				PlateIdx:     0,                  // P1
				NeighborIdx:  1,                  // P2
				BoundaryType: BoundaryConvergent, // Convergent
			},
		},
		Valid: true,
	}

	// Verify that CalculateCollisionResult gives us an IslandArc
	result := CalculateCollisionResult(p1, p2, BoundaryConvergent)
	assert.Equal(t, FeatureIslandArc, result.Feature, "Should create Island Arc")

	// Run Simulation multiple times to trigger the probability (15% chance per tick)
	// We need to ensure we hit the accretion logic

	accreted := false
	for i := 0; i < 100; i++ {
		// Pass scaleFactor 1.0
		hm = SimulateTectonicsWithCache(plates, hm, cache, topology, 1.0)

		cellData := hm.GetCellData(coord)
		if cellData.IsContinental {
			accreted = true
			elev := hm.Get(coord)

			// Verify it's rugged (not just flat -100)
			// Our code sets it to 50 + rand*450
			// Or keeps it if it was already high (unlikely here starting at -4000)

			// It should be > 0 and <= 500 roughly (plus potential tectonic uplift from collision)
			assert.Greater(t, elev, 0.0, "Accreted land should be above sea level")
			assert.NotEqual(t, -100.0, elev, "Should not be flat -100m shelf")

			t.Logf("Accretion happened at tick %d. Elevation: %.2f", i, elev)
			break
		}
	}

	assert.True(t, accreted, "Should have triggered island arc accretion within 100 ticks")
}
