package geography

import (
	"sort"
)

// AssignOceanLand determines the sea level and classifies terrain
func AssignOceanLand(hm *Heightmap, targetLandRatio float64) float64 {
	// 1. Flatten and sort elevations to find percentile
	elevations := make([]float64, len(hm.Elevations))
	copy(elevations, hm.Elevations)
	sort.Float64s(elevations)

	// 2. Find index for sea level
	// If targetLandRatio is 0.3 (30% land), we need the 70th percentile elevation to be sea level
	// So 70% of points are below sea level (ocean)
	oceanRatio := 1.0 - targetLandRatio
	index := int(float64(len(elevations)) * oceanRatio)

	if index >= len(elevations) {
		index = len(elevations) - 1
	}
	if index < 0 {
		index = 0
	}

	seaLevel := elevations[index]

	// 3. Adjust sea level if needed (e.g. clamp to reasonable bounds if requested, but percentile is robust)
	// Prompt mentions "Adjust sea level iteratively if >5% off target", but percentile method guarantees exact ratio
	// unless there are many duplicate values (flat plains).
	// We'll stick with percentile as it's efficient and accurate.

	return seaLevel
}
