package geography

import (
	"math"
)

// GenerateHeightmap creates the final heightmap
func GenerateHeightmap(width, height int, plates []TectonicPlate, seed int64) *Heightmap {
	hm := NewHeightmap(width, height)
	noise := NewPerlinGenerator(seed)

	// 1. Base Elevation based on Plate Type
	// We need the Voronoi map again to assign base elevation
	// TODO: Optimize by reusing Voronoi map from SimulateTectonics if possible,
	// but for now recalculating is cleaner separation.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Find closest plate
			minDist := math.MaxFloat64
			var closestPlate TectonicPlate

			for _, p := range plates {
				dist := distance(float64(x), float64(y), p.Centroid.X, p.Centroid.Y)
				if dist < minDist {
					minDist = dist
					closestPlate = p
				}
			}

			baseElev := -4000.0 // Oceanic default
			if closestPlate.Type == PlateContinental {
				baseElev = 100.0 // Continental default
			}

			hm.Set(x, y, baseElev)
		}
	}

	// 2. Apply Tectonic Modifiers
	tectonicMods := SimulateTectonics(plates, width, height)
	for i := 0; i < len(hm.Elevations); i++ {
		hm.Elevations[i] += tectonicMods.Elevations[i]
	}

	// 2a. Apply Volcanic Hotspots
	ApplyHotspots(hm, plates, seed)

	// 3. Apply Noise for variation
	// Scale noise to be significant but not overwhelming
	// Frequency 0.05 for broad features, 0.1 for detail
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Multiple octaves of noise
			n1 := noise.Noise2D(float64(x)*0.02, float64(y)*0.02)
			n2 := noise.Noise2D(float64(x)*0.1, float64(y)*0.1)

			variation := n1*500 + n2*100

			current := hm.Get(x, y)
			hm.Set(x, y, current+variation)
		}
	}

	// 4. Advanced Erosion
	// Thermal erosion to stabilize slopes
	ApplyThermalErosion(hm, 5, seed) // 5 iterations

	// Hydraulic erosion for river channels
	// Drops proportional to area
	numDrops := width * height * 5
	ApplyHydraulicErosion(hm, numDrops, seed)

	// 4. Smooth (Gaussian blur approximation)
	// Simple box blur for performance
	smooth(hm)

	// Update Min/Max
	minElev, maxElev := math.MaxFloat64, -math.MaxFloat64
	for _, val := range hm.Elevations {
		if val < minElev {
			minElev = val
		}
		if val > maxElev {
			maxElev = val
		}
	}
	hm.MinElev = minElev
	hm.MaxElev = maxElev

	return hm
}

func smooth(hm *Heightmap) {
	temp := make([]float64, len(hm.Elevations))
	copy(temp, hm.Elevations)

	for y := 1; y < hm.Height-1; y++ {
		for x := 1; x < hm.Width-1; x++ {
			sum := 0.0
			count := 0.0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					sum += temp[(y+dy)*hm.Width+(x+dx)]
					count++
				}
			}

			hm.Set(x, y, sum/count)
		}
	}
}
