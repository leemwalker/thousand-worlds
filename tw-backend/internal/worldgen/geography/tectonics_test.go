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
		// Verify age range is 0-200 million years
		assert.GreaterOrEqual(t, p.Age, 0.0)
		assert.LessOrEqual(t, p.Age, 200.0)
	}

	// Check ratio is approximately 30% continental (allow variance due to random assignment)
	// With 10 plates and 30% probability, expect 1-5 continental plates
	assert.GreaterOrEqual(t, continentalCount, 1, "Should have at least 1 continental plate")
	assert.LessOrEqual(t, continentalCount, 6, "Should have at most 6 continental plates")
	assert.GreaterOrEqual(t, oceanicCount, 4, "Should have at least 4 oceanic plates")
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

// =============================================================================
// Province Generation Tests (Phase 5: Geological Provinces)
// =============================================================================

func TestGenerateProvinces_OnlyContinentalPlates(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	plates := GeneratePlates(8, topology, 12345)

	provinces := GenerateProvinces(plates, topology, 12345)

	// Provinces should only be created for continental plates
	assert.NotEmpty(t, provinces, "Should generate at least some provinces")

	// Each province should belong to a continental plate
	for _, prov := range provinces {
		foundPlate := false
		for _, plate := range plates {
			if plate.ID == prov.PlateID && plate.Type == PlateContinental {
				foundPlate = true
				break
			}
		}
		assert.True(t, foundPlate, "Province %d should belong to a continental plate", prov.ID)
	}
}

func TestGenerateProvinces_HardnessValues(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	plates := GeneratePlates(8, topology, 54321)

	provinces := GenerateProvinces(plates, topology, 54321)

	for _, prov := range provinces {
		switch prov.Type {
		case ProvinceCraton:
			assert.InDelta(t, CratonHardness, prov.Hardness, 0.01, "Craton should have hardness ~0.9")
		case ProvinceFoldBelt:
			assert.InDelta(t, FoldBeltHardness, prov.Hardness, 0.01, "FoldBelt should have hardness ~0.5")
		case ProvinceBasin:
			assert.InDelta(t, BasinHardness, prov.Hardness, 0.01, "Basin should have hardness ~0.2")
		default:
			t.Errorf("Unknown province type: %s", prov.Type)
		}
	}
}

func TestGenerateProvinces_3to5PerPlate(t *testing.T) {
	resolution := 32
	topology := spatial.NewCubeSphereTopology(resolution)
	plates := GeneratePlates(8, topology, 99999)

	provinces := GenerateProvinces(plates, topology, 99999)

	// Count provinces per continental plate
	provinceCount := make(map[string]int)
	for _, prov := range provinces {
		provinceCount[prov.PlateID.String()]++
	}

	// Each continental plate should have 3-5 provinces
	continentalCount := 0
	for _, plate := range plates {
		if plate.Type == PlateContinental {
			continentalCount++
			count := provinceCount[plate.ID.String()]
			assert.GreaterOrEqual(t, count, 3, "Continental plate should have at least 3 provinces")
			assert.LessOrEqual(t, count, 5, "Continental plate should have at most 5 provinces")
		}
	}
	assert.Greater(t, continentalCount, 0, "Should have at least one continental plate")
}

func TestInitializeProvinceHardness_SetsAllContinentalCells(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	plates := GeneratePlates(5, topology, 11111)
	hm := NewSphereHeightmap(topology)

	provinces := GenerateProvinces(plates, topology, 11111)
	// 4. Initialize hardness
	InitializeProvinceHardness(hm, plates, provinces, topology, 12345)

	// All cells in continental plates should have non-zero hardness
	for _, plate := range plates {
		if plate.Type == PlateContinental {
			for coord := range plate.Region {
				hardness := hm.GetRockHardness(coord)
				assert.Greater(t, hardness, 0.0,
					"Continental cell at face %d (%d,%d) should have hardness > 0",
					coord.Face, coord.X, coord.Y)
			}
		}
	}
}
