package gamemap

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"
	"time"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/require"
)

// TestRenderHeightmapPNG_ReturnsValidPNG tests that heightmap renders as valid 16-bit PNG
func TestRenderHeightmapPNG_ReturnsValidPNG(t *testing.T) {
	// Setup
	ctx := context.Background()
	config := DefaultRenderConfig()
	pool := NewRendererPool(2)
	cache := NewMapCache(5 * time.Minute)
	r := NewRenderer(config, pool, cache)

	// Create minimal test geology using flat heightmap (like other tests)
	geo := createTestGeologyForHeightmap(t)

	// Act
	data, err := r.RenderHeightmapPNG(ctx, "test-world", geo, 256, 128)

	// Assert
	require.NoError(t, err, "RenderHeightmapPNG should not error")
	require.NotEmpty(t, data, "Should return image data")

	// Verify it's a valid PNG
	_, err = png.Decode(bytes.NewReader(data))
	require.NoError(t, err, "Should be valid PNG format")
}

// TestRenderHeightmapPNG_Is16Bit tests that heightmap is 16-bit grayscale
func TestRenderHeightmapPNG_Is16Bit(t *testing.T) {
	ctx := context.Background()
	config := DefaultRenderConfig()
	pool := NewRendererPool(2)
	cache := NewMapCache(5 * time.Minute)
	r := NewRenderer(config, pool, cache)

	geo := createTestGeologyForHeightmap(t)

	data, err := r.RenderHeightmapPNG(ctx, "test-world", geo, 256, 128)
	require.NoError(t, err)

	// Decode and check bit depth
	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	// Check if it's Gray16 format
	_, ok := img.(*image.Gray16)
	require.True(t, ok, "Heightmap should be 16-bit grayscale (Gray16)")
}

// TestRenderHeightmapPNG_NormalizesElevation tests elevation maps to 0-65535 range
func TestRenderHeightmapPNG_NormalizesElevation(t *testing.T) {
	ctx := context.Background()
	config := DefaultRenderConfig()
	pool := NewRendererPool(2)
	cache := NewMapCache(5 * time.Minute)
	r := NewRenderer(config, pool, cache)

	geo := createTestGeologyForHeightmap(t)

	data, err := r.RenderHeightmapPNG(ctx, "test-world", geo, 64, 32)
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	gray16, ok := img.(*image.Gray16)
	require.True(t, ok, "Should be Gray16")
	bounds := gray16.Bounds()

	// Check that we have both low and high values (proper normalization)
	var minPixel, maxPixel uint16 = 65535, 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			val := gray16.Gray16At(x, y).Y
			if val < minPixel {
				minPixel = val
			}
			if val > maxPixel {
				maxPixel = val
			}
		}
	}

	// Normalized range should span from near 0 to near 65535
	require.Less(t, minPixel, uint16(10000), "Min pixel should be low")
	require.Greater(t, maxPixel, uint16(50000), "Max pixel should be high")

	t.Logf("Pixel range: %d to %d", minPixel, maxPixel)
}

// createTestGeologyForHeightmap creates a minimal WorldGeology for heightmap testing
func createTestGeologyForHeightmap(t *testing.T) *ecosystem.WorldGeology {
	t.Helper()

	// Create a simple heightmap with elevation variation
	hmSize := 64
	hm := &geography.Heightmap{
		Width:      hmSize,
		Height:     hmSize,
		Elevations: make([]float64, hmSize*hmSize),
	}

	// Add elevation variation: gradient from -5000 to +5000
	for y := 0; y < hmSize; y++ {
		for x := 0; x < hmSize; x++ {
			idx := y*hmSize + x
			// Create gradient: bottom-left is low, top-right is high
			elev := float64(x+y)/float64(2*hmSize)*10000 - 5000
			hm.Elevations[idx] = elev
		}
	}

	return &ecosystem.WorldGeology{
		Heightmap: hm,
		SeaLevel:  0,
	}
}
