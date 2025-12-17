package gamemap

import (
	"context"
	"testing"

	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetRenderQuality(t *testing.T) {
	tests := []struct {
		name       string
		perception int
		expected   RenderQuality
	}{
		{"Very low perception", 0, QualityLow},
		{"Low perception", 20, QualityLow},
		{"Low threshold", 30, QualityLow},
		{"Medium threshold", 31, QualityMedium},
		{"Medium perception", 50, QualityMedium},
		{"Upper medium", 70, QualityMedium},
		{"High threshold", 71, QualityHigh},
		{"High perception", 90, QualityHigh},
		{"Max perception", 100, QualityHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRenderQuality(tt.perception)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceGetMapData_NoWorldData(t *testing.T) {
	// Create service with minimal dependencies (nil for optional ones)
	svc := &Service{}

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     uuid.New(),
		PositionX:   50.5,
		PositionY:   50.5,
	}

	ctx := context.Background()
	mapData, err := svc.GetMapData(ctx, char)

	assert.NoError(t, err)
	assert.NotNil(t, mapData)

	// Should have 81 tiles (9x9 grid)
	assert.Equal(t, 81, len(mapData.Tiles))
	assert.Equal(t, 9, mapData.GridSize)

	// Default quality is high (perception defaults to 100 for lobby users)
	assert.Equal(t, QualityHigh, mapData.RenderQuality)

	// Player position should be stored
	assert.Equal(t, 50.5, mapData.PlayerX)
	assert.Equal(t, 50.5, mapData.PlayerY)
}

func TestServiceGetMapData_GridCentering(t *testing.T) {
	svc := &Service{}

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     uuid.New(),
		PositionX:   100.0,
		PositionY:   200.0,
	}

	ctx := context.Background()
	mapData, err := svc.GetMapData(ctx, char)

	assert.NoError(t, err)

	// Find player tile (center of grid)
	var playerTile *MapTile
	for i := range mapData.Tiles {
		if mapData.Tiles[i].IsPlayer {
			playerTile = &mapData.Tiles[i]
			break
		}
	}

	assert.NotNil(t, playerTile, "Player tile should be marked")
	assert.Equal(t, 100, playerTile.X)
	assert.Equal(t, 200, playerTile.Y)

	// Check bounds - should span from 96 to 104 (100-4 to 100+4 with radius 4)
	minX, maxX := 1000, -1000
	minY, maxY := 1000, -1000
	for _, tile := range mapData.Tiles {
		if tile.X < minX {
			minX = tile.X
		}
		if tile.X > maxX {
			maxX = tile.X
		}
		if tile.Y < minY {
			minY = tile.Y
		}
		if tile.Y > maxY {
			maxY = tile.Y
		}
	}

	assert.Equal(t, 96, minX)
	assert.Equal(t, 104, maxX)
	assert.Equal(t, 196, minY)
	assert.Equal(t, 204, maxY)
}

func TestMapGridConstants(t *testing.T) {
	assert.Equal(t, 4, MapGridRadius)
	assert.Equal(t, 9, MapGridSize)
}

func TestServiceGetMapData_FlyingScale(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name          string
		altitude      float64
		isFlying      bool
		expectedScale int
		expectedGrid  int
	}{
		// Ground level: radius 4 = 9x9 grid (odd for centering)
		{"On ground", 0, false, 1, 9},
		// Flying: radius = 4 + floor(altitude/5), grid = radius*2 + 1
		{"Flying at 1m", 1, true, 1, 9},      // 4 + 0 = 4, grid = 9
		{"Flying at 5m", 5, true, 1, 11},     // 4 + 1 = 5, grid = 11
		{"Flying at 25m", 25, true, 1, 19},   // 4 + 5 = 9, grid = 19
		{"Flying at 50m", 50, true, 1, 29},   // 4 + 10 = 14, grid = 29
		{"Flying at 100m", 100, true, 1, 49}, // 4 + 20 = 24, grid = 49. Stride = 100/100 = 1
		{"Flying at 105m", 105, true, 1, 51}, // 4 + 21 = 25, grid = 51 (max). Stride = 105/100 = 1
		{"Flying at 200m", 200, true, 2, 51}, // capped at 51. Stride = 200/100 = 2
		{"Flying at 500m", 500, true, 5, 51}, // capped at 51. Stride = 500/100 = 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := &auth.Character{
				CharacterID: uuid.New(),
				WorldID:     uuid.New(),
				PositionX:   50.0,
				PositionY:   50.0,
				PositionZ:   tt.altitude,
				IsFlying:    tt.isFlying,
			}

			ctx := context.Background()
			mapData, err := svc.GetMapData(ctx, char)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedScale, mapData.Scale, "Scale should match stride")
			assert.Equal(t, tt.expectedGrid, mapData.GridSize, "Grid size should match expected")
		})
	}
}

func TestServiceGetMapData_Aggregation(t *testing.T) {
	svc := &Service{}
	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     uuid.New(),
		PositionX:   1000.0,
		PositionY:   1000.0,
		PositionZ:   500.0, // Should result in Stride 5
		IsFlying:    true,
	}

	ctx := context.Background()
	mapData, err := svc.GetMapData(ctx, char)

	assert.NoError(t, err)
	assert.Equal(t, 5, mapData.Scale)
	assert.Equal(t, 51, mapData.GridSize)

	// Check bounds
	// Center 1000. Radius 25 (tiles). Stride 5.
	// Extent = 25 * 5 = 125m.
	// MinX = 1000 - 125 = 875.
	// MaxX = 1000 + 125 = 1125.

	minX, maxX := 2000, 0
	for _, tile := range mapData.Tiles {
		if tile.X < minX {
			minX = tile.X
		}
		if tile.X > maxX {
			maxX = tile.X
		}
	}

	assert.Equal(t, 875, minX)
	assert.Equal(t, 1125, maxX)

	// Check spacing between adjacent tiles in list
	// This is approximate as list order isn't strictly guaranteed coordinate-wise but usually row-major
	// Let's just find two tiles in the same row
	var t1, t2 *MapTile
	for i := range mapData.Tiles {
		if mapData.Tiles[i].Y == 1000 {
			if mapData.Tiles[i].X == 1000 {
				t1 = &mapData.Tiles[i]
			}
			if mapData.Tiles[i].X == 1005 {
				t2 = &mapData.Tiles[i]
			}
		}
	}
	assert.NotNil(t, t1, "Should have center tile")
	assert.NotNil(t, t2, "Should have tile at center + 5")
	assert.NotNil(t, t1, "Should have center tile")
	assert.NotNil(t, t2, "Should have tile at center + 5")
	// If t2 exists, it implies stride works (since normal step is 1)
}

func TestServiceGetMapData_Occlusion(t *testing.T) {
	svc := &Service{
		worldGeology: make(map[uuid.UUID]*ecosystem.WorldGeology),
	}
	worldID := uuid.New()

	// Create a simple heightmap
	// 20x20
	// Player at 10,10.
	// Hill at 12,10 (Elev 50).
	// Target at 14,10 (Elev 0).
	// Player Elev 0.
	// Slope to Hill: (50-0)/2 = 25.
	// Slope to Target: (0-0)/4 = 0.
	// 0 < 25 -> Hidden.

	hm := &geography.Heightmap{
		Width:      20,
		Height:     20,
		Elevations: make([]float64, 400),
	}
	// Default 0
	hm.Set(12, 10, 50.0) // The Hill

	geo := &ecosystem.WorldGeology{
		Heightmap: hm,
		Biomes:    make([]geography.Biome, 400),
	}
	// Init biomes to avoid panic
	for i := range geo.Biomes {
		geo.Biomes[i] = geography.Biome{Type: "plains"}
	}
	svc.SetWorldGeology(worldID, geo)

	char := &auth.Character{
		CharacterID: uuid.New(),
		WorldID:     worldID,
		PositionX:   10.0,
		PositionY:   10.0,
		PositionZ:   0.0,
	}

	ctx := context.Background()
	mapData, err := svc.GetMapData(ctx, char)
	assert.NoError(t, err)

	var hillTile, targetTile *MapTile
	for i := range mapData.Tiles {
		tile := &mapData.Tiles[i]
		if tile.X == 12 && tile.Y == 10 {
			hillTile = tile
		}
		if tile.X == 14 && tile.Y == 10 {
			targetTile = tile
		}
	}

	assert.NotNil(t, hillTile, "Hill tile should be in grid")
	assert.NotNil(t, targetTile, "Target tile should be in grid")

	assert.False(t, hillTile.Occluded, "Hill should see player (slope 25 > -inf)")
	assert.True(t, targetTile.Occluded, "Target behind hill should be occluded (slope 0 < 25)")
}
