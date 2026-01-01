package calibration

import (
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// =============================================================================
// Statistics Collector
// =============================================================================

// CollectStats gathers comprehensive statistics from a WorldGeology instance.
// This is the primary entry point for calibration analysis.
func CollectStats(geo *ecosystem.WorldGeology) SimulationStats {
	stats := SimulationStats{
		Seed:       geo.Seed,
		Resolution: geo.Heightmap.Width,
		Years:      geo.TotalYearsSimulated,
	}

	// Collect hypsometry stats
	collectHypsometry(geo, &stats)

	// Collect climate stats
	collectClimate(geo, &stats)

	// Collect hydrology stats
	collectHydrology(geo, &stats)

	// Collect geology stats
	collectGeology(geo, &stats)

	// Collect astronomy stats
	collectAstronomy(geo, &stats)

	return stats
}

// collectHypsometry calculates elevation distribution statistics.
func collectHypsometry(geo *ecosystem.WorldGeology, stats *SimulationStats) {
	seaLevel := geo.SeaLevel

	var totalCells, landCells int
	var sumOcean, sumLand float64
	var minElev, maxElev float64 = 999999, -999999

	// Use SphereHeightmap if available, otherwise flat
	if geo.SphereHeightmap != nil {
		topo := geo.SphereHeightmap.Topology()
		res := topo.Resolution()
		totalCells = 6 * res * res

		for face := 0; face < 6; face++ {
			for y := 0; y < res; y++ {
				for x := 0; x < res; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					elev := geo.SphereHeightmap.Get(coord)

					if elev < minElev {
						minElev = elev
					}
					if elev > maxElev {
						maxElev = elev
					}

					if elev > seaLevel {
						landCells++
						sumLand += elev - seaLevel
					} else {
						sumOcean += elev - seaLevel // Negative depth
					}
				}
			}
		}
	} else if geo.Heightmap != nil {
		totalCells = len(geo.Heightmap.Elevations)

		for _, elev := range geo.Heightmap.Elevations {
			if elev < minElev {
				minElev = elev
			}
			if elev > maxElev {
				maxElev = elev
			}

			if elev > seaLevel {
				landCells++
				sumLand += elev - seaLevel
			} else {
				sumOcean += elev - seaLevel
			}
		}
	}

	oceanCells := totalCells - landCells

	// Calculate percentages and means
	if totalCells > 0 {
		stats.OceanCoveragePercent = float64(oceanCells) / float64(totalCells) * 100
	}
	if oceanCells > 0 {
		stats.MeanOceanDepthM = sumOcean / float64(oceanCells)
	}
	if landCells > 0 {
		stats.MeanLandHeightM = sumLand / float64(landCells)
	}

	stats.MinElevationM = minElev
	stats.MaxElevationM = maxElev

	// Build histogram
	buildElevationHistogram(geo, stats, minElev, maxElev)
}

// buildElevationHistogram creates a binned elevation distribution.
func buildElevationHistogram(geo *ecosystem.WorldGeology, stats *SimulationStats, minElev, maxElev float64) {
	binSize := 100.0 // 100m bins
	numBins := int((maxElev-minElev)/binSize) + 1
	if numBins < 1 {
		numBins = 1
	}
	if numBins > 200 {
		numBins = 200 // Cap at 200 bins
		binSize = (maxElev - minElev) / float64(numBins)
	}

	histogram := make([]int, numBins)
	stats.HistogramBinSize = binSize

	if geo.SphereHeightmap != nil {
		topo := geo.SphereHeightmap.Topology()
		res := topo.Resolution()

		for face := 0; face < 6; face++ {
			for y := 0; y < res; y++ {
				for x := 0; x < res; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					elev := geo.SphereHeightmap.Get(coord)
					bin := int((elev - minElev) / binSize)
					if bin >= numBins {
						bin = numBins - 1
					}
					if bin < 0 {
						bin = 0
					}
					histogram[bin]++
				}
			}
		}
	} else if geo.Heightmap != nil {
		for _, elev := range geo.Heightmap.Elevations {
			bin := int((elev - minElev) / binSize)
			if bin >= numBins {
				bin = numBins - 1
			}
			if bin < 0 {
				bin = 0
			}
			histogram[bin]++
		}
	}

	stats.ElevationHistogram = histogram
}

// collectClimate calculates temperature and rainfall statistics.
func collectClimate(geo *ecosystem.WorldGeology, stats *SimulationStats) {
	if len(geo.Biomes) == 0 {
		return
	}

	var sumTemp, minTemp, maxTemp float64 = 0, 999999, -999999
	var sumRainfall, maxRainfall float64 = 0, 0

	// Track equatorial and polar temperatures
	var equatorSum, poleSum float64
	var equatorCount, poleCount int

	height := geo.Heightmap.Height

	for i, biome := range geo.Biomes {
		temp := biome.Temperature
		sumTemp += temp
		if temp < minTemp {
			minTemp = temp
		}
		if temp > maxTemp {
			maxTemp = temp
		}

		// Estimate rainfall from biome type (since we may not have direct data)
		rainfall := estimateRainfallFromBiome(string(biome.Type))
		sumRainfall += rainfall
		if rainfall > maxRainfall {
			maxRainfall = rainfall
		}

		// Determine latitude band for equator-pole gradient
		y := i / geo.Heightmap.Width
		normalizedY := float64(y) / float64(height)
		latitude := (normalizedY - 0.5) * 2.0 // -1 to +1

		if latitude > -0.2 && latitude < 0.2 {
			equatorSum += temp
			equatorCount++
		} else if latitude < -0.8 || latitude > 0.8 {
			poleSum += temp
			poleCount++
		}
	}

	count := len(geo.Biomes)
	stats.GlobalMeanTempC = sumTemp / float64(count)
	stats.MinTempC = minTemp
	stats.MaxTempC = maxTemp
	stats.MeanRainfallMM = sumRainfall / float64(count)
	stats.MaxRainfallMM = maxRainfall

	if equatorCount > 0 {
		stats.EquatorMeanTempC = equatorSum / float64(equatorCount)
	}
	if poleCount > 0 {
		stats.PoleMeanTempC = poleSum / float64(poleCount)
	}
}

// estimateRainfallFromBiome returns estimated annual rainfall in mm.
func estimateRainfallFromBiome(biomeType string) float64 {
	switch biomeType {
	case "rainforest":
		return 2500.0
	case "swamp":
		return 2000.0
	case "forest":
		return 1200.0
	case "taiga":
		return 600.0
	case "grassland":
		return 500.0
	case "savanna":
		return 400.0
	case "tundra":
		return 250.0
	case "desert":
		return 50.0
	case "ocean", "beach":
		return 1000.0
	default:
		return 500.0
	}
}

// collectHydrology calculates river and drainage statistics.
func collectHydrology(geo *ecosystem.WorldGeology, stats *SimulationStats) {
	stats.RiverCount = len(geo.Rivers)

	// Count land cells with significant flux (river density)
	if geo.SphereHeightmap != nil {
		topo := geo.SphereHeightmap.Topology()
		res := topo.Resolution()
		seaLevel := geo.SeaLevel

		var landCells, fluxCells int
		fluxThreshold := 50.0 // Minimum flux to count as river drainage

		for face := 0; face < 6; face++ {
			for y := 0; y < res; y++ {
				for x := 0; x < res; x++ {
					coord := spatial.Coordinate{Face: face, X: x, Y: y}
					elev := geo.SphereHeightmap.Get(coord)

					if elev > seaLevel {
						landCells++
						cellData := geo.SphereHeightmap.GetCellData(coord)
						if cellData.Flux > fluxThreshold {
							fluxCells++
						}
					}
				}
			}
		}

		if landCells > 0 {
			stats.RiverDensityPercent = float64(fluxCells) / float64(landCells) * 100
		}
	}

	// Count lakes from biomes
	for _, biome := range geo.Biomes {
		if biome.Type == "lake" {
			stats.LakeCount++
		}
	}
}

// collectGeology counts geological features.
func collectGeology(geo *ecosystem.WorldGeology, stats *SimulationStats) {
	stats.PlateCount = len(geo.Plates)
	stats.ProvinceCount = len(geo.Provinces)
	stats.HotspotCount = len(geo.Hotspots)

	if geo.Caves != nil {
		stats.CaveCount = len(geo.Caves)
	}

	// Estimate continent count from continental plates
	continentCount := 0
	for _, plate := range geo.Plates {
		if plate.Type == geography.PlateContinental {
			continentCount++
		}
	}
	stats.ContinentCount = continentCount
}

// collectAstronomy counts satellite features.
func collectAstronomy(geo *ecosystem.WorldGeology, stats *SimulationStats) {
	stats.MoonCount = len(geo.Satellites)
}
