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

	plates := GeneratePlates(count, topology, seed, 0.30)

	assert.Equal(t, count, len(plates))

	continentalCount := 0
	oceanicCount := 0

	for _, p := range plates {
		if p.Type == PlateContinental {
			continentalCount++
			// Continental: 30-40km thick, granite density
			assert.GreaterOrEqual(t, p.Thickness, 30.0, "Continental thickness should be >= 30km")
			assert.LessOrEqual(t, p.Thickness, 40.0, "Continental thickness should be <= 40km")
			assert.Equal(t, DensityGranite, p.MeanDensity, "Continental density should be granite")
		} else {
			oceanicCount++
			// Oceanic: 6-10km thick, basalt density
			assert.GreaterOrEqual(t, p.Thickness, 6.0, "Oceanic thickness should be >= 6km")
			assert.LessOrEqual(t, p.Thickness, 10.0, "Oceanic thickness should be <= 10km")
			assert.Equal(t, DensityBasalt, p.MeanDensity, "Oceanic density should be basalt")
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

func TestGravityScaling(t *testing.T) {
	// 1. Setup Topology
	topology := spatial.NewCubeSphereTopology(4) // Low res for speed
	seed := int64(12345)

	// 2. Setup Plates
	plates := GeneratePlates(5, topology, seed, 0.4)

	// 3. Test Low Gravity (Mars-like)
	// Mass 0.1 -> Gravity ~0.38g. Max Elev should be ~30km?
	// Earth Max = 12000. 12000 / 0.38 = 31500.
	coreLow := NewPlanetaryCore(0.1, 4.5)
	maxElevLow := coreLow.GetMaxElevation()

	hmLow := NewSphereHeightmap(topology)
	// Force uplift with high scale factor to generate tall mountains
	hmLow = SimulateTectonics(plates, hmLow, topology, 50.0, maxElevLow)

	hmLow.UpdateMinMax()
	_, maxValLow := hmLow.MinMax()

	// 4. Test High Gravity (Super-Earth)
	// Mass 10.0 -> Gravity ~2.9g. Max Elev ~4100m.
	// 12000 / 2.9 = 4137.
	coreHigh := NewPlanetaryCore(10.0, 4.5)
	maxElevHigh := coreHigh.GetMaxElevation()

	hmHigh := NewSphereHeightmap(topology)
	hmHigh = SimulateTectonics(plates, hmHigh, topology, 50.0, maxElevHigh)

	hmHigh.UpdateMinMax()
	_, maxValHigh := hmHigh.MinMax()

	t.Logf("Low Gravity (%.2fg) Max Elev: %.2fm (Limit: %.2f)", coreLow.Gravity, maxValLow, maxElevLow)
	t.Logf("High Gravity (%.2fg) Max Elev: %.2fm (Limit: %.2f)", coreHigh.Gravity, maxValHigh, maxElevHigh)

	// Assertions
	if maxValLow <= maxValHigh {
		t.Errorf("Low gravity should allow taller mountains than high gravity. Low: %.2f, High: %.2f", maxValLow, maxValHigh)
	}

	// Check clamping
	if maxValHigh > maxElevHigh+1.0 { // Tolerance
		t.Errorf("High gravity mountains exceeded limit! Val: %.2f, Limit: %.2f", maxValHigh, maxElevHigh)
	}
}

func TestSimulateTectonics(t *testing.T) {
	resolution := 16
	topology := spatial.NewCubeSphereTopology(resolution)
	count := 5
	seed := int64(12345)

	plates := GeneratePlates(count, topology, seed, 0.30)
	hm := NewSphereHeightmap(topology)

	// Initialize with zeros
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				hm.Set(spatial.Coordinate{Face: face, X: x, Y: y}, 0)
			}
		}
	}

	result := SimulateTectonics(plates, hm, topology, 1.0, 15000.0)

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
	plates := GeneratePlates(8, topology, 12345, 0.30)

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
	plates := GeneratePlates(8, topology, 54321, 0.30)

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
	plates := GeneratePlates(8, topology, 99999, 0.30)

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
	plates := GeneratePlates(5, topology, 11111, 0.30)
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
