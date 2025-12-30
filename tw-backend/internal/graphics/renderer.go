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

// elevationToColor converts elevation to satellite-style colors matching the WebGL shader.
// Uses the "Lifeless Protoplanet" palette: basalt, granite, clay, ice materials.
func elevationToColor(elev float64) color.RGBA {
	// Normalize elevation: sea level = 0, range roughly -6000 to +8848
	// Browser uses 0.5 as sea level in normalized space
	seaLevel := 0.0
	relativeElev := elev - seaLevel

	if relativeElev <= 0 {
		// Water: Multiple depth zones (matching WebGL C_WATER_* constants)
		// depth: 0=coast, 1=deepest (-6000m)
		depth := math.Min(math.Abs(relativeElev)/6000.0, 1.0)

		// Five-zone water gradient matching shader
		if depth < 0.1 {
			// Coastal zone: C_WATER_COASTAL -> C_WATER_SHALLOW
			t := depth * 10.0
			return lerpColor(
				color.RGBA{R: 71, G: 133, B: 184, A: 255}, // C_WATER_COASTAL (0.28, 0.52, 0.72)
				color.RGBA{R: 46, G: 94, B: 171, A: 255},  // C_WATER_SHALLOW (0.18, 0.37, 0.67)
				t,
			)
		} else if depth < 0.3 {
			// Shelf: C_WATER_SHALLOW -> C_WATER_MID
			t := (depth - 0.1) * 5.0
			return lerpColor(
				color.RGBA{R: 46, G: 94, B: 171, A: 255}, // C_WATER_SHALLOW
				color.RGBA{R: 20, G: 46, B: 89, A: 255},  // C_WATER_MID (0.08, 0.18, 0.35)
				t,
			)
		} else if depth < 0.7 {
			// Mid-ocean: C_WATER_MID -> C_WATER_DEEP
			t := (depth - 0.3) * 2.5
			return lerpColor(
				color.RGBA{R: 20, G: 46, B: 89, A: 255}, // C_WATER_MID
				color.RGBA{R: 5, G: 15, B: 36, A: 255},  // C_WATER_DEEP (0.02, 0.06, 0.14)
				t,
			)
		}
		// Abyssal: C_WATER_DEEP -> C_WATER_ABYSSAL
		t := math.Min((depth-0.7)*3.3, 1.0)
		return lerpColor(
			color.RGBA{R: 5, G: 15, B: 36, A: 255}, // C_WATER_DEEP
			color.RGBA{R: 3, G: 8, B: 20, A: 255},  // C_WATER_ABYSSAL (0.01, 0.03, 0.08)
			t,
		)
	}

	// Land materials based on height (matching WebGL C_* material constants)
	// t: 0=sea level, 1=highest peak (8848m)
	t := math.Min(relativeElev/8848.0, 1.0)

	if t < 0.03 {
		// Wet coastal sand -> dry sand
		t2 := t * 33.0
		return lerpColor(
			color.RGBA{R: 166, G: 148, B: 107, A: 255}, // C_SAND_WET (0.65, 0.58, 0.42)
			color.RGBA{R: 209, G: 194, B: 148, A: 255}, // C_SAND_DRY (0.82, 0.76, 0.58)
			t2,
		)
	} else if t < 0.08 {
		// Dry sand -> sediment
		t2 := (t - 0.03) * 20.0
		return lerpColor(
			color.RGBA{R: 209, G: 194, B: 148, A: 255}, // C_SAND_DRY
			color.RGBA{R: 125, G: 107, B: 87, A: 255},  // C_SEDIMENT (0.49, 0.42, 0.34)
			t2,
		)
	} else if t < 0.18 {
		// Sediment -> clay
		t2 := (t - 0.08) * 10.0
		return lerpColor(
			color.RGBA{R: 125, G: 107, B: 87, A: 255}, // C_SEDIMENT
			color.RGBA{R: 148, G: 97, B: 71, A: 255},  // C_CLAY (0.58, 0.38, 0.28)
			t2,
		)
	} else if t < 0.35 {
		// Clay -> basalt
		t2 := (t - 0.18) * 5.9
		return lerpColor(
			color.RGBA{R: 148, G: 97, B: 71, A: 255}, // C_CLAY
			color.RGBA{R: 56, G: 54, B: 51, A: 255},  // C_ROCK_BASALT (0.22, 0.21, 0.20)
			t2,
		)
	} else if t < 0.55 {
		// Basalt -> granite
		t2 := (t - 0.35) * 5.0
		return lerpColor(
			color.RGBA{R: 56, G: 54, B: 51, A: 255}, // C_ROCK_BASALT
			color.RGBA{R: 97, G: 92, B: 87, A: 255}, // C_ROCK_GRANITE (0.38, 0.36, 0.34)
			t2,
		)
	} else if t < 0.72 {
		// Granite -> rocky ice
		t2 := (t - 0.55) * 5.9
		return lerpColor(
			color.RGBA{R: 97, G: 92, B: 87, A: 255},    // C_ROCK_GRANITE
			color.RGBA{R: 140, G: 148, B: 158, A: 255}, // C_ICE_ROCK (0.55, 0.58, 0.62)
			t2,
		)
	} else if t < 0.85 {
		// Rocky ice -> glacier
		t2 := (t - 0.72) * 7.7
		return lerpColor(
			color.RGBA{R: 140, G: 148, B: 158, A: 255}, // C_ICE_ROCK
			color.RGBA{R: 191, G: 209, B: 224, A: 255}, // C_ICE_GLACIER (0.75, 0.82, 0.88)
			t2,
		)
	}
	// Glacier -> snow (peaks)
	t2 := math.Min((t-0.85)*6.7, 1.0)
	return lerpColor(
		color.RGBA{R: 191, G: 209, B: 224, A: 255}, // C_ICE_GLACIER
		color.RGBA{R: 242, G: 242, B: 250, A: 255}, // C_SNOW (0.95, 0.95, 0.98)
		t2,
	)
}

// lerpColor linearly interpolates between two colors
func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
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
