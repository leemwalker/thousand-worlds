package graphics

import (
	"image/color"
	"testing"

	"tw-backend/internal/worldgen/geography"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "standard 1024 width",
			width:          1024,
			expectedWidth:  1024,
			expectedHeight: 512,
		},
		{
			name:           "small 256 width",
			width:          256,
			expectedWidth:  256,
			expectedHeight: 128,
		},
		{
			name:           "odd width",
			width:          100,
			expectedWidth:  100,
			expectedHeight: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(tt.width)
			if r.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", r.Width, tt.expectedWidth)
			}
			if r.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", r.Height, tt.expectedHeight)
			}
		})
	}
}

func TestElevationToColor_Ocean(t *testing.T) {
	tests := []struct {
		name  string
		elev  float64
		check func(c color.RGBA) bool
	}{
		{
			name: "deep ocean is dark blue",
			elev: -10000,
			check: func(c color.RGBA) bool {
				// Deep ocean should be darker (lower RGB values)
				return c.B > c.R && c.B > c.G && c.B < 120
			},
		},
		{
			name: "shallow ocean is lighter blue",
			elev: -100,
			check: func(c color.RGBA) bool {
				// Shallow ocean should be lighter blue
				return c.B > c.R && c.B > c.G && c.B > 150
			},
		},
		{
			name: "sea level is lightest blue",
			elev: -1,
			check: func(c color.RGBA) bool {
				return c.B > c.R && c.B > c.G
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := elevationToColor(tt.elev)
			if !tt.check(c) {
				t.Errorf("elevationToColor(%f) = %v, failed color check", tt.elev, c)
			}
		})
	}
}

func TestElevationToColor_Land(t *testing.T) {
	tests := []struct {
		name  string
		elev  float64
		check func(c color.RGBA) bool
	}{
		{
			name: "lowland is sandy",
			elev: 50,
			check: func(c color.RGBA) bool {
				// Sandy lowlands - warm tan/beige colors
				return c.R > 150 && c.G > 130 && c.B > 90
			},
		},
		{
			name: "midland is clay/sediment",
			elev: 1000,
			check: func(c color.RGBA) bool {
				// Clay/sediment zone - brownish colors
				return c.R > 100 && c.G > 50
			},
		},
		{
			name: "highland is rocky basalt/granite",
			elev: 4000,
			check: func(c color.RGBA) bool {
				// Rocky highlands - grayish rock colors
				return c.R > 50 && c.G > 50 && c.B > 50 && c.R < 180
			},
		},
		{
			name: "mountain peak is snow/ice",
			elev: 8000,
			check: func(c color.RGBA) bool {
				// Snow/ice caps should be bright and bluish-white
				return c.R > 180 && c.G > 180 && c.B > 180
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := elevationToColor(tt.elev)
			if !tt.check(c) {
				t.Errorf("elevationToColor(%f) = %v, failed color check", tt.elev, c)
			}
		})
	}
}

func TestRenderFrame_Dimensions(t *testing.T) {
	tests := []struct {
		name           string
		rendererWidth  int
		heightmapW     int
		heightmapH     int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "512x256 output from 64x32 heightmap",
			rendererWidth:  512,
			heightmapW:     64,
			heightmapH:     32,
			expectedWidth:  512,
			expectedHeight: 256,
		},
		{
			name:           "1024x512 output from 128x64 heightmap",
			rendererWidth:  1024,
			heightmapW:     128,
			heightmapH:     64,
			expectedWidth:  1024,
			expectedHeight: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(tt.rendererWidth)
			hm := geography.NewHeightmap(tt.heightmapW, tt.heightmapH)

			// Set some test elevations
			for y := 0; y < tt.heightmapH; y++ {
				for x := 0; x < tt.heightmapW; x++ {
					// Gradient from ocean to mountain
					elev := float64(x-tt.heightmapW/2) * 100
					hm.Set(x, y, elev)
				}
			}

			img := r.RenderFrame(hm, FrameInfo{Year: 1000000, PlateCount: 5})
			bounds := img.Bounds()

			if bounds.Dx() != tt.expectedWidth {
				t.Errorf("Image width = %d, want %d", bounds.Dx(), tt.expectedWidth)
			}
			if bounds.Dy() != tt.expectedHeight {
				t.Errorf("Image height = %d, want %d", bounds.Dy(), tt.expectedHeight)
			}
		})
	}
}

func TestFormatYear(t *testing.T) {
	tests := []struct {
		year     int64
		expected string
	}{
		{100, "100 years"},
		{1000, "1K years"},
		{5000, "5K years"},
		{1_000_000, "1M years"},
		{120_000_000, "120M years"},
		{1_000_000_000, "1B years"},
		{4_500_000_000, "4B years"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatYear(tt.year)
			if result != tt.expected {
				t.Errorf("formatYear(%d) = %q, want %q", tt.year, result, tt.expected)
			}
		})
	}
}

func TestClampUint8(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{-10, 0},
		{0, 0},
		{128, 128},
		{255, 255},
		{300, 255},
	}

	for _, tt := range tests {
		result := clampUint8(tt.input)
		if result != tt.expected {
			t.Errorf("clampUint8(%f) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
