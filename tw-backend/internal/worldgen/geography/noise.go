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
