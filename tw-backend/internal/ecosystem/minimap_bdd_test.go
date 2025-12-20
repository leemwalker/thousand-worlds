package ecosystem_test

import (
	"testing"

	"tw-backend/internal/ecosystem"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// BDD Tests: Minimap Rendering
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Biome Visual Resolution (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various biome types
// When: GetBiomeVisual is called
// Then: Correct visual representation should be returned
func TestBDD_Minimap_BiomeVisual(t *testing.T) {
	scenarios := []struct {
		name          string
		biome         string
		expectedEmoji string
		expectedChar  string
	}{
		{"Ocean", "ocean", "🌊", "~"},
		{"Rainforest", "rainforest", "🌴", "%"},
		{"Grassland", "grassland", "🌾", "\""},
		{"Deciduous", "deciduous", "🌳", "&"},
		{"Alpine", "alpine", "🏔️", "^"},
		{"Taiga", "taiga", "🌲", "*"},
		{"Desert", "desert", "🌵", "."},
		{"Tundra", "tundra", "❄️", "-"},
		{"Unknown (default)", "unknown", "🌾", "\""}, // Defaults to grassland
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			visual := ecosystem.GetBiomeVisual(sc.biome)
			assert.Equal(t, sc.expectedEmoji, visual.Emoji,
				"Biome %s should have emoji %s", sc.biome, sc.expectedEmoji)
			assert.Equal(t, sc.expectedChar, visual.Char,
				"Biome %s should have char %s", sc.biome, sc.expectedChar)
			assert.NotEmpty(t, visual.Color, "Should have color")
			assert.NotEmpty(t, visual.Tailwind, "Should have tailwind class")
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Elevation Visual Resolution (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various elevations
// When: GetElevationVisual is called
// Then: Correct tier should be returned based on elevation
func TestBDD_Minimap_ElevationVisual(t *testing.T) {
	scenarios := []struct {
		name         string
		elevation    float64
		expectedName string
	}{
		{"Deep Ocean (-2000)", -2000, "deep_ocean"},
		{"Deep Ocean Edge (-1000)", -1000, "deep_ocean"},
		{"Shallow Water (-500)", -500, "shallow_water"},
		{"Sea Level (0)", 0, "shallow_water"},
		{"Lowland (200)", 200, "lowland"},
		{"Highland (1500)", 1500, "highland"},
		{"Peak (5000)", 5000, "peak"},
		{"Extreme Peak (9000)", 9000, "peak"},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			visual := ecosystem.GetElevationVisual(sc.elevation)
			assert.Equal(t, sc.expectedName, visual.Name,
				"Elevation %.0f should be %s tier", sc.elevation, sc.expectedName)
			assert.NotEmpty(t, visual.Color, "Should have color")
			assert.NotEmpty(t, visual.Tailwind, "Should have tailwind class")
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Catastrophe Visual Overlay (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various catastrophe types
// When: GetCatastropheVisual is called
// Then: Correct overlay or nil should be returned
func TestBDD_Minimap_CatastropheVisual(t *testing.T) {
	scenarios := []struct {
		name          string
		catastrophe   string
		expectedEmoji string
		expectNil     bool
	}{
		{"Volcano", "volcano", "🌋", false},
		{"Asteroid", "asteroid", "☄️", false},
		{"Flood Basalt", "flood_basalt", "♨️", false},
		{"Anoxia", "anoxia", "🦠", false},
		{"Ice Age", "ice_age", "🧊", false},
		{"Unknown", "unknown_type", "", true},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			visual := ecosystem.GetCatastropheVisual(sc.catastrophe)
			if sc.expectNil {
				assert.Nil(t, visual, "Unknown catastrophe should return nil")
			} else {
				assert.NotNil(t, visual, "Known catastrophe should return visual")
				assert.Equal(t, sc.expectedEmoji, visual.Emoji)
				assert.NotEmpty(t, visual.Tailwind, "Should have animation class")
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Minimap Cell Creation
// -----------------------------------------------------------------------------
// Given: Biome, elevation, and catastrophe parameters
// When: NewMinimapCell is called
// Then: All visual fields should be populated correctly
func TestBDD_Minimap_CellCreation(t *testing.T) {
	cell := ecosystem.NewMinimapCell(5, 10, "rainforest", 200, "volcano")

	// Position
	assert.Equal(t, 5, cell.Q)
	assert.Equal(t, 10, cell.R)

	// Biome visuals
	assert.Equal(t, "rainforest", cell.BiomeType)
	assert.Equal(t, "🌴", cell.BiomeEmoji)
	assert.Equal(t, "%", cell.BiomeChar)
	assert.NotEmpty(t, cell.BiomeColor)
	assert.NotEmpty(t, cell.BiomeTailwind)

	// Elevation visuals
	assert.Equal(t, 200.0, cell.Elevation)
	assert.Equal(t, "lowland", cell.ElevName)
	assert.NotEmpty(t, cell.ElevColor)
	assert.NotEmpty(t, cell.ElevTailwind)

	// Catastrophe overlay
	assert.Equal(t, "volcano", cell.Catastrophe)
	assert.Equal(t, "🌋", cell.CatastropheEmoji)
	assert.Equal(t, "A", cell.CatastropheChar)
	assert.NotEmpty(t, cell.CatastropheColor)
	assert.Contains(t, cell.CatastropheTailwind, "animate")
}

// -----------------------------------------------------------------------------
// Scenario: Minimap Cell Without Catastrophe
// -----------------------------------------------------------------------------
// Given: Normal cell without catastrophe
// When: NewMinimapCell is called with empty catastrophe
// Then: Catastrophe fields should be empty
func TestBDD_Minimap_CellNoCatastrophe(t *testing.T) {
	cell := ecosystem.NewMinimapCell(0, 0, "ocean", -500, "")

	assert.Equal(t, "ocean", cell.BiomeType)
	assert.Equal(t, "🌊", cell.BiomeEmoji)
	assert.Equal(t, "", cell.Catastrophe)
	assert.Equal(t, "", cell.CatastropheEmoji)
}

// -----------------------------------------------------------------------------
// Scenario: Visual Priority - Catastrophe Overrides Biome
// -----------------------------------------------------------------------------
// Given: Cell with both biome and catastrophe
// When: Visual is resolved
// Then: Catastrophe should have separate overlay (both visible)
func TestBDD_Minimap_VisualPriority(t *testing.T) {
	// Volcano on grassland
	cell := ecosystem.NewMinimapCell(0, 0, "grassland", 500, "volcano")

	// Both biome and catastrophe visuals should be present
	assert.Equal(t, "🌾", cell.BiomeEmoji, "Biome emoji preserved")
	assert.Equal(t, "🌋", cell.CatastropheEmoji, "Catastrophe overlay present")

	// Frontend would prioritize CatastropheEmoji when rendering
}

// -----------------------------------------------------------------------------
// Scenario: MinimapBatch Structure
// -----------------------------------------------------------------------------
// Given: Batch of minimap cells
// When: MinimapBatch is created
// Then: Should contain world ID, year, and cells
func TestBDD_Minimap_BatchStructure(t *testing.T) {
	cells := []ecosystem.MinimapCell{
		ecosystem.NewMinimapCell(0, 0, "ocean", -100, ""),
		ecosystem.NewMinimapCell(1, 0, "grassland", 100, ""),
	}

	batch := ecosystem.MinimapBatch{
		Year:  1000000,
		Cells: cells,
	}

	assert.Equal(t, int64(1000000), batch.Year)
	assert.Len(t, batch.Cells, 2)
	assert.Equal(t, "ocean", batch.Cells[0].BiomeType)
	assert.Equal(t, "grassland", batch.Cells[1].BiomeType)
}

// -----------------------------------------------------------------------------
// Scenario: Line of Sight - Horizon Culling
// -----------------------------------------------------------------------------
// Given: Player at elevation 100m
// When: Tile is behind higher elevation
// Then: Tile should be marked as occluded
//
//	AND Occluded tiles should render darker
func TestBDD_Minimap_LineOfSight_Occlusion(t *testing.T) {
	assert.Fail(t, "BDD RED: Horizon culling is in gamemap.Service.computeOcclusion - see service_test.go")
	// Pseudocode:
	// grid := [][]*MapTile{{playerTile, hillTile, behindHillTile}}
	// computeOcclusion(grid, radius, playerAlt: 100, stride: 1)
	// assert behindHillTile.Occluded == true
}

// -----------------------------------------------------------------------------
// Scenario: Perception-Based Rendering Quality
// -----------------------------------------------------------------------------
// Given: Player with high perception (0.9)
// When: Minimap is rendered
// Then: Should see emoji overlays
//
//	AND Details should be visible at greater distance
func TestBDD_Minimap_PerceptionRendering_High(t *testing.T) {
	assert.Fail(t, "BDD RED: Perception-based rendering in gamemap.GetRenderQuality - see worldmap tests")
	// Pseudocode:
	// player := Character{Perception: 0.9}
	// rendering := getRenderingMode(player)
	// assert rendering.UseEmoji == true
	// assert rendering.VisibleRadius > 5
}

// -----------------------------------------------------------------------------
// Scenario: Player Movement & Grid Shifting
// -----------------------------------------------------------------------------
// Given: A player at [10, 10] with a visible landmark at [10, 11] (North)
// When: Player moves North to [10, 11]
// Then: The landmark should shift from the "North" cell to the "Center" cell
func TestBDD_Minimap_MovementShift(t *testing.T) {
	assert.Fail(t, "BDD RED: Grid centering requires full service integration test")
	// Pseudocode:
	// world := MockWorldWithLandmarkAt(10, 11, "Tower")
	// player := &Character{X: 10, Y: 10}

	// Step 1: Initial Look
	// grid1 := service.GenerateMinimap(player, radius: 1)
	// assert grid1.GetCell(0, 1).Content == "Tower" // 0,1 is North relative to center

	// Step 2: Move
	// player.Y = 11
	// grid2 := service.GenerateMinimap(player, radius: 1)
	// assert grid2.GetCell(0, 0).Content == "Tower" // 0,0 is now the Center
}

// -----------------------------------------------------------------------------
// Scenario: Spherical Seam Wrapping
// -----------------------------------------------------------------------------
// Given: A world with width 100, and player at East edge [99, 50]
// When: Minimap is generated with radius 2
// Then: The Eastern-most cells should map to x=0 and x=1
func TestBDD_Minimap_SphericalWrapping(t *testing.T) {
	assert.Fail(t, "BDD RED: Spherical wrapping in spatial service - see spatial_service_test.go")
	// Pseudocode:
	// player := &Character{X: 99, Y: 50}
	// grid := service.GenerateMinimap(player, radius: 2)

	// // Expected grid X coords: [97, 98, 99, 0, 1]
	// eastCell := grid.GetRelativeCell(2, 0) // 2 units East
	// assert eastCell.WorldX == 1
}

// -----------------------------------------------------------------------------
// Scenario: Dynamic Perception Change
// -----------------------------------------------------------------------------
// Given: A player with High perception rendering Emoji map
// When: Player receives "Blinded" status effect (reducing perception to 0)
// Then: The *next* render request should downgrade to ASCII or "Fog"
func TestBDD_Minimap_StatusEffectImpact(t *testing.T) {
	assert.Fail(t, "BDD RED: Integration with status effects not yet implemented")
	// Pseudocode:
	// player.Perception = 10
	// view1 := service.Render(player)
	// assert view1.Mode == "HighRes"

	// player.AddEffect("blind")
	// view2 := service.Render(player)
	// assert view2.Mode == "ASCII" || view2.Mode == "Obscured"
}

// -----------------------------------------------------------------------------
// Scenario: BiomeMap Completeness
// -----------------------------------------------------------------------------
// Given: The BiomeMap constant
// When: All expected biomes are checked
// Then: All should have valid visual data
func TestBDD_Minimap_BiomeMapComplete(t *testing.T) {
	expectedBiomes := []string{
		"ocean", "rainforest", "grassland", "deciduous",
		"alpine", "taiga", "desert", "tundra",
	}

	for _, biome := range expectedBiomes {
		t.Run(biome, func(t *testing.T) {
			visual, ok := ecosystem.BiomeMap[biome]
			assert.True(t, ok, "BiomeMap should contain %s", biome)
			assert.NotEmpty(t, visual.Emoji, "%s should have emoji", biome)
			assert.NotEmpty(t, visual.Char, "%s should have char", biome)
			assert.NotEmpty(t, visual.Color, "%s should have color", biome)
			assert.NotEmpty(t, visual.Tailwind, "%s should have tailwind", biome)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: CatastropheMap Completeness
// -----------------------------------------------------------------------------
// Given: The CatastropheMap constant
// When: All expected catastrophes are checked
// Then: All should have visual data, animated ones should have animation classes
func TestBDD_Minimap_CatastropheMapComplete(t *testing.T) {
	expectedCatastrophes := []struct {
		name     string
		animated bool
	}{
		{"volcano", true},
		{"asteroid", true},
		{"flood_basalt", false},
		{"anoxia", false},
		{"ice_age", false},
	}

	for _, cat := range expectedCatastrophes {
		t.Run(cat.name, func(t *testing.T) {
			visual, ok := ecosystem.CatastropheMap[cat.name]
			assert.True(t, ok, "CatastropheMap should contain %s", cat.name)
			assert.NotEmpty(t, visual.Emoji, "%s should have emoji", cat.name)
			assert.NotEmpty(t, visual.Tailwind, "%s should have tailwind class", cat.name)
			if cat.animated {
				assert.Contains(t, visual.Tailwind, "animate",
					"%s should have animation class", cat.name)
			}
		})
	}
}
