package geography

import (
	"sync"
	"tw-backend/internal/spatial"
)

// =============================================================================
// Hypsometric Curve Constants
// =============================================================================

// Shelf zone boundaries in meters (relative to sea level)
const (
	ShelfLowerBound  = -600.0 // Deepest part of continental shelf
	ShelfUpperBound  = 200.0  // Coastal plain edge
	ShelfCompression = 0.3    // Compression factor for shelf zone (0.3 = flatten to 30%)
)

// ApplyHypsometricCurve remaps height values to create realistic continental shelves.
// The shelf zone (-600m to +200m relative to sea level) is compressed/flattened,
// while deep ocean and high land pass through with minimal change.
// This creates the characteristic "step" in real Earth topography.
func ApplyHypsometricCurve(height, seaLevel float64) float64 {
	// Calculate relative height from sea level
	relativeHeight := height - seaLevel

	// Shelf zone: compress the gradient
	if relativeHeight > ShelfLowerBound && relativeHeight < ShelfUpperBound {
		// How far into the shelf zone are we? (0 = lower, 1 = upper)
		shelfRange := ShelfUpperBound - ShelfLowerBound
		normalizedPos := (relativeHeight - ShelfLowerBound) / shelfRange

		// Map back with compression
		compressedHeight := ShelfLowerBound + (normalizedPos * shelfRange * ShelfCompression)
		return seaLevel + compressedHeight
	}

	// Deep ocean: pass through (slightly boost depth for drama)
	if relativeHeight <= ShelfLowerBound {
		// Below shelf edge, keep the depth but shift by shelf compression effect
		offset := ShelfLowerBound * (1.0 - ShelfCompression)
		return height + offset
	}

	// Land: pass through (slightly boost height for contrast)
	if relativeHeight >= ShelfUpperBound {
		// Above shelf edge, keep the height but shift by shelf compression effect
		offset := ShelfUpperBound * (1.0 - ShelfCompression)
		return height - offset
	}

	return height
}

// GenerateHeightmap creates the final heightmap for a spherical world.
// Uses SphereHeightmap and spherical topology for all calculations.
// DEPRECATED: Use GenerateHeightmapWithTidalStress for satellite-aware generation.
func GenerateHeightmap(plates []TectonicPlate, heightmap *SphereHeightmap, topology spatial.Topology, seed int64, erosionRate float64, rainfallFactor float64) *SphereHeightmap {
	// Default to Earth-Moon baseline tidal stress and modern Earth heat for backward compatibility
	return GenerateHeightmapWithTidalStress(plates, heightmap, topology, seed, erosionRate, rainfallFactor, 1.0, 1.0, 10000.0) // Added default maxElevation
}

// GenerateHeightmapWithTidalStress creates the final heightmap with satellite-aware volcanism.
// The tidalStress parameter affects volcanic activity (0.0 = no moons, 1.0 = Earth-Moon, >1.0 = multiple/close moons).
// The heatMultiplier parameter scales volcanic activity based on planetary age (1.0 = modern, 10.0 = early Earth).
func GenerateHeightmapWithTidalStress(plates []TectonicPlate, heightmap *SphereHeightmap, topology spatial.Topology, seed int64, erosionRate float64, rainfallFactor float64, tidalStress float64, heatMultiplier float64, maxElevation float64) *SphereHeightmap {
	// 1. Initial Tectonics Simulation
	// We simulate basic plate collisions to form mountain ranges
	heightmap = SimulateTectonics(plates, heightmap, topology, 1.0, maxElevation)
	fbm := NewFBMGenerator(seed, DefaultTerrainFBMConfig())
	resolution := topology.Resolution()

	// 1. Base Elevation based on Plate Type
	// Use the Region map from plates to assign base elevation
	for i := range plates {
		plate := &plates[i]
		baseElev := OceanicBaseElevation // -4500.0
		if plate.Type == PlateContinental {
			baseElev = ContinentalBaseElevation // 300.0
		}

		for coord := range plate.Region {
			heightmap.Set(coord, baseElev)
		}
	}

	// 2. Apply Tectonic Modifiers
	SimulateTectonics(plates, heightmap, topology, 1.0, maxElevation)

	// 2a. Apply Volcanic Hotspots (scaled by tidal stress and planetary heat)
	ApplyHotspots(heightmap, plates, topology, seed, tidalStress, heatMultiplier)

	// 3. Apply FBM/Ridge Noise for natural terrain variation
	// Use standard FBM for mid-levels, Ridge noise for extremes (ocean floor/peaks)
	// OPTIMIZED: Parallelized by face
	var wg sync.WaitGroup
	wg.Add(6)
	for face := 0; face < 6; face++ {
		go func(f int) {
			defer wg.Done()
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: f, X: x, Y: y}

					// Get sphere position for 3D noise sampling
					sx, sy, sz := topology.ToSphere(coord)

					current := heightmap.Get(coord)

					// Use Ridge Noise for deep ocean (< -2000m) and high peaks (> 2000m)
					// Creates sharp ridges/valleys instead of smooth rolling terrain
					var variation float64
					if current < -2000.0 || current > 2000.0 {
						// Ridge noise: creates sharp mid-ocean ridges and rugged peaks
						// Range [0,1] * 800 - 400 gives [-400, +400] variation
						ridgeNoise := fbm.RidgeFBM3D(sx, sy, sz)
						variation = (ridgeNoise * 800.0) - 400.0
					} else {
						// Standard FBM for coastal/mid-level terrain
						// 600m variation provides natural hills/valleys without overwhelming tectonics
						variation = fbm.FBM3D(sx, sy, sz) * 600.0
					}

					heightmap.Set(coord, current+variation)
				}
			}
		}(face)
	}
	wg.Wait()

	// 4. Thermal Erosion (slope stabilization)
	if resolution > 32 {
		iterations := int(5.0 * erosionRate)
		if iterations < 1 {
			iterations = 1
		}
		ApplyThermalErosionSpherical(heightmap, topology, iterations, seed)

		// 5. Hydraulic Erosion (rain carving valleys)
		// Use significantly more droplets for visible valley formation
		effectiveRainfall := rainfallFactor
		if effectiveRainfall <= 0 {
			effectiveRainfall = 1.0
		}
		totalCells := 6 * resolution * resolution
		// Minimum 10,000 droplets, scaled with erosionRate and rainfall
		numDrops := int(float64(totalCells) * 0.15 * erosionRate * effectiveRainfall)
		if numDrops < 10000 {
			numDrops = 10000
		}
		ApplyHydraulicErosionSpherical(heightmap, topology, numDrops, seed)

		// 6. Smooth (slight blur to blend erosion artifacts)
		SmoothSpherical(heightmap, topology)
	}

	// 7. Apply Isostatic Relaxation
	// This replaces NormalizeLandRatio - land/water ratio now comes naturally from plate physics.
	// Continental crust floats at +150m, oceanic sinks to -4000m.
	// The ~30% continental plate area = ~30% land (before erosion/sedimentation adjustments).
	ApplyIsostaticRelaxation(plates, heightmap, topology, IsostaticRelaxationRate)

	// 8. Apply Hypsometric Curve for continental shelf flattening
	// Sea level is at 0 (the boundary between positive/negative elevations)
	// OPTIMIZED: Parallelized by face
	wg.Add(6)
	for face := 0; face < 6; face++ {
		go func(f int) {
			defer wg.Done()
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: f, X: x, Y: y}
					current := heightmap.Get(coord)
					remapped := ApplyHypsometricCurve(current, 0.0) // Sea level at 0
					heightmap.Set(coord, remapped)
				}
			}
		}(face)
	}
	wg.Wait()

	// 9. Add Micro-Roughness for land texture
	// High-frequency noise only on land (>0) to add small hills, bumps
	// This prevents smooth polygons on flat terrain
	if resolution > 32 {
		microFbm := NewFBMGenerator(seed+9999, FBMConfig{
			Octaves:      4,
			Frequency:    2.0, // High frequency for small detail
			Lacunarity:   2.0,
			Persistence:  0.5,
			WarpStrength: 0.1,
		})
		// OPTIMIZED: Parallelized by face
		wg.Add(6)
		for face := 0; face < 6; face++ {
			go func(f int) {
				defer wg.Done()
				for y := 0; y < resolution; y++ {
					for x := 0; x < resolution; x++ {
						coord := spatial.Coordinate{Face: f, X: x, Y: y}
						current := heightmap.Get(coord)

						// Only add micro-roughness to land (> 0)
						if current > 0 {
							sx, sy, sz := topology.ToSphere(coord)
							// Small amplitude (50m) detail noise
							microNoise := microFbm.FBM3D(sx*50, sy*50, sz*50) * 50.0
							heightmap.Set(coord, current+microNoise)
						}
					}
				}
			}(face)
		}
		wg.Wait()
	}

	// Update Min/Max
	heightmap.UpdateMinMax()

	return heightmap
}

// SmoothSpherical applies a box blur to the sphere heightmap
// OPTIMIZED: Uses fast slice copying instead of map-based buffering.
func SmoothSpherical(hm *SphereHeightmap, topology spatial.Topology) {
	resolution := topology.Resolution()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// Create a fast copy of current values using slices
	// original[face][index]
	original := make([][]float64, 6)
	for f := 0; f < 6; f++ {
		original[f] = make([]float64, resolution*resolution)
		// Access underlying face data if possible, or copy manually if private
		// Since we are in the same package (geography), we can rely on Get/Set for now,
		// but let's copy efficiently.
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				original[f][y*resolution+x] = hm.Get(spatial.Coordinate{Face: f, X: x, Y: y})
			}
		}
	}

	// Apply smoothing (Parallelized by face)
	var wg sync.WaitGroup
	wg.Add(6)
	for face := 0; face < 6; face++ {
		go func(f int) {
			defer wg.Done()
			for y := 0; y < resolution; y++ {
				for x := 0; x < resolution; x++ {
					coord := spatial.Coordinate{Face: f, X: x, Y: y}
					idx := y*resolution + x

					sum := original[f][idx]
					count := 1.0

					for _, dir := range directions {
						neighbor := topology.GetNeighbor(coord, dir)

						// Fast lookup from original slice
						nVal := original[neighbor.Face][neighbor.Y*resolution+neighbor.X]
						sum += nVal
						count++
					}

					hm.Set(coord, sum/count)
				}
			}
		}(face)
	}
	wg.Wait()
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
