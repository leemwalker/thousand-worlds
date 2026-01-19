package ecosystem

import (
	"testing"

	"github.com/google/uuid"
)

// BenchmarkSimulateGeology profiles the main geological simulation loop.
// Run with CPU profiling:
//
//	go test -bench=BenchmarkSimulateGeology -cpuprofile=cpu.prof ./internal/ecosystem/... -benchmem
//
// Analyze with:
//
//	go tool pprof -http=:8080 cpu.prof
func BenchmarkSimulateGeology(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0 // ~Earth

	geology := NewWorldGeology(worldID, seed, circumference)
	// Use 128 resolution for reasonable benchmark time
	geology.InitializeGeology(128)

	dt := int64(100_000) // 100k years per step
	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		geology.SimulateGeology(dt, globalTempMod)
	}
}

// BenchmarkSimulateGeology_Hadean profiles early Earth (molten) simulation.
// This should be faster due to Hadean optimizations skipping surface processes.
func BenchmarkSimulateGeology_Hadean(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology(128)
	// Keep TotalYearsSimulated at 0 to stay in Hadean (heat > 4.0)

	dt := int64(1_000_000) // 1M years per step (fast Hadean mode)
	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		geology.SimulateGeology(dt, globalTempMod)
	}
}

// BenchmarkSimulateGeology_ModernEarth profiles mature Earth simulation.
// This exercises all surface processes (erosion, rivers, biomes).
func BenchmarkSimulateGeology_ModernEarth(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology(128)
	// Skip to modern era to exercise all surface processes
	geology.TotalYearsSimulated = 4_500_000_000 // 4.5 billion years

	dt := int64(100_000) // 100k years per step
	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		geology.SimulateGeology(dt, globalTempMod)
	}
}

// BenchmarkSimulateGeology_HighRes profiles high resolution simulation.
// Tests performance at 4096 resolution (6 * 4096 * 4096 = 100,663,296 cells).
func BenchmarkSimulateGeology_HighRes(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology(256) // High resolution
	geology.TotalYearsSimulated = 4_500_000_000

	dt := int64(100_000)
	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		geology.SimulateGeology(dt, globalTempMod)
	}
}

// BenchmarkSimulateGeology_UltraRes profiles performance at 1024 resolution (6.3M cells).
func BenchmarkSimulateGeology_UltraRes(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology(1024) // 1024 resolution
	geology.TotalYearsSimulated = 4_500_000_000

	dt := int64(100_000)
	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		geology.SimulateGeology(dt, globalTempMod)
	}
}

// BenchmarkLongSimulationRun_UltraRes simulates a single 10M year step at 1024 res.
// This forces all systems (erosion, weather, tectonics, magma, caves) to run at least once.
func BenchmarkLongSimulationRun_UltraRes(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology(1024)
	// Set TotalYearsSimulated such that the first step brings it to 4.51B
	dt := int64(10_000_000)
	geology.TotalYearsSimulated = 4_500_000_000

	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Run a single massive step to trigger all accumulators and periodic systems
		geology.SimulateGeology(dt, globalTempMod)
	}
}

// BenchmarkLongSimulationRun_ExtremeRes simulates a single 10M year step at 2048 res (25M cells).
// This confirms bottlenecks at extreme resolution.
func BenchmarkLongSimulationRun_ExtremeRes(b *testing.B) {
	worldID := uuid.New()
	seed := int64(12345)
	circumference := 40_000_000.0

	geology := NewWorldGeology(worldID, seed, circumference)
	geology.InitializeGeology(4096) // 6 * 4096 * 4096 = 100M cells
	dt := int64(10_000_000)
	geology.TotalYearsSimulated = 4_500_000_000

	globalTempMod := 0.0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		geology.SimulateGeology(dt, globalTempMod)
	}
}
