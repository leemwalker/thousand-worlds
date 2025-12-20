package underground_test

import (
	"testing"

	"tw-backend/internal/worldgen/underground"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: Underground System
// =============================================================================
// These tests verify underground geological features including caves,
// strata, fossils, and mining mechanics.

// -----------------------------------------------------------------------------
// Scenario: Karst Topography (Limestone Dissolution)
// -----------------------------------------------------------------------------
// Given: Area with thick limestone strata
// When: High rainfall + CO2 over millions of years
// Then: Karst caves should form
//
//	AND Caves should have connected chambers
func TestBDD_Karst_LimestoneDissolution(t *testing.T) {
	// Create column grid with limestone
	grid := underground.NewColumnGrid(10, 10)
	col := grid.Get(5, 5)
	require.NotNil(t, col, "Column should exist")

	// Add thick limestone layer
	col.Strata = []underground.StrataLayer{
		{TopZ: 0, BottomZ: -100, Material: "limestone", Hardness: 3.0, Porosity: 0.3},
	}
	col.Surface = 0

	// Simulate with high rainfall
	rainfall := make([]float64, 100)
	for i := range rainfall {
		rainfall[i] = 0.8 // High rainfall
	}

	config := underground.DefaultCaveConfig()
	config.DissolutionRate = 0.01 // Increased for testing

	caves := underground.SimulateCaveFormation(grid, rainfall, 1_000_000, 42, config)

	assert.Greater(t, len(caves), 0, "Should form at least one karst cave")
	if len(caves) > 0 {
		assert.Equal(t, "karst", caves[0].CaveType, "Cave should be karst type")
	}
}

// -----------------------------------------------------------------------------
// Scenario: Strata Composition (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various geological contexts
// When: Strata are generated
// Then: Layer composition should match expectations
func TestBDD_Strata_Composition(t *testing.T) {
	scenarios := []struct {
		name           string
		isVolcanic     bool
		isAncient      bool
		expectMaterial string // Expected primary material
	}{
		{"Volcanic region", true, false, "basalt"},
		{"Ancient craton", false, true, "granite"},
		{"Sedimentary basin", false, false, "sandstone"},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			ctx := underground.StrataContext{
				IsVolcanic: sc.isVolcanic,
				IsAncient:  sc.isAncient,
			}

			strata := underground.GenerateStrataForContext(ctx)

			require.Greater(t, len(strata), 0, "Should generate at least one layer")

			// Find the primary rock layer (not soil)
			var primaryMaterial string
			for _, layer := range strata {
				if layer.Material != "soil" {
					primaryMaterial = layer.Material
					break
				}
			}
			assert.Equal(t, sc.expectMaterial, primaryMaterial,
				"Primary material should match geological context")
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Fossil Formation Timeline
// -----------------------------------------------------------------------------
// Given: Dead organism buried in sediment
// When: 10,000+ years pass with proper conditions
// Then: Fossil deposit should form
//
//	AND Fossil should retain species info
func TestBDD_Fossil_FormationTimeline(t *testing.T) {
	grid := underground.NewColumnGrid(10, 10)
	col := grid.Get(5, 5)
	require.NotNil(t, col, "Column should exist")

	// Create organic deposit representing dead organism
	deadOrganism := underground.Deposit{
		ID:       uuid.New(),
		Type:     "organic",
		DepthZ:   -10,
		Quantity: 100,
		Source: &underground.OrganicSource{
			OriginalEntityID: uuid.New(),
			Species:          "Trilobite",
			DeathYear:        0,
			BurialYear:       0,
		},
	}

	fossilCount := underground.SimulateFossilFormation(grid, deadOrganism, 100_000)

	assert.Greater(t, fossilCount, 0, "Should form at least one fossil deposit")
}

// -----------------------------------------------------------------------------
// Scenario: Oil Formation (Deep Burial + Heat)
// -----------------------------------------------------------------------------
// Given: Ancient organic deposits (algae/plankton)
// When: Buried deep with heat + cap rock
// Then: Oil deposits should form
//
//	AND Oil should be trapped under impermeable layer
func TestBDD_Oil_Formation(t *testing.T) {
	grid := underground.NewColumnGrid(10, 10)
	col := grid.Get(5, 5)
	require.NotNil(t, col, "Column should exist")

	// Setup organic source and cap rock
	col.Strata = []underground.StrataLayer{
		{TopZ: 0, BottomZ: -100, Material: "shale", Hardness: 4.0, Porosity: 0.05}, // Cap rock
		{TopZ: -100, BottomZ: -500, Material: "sandstone", Hardness: 5.0, Porosity: 0.2},
	}
	col.Resources = []underground.Deposit{
		{ID: uuid.New(), Type: "organic", DepthZ: -200, Quantity: 1000},
	}

	oilDeposits := underground.SimulateOilFormation(grid, 100_000_000) // 100M years

	require.NotNil(t, oilDeposits, "Should generate oil deposits, not nil")
	assert.Greater(t, len(oilDeposits), 0, "Should form at least one oil deposit")
}

// -----------------------------------------------------------------------------
// Scenario: Mining Hardness Requirement
// -----------------------------------------------------------------------------
// Given: Rock with hardness 7 (granite)
// When: Mining with tool strength 5
// Then: Mining should fail or be extremely slow
//
//	AND Should require tool strength >= rock hardness
func TestBDD_Mining_HardnessRequirement(t *testing.T) {
	// Stone pick has max hardness 4
	stonePick := underground.StandardTools["stone_pick"]

	// Granite has hardness 7.5
	graniteHardness := 7.5

	// Iron pick has max hardness 6
	ironPick := underground.StandardTools["iron_pick"]

	// Calculate mining speeds
	stoneSpeed := underground.CalculateMiningSpeed(stonePick, graniteHardness)
	ironSpeed := underground.CalculateMiningSpeed(ironPick, graniteHardness)
	ironOnLimestone := underground.CalculateMiningSpeed(ironPick, 4.0)

	// Stone pick should fail on granite (hardness 7.5 > 4)
	assert.Equal(t, 0.0, stoneSpeed, "Stone pick should not mine granite")

	// Iron pick should also fail on granite (hardness 7.5 > 6)
	assert.Equal(t, 0.0, ironSpeed, "Iron pick should not mine granite")

	// Iron pick should work on limestone (hardness 4 <= 6)
	assert.Greater(t, ironOnLimestone, 0.0, "Iron pick should mine limestone")
}

// -----------------------------------------------------------------------------
// Scenario: Burrow Creation
// -----------------------------------------------------------------------------
// Given: Creature with digging ability
// When: Creating tunnel in soft soil
// Then: Void space should be created
//
//	AND Should fail in rock without proper tools
func TestBDD_Burrow_Creation(t *testing.T) {
	col := &underground.WorldColumn{
		X: 5, Y: 5,
		Surface: 0,
		Strata: []underground.StrataLayer{
			{TopZ: 0, BottomZ: -10, Material: "soil", Hardness: 1.0, Porosity: 0.5},
		},
	}

	// Creature attempts to burrow
	burrow := underground.SimulateBurrowCreation(col, -5, 2.0) // Tool strength 2

	require.NotNil(t, burrow, "Burrow should be created in soft soil")
	assert.Equal(t, "burrow", burrow.VoidType, "Should be a burrow type void")
}

// -----------------------------------------------------------------------------
// Scenario: Rock Cycle (Metamorphism)
// -----------------------------------------------------------------------------
// Given: Sedimentary rock under high heat/pressure
// When: Tectonic forces apply
// Then: Rock should transform to metamorphic type
//
//	AND Limestone -> Marble, Sandstone -> Quartzite
func TestBDD_RockCycle_Metamorphism(t *testing.T) {
	grid := underground.NewColumnGrid(10, 10)
	col := grid.Get(5, 5)
	require.NotNil(t, col, "Column should exist")

	col.Strata = []underground.StrataLayer{
		{TopZ: -500, BottomZ: -1000, Material: "limestone", Hardness: 3.0, Porosity: 0.2},
	}

	// Apply high temperature and pressure
	transformed := underground.SimulateRockCycle(grid, 10_000_000, 600, 200) // 600°C, 200 MPa

	assert.Greater(t, transformed, 0, "Should transform at least one stratum")
}

// -----------------------------------------------------------------------------
// Scenario: Ley Line Nodes
// -----------------------------------------------------------------------------
// Given: Underground intersection of ley lines
// When: Magical energy accumulates
// Then: Ley line node should form
//
//	AND Node should have magical properties
func TestBDD_LeyLine_Nodes(t *testing.T) {
	grid := underground.NewColumnGrid(20, 20)

	// Generate ley line nodes with high magic level
	nodes := underground.GenerateLeyLineNodes(grid, 0.5, 42)

	assert.Greater(t, len(nodes), 0, "Should generate at least one ley line node")
	if len(nodes) > 0 {
		assert.Greater(t, nodes[0].Power, 0.0, "Node should have positive power")
		assert.Greater(t, nodes[0].Connections, 0, "Node should have connections")
	}
}

// -----------------------------------------------------------------------------
// Scenario: Sedimentary Deposition
// -----------------------------------------------------------------------------
// Given: River delta or lake bottom
// When: Sediments accumulate over time
// Then: New sedimentary strata should form
//
//	AND New layers should be on top of existing
func TestBDD_Sediment_Deposition(t *testing.T) {
	grid := underground.NewColumnGrid(10, 10)
	col := grid.Get(5, 5)
	require.NotNil(t, col, "Column should exist")

	initialLayerCount := len(col.Strata)

	// Simulate sediment deposition: 0.001m/year for 10000 years = 10m of sediment
	thickness := underground.SimulateSedimentDeposition(grid, 0.001, 10000)

	assert.Equal(t, 10.0, thickness, "Should deposit 10m of sediment")
	assert.Greater(t, len(col.Strata), initialLayerCount, "Should add new strata layers")
}

// -----------------------------------------------------------------------------
// Scenario: Aquifer Puncture
// -----------------------------------------------------------------------------
// Given: Underground water table
// When: Mining breaks through to aquifer
// Then: Water should flood the mine
//
//	AND Flow rate should depend on porosity
func TestBDD_Aquifer_Puncture(t *testing.T) {
	col := &underground.WorldColumn{
		X: 5, Y: 5,
		Surface: 0,
		Strata: []underground.StrataLayer{
			{TopZ: 0, BottomZ: -50, Material: "sandstone", Hardness: 5.0, Porosity: 0.3},
		},
	}

	flowRate := underground.PunctureAquifer(col, -30)

	assert.Greater(t, flowRate, 0.0, "Should return positive water flow rate")
}

// -----------------------------------------------------------------------------
// Scenario: Tectonic Faulting
// -----------------------------------------------------------------------------
// Given: Strata layers at fault line
// When: Earthquake occurs
// Then: Layers should shift vertically
//
//	AND Cave passages may collapse or open
func TestBDD_Tectonic_Faulting(t *testing.T) {
	grid := underground.NewColumnGrid(10, 10)

	// Get a column to track before/after
	col := grid.Get(7, 5)
	require.NotNil(t, col, "Column should exist")

	// Add a stratum to track displacement
	col.Strata = []underground.StrataLayer{
		{TopZ: 0, BottomZ: -50, Material: "granite", Hardness: 7.0, Porosity: 0.1},
	}
	initialTopZ := col.Strata[0].TopZ

	// Simulate fault at x=5 with 10m slip
	affected := underground.SimulateFaulting(grid, 5, 10.0)

	assert.Greater(t, affected, 0, "Should affect at least one column")
	assert.Equal(t, initialTopZ+10.0, col.Strata[0].TopZ, "Strata should be shifted by slip amount")
}

// -----------------------------------------------------------------------------
// Scenario: Geode Formation
// -----------------------------------------------------------------------------
// Given: Volcanic void with mineral-rich water
// When: Slow crystallization over millions of years
// Then: Geode deposit should form
//
//	AND Should contain crystal minerals
func TestBDD_Geode_Formation(t *testing.T) {
	grid := underground.NewColumnGrid(10, 10)
	col := grid.Get(5, 5)
	require.NotNil(t, col, "Column should exist")

	// Setup volcanic conditions with fluid
	col.Strata = []underground.StrataLayer{
		{TopZ: 0, BottomZ: -100, Material: "basalt", Hardness: 6.0, Porosity: 0.3},
	}

	geodes := underground.GenerateGeodes(grid, 42)

	assert.Greater(t, len(geodes), 0, "Should generate at least one geode")
	if len(geodes) > 0 {
		assert.NotEmpty(t, string(geodes[0].Type), "Geode should have a type")
	}
}

// -----------------------------------------------------------------------------
// Scenario: Roof Stability / Collapse
// -----------------------------------------------------------------------------
// Given: Cave with large unsupported span
// When: Roof stability calculated
// Then: Low stability areas should be identified
//
//	AND Collapse risk should increase with span
func TestBDD_Roof_Stability(t *testing.T) {
	cave := underground.NewCave("karst", 0)

	// Add large chamber
	largePos := underground.Vector3{X: 50, Y: 50, Z: -100}
	nodeID := cave.AddNode(largePos, 30.0, 20.0) // Very large: 30m radius, 20m high

	stability := underground.CalculateRoofStability(cave, nodeID)

	assert.Greater(t, stability, 0.0, "Should return positive stability value")
	assert.Less(t, stability, 0.5, "Large unsupported span should have low stability")
}
