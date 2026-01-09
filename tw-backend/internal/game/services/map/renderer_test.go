package gamemap

import (
	"bytes"
	"context"
	"image/png"
	"testing"
	"time"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRenderNormalMapPNG(t *testing.T) {
	// Setup Renderer
	config := RenderConfig{
		MaxHeight:        100,
		MaxWidth:         100,
		DefaultHeight:    10,
		DefaultWidth:     10,
		WebPQuality:      80,
		RenderTimeout:    time.Second,
		ConcurrencyLimit: 1,
	}
	pool := NewRendererPool(1)
	cache := NewMapCache(time.Minute)
	renderer := NewRenderer(config, pool, cache)

	// Create topology (resolution 10)
	topo := spatial.NewCubeSphereTopology(10)

	// Setup Mock Geology with Heightmap
	geo := &ecosystem.WorldGeology{
		SphereHeightmap: geography.NewSphereHeightmap(topo),
	}

	ctx := context.Background()
	worldID := uuid.New().String()

	pngBytes, err := renderer.RenderNormalMapPNG(ctx, worldID, geo, 20, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, pngBytes)

	// Decode PNG
	img, err := png.Decode(bytes.NewReader(pngBytes))
	assert.NoError(t, err)
	assert.Equal(t, 20, img.Bounds().Dx())
	assert.Equal(t, 10, img.Bounds().Dy())

	// Check pixel values for flat terrain (should be Z up = 1.0 -> R=128, G=128, B=255)
	// R = (0+1)*0.5*255 = 127.5 ~ 127
	// G = (0+1)*0.5*255 = 127.5 ~ 127
	// B = 1*255 = 255
	r, g, b, _ := img.At(5, 5).RGBA()
	// Go RGBA returns 0-65535, so divide by 256 for 8-bit comparison
	assert.InDelta(t, 127, r/257, 2, "Red channel should be ~127 for flat normal")
	assert.InDelta(t, 127, g/257, 2, "Green channel should be ~127 for flat normal")
	assert.InDelta(t, 255, b/257, 2, "Blue channel should be ~255 for flat normal")
}
