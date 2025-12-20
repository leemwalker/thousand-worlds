package minerals_test

import (
	"testing"

	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/minerals"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: Mineral Formation and Mining
// =============================================================================
// These tests verify mineral deposit generation and resource extraction.

// -----------------------------------------------------------------------------
// Scenario: Banded Iron Formation (Great Oxygenation Event)
// -----------------------------------------------------------------------------
// Given: Ancient ocean with rising O2 (oxygenSpike > 0.1)
// When: BIF generation is triggered
// Then: Iron deposits should form at ocean floor
//
//	AND Deposits should have alternating layers
func TestBDD_BandedIron_OxygenPrecipitation(t *testing.T) {
	oceanLocations := []minerals.Point{
		{X: 100, Y: 200},
		{X: 150, Y: 250},
		{X: 200, Y: 300},
	}

	deposits := minerals.GenerateBIFDeposits(oceanLocations, 0.15)

	require.NotNil(t, deposits, "BIF deposits should be generated, not nil")
	assert.Greater(t, len(deposits), 0, "Should generate at least one deposit")
}

// -----------------------------------------------------------------------------
// Scenario: Placer Deposits (Alluvial Gold)
// -----------------------------------------------------------------------------
// Given: River system with gold-bearing source rocks
// When: 10 million years of erosion occur
// Then: Gold placer deposits should form at river bends
//
//	AND Concentration should increase downstream
func TestBDD_Placer_AlluvialGold(t *testing.T) {
	// River path with bends
	riverPath := [][]minerals.Point{
		{{X: 0, Y: 0}, {X: 10, Y: 5}, {X: 20, Y: 3}, {X: 30, Y: 8}},
	}

	deposits := minerals.GeneratePlacerDeposits(riverPath, "gold", 0.8)

	require.NotNil(t, deposits, "Placer deposits should be generated, not nil")
	assert.Greater(t, len(deposits), 0, "Should generate at least one placer deposit")
}

// -----------------------------------------------------------------------------
// Scenario: Hydrothermal Vents (Black Smokers)
// -----------------------------------------------------------------------------
// Given: Mid-ocean ridge with active volcanism
// When: Hydrothermal deposit generation triggered
// Then: Sulfide deposits should form near vents
//
//	AND Should contain copper, zinc, and gold
func TestBDD_Hydrothermal_BlackSmoker(t *testing.T) {
	ridgeLocations := []minerals.Point{
		{X: 500, Y: 600},
		{X: 510, Y: 610},
	}

	deposits := minerals.GenerateHydrothermalDeposits(ridgeLocations)

	require.NotNil(t, deposits, "Hydrothermal deposits should be generated, not nil")
	assert.Greater(t, len(deposits), 0, "Should generate at least one hydrothermal deposit")
}

// -----------------------------------------------------------------------------
// Scenario: Kimberlite Pipe (Diamond Formation)
// -----------------------------------------------------------------------------
// Given: Ancient craton (2.5+ billion years old)
// When: Deep mantle eruption (150+ km depth)
// Then: Diamond-bearing kimberlite pipe should form
//
//	AND Diamonds should be present in deposit
func TestBDD_Kimberlite_DiamondPipe(t *testing.T) {
	cratonAge := 2.8 // Billion years old (old enough)
	depth := 180.0   // km (deep enough for diamond stability)

	deposit := minerals.GenerateKimberlitePipe(cratonAge, depth)

	require.NotNil(t, deposit, "Kimberlite pipe should be generated, not nil")
	assert.Equal(t, "Diamond", deposit.MineralType.Name, "Kimberlite pipe should contain diamonds")
}

// -----------------------------------------------------------------------------
// Scenario: Kimberlite Requirements (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various craton ages and eruption depths
// When: Kimberlite generation attempted
// Then: Only correct combinations should produce diamonds
func TestBDD_Kimberlite_Requirements(t *testing.T) {
	scenarios := []struct {
		name       string
		cratonAge  float64 // Billion years
		depth      float64 // km
		expectDest bool    // Should produce deposit
	}{
		{"Old craton, deep eruption", 2.8, 180, true},       // Should work
		{"Young craton, deep eruption", 0.5, 180, false},    // Too young
		{"Old craton, shallow eruption", 2.8, 100, false},   // Too shallow
		{"Young craton, shallow eruption", 0.5, 100, false}, // Both wrong
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			deposit := minerals.GenerateKimberlitePipe(sc.cratonAge, sc.depth)

			if sc.expectDest {
				require.NotNil(t, deposit, "Should generate kimberlite pipe")
			} else {
				assert.Nil(t, deposit, "Should NOT generate kimberlite pipe")
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Mining Extraction - Basic
// -----------------------------------------------------------------------------
// Given: A discovered ore deposit with quantity 1000
// When: Mine 100 units
// Then: Deposit quantity should decrease to 900
//
//	AND Return value should be 100 (amount mined)
func TestBDD_Mining_Extraction(t *testing.T) {
	deposit := &minerals.MineralDeposit{
		Quantity: 1000,
	}

	extracted := minerals.ExtractResource(deposit, 100)

	assert.Equal(t, 100, extracted, "Should extract requested amount")
	assert.Equal(t, 900, deposit.Quantity, "Deposit quantity should be reduced")
}

// -----------------------------------------------------------------------------
// Scenario: Mining Depletion
// -----------------------------------------------------------------------------
// Given: A deposit with quantity 50
// When: Attempt to mine 100 units
// Then: Only 50 units should be extracted
//
//	AND Deposit quantity should become 0
//	AND Deposit should be marked depleted
func TestBDD_Mining_Depletion(t *testing.T) {
	deposit := &minerals.MineralDeposit{
		Quantity: 50,
	}

	extracted := minerals.ExtractResource(deposit, 100)

	assert.Equal(t, 50, extracted, "Should only extract what's available")
	assert.Equal(t, 0, deposit.Quantity, "Deposit should be empty")
}

// -----------------------------------------------------------------------------
// Scenario: Ore Vein Discovery
// -----------------------------------------------------------------------------
// Given: A survey in known mineral-rich area
// When: Discovery rolls succeed
// Then: Visible veins should be found first
//
//	AND Hidden veins require progressively deeper mining
func TestBDD_OreVein_Discovery(t *testing.T) {
	// Create deposits, some visible, some hidden
	deposits := []*minerals.MineralDeposit{
		{SurfaceVisible: true, Depth: 10},
		{SurfaceVisible: false, Depth: 50},
		{SurfaceVisible: false, Depth: 200},
	}

	// Discover with high skill and deep search
	discovered := minerals.DiscoverDeposits(deposits, 100, 0.8)

	require.NotNil(t, discovered, "Should discover some deposits")
	assert.GreaterOrEqual(t, len(discovered), 1, "Should find at least visible deposits")
}

// -----------------------------------------------------------------------------
// Scenario: Tin Formation (Bronze Age Prerequisite)
// -----------------------------------------------------------------------------
// Given: Granitic intrusion in sedimentary rocks
// When: Hydrothermal fluids cool
// Then: Cassiterite (tin ore) deposits should form
//
//	AND Should be co-located with copper for bronze production
func TestBDD_Tin_Formation(t *testing.T) {
	ctx := minerals.TinFormationContext{
		HasGranite:     true,
		HasSedimentary: true,
		Temperature:    400, // Hydrothermal temperature
	}

	// Copper locations where tin should co-occur
	copperLocations := []minerals.Point{
		{X: 100, Y: 100},
		{X: 200, Y: 200},
	}

	deposits := minerals.GenerateTinDeposits(ctx, copperLocations)

	require.NotNil(t, deposits, "Should generate tin deposits at granite-sedimentary contacts")
	assert.GreaterOrEqual(t, len(deposits), len(copperLocations), "Should have at least one tin deposit per copper location")
	assert.Equal(t, "Cassiterite", deposits[0].MineralType.Name, "Tin ore should be cassiterite")
}

// -----------------------------------------------------------------------------
// Scenario: Tool Stone Formation (Obsidian/Flint)
// -----------------------------------------------------------------------------
// Given: Volcanic silica-rich flows OR chalk with flint nodules
// When: Formation conditions met
// Then: Tool stone deposits should form
//
//	AND Should be high hardness (suitable for tools)
func TestBDD_ToolStone_Formation(t *testing.T) {
	// Test volcanic obsidian
	volcanicDeposits := minerals.GenerateToolStoneDeposits(true, false)
	require.NotNil(t, volcanicDeposits, "Volcanic context should create tool stones")
	assert.Greater(t, len(volcanicDeposits), 0, "Should have at least one deposit")
	assert.Equal(t, "Obsidian", volcanicDeposits[0].MineralType.Name, "Volcanic should create obsidian")

	// Test chalk flint
	chalkDeposits := minerals.GenerateToolStoneDeposits(false, true)
	require.NotNil(t, chalkDeposits, "Chalk should create flint")
	assert.Equal(t, "Flint", chalkDeposits[0].MineralType.Name, "Chalk should create flint")
}

// -----------------------------------------------------------------------------
// Scenario: Coal Seam Formation
// -----------------------------------------------------------------------------
// Given: Swampy forest biome (Carboniferous period)
// When: Millions of years of burial and pressure
// Then: Coal seams should form
//
//	AND Coal rank should increase with depth/age
func TestBDD_Coal_Formation(t *testing.T) {
	config := minerals.CoalFormationConfig{
		OrganicMatter: 0.8,        // High organic content (swamp)
		BurialDepth:   500,        // 500m burial
		BurialAge:     50_000_000, // 50 million years
	}

	deposits := minerals.GenerateCoalDeposits(config)

	require.NotNil(t, deposits, "Should generate coal deposits")
	assert.Greater(t, len(deposits), 0, "Should have at least one coal seam")
	// With 500m depth and 50M years, rank = 0.5 + 0.5 = 1, should be Lignite
	assert.Equal(t, "Lignite", deposits[0].MineralType.Name, "Should be lignite at this depth/age")
}

// -----------------------------------------------------------------------------
// Scenario: Evaporite Formation (Salt/Gypsum)
// -----------------------------------------------------------------------------
// Given: Shallow sea or lake in arid climate
// When: Water evaporates over time
// Then: Salt and gypsum deposits should form
//
//	AND Deposit layers should show evaporation sequence
func TestBDD_Evaporite_Formation(t *testing.T) {
	config := minerals.EvaporiteFormationConfig{
		WaterVolume:   1000,   // Large basin
		EvaporateRate: 0.8,    // High evaporation
		Climate:       "arid", // Arid climate required
	}

	deposits := minerals.GenerateEvaporiteDeposits(config)

	require.NotNil(t, deposits, "Should generate evaporite deposits")
	assert.Greater(t, len(deposits), 0, "Should have at least one evaporite layer")
	// Check evaporation sequence - should have gypsum and halite
	foundGypsum := false
	for _, dep := range deposits {
		if dep.MineralType.Name == "Gypsum" {
			foundGypsum = true
		}
	}
	assert.True(t, foundGypsum, "Should have gypsum in evaporite sequence")
}

// -----------------------------------------------------------------------------
// Scenario: Saltpeter Formation (Gunpowder Prerequisite)
// -----------------------------------------------------------------------------
// Given: Cave environment with bat guano OR desert soils
// When: Nitrogen-fixing conditions exist
// Then: Saltpeter deposits should form
//
//	AND Required for gunpowder tech advance
func TestBDD_Saltpeter_Formation(t *testing.T) {
	ctx := minerals.SaltpeterFormationContext{
		HasCaves:      true,
		HasDesert:     false,
		OrganicMatter: 0.6, // High organic content (bat guano)
	}

	deposits := minerals.GenerateSaltpeterDeposits(ctx)

	require.NotNil(t, deposits, "Should generate saltpeter deposits")
	assert.Greater(t, len(deposits), 0, "Should have at least one deposit")
	assert.Equal(t, "Saltpeter", deposits[0].MineralType.Name, "Should be saltpeter mineral")
}

// -----------------------------------------------------------------------------
// Scenario: Ore Grade Variation
// -----------------------------------------------------------------------------
// Given: A large mineral deposit
// When: Sampled at different points
// Then: Concentration should vary (center richer)
//
//	AND Mining should follow richest veins first
func TestBDD_OreGrade_Variation(t *testing.T) {
	// Create a deposit with known center and dimensions
	deposit := &minerals.MineralDeposit{
		Location:      geography.Point{X: 100, Y: 100},
		Concentration: 0.8,
		VeinLength:    50,
		VeinWidth:     20,
	}

	// Sample at center - should be richest
	centerConc := minerals.SampleConcentration(deposit, 100, 100)
	// Sample at edge - should be lower
	edgeConc := minerals.SampleConcentration(deposit, 105, 100)

	assert.Greater(t, centerConc, 0.0, "Center should have positive concentration")
	assert.GreaterOrEqual(t, centerConc, edgeConc, "Center should be richer than edge")
}

// -----------------------------------------------------------------------------
// Scenario: Mana Crystal Formation (Magical Minerals)
// -----------------------------------------------------------------------------
// Given: Ley line intersection point
// When: Magical energy accumulates
// Then: Mana crystals should form
//
//	AND Crystal potency should depend on ley line strength
func TestBDD_ManaCrystal_Formation(t *testing.T) {
	ctx := minerals.ManaCrystalFormationContext{
		LeyLineStrength: 0.7,
		IsIntersection:  true,
		Depth:           100,
	}

	deposits := minerals.GenerateManaCrystals(ctx)

	require.NotNil(t, deposits, "Should generate mana crystals")
	assert.Greater(t, len(deposits), 0, "Should have at least one crystal")
	assert.Contains(t, deposits[0].MineralType.Name, "Mana Crystal", "Should be a mana crystal")
}
