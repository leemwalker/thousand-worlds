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

	// Check bounds - should span from 96 to 104 (100-4 to 100+4)
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
		{"On ground", 0, false, 1, 9},
		{"Flying at 1m", 1, true, 1, 17},
		{"Flying at 20m", 20, true, 2, 17},
		{"Flying at 40m", 40, true, 3, 17},
		{"Flying at 60m", 60, true, 4, 17},
		{"Flying at 100m", 100, true, 6, 17},
		{"Flying at 200m", 200, true, 11, 17},
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
			assert.Equal(t, tt.expectedScale, mapData.Scale, "Scale should match expected")
			assert.Equal(t, tt.expectedGrid, mapData.GridSize, "Grid size should match expected")
		})
	}
}
