package geography

import (
	"math"

	"github.com/google/uuid"
)

// AssignBiomes determines the biome for each cell
func AssignBiomes(hm *Heightmap, seaLevel float64, seed int64) []Biome {
	biomes := make([]Biome, hm.Width*hm.Height)
	noise := NewPerlinGenerator(seed)

	for y := 0; y < hm.Height; y++ {
		for x := 0; x < hm.Width; x++ {
			elev := hm.Get(x, y)

			// 1. Determine base type by elevation
			var bType BiomeType
			if elev <= seaLevel {
				bType = BiomeOcean
			} else if elev < seaLevel+200 {
				bType = BiomeLowland
			} else if elev < seaLevel+1000 {
				bType = BiomeHighland
			} else if elev < seaLevel+3000 {
				bType = BiomeMountain
			} else {
				bType = BiomeHighMountain
			}

			// 2. Latitude factor (0 at equator, 1 at poles)
			// Assume map covers full hemisphere or world?
			// Let's assume y=0 is North Pole, y=Height is South Pole?
			// Or y=Height/2 is Equator. Let's assume Equator at center.
			normalizedY := float64(y) / float64(hm.Height)
			latitude := math.Abs(normalizedY-0.5) * 2.0 // 0.0 (center) to 1.0 (edges)

			// 3. Moisture
			// Use Perlin noise for moisture map
			// Scale: 0.0 to 1.0
			n := noise.Noise2D(float64(x)*0.05, float64(y)*0.05)
			moisture := (n + 1.0) / 2.0 // Normalize to 0-1

			// 4. Combine
			finalBiome := resolveBiome(bType, latitude, moisture)

			biomes[y*hm.Width+x] = Biome{
				BiomeID:       uuid.New(),
				Name:          string(finalBiome),
				Type:          finalBiome,
				Temperature:   calculateTemperature(latitude, elev, seaLevel),
				Precipitation: moisture * 2000, // mm/year
			}
		}
	}

	return biomes
}

func resolveBiome(elevType BiomeType, lat float64, moisture float64) BiomeType {
	if elevType == BiomeOcean {
		return BiomeOcean
	}
	if elevType == BiomeHighMountain {
		return BiomeAlpine
	}
	if elevType == BiomeMountain {
		return BiomeAlpine
	}

	// Polar
	if lat > 0.8 {
		return BiomeTundra
	}

	// Subarctic
	if lat > 0.6 {
		if moisture > 0.3 {
			return BiomeTaiga
		}
		return BiomeTundra
	}

	// Temperate
	if lat > 0.3 {
		if moisture < 0.3 {
			return BiomeGrassland // Or Desert if very dry
		}
		if elevType == BiomeHighland {
			return BiomeDeciduousForest
		}
		if moisture > 0.6 {
			return BiomeDeciduousForest
		}
		return BiomeGrassland
	}

	// Tropical/Subtropical
	if moisture > 0.6 {
		return BiomeRainforest
	}
	if moisture < 0.3 {
		return BiomeDesert
	}
	return BiomeGrassland // Savanna
}

func calculateTemperature(lat float64, elev float64, seaLevel float64) float64 {
	// Base temp at equator: 30C
	// Base temp at poles: -20C
	baseTemp := 30.0 - (lat * 50.0)

	// Lapse rate: -6.5C per 1000m
	altitude := elev - seaLevel
	if altitude < 0 {
		altitude = 0
	}
	temp := baseTemp - (altitude/1000.0)*6.5

	return temp
}
