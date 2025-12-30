package graphics

import (
	"image"
	"image/color"
	"math"

	"tw-backend/internal/worldgen/geography"

	"github.com/fogleman/gg"
)

// Renderer converts spherical heightmap data to 2D images with hypsometric coloring
type Renderer struct {
	Width  int
	Height int
}

// NewRenderer creates a new renderer with the specified output dimensions
func NewRenderer(width int) *Renderer {
	return &Renderer{
		Width:  width,
		Height: width / 2, // Equirectangular projection: 2:1 aspect ratio
	}
}

// FrameInfo contains metadata to overlay on the rendered frame
type FrameInfo struct {
	Year       int64
	PlateCount int
	Events     []string
}

// RenderFrame renders a single frame from the heightmap with optional info overlay
func (r *Renderer) RenderFrame(heightmap *geography.Heightmap, info FrameInfo) image.Image {
	ctx := gg.NewContext(r.Width, r.Height)

	// Convert heightmap to image using hypsometric tinting
	r.renderTerrain(ctx, heightmap)

	// Apply hillshading for depth
	r.applyHillshading(ctx, heightmap)

	// Overlay text info
	r.renderOverlay(ctx, info)

	return ctx.Image()
}

// renderTerrain converts elevation to hypsometric colors
func (r *Renderer) renderTerrain(ctx *gg.Context, heightmap *geography.Heightmap) {
	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			// Map renderer coordinates to heightmap coordinates
			hx := x * heightmap.Width / r.Width
			hy := y * heightmap.Height / r.Height

			elev := heightmap.Get(hx, hy)
			c := elevationToColor(elev)
			ctx.SetColor(c)
			ctx.SetPixel(x, y)
		}
	}
}

// applyHillshading adds depth by computing normals from neighbor elevations
func (r *Renderer) applyHillshading(ctx *gg.Context, heightmap *geography.Heightmap) {
	// Light direction (northwest, above)
	lightX, lightY, lightZ := -0.5, -0.5, 0.7
	lightMag := math.Sqrt(lightX*lightX + lightY*lightY + lightZ*lightZ)
	lightX, lightY, lightZ = lightX/lightMag, lightY/lightMag, lightZ/lightMag

	img := ctx.Image()

	for y := 1; y < r.Height-1; y++ {
		for x := 1; x < r.Width-1; x++ {
			// Map to heightmap coordinates
			hx := x * heightmap.Width / r.Width
			hy := y * heightmap.Height / r.Height

			// Skip edge cases
			if hx <= 0 || hx >= heightmap.Width-1 || hy <= 0 || hy >= heightmap.Height-1 {
				continue
			}

			// Compute gradient using central difference
			dzdx := (heightmap.Get(hx+1, hy) - heightmap.Get(hx-1, hy)) / 2.0
			dzdy := (heightmap.Get(hx, hy+1) - heightmap.Get(hx, hy-1)) / 2.0

			// Normal vector (unnormalized is fine for dot product direction)
			// Surface normal: (-dzdx, -dzdy, 1) (pointing up)
			scale := 0.0001 // Scale factor for elevation gradient
			nx, ny, nz := -dzdx*scale, -dzdy*scale, 1.0
			nmag := math.Sqrt(nx*nx + ny*ny + nz*nz)
			nx, ny, nz = nx/nmag, ny/nmag, nz/nmag

			// Dot product with light direction
			dot := nx*lightX + ny*lightY + nz*lightZ

			// Clamp and convert to shading factor (0.5 to 1.5)
			shade := 0.5 + dot*0.5
			if shade < 0.3 {
				shade = 0.3
			}
			if shade > 1.2 {
				shade = 1.2
			}

			// Apply shading to existing color
			origColor := img.At(x, y)
			r, g, b, a := origColor.RGBA()
			ctx.SetRGBA255(
				clampUint8(float64(r>>8)*shade),
				clampUint8(float64(g>>8)*shade),
				clampUint8(float64(b>>8)*shade),
				int(a>>8),
			)
			ctx.SetPixel(x, y)
		}
	}
}

// renderOverlay draws text information on the image
func (r *Renderer) renderOverlay(ctx *gg.Context, info FrameInfo) {
	// Semi-transparent background for text
	ctx.SetRGBA(0, 0, 0, 0.6)
	ctx.DrawRectangle(10, 10, 300, 30)
	ctx.Fill()

	// Text color
	ctx.SetRGB(1, 1, 1)

	// Format year for display
	yearStr := formatYear(info.Year)
	text := yearStr
	if info.PlateCount > 0 {
		text += " | Plates: " + itoa(info.PlateCount)
	}
	if len(info.Events) > 0 {
		text += " | " + info.Events[0]
	}

	ctx.DrawString(text, 15, 30)
}

// elevationToColor converts elevation to a hypsometric color
func elevationToColor(elev float64) color.RGBA {
	if elev < 0 {
		// Ocean: Dark Blue -> Light Blue
		// Deepest: -10000m, Shallowest: 0m
		depth := math.Min(math.Abs(elev), 10000)
		t := 1.0 - (depth / 10000.0) // 0 at deep, 1 at surface

		return color.RGBA{
			R: uint8(10 + t*50),
			G: uint8(30 + t*100),
			B: uint8(80 + t*120),
			A: 255,
		}
	}

	// Land: Sand -> Green -> Gray -> White
	// 0-100m: Sandy lowlands
	// 100-1000m: Green vegetation
	// 1000-3000m: Gray rock
	// 3000m+: White snow

	if elev < 100 {
		t := elev / 100.0
		return color.RGBA{
			R: uint8(210 - t*60),
			G: uint8(180 - t*30),
			B: uint8(140 - t*80),
			A: 255,
		}
	}

	if elev < 1000 {
		t := (elev - 100) / 900.0
		return color.RGBA{
			R: uint8(150 - t*80),
			G: uint8(150 + t*30),
			B: uint8(60 + t*30),
			A: 255,
		}
	}

	if elev < 3000 {
		t := (elev - 1000) / 2000.0
		return color.RGBA{
			R: uint8(70 + t*100),
			G: uint8(180 - t*80),
			B: uint8(90 + t*80),
			A: 255,
		}
	}

	// Snow caps
	t := math.Min((elev-3000)/2000.0, 1.0)
	return color.RGBA{
		R: uint8(170 + t*85),
		G: uint8(100 + t*155),
		B: uint8(170 + t*85),
		A: 255,
	}
}

// clampUint8 clamps a float to 0-255 and returns as int
func clampUint8(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return int(v)
}

// formatYear formats year for display (e.g., "120M years")
func formatYear(year int64) string {
	if year >= 1_000_000_000 {
		return itoa64(year/1_000_000_000) + "B years"
	}
	if year >= 1_000_000 {
		return itoa64(year/1_000_000) + "M years"
	}
	if year >= 1_000 {
		return itoa64(year/1_000) + "K years"
	}
	return itoa64(year) + " years"
}

// itoa converts int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

// itoa64 converts int64 to string
func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}
