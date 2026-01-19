package geography

import (
	"math/rand"
	"testing"
	"time"
	"tw-backend/internal/spatial"

	"github.com/stretchr/testify/assert"
)

func TestClampFloat(t *testing.T) {
	assert.Equal(t, 0.5, clampFloat(0.5, 0.0, 1.0))
	assert.Equal(t, 1.0, clampFloat(1.5, 0.0, 1.0))
	assert.Equal(t, 0.0, clampFloat(-0.5, 0.0, 1.0))
}

func TestSimulateSoilFertility(t *testing.T) {
	assert.Equal(t, 0.9, SimulateSoilFertility(600, true))
	assert.Equal(t, 0.6, SimulateSoilFertility(100, true))
	assert.Equal(t, 0.5, SimulateSoilFertility(100, false))
}

func TestSimulateAtollFormation(t *testing.T) {
	elev, isReef := SimulateAtollFormation(3.0, "volcanic")
	assert.Equal(t, -10.0, elev)
	assert.True(t, isReef)

	elev, isReef = SimulateAtollFormation(1.0, "volcanic")
	assert.Equal(t, 100.0, elev)
	assert.False(t, isReef)

	elev, isReef = SimulateAtollFormation(3.0, "continental")
	assert.Equal(t, 100.0, elev)
	assert.False(t, isReef)
}

func TestSimulateExtension(t *testing.T) {
	terrain, thick := SimulateExtension(0.6)
	assert.Equal(t, "alternating_ridge_valley", terrain)
	assert.Equal(t, 0.8, thick)

	terrain, thick = SimulateExtension(0.4)
	assert.Equal(t, "flat", terrain)
	assert.Equal(t, 1.0, thick)
}

func TestSimulateTerraneAccretion(t *testing.T) {
	assert.True(t, SimulateTerraneAccretion(100.0, 5.0))
	assert.False(t, SimulateTerraneAccretion(10.0, 5.0))
}

func TestCalculateFragmentationEffects(t *testing.T) {
	spec, size := CalculateFragmentationEffects(0.8)
	assert.Equal(t, 2.0, spec)
	assert.Equal(t, 0.8, size)

	spec, size = CalculateFragmentationEffects(0.5)
	assert.Equal(t, 1.0, spec)
	assert.Equal(t, 1.0, size)
}

func TestGenerateHeightmapWithTidalStress(t *testing.T) {
	res := 16
	topo := spatial.NewCubeSphereTopology(res)
	// Args: count=5, topo, seed=123, continentalPerc=0.3
	plates := GeneratePlates(5, topo, 123, 0.3)

	shm := NewSphereHeightmap(topo)

	// Test Run
	result := GenerateHeightmapWithTidalStress(plates, shm, topo, 123, 1.0, 1.0, 1.0, 1.0, 9000.0)

	assert.NotNil(t, result)
	min, max := result.MinMax()
	assert.Less(t, min, max)
}

func TestIsSinkAndGetFlux(t *testing.T) {
	res := 16
	totalCells := 6 * res * res

	coord := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	coordIdx := 0*res*res + 5*res + 5

	hl := &HydrologyLayer{
		FlowDirection: make([]int, totalCells),
		Flux:          make([]float64, totalCells),
		Resolution:    res,
	}

	// Initialize FlowDirection to -1 (Sink)
	for i := range hl.FlowDirection {
		hl.FlowDirection[i] = -1
	}

	// Case 1: Sink (FlowDirection is -1)
	assert.True(t, hl.IsSink(coord))

	// Case 2: Flowing somewhere else
	hl.FlowDirection[coordIdx] = 0 // Pointing to index 0
	// IsSink logic: returns h.FlowDirection[idx] == -1
	assert.False(t, hl.IsSink(coord))

	// Flux
	hl.Flux[coordIdx] = 100.0
	assert.Equal(t, 100.0, hl.GetFlux(coord))
	assert.Equal(t, 0.0, hl.GetFlux(spatial.Coordinate{Face: 1, X: 0, Y: 0}))
}

func TestSphereHeightmap_Bounds(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(16)
	shm := NewSphereHeightmap(topo)

	invalidCoord := spatial.Coordinate{Face: -1, X: 0, Y: 0}

	// Get/Set bounds
	assert.Equal(t, 0.0, shm.Get(invalidCoord))
	shm.Set(invalidCoord, 100.0) // Should not panic or set

	// Cell Data bounds
	emptyData := shm.GetCellData(invalidCoord)
	assert.Equal(t, 0.0, emptyData.RockHardness)

	shm.SetCellData(invalidCoord, CellData{RockHardness: 1.0}) // Should not panic
}

func TestBiomes_Classify(t *testing.T) {
	// Tests for ClassifyBiome(tempC, rainfallMM, drainage, elevation, seaLevel, flux, isLake)
	// 1. Lake
	assert.Equal(t, BiomeLake, ClassifyBiome(20, 1000, 0.5, 100, 0, 10, true))
	// 2. Ocean
	assert.Equal(t, BiomeOcean, ClassifyBiome(20, 1000, 0.5, -100, 0, 10, false))
	// 3. Alpine
	assert.Equal(t, BiomeAlpine, ClassifyBiome(0, 1000, 0.5, 4000, 0, 10, false))
	// 4. Rainforest
	assert.Equal(t, BiomeRainforest, ClassifyBiome(25, 2500, 0.5, 100, 0, 10, false))
	// 5. Desert
	assert.Equal(t, BiomeDesert, ClassifyBiome(30, 100, 0.5, 100, 0, 10, false))
	// 6. Tundra
	assert.Equal(t, BiomeTundra, ClassifyBiome(-10, 200, 0.5, 100, 0, 10, false))
	// 7. Wetland
	assert.Equal(t, BiomeWetland, ClassifyBiome(15, 1000, 0.1, 100, 0, 500, false))
}

func TestErosion_StreamPower(t *testing.T) {
	res := 16
	topo := spatial.NewCubeSphereTopology(res)
	shm := NewSphereHeightmap(topo)

	c1 := spatial.Coordinate{Face: 0, X: 0, Y: 0}
	c2 := spatial.Coordinate{Face: 0, X: 0, Y: 1}
	c3 := spatial.Coordinate{Face: 0, X: 0, Y: 2}

	shm.Set(c1, 100.0)
	shm.Set(c2, 80.0)
	shm.Set(c3, 60.0)

	totalCells := 6 * res * res
	hydro := &HydrologyLayer{
		FlowDirection: make([]int, totalCells),
		Flux:          make([]float64, totalCells),
		Resolution:    res,
	}
	for i := range hydro.FlowDirection {
		hydro.FlowDirection[i] = -1
	}

	idx1 := 0*res*res + 0*res + 0
	idx2 := 0*res*res + 1*res + 0
	idx3 := 0*res*res + 2*res + 0

	hydro.FlowDirection[idx1] = idx2
	hydro.FlowDirection[idx2] = idx3
	hydro.FlowDirection[idx3] = -1

	hydro.Flux[idx1] = 10000.0
	hydro.Flux[idx2] = 20000.0
	hydro.Flux[idx3] = 30000.0

	ApplyStreamPowerErosion(shm, hydro, nil, 1.0, 0.0)

	assert.True(t, shm.Get(c1) < 100.0, "c1 should erode")
}

func TestErosion_Differential(t *testing.T) {
	res := 16
	topo := spatial.NewCubeSphereTopology(res)
	shm := NewSphereHeightmap(topo)

	for i := 0; i < 6; i++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				shm.Set(spatial.Coordinate{Face: i, X: x, Y: y}, 50.0+float64(x))
			}
		}
	}
	ApplyDifferentialErosion(shm, topo, 100, 123, 0.0)

	changed := false
	for i := 0; i < 6; i++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				if shm.Get(spatial.Coordinate{Face: i, X: x, Y: y}) != 50.0+float64(x) {
					changed = true
					break
				}
			}
		}
	}
	assert.True(t, changed, "Differential erosion should modify terrain")
}

func TestCatastrophes(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	topo := spatial.NewCubeSphereTopology(16)
	shm := NewSphereHeightmap(topo)

	// 1. Volcanoes
	ApplyVolcanicMountains(shm, topo, 1.0, rng)
	shm.UpdateMinMax()
	min, max := shm.MinMax()
	_ = min
	assert.NotEqual(t, 0.0, max)

	// 2. Impact Crater
	shm2 := NewSphereHeightmap(topo)
	// Set base level
	for f := 0; f < 6; f++ {
		for i := 0; i < 16*16; i++ {
			shm2.SetCellData(spatial.Coordinate{Face: f, X: i % 16, Y: i / 16}, CellData{})
		}
	}
	shm2.UpdateMinMax()
	ApplyImpactCrater(shm2, topo, 0.5, rng)
	shm2.UpdateMinMax()
	min2, max2 := shm2.MinMax()
	assert.True(t, min2 < 0 || max2 > 0, "Crater should modify terrain")

	// 3. Flood Basalt
	ApplyFloodBasalt(shm2, topo, 0.5, rng)
	shm2.UpdateMinMax()
	min3, max3 := shm2.MinMax()
	assert.True(t, max3 > max2 || min3 > min2, "Flood basalt should add elevation")

	// 4. Ice Age
	// Need elevation > threshold
	c := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	shm2.Set(c, 5000.0)
	shm2.UpdateMinMax()
	shm2.MaxElev = 6000.0
	ApplyIceAgeEffects(shm2, topo, 1.0)
	assert.Less(t, shm2.Get(c), 5000.0, "Ice age should erode high peaks")
}

func TestCoastal_Simulation(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(16)
	shm := NewSphereHeightmap(topo)

	// Set up a coastline
	// Face 0: Land (10m), except Y=15 (0m - Coast)
	// Face 1: Ocean (-50m)
	res := 16
	for y := 0; y < res; y++ {
		for x := 0; x < res; x++ {
			c := spatial.Coordinate{Face: 0, X: x, Y: y}
			shm.Set(c, 10.0)
		}
	}
	// Deep water neighbor
	cWater := spatial.Coordinate{Face: 0, X: 8, Y: 8}
	shm.Set(cWater, -100.0) // Deep hole in middle of land

	config := DefaultCoastalConfig()
	SimulateCoastalErosion(shm, topo, 1.0, 0.0, config)

	// Check neighbor erosion
	neighbor := spatial.Coordinate{Face: 0, X: 7, Y: 8}
	// Should erode from 10.0
	assert.Less(t, shm.Get(neighbor), 10.0)
}

func TestCoastal_Features(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(16)
	shm := NewSphereHeightmap(topo)

	// 1. Beaches
	c := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	shm.Set(c, -1.0)                            // Shallow water
	shm.SetCellData(c, CellData{Sediment: 5.0}) // Lots of sediment
	FormBeaches(shm, topo, 0.0)
	assert.Greater(t, shm.Get(c), -1.0, "Beach should form (raise elevation)")

	// 2. Intertidal
	c2 := spatial.Coordinate{Face: 0, X: 6, Y: 6}
	shm.Set(c2, -1.5)
	// Need deep neighbor
	shm.Set(spatial.Coordinate{Face: 0, X: 6, Y: 7}, -10.0)

	MarkIntertidalZones(shm, topo, 0.0, 4.0)
	data := shm.GetCellData(c2)
	assert.True(t, data.IsIntertidal, "Should be intertidal")

	// 3. Estuaries
	c3 := spatial.Coordinate{Face: 0, X: 2, Y: 2}
	shm.Set(c3, 0.0) // Sea Level
	shm.SetCellData(c3, CellData{Flux: 1000.0})
	// Neighbor ocean
	shm.Set(spatial.Coordinate{Face: 0, X: 2, Y: 3}, -60.0)

	FormEstuaries(shm, topo, 0.0, 500.0)
	data3 := shm.GetCellData(c3)
	assert.True(t, data3.IsEstuary, "Should be estuary")
}

func TestCoastal_Spits(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(16)
	shm := NewSphereHeightmap(topo)

	// Setup for Spit:
	// c1: Beach with sediment (Face 0, X 5, Y 5)
	// c2: Shallow water ahead (Face 0, X 5, Y 6) (Extension direction)
	// c3: Deep water to the side of c1 (Face 0, X 6, Y 5) (Bay)

	c1 := spatial.Coordinate{Face: 0, X: 5, Y: 5}
	c2 := spatial.Coordinate{Face: 0, X: 5, Y: 6}
	c3 := spatial.Coordinate{Face: 0, X: 6, Y: 5}

	shm.Set(c1, 0.0)                              // Sea Level
	shm.SetCellData(c1, CellData{Sediment: 20.0}) // Plenty of sediment

	shm.Set(c2, -2.0)  // Shallow water ahead (valid for spit extension)
	shm.Set(c3, -30.0) // Deep water to side (Bay)

	// Run
	FormSpitsAndBars(shm, topo, 0.0, 123)

	// Expect c2 to become a spit
	data2 := shm.GetCellData(c2)
	assert.True(t, data2.IsSpit, "Spit should extend to c2")
	assert.Greater(t, shm.Get(c2), -2.0, "Spit should raise elevation")
}
