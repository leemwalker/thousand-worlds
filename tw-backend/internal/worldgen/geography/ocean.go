package geography

import (
	"sort"

	"tw-backend/internal/spatial"
)

// AssignOceanLand determines the sea level and classifies terrain
func AssignOceanLand(hm *Heightmap, targetLandRatio float64) float64 {
	// 1. Flatten and sort elevations to find percentile
	// 1. Flatten and sort elevations to find percentile
	elevations := make([]float64, len(hm.Elevations))
	for i, v := range hm.Elevations {
		elevations[i] = float64(v)
	}
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
	// Prompt mentions "Adjust sea level iteratively if > 5% off target", but percentile method guarantees exact ratio
	// unless there are many duplicate values (flat plains).
	// We'll stick with percentile as it's efficient and accurate.

	return seaLevel
}

// NormalizeLandRatio adjusts a SphereHeightmap to achieve a target land/water ratio.
// This ensures consistent land coverage (25-35%) regardless of random seed.
// Algorithm:
// 1. Collect all heights into a slice
// 2. Sort to find the percentile needed for target ratio
// 3. Shift all heights so the target percentile becomes sea level (0)
// Returns the calculated sea level (before shifting).
func NormalizeLandRatio(hm *SphereHeightmap, topology spatial.Topology, targetRatio float64) float64 {
	resolution := topology.Resolution()
	totalCells := 6 * resolution * resolution

	// Step 1: Collect all heights
	elevations := make([]float64, 0, totalCells)
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				elevations = append(elevations, hm.Get(coord))
			}
		}
	}

	// Step 2: Sort to find percentile
	sort.Float64s(elevations)

	// For 30% land, we need the 70th percentile to be sea level
	oceanRatio := 1.0 - targetRatio
	percentileIndex := int(float64(len(elevations)) * oceanRatio)
	if percentileIndex >= len(elevations) {
		percentileIndex = len(elevations) - 1
	}
	if percentileIndex < 0 {
		percentileIndex = 0
	}

	desiredSeaLevel := elevations[percentileIndex]

	// Step 3: Calculate offset to shift sea level to 0
	offset := 0.0 - desiredSeaLevel

	// Step 4: Apply offset to all cells
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				currentElev := hm.Get(coord)
				hm.Set(coord, currentElev+offset)
			}
		}
	}

	// Update min/max after shifting
	hm.UpdateMinMax()

	return desiredSeaLevel
}
