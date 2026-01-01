package gamemap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TDD Red Phase: Tile Types Tests
// =============================================================================

func TestCubeFace_String(t *testing.T) {
	tests := []struct {
		face     CubeFace
		expected string
	}{
		{FaceFront, "Front"},
		{FaceBack, "Back"},
		{FaceLeft, "Left"},
		{FaceRight, "Right"},
		{FaceTop, "Top"},
		{FaceBottom, "Bottom"},
		{CubeFace(99), "Unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.face.String())
	}
}

func TestTilesPerSide(t *testing.T) {
	assert.Equal(t, 1, TilesPerSide(0))
	assert.Equal(t, 2, TilesPerSide(1))
	assert.Equal(t, 4, TilesPerSide(2))
	assert.Equal(t, 8, TilesPerSide(3))
	assert.Equal(t, 1, TilesPerSide(-1)) // Edge case
}

func TestTotalTiles(t *testing.T) {
	assert.Equal(t, 1, TotalTiles(0))
	assert.Equal(t, 4, TotalTiles(1))
	assert.Equal(t, 16, TotalTiles(2))
	assert.Equal(t, 64, TotalTiles(3))
}

// =============================================================================
// TDD Red Phase: Tile Renderer Tests
// =============================================================================

func TestNewTileRenderer(t *testing.T) {
	renderer := NewTileRenderer(RenderConfig{
		MaxWidth:  2048,
		MaxHeight: 2048,
	})

	require.NotNil(t, renderer)
	assert.Equal(t, 2048, renderer.config.MaxWidth)
}

func TestRenderTile_Level0_FullFace(t *testing.T) {
	renderer := NewTileRenderer(RenderConfig{
		MaxWidth:  256,
		MaxHeight: 256,
	})

	ctx := context.Background()
	req := TileRequest{
		Face:  FaceFront,
		Level: 0,
		X:     0,
		Y:     0,
		Size:  256,
	}

	tile, err := renderer.RenderTile(ctx, req, nil)
	require.NoError(t, err)
	require.NotNil(t, tile)

	assert.Equal(t, FaceFront, tile.Coord.Face)
	assert.Equal(t, 0, tile.Coord.Level)
	assert.Equal(t, 0, tile.Coord.X)
	assert.Equal(t, 0, tile.Coord.Y)
	assert.Equal(t, 256, tile.Width)
	assert.Equal(t, 256, tile.Height)
	assert.NotEmpty(t, tile.Image, "Tile image should not be empty")
}

func TestRenderTile_Level1_Subdivided(t *testing.T) {
	renderer := NewTileRenderer(RenderConfig{
		MaxWidth:  256,
		MaxHeight: 256,
	})

	ctx := context.Background()

	// At level 1, we should have 4 tiles (2x2 grid)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			req := TileRequest{
				Face:  FaceFront,
				Level: 1,
				X:     x,
				Y:     y,
				Size:  256,
			}

			tile, err := renderer.RenderTile(ctx, req, nil)
			require.NoError(t, err)
			require.NotNil(t, tile)

			assert.Equal(t, x, tile.Coord.X)
			assert.Equal(t, y, tile.Coord.Y)
			assert.Equal(t, 1, tile.Coord.Level)
		}
	}
}

func TestRenderTile_InvalidCoords_ReturnsError(t *testing.T) {
	renderer := NewTileRenderer(RenderConfig{})

	ctx := context.Background()

	// X out of range for level 1 (should be 0 or 1)
	req := TileRequest{
		Face:  FaceFront,
		Level: 1,
		X:     5, // Invalid: only 0-1 valid at level 1
		Y:     0,
		Size:  256,
	}

	_, err := renderer.RenderTile(ctx, req, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func TestRenderTileHeightmap_IncludesElevationData(t *testing.T) {
	renderer := NewTileRenderer(RenderConfig{
		MaxWidth:  256,
		MaxHeight: 256,
	})

	ctx := context.Background()
	req := TileRequest{
		Face:  FaceFront,
		Level: 0,
		X:     0,
		Y:     0,
		Size:  256,
	}

	// Pass nil for worldGeo to test stub behavior
	tile, err := renderer.RenderTile(ctx, req, nil)
	require.NoError(t, err)
	require.NotNil(t, tile)

	// Heightmap should be included
	assert.NotEmpty(t, tile.Heightmap, "Heightmap should be included in tile data")
}

// =============================================================================
// TDD Red Phase: Cube to Sphere Projection Tests
// =============================================================================

func TestCubeToSphere_FrontFaceCenter(t *testing.T) {
	// Center of front face (Z+) should map to equator at 90° longitude
	lat, lon := CubeToSphere(FaceFront, 0.5, 0.5)

	assert.InDelta(t, 0.0, lat, 1.0, "Center of front face should be near equator")
	assert.InDelta(t, 90.0, lon, 1.0, "Center of front face (Z+) should be at 90° longitude")
}

func TestCubeToSphere_TopFaceCenter(t *testing.T) {
	// Center of top face should map to north pole
	lat, lon := CubeToSphere(FaceTop, 0.5, 0.5)

	assert.InDelta(t, 90.0, lat, 1.0, "Center of top face should be at north pole")
	_ = lon // Longitude is undefined at poles
}

func TestCubeToSphere_BottomFaceCenter(t *testing.T) {
	// Center of bottom face should map to south pole
	lat, lon := CubeToSphere(FaceBottom, 0.5, 0.5)

	assert.InDelta(t, -90.0, lat, 1.0, "Center of bottom face should be at south pole")
	_ = lon // Longitude is undefined at poles
}
