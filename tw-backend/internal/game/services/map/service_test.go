package gamemap

import (
	"context"
	"testing"

	"tw-backend/internal/auth"

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

	// Should have 100 tiles (10x10 grid)
	assert.Equal(t, 100, len(mapData.Tiles))
	assert.Equal(t, 10, mapData.GridSize)

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

	// Check bounds - should span from 95 to 104 (100-5 to 100+4 with radius 5)
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

	assert.Equal(t, 95, minX)
	assert.Equal(t, 104, maxX)
	assert.Equal(t, 195, minY)
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
		// Ground level: radius 5 = 10x10 grid
		{"On ground", 0, false, 1, 10},
		// Flying: radius = 5 + floor(altitude/5), grid = radius * 2
		{"Flying at 1m", 1, true, 1, 10},      // 5 + 0 = 5, grid = 10
		{"Flying at 5m", 5, true, 1, 12},      // 5 + 1 = 6, grid = 12
		{"Flying at 25m", 25, true, 1, 20},    // 5 + 5 = 10, grid = 20
		{"Flying at 50m", 50, true, 1, 30},    // 5 + 10 = 15, grid = 30
		{"Flying at 100m", 100, true, 1, 50},  // 5 + 20 = 25, grid = 50
		{"Flying at 355m", 355, true, 1, 152}, // 5 + 71 = 76, grid = 152 (max)
		{"Flying at 500m", 500, true, 1, 152}, // capped at 152
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
			assert.Equal(t, tt.expectedScale, mapData.Scale, "Scale should always be 1")
			assert.Equal(t, tt.expectedGrid, mapData.GridSize, "Grid size should match expected")
		})
	}
}
