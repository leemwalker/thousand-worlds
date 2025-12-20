package underground_test

import (
	"testing"

	"tw-backend/internal/worldgen/underground"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_FullGeologicalSimulation tests the complete underground system
// working together: column grid, strata, cave formation, magma, deposits, mining
func TestIntegration_FullGeologicalSimulation(t *testing.T) {
	// Create a 10x10 world grid
	grid := underground.NewColumnGrid(10, 10)
	require.NotNil(t, grid)

	// Initialize all columns with surface elevation and strata
	for _, col := range grid.AllColumns() {
		col.Surface = 100
		col.Bedrock = -5000

		// Add continental-style strata
		col.AddStratum("soil", 100, 90, 2, 0, 0.4)
		col.AddStratum("limestone", 90, 0, 4, 100000, 0.3)
		col.AddStratum("granite", 0, -2000, 8, 500000, 0.05)
		col.AddStratum("basalt", -2000, -5000, 7, 1000000, 0.03)
	}

	// PHASE 4: Cave Formation
	t.Run("CaveFormation", func(t *testing.T) {
		config := underground.DefaultCaveConfig()
		config.DissolutionRate = 0.1 // High rate for testing

		rainfall := make([]float64, 100)
		for i := range rainfall {
			rainfall[i] = 0.8 // High rainfall
		}

		caves := underground.SimulateCaveFormation(grid, rainfall, 10_000_000, 42, config)

		// With high dissolution rate and limestone, we should get caves
		assert.NotEmpty(t, caves, "Should form caves in limestone")

		// Verify caves are registered in columns
		totalVoids := 0
		for _, col := range grid.AllColumns() {
			totalVoids += len(col.Voids)
		}
		assert.Greater(t, totalVoids, 0, "Caves should create voids in columns")
	})

	// PHASE 5: Magma Simulation
	t.Run("MagmaSimulation", func(t *testing.T) {
		// Set up hotspot column with magma
		hotspotCol := grid.Get(5, 5)
		hotspotCol.Magma = &underground.MagmaInfo{
			TopZ:        -1000,
			BottomZ:     -3000,
			Temperature: 1500,
			Pressure:    85, // Above eruption threshold
			Viscosity:   0.5,
		}

		// Create tectonic boundaries
		centroids := []underground.Vector3{
			{X: 3, Y: 5, Z: 0},
			{X: 7, Y: 5, Z: 0},
		}
		movements := []underground.Vector3{
			{X: 0.5, Y: 0, Z: 0},
			{X: -0.5, Y: 0, Z: 0},
		}

		boundaries := underground.GetTectonicBoundaries(10, 10, centroids, movements)
		assert.NotEmpty(t, boundaries, "Should detect plate boundaries")

		// Check for convergent boundaries
		hasConvergent := false
		for _, b := range boundaries {
			if b.BoundaryType == "convergent" {
				hasConvergent = true
				break
			}
		}
		assert.True(t, hasConvergent, "Should have convergent boundaries")
	})

	// PHASE 6: Fossil/Oil Formation
	t.Run("DepositEvolution", func(t *testing.T) {
		// Add organic deposits to columns
		col := grid.Get(3, 3)
		deposit := underground.CreateOrganicDeposit(uuid.New(), "fish", 3, 3, 100, 0, false)
		col.AddOrganicDeposit(deposit)

		assert.Equal(t, 1, len(col.Resources))
		assert.Equal(t, "remains", col.Resources[0].Type)

		// Simulate deposit evolution
		config := underground.DefaultDepositConfig()
		config.SedimentRatePerYear = 1.0 // Fast sedimentation

		underground.SimulateDepositEvolution(grid, 5000, config, nil, 42)

		// Deposit should have been buried
		assert.Less(t, col.Resources[0].DepthZ, 100.0, "Deposit should be buried")
	})

	// PHASE 7: Mining
	t.Run("MiningOperations", func(t *testing.T) {
		col := grid.Get(7, 7)

		// Add a resource at depth
		col.Resources = append(col.Resources, underground.Deposit{
			Type:     "iron",
			DepthZ:   80,
			Quantity: 50,
		})

		// Try mining with different tools
		handsTool := underground.StandardTools["hands"]
		ironTool := underground.StandardTools["iron_pick"]

		// Hands can't mine granite at depth
		graniteDepth := -500.0
		result := underground.Mine(col, graniteDepth, handsTool, false)
		assert.False(t, result.Success, "Hands shouldn't mine granite")

		// Iron pick can mine soil
		soilDepth := 95.0
		result = underground.Mine(col, soilDepth, ironTool, true)
		assert.True(t, result.Success, "Iron pick should mine soil")
		assert.NotNil(t, result.VoidCreated, "Should create tunnel")

		// Extract resource
		resourceFound := false
		for i := range col.Resources {
			if col.Resources[i].Type == "iron" {
				extracted, ok := underground.ExtractResource(&col.Resources[i], 10)
				assert.True(t, ok)
				assert.Equal(t, 10.0, extracted)
				assert.Equal(t, 40.0, col.Resources[i].Quantity)
				resourceFound = true
				break
			}
		}
		assert.True(t, resourceFound, "Should find iron resource")
	})

	// FULL INTEGRATION: Burrow Creation
	t.Run("BurrowCreation", func(t *testing.T) {
		col := grid.Get(2, 2)
		ownerID := uuid.New()

		burrow, err := underground.CreateBurrow(col, ownerID, 100, 10, 3)

		assert.NoError(t, err)
		assert.NotNil(t, burrow)
		assert.Equal(t, 3, len(burrow.Chambers))
		assert.Equal(t, 3, len(burrow.Tunnels))
	})

	// Verify final grid state
	t.Run("FinalGridState", func(t *testing.T) {
		totalStrata := 0
		totalVoids := 0
		totalResources := 0

		for _, col := range grid.AllColumns() {
			totalStrata += len(col.Strata)
			totalVoids += len(col.Voids)
			totalResources += len(col.Resources)
		}

		assert.Equal(t, 400, totalStrata, "All columns should have 4 strata")
		assert.Greater(t, totalVoids, 0, "Should have voids from caves/mining/burrows")
		assert.Greater(t, totalResources, 0, "Should have resources")
	})
}

// TestIntegration_CaveNetwork tests cave network creation and connectivity
func TestIntegration_CaveNetwork(t *testing.T) {
	// Create several caves and connect them
	caves := []*underground.Cave{}

	// Create a chain of caves
	for i := 0; i < 5; i++ {
		cave := underground.NewCave("karst", 1000000)
		pos := underground.Vector3{X: float64(i * 20), Y: 0, Z: -50}
		cave.AddNode(pos, 10, 8)
		caves = append(caves, cave)
	}

	// Connect adjacent caves
	underground.ConnectAdjacentCaves(caves, 25) // 25m connection distance

	// First cave should have merged with second
	assert.GreaterOrEqual(t, len(caves[0].Nodes), 2, "Caves should be connected")
}

// TestIntegration_ColumnGridOperations tests thread-safe grid operations
func TestIntegration_ColumnGridOperations(t *testing.T) {
	grid := underground.NewColumnGrid(5, 5)

	// Concurrent reads should work
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			col := grid.Get(2, 2)
			_ = col.Surface
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
