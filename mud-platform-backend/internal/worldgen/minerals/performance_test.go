package minerals

import (
	"testing"
	"time"

	"mud-platform-backend/internal/worldgen/geography"
)

func TestPerformance1000Deposits(t *testing.T) {
	ctx := &TectonicContext{
		PlateBoundaryType:  geography.BoundaryConvergent,
		MagmaFlowDirection: geography.Vector{X: 1, Y: 0},
		FaultLineDirection: geography.Vector{X: 0, Y: 1},
		ErosionLevel:       0.5,
		Age:                50,
		Elevation:          2500,
		IsVolcanic:         true,
		IsSedimentaryBasin: false,
	}

	start := time.Now()
	deposits := make([]*MineralDeposit, 1000)

	for i := 0; i < 1000; i++ {
		epicenter := geography.Point{X: float64(i % 100), Y: float64(i / 100)}
		deposits[i] = GenerateMineralVein(ctx, MineralIron, epicenter)
	}

	elapsed := time.Since(start)

	// Must complete in under 5 seconds
	if elapsed > 5*time.Second {
		t.Errorf("Performance test failed: took %v to generate 1000 deposits (max 5s)", elapsed)
	}

	// Verify all deposits are valid
	for _, d := range deposits {
		if d == nil || d.Quantity <= 0 {
			t.Error("Invalid deposit generated")
		}
	}

	t.Logf("Generated 1000 deposits in %v", elapsed)
}

func BenchmarkGenerateMineralVein(b *testing.B) {
	ctx := &TectonicContext{
		IsVolcanic:         true,
		MagmaFlowDirection: geography.Vector{X: 1, Y: 0},
	}
	epicenter := geography.Point{X: 0, Y: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateMineralVein(ctx, MineralGold, epicenter)
	}
}

func BenchmarkGenerateCluster(b *testing.B) {
	ctx := &TectonicContext{
		IsVolcanic:         true,
		MagmaFlowDirection: geography.Vector{X: 1, Y: 0},
	}
	primary := GenerateMineralVein(ctx, MineralGold, geography.Point{X: 0, Y: 0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCluster(primary, ctx)
	}
}
