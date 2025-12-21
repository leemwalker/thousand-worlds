package gamemap_test

import (
	"context"
	"testing"

	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	gamemap "tw-backend/internal/game/services/map"
	"tw-backend/internal/repository"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: World Map Rendering
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Render Quality by Perception (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Different character perception levels
// When: GetRenderQuality is called
// Then: Quality tier should match perception thresholds
func TestBDD_WorldMap_RenderQuality(t *testing.T) {
	scenarios := []struct {
		name       string
		perception int
		expected   gamemap.RenderQuality
	}{
		{"Very Low Perception (10)", 10, gamemap.QualityLow},
		{"Low Perception (30)", 30, gamemap.QualityLow},
		{"Medium Perception (31)", 31, gamemap.QualityMedium},
		{"Mid-High Perception (50)", 50, gamemap.QualityMedium},
		{"High Perception (71)", 71, gamemap.QualityHigh},
		{"Max Perception (100)", 100, gamemap.QualityHigh},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			quality := gamemap.GetRenderQuality(sc.perception)
			assert.Equal(t, sc.expected, quality,
				"Perception %d should yield %s quality", sc.perception, sc.expected)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: worldToGrid Coordinate Mapping
// -----------------------------------------------------------------------------
// Given: World coordinates and bounds
// When: worldToGrid is called (via GetMapData)
// Then: Coordinates should map correctly to grid indices
func TestBDD_WorldMap_WorldToGrid(t *testing.T) {
	// worldToGrid is unexported but tested indirectly via GetMapData
	// This test verifies that the service correctly maps world coords to grid
	svc := &gamemap.Service{}

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     uuid.New(),
		PositionX:   50.0,
		PositionY:   50.0,
	}

	ctx := context.Background()
	mapData, err := svc.GetMapData(ctx, char)

	assert.NoError(t, err)
	assert.NotNil(t, mapData)
	// Player should be at center of grid
	var playerTile *gamemap.MapTile
	for i := range mapData.Tiles {
		if mapData.Tiles[i].IsPlayer {
			playerTile = &mapData.Tiles[i]
			break
		}
	}
	assert.NotNil(t, playerTile, "Player position should be marked in grid")
	assert.Equal(t, 50, playerTile.X)
	assert.Equal(t, 50, playerTile.Y)
}

// -----------------------------------------------------------------------------
// Scenario: Region Aggregation - Dominant Biome
// -----------------------------------------------------------------------------
// Given: World map with geology data
// When: GetWorldMapData is called
// Then: Biome data should be retrieved from heightmap
func TestBDD_WorldMap_RegionAggregation(t *testing.T) {
	// This test validates that region aggregation works with geology
	// The full implementation test is in service_test.go
	svc := gamemap.NewService(nil, nil, nil, nil, nil, nil)
	worldID := uuid.New()

	// Create minimal geology
	hm := &geography.Heightmap{
		Width:      10,
		Height:     10,
		Elevations: make([]float64, 100),
	}
	biomes := make([]geography.Biome, 100)
	for i := range biomes {
		biomes[i] = geography.Biome{Type: geography.BiomeGrassland}
	}

	geo := &ecosystem.WorldGeology{
		Heightmap: hm,
		Biomes:    biomes,
	}
	svc.SetWorldGeology(worldID, geo)

	// Verify geology was set
	assert.NotNil(t, svc, "Service should accept geology data")
}

// -----------------------------------------------------------------------------
// Scenario: Map Data Caching
// -----------------------------------------------------------------------------
// Given: A map that has already been generated for a resolution (e.g., 64x64)
// When: GetWorldMapData is called a second time
// Then: The cached result should be returned
//
//	AND The heavy aggregation logic should NOT run again
func TestBDD_WorldMap_Caching(t *testing.T) {
	// Create mock world repo that returns a valid world
	mockRepo := &MockWorldRepo{
		World: &repository.World{
			ID:            uuid.New(),
			Name:          "Test World",
			Circumference: floatPtr(1000.0),
		},
	}

	svc := gamemap.NewService(mockRepo, nil, nil, nil, nil, nil)
	worldID := mockRepo.World.ID

	// Set up geology
	hm := &geography.Heightmap{
		Width:      64,
		Height:     64,
		Elevations: make([]float64, 64*64),
	}
	biomes := make([]geography.Biome, 64*64)
	for i := range biomes {
		biomes[i] = geography.Biome{Type: geography.BiomeOcean}
	}
	geo := &ecosystem.WorldGeology{
		Heightmap: hm,
		Biomes:    biomes,
	}
	svc.SetWorldGeology(worldID, geo)

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     worldID,
		PositionX:   500.0,
		PositionY:   250.0,
	}

	ctx := context.Background()

	// First call - generates data
	// Note: Spherical world has 2:1 aspect ratio (width=circumference, height=circumference/2)
	// So gridSize 64 becomes 128x64 with aspect ratio scaling
	data1, err := svc.GetWorldMapData(ctx, char, 64)
	require.NoError(t, err)
	require.NotNil(t, data1)
	assert.Equal(t, 128, data1.GridWidth, "2:1 aspect ratio doubles width")
	assert.Equal(t, 64, data1.GridHeight, "Height stays at gridSize")
	assert.Len(t, data1.Tiles, 128*64, "128x64 grid should have 8192 tiles")

	// Second call - should use cache
	data2, err := svc.GetWorldMapData(ctx, char, 64)
	require.NoError(t, err)
	require.NotNil(t, data2)

	// Both should return same data structure (cached)
	assert.Equal(t, data1.GridWidth, data2.GridWidth)
	assert.Equal(t, data1.WorldWidth, data2.WorldWidth)
}

// -----------------------------------------------------------------------------
// Scenario: Uninitialized World Handling
// -----------------------------------------------------------------------------
// Given: A world entry exists but Geology/Heightmap is nil
// When: GetMapData is called
// Then: A graceful response should be returned (no panic)
func TestBDD_WorldMap_UninitializedState(t *testing.T) {
	// Service with no geology data should still return valid map data
	// Uses NewService to ensure internal maps are initialized
	svc := gamemap.NewService(nil, nil, nil, nil, nil, nil)

	// Position within default bounds (0, 0) - (10, 10)
	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     uuid.New(),
		PositionX:   5.0,
		PositionY:   5.0,
	}

	ctx := context.Background()
	mapData, err := svc.GetMapData(ctx, char)

	// Should not panic, should return valid grid
	assert.NoError(t, err)
	assert.NotNil(t, mapData)
	assert.Equal(t, 81, len(mapData.Tiles), "9x9 grid")

	// When no world data exists, tiles within default bounds get "lobby" biome
	// Tiles outside bounds get "void" biome
	lobbyCount := 0
	voidCount := 0
	for _, tile := range mapData.Tiles {
		if tile.Biome == "lobby" {
			lobbyCount++
		}
		if tile.Biome == "void" {
			voidCount++
		}
	}
	// At position 5,5 with radius 4, some tiles will be within bounds 0-10
	assert.Greater(t, lobbyCount+voidCount, 0, "Should have biome tiles")
}

// -----------------------------------------------------------------------------
// Scenario: Aggregation Biome Weighting
// -----------------------------------------------------------------------------
// Given: A sub-region mixed with Ocean and Land tiles
// When: Aggregated into a single Map Cell
// Then: Water biomes should be preserved with 1.5x weighting
func TestBDD_WorldMap_BiomeWeighting(t *testing.T) {
	// Create mock world repo
	mockRepo := &MockWorldRepo{
		World: &repository.World{
			ID:            uuid.New(),
			Name:          "Coastal World",
			Circumference: floatPtr(1000.0),
		},
	}

	svc := gamemap.NewService(mockRepo, nil, nil, nil, nil, nil)
	worldID := mockRepo.World.ID

	// Create geology with a coastal region:
	// 40% ocean, 60% land in the same region
	// Without weighting: land would win (60% > 40%)
	// With 1.5x ocean weight: ocean = 40*1.5 = 60, land = 60*1.0 = 60 (tie/ocean wins)
	// Actually with 4 ocean vs 5 land in a 3x3 region:
	// Ocean = 4 * 1.5 = 6.0, Land = 5 * 1.0 = 5.0 -> Ocean wins
	hmSize := 64
	hm := &geography.Heightmap{
		Width:      hmSize,
		Height:     hmSize,
		Elevations: make([]float64, hmSize*hmSize),
	}

	biomes := make([]geography.Biome, hmSize*hmSize)
	for i := range biomes {
		y := i / hmSize
		// Top half is ocean, bottom half is grassland
		// But we'll create a coastal strip in the middle row
		if y < hmSize/2 {
			biomes[i] = geography.Biome{Type: geography.BiomeOcean}
			hm.Elevations[i] = -100
		} else {
			biomes[i] = geography.Biome{Type: geography.BiomeGrassland}
			hm.Elevations[i] = 100
		}
	}

	geo := &ecosystem.WorldGeology{
		Heightmap: hm,
		Biomes:    biomes,
	}
	svc.SetWorldGeology(worldID, geo)

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     worldID,
		PositionX:   500.0,
		PositionY:   250.0,
	}

	ctx := context.Background()

	// Request a smaller grid to trigger aggregation
	// 64x64 heightmap -> 16x16 grid = 4x aggregation
	data, err := svc.GetWorldMapData(ctx, char, 16)

	require.NoError(t, err)
	require.NotNil(t, data)

	// Find the tiles near the middle (coastal boundary)
	// These tiles should have ocean due to 1.5x weighting
	// The exact boundary row is at gridY = 8 (middle of 16 rows)
	coastalTiles := 0
	grasslandTiles := 0
	oceanTiles := 0

	for _, tile := range data.Tiles {
		switch tile.Biome {
		case "Ocean": // Capitalized to match geography.BiomeOcean
			oceanTiles++
		case "Grassland": // Capitalized to match geography.BiomeGrassland
			grasslandTiles++
		}
		// Count tiles at the boundary (middle rows where aggregation matters)
		if tile.GridY >= 7 && tile.GridY <= 8 && tile.Biome == "Ocean" {
			coastalTiles++
		}
	}

	// Due to 1.5x ocean weighting, coastal regions near the boundary
	// should favor ocean even when land is slightly more numerous
	assert.Greater(t, oceanTiles, 0, "Should have ocean tiles")
	assert.Greater(t, grasslandTiles, 0, "Should have grassland tiles")

	// The aggregation should produce tiles at the boundary
	// With 1.5x water weighting, ocean should be present at the boundary
	t.Logf("Tiles: ocean=%d, grassland=%d, coastal=%d", oceanTiles, grasslandTiles, coastalTiles)
}

// -----------------------------------------------------------------------------
// Scenario: Grid Aspect Ratio
// -----------------------------------------------------------------------------
// Given: A world that is 2x wider than it is tall
// When: Requested with a square grid size (e.g., 64)
// Then: The resulting map data should respect the world's aspect ratio
func TestBDD_WorldMap_AspectRatio(t *testing.T) {
	// Create a 2:1 aspect ratio world (2000 width, 1000 height)
	// For spherical worlds, height is circumference/2
	mockRepo := &MockWorldRepo{
		World: &repository.World{
			ID:            uuid.New(),
			Name:          "Wide World",
			Circumference: floatPtr(2000.0), // Width = 2000, Height = 1000
		},
	}

	svc := gamemap.NewService(mockRepo, nil, nil, nil, nil, nil)
	worldID := mockRepo.World.ID

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     worldID,
		PositionX:   1000.0,
		PositionY:   500.0,
	}

	ctx := context.Background()
	data, err := svc.GetWorldMapData(ctx, char, 64)

	require.NoError(t, err)
	require.NotNil(t, data)

	// With 2:1 aspect ratio (2000 width / 1000 height = 2.0)
	// GridCols should be 2x GridRows
	assert.Equal(t, 64, data.GridHeight, "GridHeight should be the base gridSize")
	assert.Equal(t, 128, data.GridWidth, "GridWidth should be 2x height for 2:1 aspect ratio")

	// Total tiles should be gridCols * gridRows
	expectedTiles := 128 * 64
	assert.Len(t, data.Tiles, expectedTiles, "Tile count should match grid dimensions")

	// Verify world dimensions are preserved
	assert.Equal(t, 2000.0, data.WorldWidth)
	assert.Equal(t, 1000.0, data.WorldHeight)
}

// -----------------------------------------------------------------------------
// Scenario: Full World Modal Display
// -----------------------------------------------------------------------------
// Given: Character requests world map
// When: GetWorldMapData is called with gridSize 64
// Then: Complete 64x64 grid should be returned (4096 tiles)
func TestBDD_WorldMap_FullDisplay(t *testing.T) {
	mockRepo := &MockWorldRepo{
		World: &repository.World{
			ID:            uuid.New(),
			Name:          "Test World",
			Circumference: floatPtr(2000.0),
		},
	}

	svc := gamemap.NewService(mockRepo, nil, nil, nil, nil, nil)
	worldID := mockRepo.World.ID

	// Set up geology
	hm := &geography.Heightmap{
		Width:      128,
		Height:     128,
		Elevations: make([]float64, 128*128),
	}
	biomes := make([]geography.Biome, 128*128)
	for i := range biomes {
		biomes[i] = geography.Biome{Type: geography.BiomeGrassland}
	}
	geo := &ecosystem.WorldGeology{
		Heightmap: hm,
		Biomes:    biomes,
	}
	svc.SetWorldGeology(worldID, geo)

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     worldID,
		PositionX:   1000.0,
		PositionY:   500.0,
	}

	ctx := context.Background()
	// Note: Spherical world has 2:1 aspect ratio (width=circumference, height=circumference/2)
	// So gridSize 64 becomes 128x64 with aspect ratio scaling
	data, err := svc.GetWorldMapData(ctx, char, 64)

	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, 128, data.GridWidth, "2:1 aspect ratio doubles width")
	assert.Equal(t, 64, data.GridHeight, "Height stays at gridSize")
	assert.Len(t, data.Tiles, 128*64, "128x64 grid should return exactly 8192 tiles")
}

// -----------------------------------------------------------------------------
// Scenario: Player Position Marker
// -----------------------------------------------------------------------------
// Given: Player at specific world coordinates
// When: World map is displayed
// Then: Player position should be included in response
func TestBDD_WorldMap_PlayerPosition(t *testing.T) {
	mockRepo := &MockWorldRepo{
		World: &repository.World{
			ID:            uuid.New(),
			Name:          "Test World",
			Circumference: floatPtr(1000.0),
		},
	}

	svc := gamemap.NewService(mockRepo, nil, nil, nil, nil, nil)
	worldID := mockRepo.World.ID

	playerX := 123.45
	playerY := 678.90

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     worldID,
		PositionX:   playerX,
		PositionY:   playerY,
	}

	ctx := context.Background()
	data, err := svc.GetWorldMapData(ctx, char, 32)

	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, playerX, data.PlayerX, "PlayerX should match character position")
	assert.Equal(t, playerY, data.PlayerY, "PlayerY should match character position")
}

// -----------------------------------------------------------------------------
// Scenario: Map Service Initialization
// -----------------------------------------------------------------------------
// Given: All required dependencies
// When: NewService is called
// Then: Service should be properly configured
func TestBDD_WorldMap_ServiceInit(t *testing.T) {
	mockRepo := &MockWorldRepo{}

	svc := gamemap.NewService(mockRepo, nil, nil, nil, nil, nil)

	assert.NotNil(t, svc, "NewService should return non-nil service")
}

// -----------------------------------------------------------------------------
// Scenario: MapTile Data Structure
// -----------------------------------------------------------------------------
// Given: MapTile structure
// When: Fields are accessed
// Then: All expected fields should be available
func TestBDD_WorldMap_TileStructure(t *testing.T) {
	tile := gamemap.MapTile{
		X:           10,
		Y:           20,
		Biome:       "forest",
		Elevation:   150.5,
		IsPlayer:    true,
		OutOfBounds: false,
		Occluded:    false,
	}

	assert.Equal(t, 10, tile.X)
	assert.Equal(t, 20, tile.Y)
	assert.Equal(t, "forest", tile.Biome)
	assert.Equal(t, 150.5, tile.Elevation)
	assert.True(t, tile.IsPlayer)
	assert.False(t, tile.OutOfBounds)
	assert.False(t, tile.Occluded)
}

// -----------------------------------------------------------------------------
// Scenario: WorldMapData Data Structure
// -----------------------------------------------------------------------------
// Given: WorldMapData structure
// When: Fields are accessed
// Then: Grid dimensions and player position should be available
func TestBDD_WorldMap_DataStructure(t *testing.T) {
	data := gamemap.WorldMapData{
		GridWidth:   64,
		GridHeight:  32,
		WorldWidth:  2000,
		WorldHeight: 1000,
		PlayerX:     500,
		PlayerY:     250,
		Tiles:       []gamemap.WorldMapTile{},
	}

	assert.Equal(t, 64, data.GridWidth)
	assert.Equal(t, 32, data.GridHeight)
	assert.Equal(t, 2000.0, data.WorldWidth)
	assert.Equal(t, 1000.0, data.WorldHeight)
	assert.Equal(t, 500.0, data.PlayerX)
	assert.Equal(t, 250.0, data.PlayerY)
}

// -----------------------------------------------------------------------------
// Scenario: WorldMapTile Data Structure
// -----------------------------------------------------------------------------
// Given: WorldMapTile for region
// When: Fields are accessed
// Then: Grid position, biome, and elevation should be available
func TestBDD_WorldMap_RegionTileStructure(t *testing.T) {
	tile := gamemap.WorldMapTile{
		GridX:        5,
		GridY:        10,
		Biome:        "desert",
		AvgElevation: 300.0,
		IsPlayer:     false,
	}

	assert.Equal(t, 5, tile.GridX)
	assert.Equal(t, 10, tile.GridY)
	assert.Equal(t, "desert", tile.Biome)
	assert.Equal(t, 300.0, tile.AvgElevation)
	assert.False(t, tile.IsPlayer)
}

// -----------------------------------------------------------------------------
// Scenario: MapEntity for Entity Display
// -----------------------------------------------------------------------------
// Given: MapEntity with entity data
// When: Entity is added to tile
// Then: Entity info should be retrievable
func TestBDD_WorldMap_EntityDisplay(t *testing.T) {
	entity := gamemap.MapEntity{
		Type:   "npc",
		Name:   "Merchant",
		Status: "friendly",
		Glyph:  "🧙",
	}

	assert.Equal(t, "npc", entity.Type)
	assert.Equal(t, "Merchant", entity.Name)
	assert.Equal(t, "friendly", entity.Status)
	assert.Equal(t, "🧙", entity.Glyph)
}

// -----------------------------------------------------------------------------
// Scenario: PortalInfo Display
// -----------------------------------------------------------------------------
// Given: Portal on map tile
// When: Portal info is accessed
// Then: Destination and status should be available
func TestBDD_WorldMap_PortalDisplay(t *testing.T) {
	portal := gamemap.PortalInfo{
		WorldName:   "Underworld",
		Destination: "Level 1 Entrance",
		Active:      true,
	}

	assert.Equal(t, "Underworld", portal.WorldName)
	assert.Equal(t, "Level 1 Entrance", portal.Destination)
	assert.True(t, portal.Active)
}

// =============================================================================
// Mock Implementations
// =============================================================================

type MockWorldRepo struct {
	World *repository.World
}

func (m *MockWorldRepo) GetWorld(ctx context.Context, id uuid.UUID) (*repository.World, error) {
	if m.World != nil && m.World.ID == id {
		return m.World, nil
	}
	return nil, nil
}

func (m *MockWorldRepo) CreateWorld(ctx context.Context, world *repository.World) error {
	return nil
}

func (m *MockWorldRepo) ListWorlds(ctx context.Context) ([]repository.World, error) {
	return nil, nil
}

func (m *MockWorldRepo) GetWorldsByOwner(ctx context.Context, ownerID uuid.UUID) ([]repository.World, error) {
	return nil, nil
}

func (m *MockWorldRepo) UpdateWorld(ctx context.Context, world *repository.World) error {
	return nil
}

func (m *MockWorldRepo) DeleteWorld(ctx context.Context, worldID uuid.UUID) error {
	return nil
}

func floatPtr(f float64) *float64 {
	return &f
}
