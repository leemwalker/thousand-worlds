package geography

import (
	"github.com/aquilax/go-perlin"
)

// PerlinGenerator generates 2D Perlin noise
type PerlinGenerator struct {
	p *perlin.Perlin
}

// NewPerlinGenerator creates a new generator with a seed
func NewPerlinGenerator(seed int64) *PerlinGenerator {
	// alpha, beta, n (iterations)
	// alpha: weight when sum is formed (default 2)
	// beta: harmonic scaling/lacunarity (default 2)
	// n: number of octaves (default 3)
	p := perlin.NewPerlin(2, 2, 3, seed)
	return &PerlinGenerator{p: p}
}

// Noise2D returns a value between -1 and 1
func (g *PerlinGenerator) Noise2D(x, y float64) float64 {
	return g.p.Noise2D(x, y)
}

// Noise3D returns a value between -1 and 1 for 3D coordinates
func (g *PerlinGenerator) Noise3D(x, y, z float64) float64 {
	return g.p.Noise3D(x, y, z)
}

// =============================================================================
// FBM (Fractal Brownian Motion) Noise System
// =============================================================================

// FBMConfig contains parameters for Fractal Brownian Motion noise generation.
// These parameters control the character of generated terrain.
type FBMConfig struct {
	Octaves      int     // Number of noise layers (more = more detail)
	Frequency    float64 // Base frequency (lower = larger features)
	Lacunarity   float64 // Frequency multiplier per octave (typically 2.0)
	Persistence  float64 // Amplitude multiplier per octave (typically 0.5)
	WarpStrength float64 // Domain warping intensity (0 = none, 0.5 = strong)
}

// DefaultTerrainFBMConfig returns the recommended configuration for natural terrain.
// These values produce Earth-like terrain with organic, non-repeating patterns.
// Tuned for continental-scale features with domain warping to break grid artifacts.
func DefaultTerrainFBMConfig() FBMConfig {
	return FBMConfig{
		Octaves:      6,
		Frequency:    0.3, // Low frequency for large continental features (sphere coords are -1 to 1)
		Lacunarity:   2.0,
		Persistence:  0.5,
		WarpStrength: 0.4, // Moderate distortion to break diamond patterns
	}
}

// FBMGenerator generates Fractal Brownian Motion noise with optional domain warping.
// Domain warping breaks grid alignment artifacts (diamond patterns) by distorting
// the coordinate space before sampling.
type FBMGenerator struct {
	primary *perlin.Perlin // Main noise source
	warpX   *perlin.Perlin // Domain warp X offset
	warpY   *perlin.Perlin // Domain warp Y offset
	warpZ   *perlin.Perlin // Domain warp Z offset
	config  FBMConfig
	maxAmp  float64 // Precomputed max amplitude for normalization
}

// NewFBMGenerator creates a new FBM noise generator with the given seed and config.
// Uses separate Perlin instances for warping to avoid correlation artifacts.
func NewFBMGenerator(seed int64, config FBMConfig) *FBMGenerator {
	g := &FBMGenerator{
		primary: perlin.NewPerlin(2, 2, 1, seed),
		warpX:   perlin.NewPerlin(2, 2, 1, seed+1000),
		warpY:   perlin.NewPerlin(2, 2, 1, seed+2000),
		warpZ:   perlin.NewPerlin(2, 2, 1, seed+3000),
		config:  config,
	}

	// Precompute maximum possible amplitude for normalization
	// Sum of geometric series: 1 + p + p² + ... + p^(n-1) = (1 - p^n) / (1 - p)
	g.maxAmp = 0
	amp := 1.0
	for i := 0; i < config.Octaves; i++ {
		g.maxAmp += amp
		amp *= config.Persistence
	}

	return g
}

// FBM3D generates 3D fractal noise with domain warping.
// Returns a value normalized to approximately [-1, 1].
func (g *FBMGenerator) FBM3D(x, y, z float64) float64 {
	// === Domain Warping ===
	// Sample low-frequency noise to get warp offsets
	// This "twists" the coordinate space, breaking grid alignment
	if g.config.WarpStrength > 0 {
		warpFreq := g.config.Frequency * 0.5 // Lower freq for smoother warping
		dx := g.warpX.Noise3D(x*warpFreq, y*warpFreq, z*warpFreq) * g.config.WarpStrength
		dy := g.warpY.Noise3D(x*warpFreq, y*warpFreq, z*warpFreq) * g.config.WarpStrength
		dz := g.warpZ.Noise3D(x*warpFreq, y*warpFreq, z*warpFreq) * g.config.WarpStrength

		x += dx
		y += dy
		z += dz
	}

	// === Fractal Loop ===
	total := 0.0
	freq := g.config.Frequency
	amp := 1.0

	for i := 0; i < g.config.Octaves; i++ {
		total += g.primary.Noise3D(x*freq, y*freq, z*freq) * amp
		freq *= g.config.Lacunarity
		amp *= g.config.Persistence
	}

	// === Normalization ===
	// Divide by max amplitude to get roughly [-1, 1] range
	return total / g.maxAmp
}

// FBM2D generates 2D fractal noise with domain warping.
// Returns a value normalized to approximately [-1, 1].
func (g *FBMGenerator) FBM2D(x, y float64) float64 {
	// Domain Warping in 2D
	if g.config.WarpStrength > 0 {
		warpFreq := g.config.Frequency * 0.5
		dx := g.warpX.Noise2D(x*warpFreq, y*warpFreq) * g.config.WarpStrength
		dy := g.warpY.Noise2D(x*warpFreq, y*warpFreq) * g.config.WarpStrength

		x += dx
		y += dy
	}

	// Fractal Loop
	total := 0.0
	freq := g.config.Frequency
	amp := 1.0

	for i := 0; i < g.config.Octaves; i++ {
		total += g.primary.Noise2D(x*freq, y*freq) * amp
		freq *= g.config.Lacunarity
		amp *= g.config.Persistence
	}

	return total / g.maxAmp
}

// Config returns the current FBM configuration.
func (g *FBMGenerator) Config() FBMConfig {
	return g.config
}
