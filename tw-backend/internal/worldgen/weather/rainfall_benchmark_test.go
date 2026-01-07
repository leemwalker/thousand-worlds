package weather

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// BenchmarkGenerateRainfallMap measures performance of rainfall generation
// Run with: go test -bench=BenchmarkGenerateRainfallMap ./internal/worldgen/weather/... -benchmem
func BenchmarkGenerateRainfallMap(b *testing.B) {
	// Use resolution 128 for quick benchmarks
	topo := spatial.NewCubeSphereTopology(128)
	hm := geography.NewSphereHeightmap(topo)
	seaLevel := 0.0

	// Create simple terrain: some land, some ocean
	for face := 0; face < 6; face++ {
		for y := 0; y < 128; y++ {
			for x := 0; x < 128; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				// Land on faces 0-2, ocean on faces 3-5
				if face < 3 {
					hm.Set(coord, 500.0)
				} else {
					hm.Set(coord, -200.0)
				}
			}
		}
	}

	config := DefaultRainfallConfig(seaLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRainfallMap(hm, topo, config)
	}
}

// BenchmarkGenerateRainfallMap_HighRes measures performance at higher resolution
func BenchmarkGenerateRainfallMap_HighRes(b *testing.B) {
	// Use resolution 256 for more realistic benchmark
	topo := spatial.NewCubeSphereTopology(256)
	hm := geography.NewSphereHeightmap(topo)
	seaLevel := 0.0

	// Create simple terrain: some land, some ocean
	for face := 0; face < 6; face++ {
		for y := 0; y < 256; y++ {
			for x := 0; x < 256; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				// Land on faces 0-2, ocean on faces 3-5
				if face < 3 {
					hm.Set(coord, 500.0)
				} else {
					hm.Set(coord, -200.0)
				}
			}
		}
	}

	config := DefaultRainfallConfig(seaLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRainfallMap(hm, topo, config)
	}
}
