package integration_test

import (
	"math"
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

func TestRainfallErosionImpact(t *testing.T) {
	// Setup parameters
	width, height := 100, 100
	seed := int64(42) // Deterministic seed
	plateCount := 5
	erosionRate := 1.0 // Standard erosion

	// Generate plates once to share (though GenerateHeightmap re-calculates/uses them)
	// We need to pass the same plates to ensure base terrain is similar before erosion
	plates := geography.GeneratePlates(plateCount, width, height, seed)

	// 1. Generate Arid World (RainfallFactor = 0.25)
	hmArid := geography.GenerateHeightmap(width, height, plates, seed, erosionRate, 0.25)

	// 2. Generate Wet World (RainfallFactor = 2.0)
	hmWet := geography.GenerateHeightmap(width, height, plates, seed, erosionRate, 2.0)

	// Compare Statistics
	// Hydraulic erosion generally fills depressions and carves channels.
	// In a simplistic model, it often *lowers* peaks and *raises* valleys, helping transport sediment.
	// But mostly it *carves*, so average elevation might drop or staying similar while variance changes.

	// Let's check "Roughness" or "Average Change from Base".
	// Since we can't easily get Base without erosion here (unless we run a 0.0 run), let's compare the two.

	// Hypothesis: Wet world should be "smoother" in some sense (sediment deposition) OR more "carved" (valleys).
	// Let's measure:
	// A. Average Elevation (Wet might be lower due to mass loss if system isn't perfectly conservationist, or just moved around)
	// B. Standard Deviation (Wet might be lower if peaks are eroded and valleys filled)

	statsArid := calculateStats(hmArid)
	statsWet := calculateStats(hmWet)

	t.Logf("Arid Stats: Avg=%f, StdDev=%f, Min=%f, Max=%f", statsArid.Avg, statsArid.StdDev, statsArid.Min, statsArid.Max)
	t.Logf("Wet Stats:  Avg=%f, StdDev=%f, Min=%f, Max=%f", statsWet.Avg, statsWet.StdDev, statsWet.Min, statsWet.Max)

	// Assertion: Wet world should have statistically significant differences.
	// We expect Wet world to be MORE eroded.
	// In this implementation, Hydraulic erosion moves material downhill.
	// If it works, the "Wet" world should have specific characteristics compared to "Arid".

	// E.g. Lower Max Elevation? (Peaks eroded)
	assert.True(t, statsWet.Max <= statsArid.Max, "Wet world peaks should be eroded more or equal to Arid")

	// Check that they are NOT identical
	assert.NotEqual(t, statsArid.Avg, statsWet.Avg, "Average elevation should differ")
}

type Stats struct {
	Avg    float64
	StdDev float64
	Min    float64
	Max    float64
}

func calculateStats(hm *geography.Heightmap) Stats {
	sum := 0.0
	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64

	for _, v := range hm.Elevations {
		sum += v
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	avg := sum / float64(len(hm.Elevations))

	variance := 0.0
	for _, v := range hm.Elevations {
		variance += (v - avg) * (v - avg)
	}
	stdDev := math.Sqrt(variance / float64(len(hm.Elevations)))

	return Stats{Avg: avg, StdDev: stdDev, Min: minVal, Max: maxVal}
}
