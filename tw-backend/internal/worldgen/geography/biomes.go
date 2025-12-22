package geography

import (
	"math"

	"github.com/google/uuid"
)

// ClassifyBiome determines biome type from climate and elevation data.
// This is a PURE classification function with NO latitude/coordinate math.
// Temperature and rainfall should be computed by the Weather service.
//
// Parameters:
//   - tempC: Annual average temperature in Celsius (from Weather)
//   - rainfallMM: Annual rainfall in millimeters (from Weather)
//   - drainage: Soil drainage factor 0-1 (0=waterlogged, 1=well-drained)
//   - elevation: Elevation in meters
//   - seaLevel: Current sea level in meters
//
// Returns the appropriate BiomeType for these conditions.
func ClassifyBiome(tempC, rainfallMM, drainage, elevation, seaLevel float64) BiomeType {
	// 1. Check for ocean (below sea level)
	if elevation <= seaLevel {
		return BiomeOcean
	}

	// 2. Check for extreme elevation biomes
	altitudeAboveSea := elevation - seaLevel
	if altitudeAboveSea >= 3000 {
		return BiomeAlpine // Very high mountain
	}
	if altitudeAboveSea >= 1000 && tempC < 0 {
		return BiomeAlpine // Cold mountain
	}

	// 3. Convert rainfall to moisture factor (0-1 scale)
	// Using 2000mm/year as baseline for "very wet"
	moisture := rainfallMM / 2000.0
	if moisture > 1.0 {
		moisture = 1.0
	}
	if moisture < 0.0 {
		moisture = 0.0
	}

	// 4. Temperature + Moisture classification (Whittaker-style)
	return classifyByClimate(tempC, moisture)
}

// classifyByClimate determines biome from temperature and moisture.
// This implements a Whittaker biome classification approximation.
func classifyByClimate(tempC, moisture float64) BiomeType {
	// Cold climates (< -5°C)
	if tempC < -5 {
		if moisture > 0.5 {
			return BiomeTaiga // Cold but wet enough for trees
		}
		return BiomeTundra // Frozen wasteland
	}

	// Cool climates (-5 to 10°C)
	if tempC < 10 {
		if moisture > 0.6 {
			return BiomeTaiga // Boreal forest
		} else if moisture > 0.3 {
			return BiomeGrassland // Steppe/Tundra transition
		}
		return BiomeTundra // Cold desert
	}

	// Temperate climates (10 to 20°C)
	if tempC < 20 {
		if moisture > 0.6 {
			return BiomeDeciduousForest
		} else if moisture > 0.3 {
			return BiomeGrassland
		}
		return BiomeDesert // Cold desert
	}

	// Warm/Tropical climates (> 20°C)
	if moisture > 0.7 {
		return BiomeRainforest
	} else if moisture > 0.4 {
		return BiomeDeciduousForest // Seasonal tropical / Savanna with trees
	} else if moisture > 0.2 {
		return BiomeGrassland // Savanna
	}
	return BiomeDesert
}

// AssignBiomes determines the biome for each cell
// Deprecated: This function calculates temperature internally using latitude.
// New code should use ClassifyBiome with temperature from the Weather service.
func AssignBiomes(hm *Heightmap, seaLevel float64, seed int64, globalTempMod float64) []Biome {
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
			normalizedY := float64(y) / float64(hm.Height)
			latitude := math.Abs(normalizedY-0.5) * 2.0 // 0.0 (center) to 1.0 (edges)

			// 3. Moisture
			// Use Perlin noise for moisture map
			// Scale: 0.0 to 1.0
			n := noise.Noise2D(float64(x)*0.05, float64(y)*0.05)
			moisture := (n + 1.0) / 2.0 // Normalize to 0-1

			// 4. Temperature
			temp := calculateTemperature(latitude, elev, seaLevel, globalTempMod)

			// 5. Combine
			finalBiome := resolveBiome(bType, temp, moisture)

			biomes[y*hm.Width+x] = Biome{
				BiomeID:       uuid.New(),
				Name:          string(finalBiome),
				Type:          finalBiome,
				Temperature:   temp,
				Precipitation: moisture * 2000, // mm/year
			}
		}
	}

	return biomes
}

func resolveBiome(elevType BiomeType, temp float64, moisture float64) BiomeType {
	if elevType == BiomeOcean {
		return BiomeOcean
	}
	if elevType == BiomeHighMountain {
		return BiomeAlpine
	}
	// Mountains can be various biomes depending on temp, but Alpine at top
	if elevType == BiomeMountain && temp < 0 {
		return BiomeAlpine
	}

	// Determine biome based on Temperature and Moisture (Whitaker classification approximation)

	// Cold climates
	if temp < -5 {
		if moisture > 0.5 {
			return BiomeTaiga // Cold but wet enough for trees
		}
		return BiomeTundra // Frozen wasteland
	}

	// Cool climates (-5 to 10 C)
	if temp < 10 {
		if moisture > 0.6 {
			return BiomeTaiga // Boreal forest
		} else if moisture > 0.3 {
			return BiomeGrassland // Steppe/Tundra transition
		}
		return BiomeTundra // Cold desert
	}

	// Temperate climates (10 to 20 C)
	if temp < 20 {
		if moisture > 0.6 {
			return BiomeDeciduousForest
		} else if moisture > 0.3 {
			return BiomeGrassland
		}
		return BiomeDesert // Cold desert
	}

	// Warm/Tropical climates (> 20 C)
	if moisture > 0.7 {
		return BiomeRainforest
	} else if moisture > 0.4 {
		return BiomeDeciduousForest // Seasonal tropical forest / Savanna with trees
	} else if moisture > 0.2 {
		return BiomeGrassland // Savanna
	}
	return BiomeDesert
}

func calculateTemperature(lat float64, elev float64, seaLevel float64, mood float64) float64 {
	// Base temp at equator: 30C
	// Base temp at poles: -20C
	baseTemp := 30.0 - (lat * 50.0)

	// Lapse rate: -6.5C per 1000m
	altitude := elev - seaLevel
	if altitude < 0 {
		altitude = 0
	}
	temp := baseTemp - (altitude/1000.0)*6.5

	// Apply global modifier
	temp += mood

	return temp
}
