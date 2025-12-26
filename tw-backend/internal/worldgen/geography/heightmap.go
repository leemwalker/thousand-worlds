package geography

import (
	"tw-backend/internal/spatial"
)

// GenerateHeightmap creates the final heightmap for a spherical world.
// Uses SphereHeightmap and spherical topology for all calculations.
// DEPRECATED: Use GenerateHeightmapWithTidalStress for satellite-aware generation.
func GenerateHeightmap(plates []TectonicPlate, heightmap *SphereHeightmap, topology spatial.Topology, seed int64, erosionRate float64, rainfallFactor float64) *SphereHeightmap {
	// Default to Earth-Moon baseline tidal stress and modern Earth heat for backward compatibility
	return GenerateHeightmapWithTidalStress(plates, heightmap, topology, seed, erosionRate, rainfallFactor, 1.0, 1.0)
}

// GenerateHeightmapWithTidalStress creates the final heightmap with satellite-aware volcanism.
// The tidalStress parameter affects volcanic activity (0.0 = no moons, 1.0 = Earth-Moon, >1.0 = multiple/close moons).
// The heatMultiplier parameter scales volcanic activity based on planetary age (1.0 = modern, 10.0 = early Earth).
func GenerateHeightmapWithTidalStress(plates []TectonicPlate, heightmap *SphereHeightmap, topology spatial.Topology, seed int64, erosionRate float64, rainfallFactor float64, tidalStress float64, heatMultiplier float64) *SphereHeightmap {
	noise := NewPerlinGenerator(seed)
	resolution := topology.Resolution()

	// 1. Base Elevation based on Plate Type
	// Use the Region map from plates to assign base elevation
	for i := range plates {
		plate := &plates[i]
		baseElev := -4000.0 // Oceanic default
		if plate.Type == PlateContinental {
			baseElev = 100.0 // Continental default
		}

		for coord := range plate.Region {
			heightmap.Set(coord, baseElev)
		}
	}

	// 2. Apply Tectonic Modifiers
	SimulateTectonics(plates, heightmap, topology, 1.0)

	// 2a. Apply Volcanic Hotspots (scaled by tidal stress and planetary heat)
	ApplyHotspots(heightmap, plates, topology, seed, tidalStress, heatMultiplier)

	// 3. Apply Noise for variation
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				// Multiple octaves of noise using sphere position
				sx, sy, sz := topology.ToSphere(coord)
				n1 := noise.Noise3D(sx*2, sy*2, sz*2)
				n2 := noise.Noise3D(sx*10, sy*10, sz*10)

				variation := n1*500 + n2*100

				current := heightmap.Get(coord)
				heightmap.Set(coord, current+variation)
			}
		}
	}

	// 4. Advanced Erosion
	// Scale iterations by erosionRate
	iterations := int(5.0 * erosionRate)
	if iterations < 1 {
		iterations = 1
	}
	ApplyThermalErosionSpherical(heightmap, topology, iterations, seed)

	// Hydraulic erosion
	effectiveRainfall := rainfallFactor
	if effectiveRainfall <= 0 {
		effectiveRainfall = 1.0
	}
	totalCells := 6 * resolution * resolution
	numDrops := int(float64(totalCells) * 0.05 * erosionRate * effectiveRainfall)
	ApplyHydraulicErosionSpherical(heightmap, topology, numDrops, seed)

	// 5. Smooth
	SmoothSpherical(heightmap, topology)

	// Update Min/Max
	heightmap.UpdateMinMax()

	return heightmap
}

// SmoothSpherical applies a box blur to the sphere heightmap
func SmoothSpherical(hm *SphereHeightmap, topology spatial.Topology) {
	resolution := topology.Resolution()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Create a copy of current values
	original := make(map[spatial.Coordinate]float64)
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				original[coord] = hm.Get(coord)
			}
		}
	}

	// Apply smoothing
	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				sum := original[coord]
				count := 1.0

				for _, dir := range directions {
					neighbor := topology.GetNeighbor(coord, dir)
					if val, exists := original[neighbor]; exists {
						sum += val
						count++
					}
				}

				hm.Set(coord, sum/count)
			}
		}
	}
}

// ApplyThermalErosionSpherical applies thermal erosion on a sphere
func ApplyThermalErosionSpherical(hm *SphereHeightmap, topology spatial.Topology, iterations int, seed int64) {
	resolution := topology.Resolution()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}
	talusAngle := 0.5 // Maximum stable slope

	for iter := 0; iter < iterations; iter++ {
		for face := 0; face < 6; face++ {
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					currentElev := hm.Get(coord)

					for _, dir := range directions {
						neighbor := topology.GetNeighbor(coord, dir)
						neighborElev := hm.Get(neighbor)
						diff := currentElev - neighborElev

						if diff > talusAngle {
							// Transfer material
							transfer := diff * 0.25
							hm.Set(coord, currentElev-transfer)
							hm.Set(neighbor, neighborElev+transfer)
							currentElev -= transfer
						}
					}
				}
			}
		}
	}
}

// ApplyHydraulicErosionSpherical simulates water erosion on a sphere
func ApplyHydraulicErosionSpherical(hm *SphereHeightmap, topology spatial.Topology, numDrops int, seed int64) {
	// Simplified hydraulic erosion - trace water droplets downhill
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for drop := int64(0); drop < int64(numDrops); drop++ {
		// Start at random position
		startPoint := spatial.RandomPointOnSphere(seed + drop)
		coord := topology.FromVector(startPoint.X, startPoint.Y, startPoint.Z)

		sediment := 0.0
		capacity := 1.0
		erosionRate := 0.1
		depositionRate := 0.1

		// Trace downhill for max 50 steps
		for step := 0; step < 50; step++ {
			currentElev := hm.Get(coord)

			// Find steepest descent
			var lowestNeighbor *spatial.Coordinate
			lowestElev := currentElev

			for _, dir := range directions {
				neighbor := topology.GetNeighbor(coord, dir)
				neighborElev := hm.Get(neighbor)
				if neighborElev < lowestElev {
					lowestElev = neighborElev
					neighborCopy := neighbor
					lowestNeighbor = &neighborCopy
				}
			}

			if lowestNeighbor == nil {
				// Local minimum - deposit all sediment
				hm.Set(coord, currentElev+sediment)
				break
			}

			// Calculate slope
			slope := currentElev - lowestElev
			newCapacity := slope * capacity

			if sediment > newCapacity {
				// Deposit excess
				deposit := (sediment - newCapacity) * depositionRate
				hm.Set(coord, currentElev+deposit)
				sediment -= deposit
			} else {
				// Erode
				erode := (newCapacity - sediment) * erosionRate
				if erode > slope*0.5 {
					erode = slope * 0.5
				}
				hm.Set(coord, currentElev-erode)
				sediment += erode
			}

			coord = *lowestNeighbor
		}
	}
}
