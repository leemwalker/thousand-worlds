package geography

import (
	"math"

	"tw-backend/internal/spatial"
)

// =============================================================================
// Coastal Erosion Engine (Photorealistic Coastlines)
// =============================================================================
//
// Physics-based coastal erosion simulation for realistic coastline shapes.
// All features are backed by simulation data - no procedural noise for shapes.
//
// Key processes:
// 1. Wave Energy: Fetch distance determines wave power
// 2. Cliff Retreat: Rock hardness controls erosion rate
// 3. Sediment Transport: Eroded material moves alongshore (longshore drift)
// 4. Beach Formation: Sediment deposits in low-energy zones
// 5. Platform Formation: Wave-cut platforms at base of cliffs

// CoastalErosionConfig contains parameters for coastal erosion simulation
type CoastalErosionConfig struct {
	// Wave energy scaling (higher = more erosion)
	WaveEnergyScale float64

	// Maximum fetch distance for wave energy calculation (meters)
	MaxFetchDistance float64

	// Erosion rate multiplier (m/year at full wave energy, soft rock)
	BaseErosionRate float64

	// Sediment transport efficiency (0-1, how much eroded material becomes beach)
	SedimentRetention float64

	// Tidal range affects intertidal zone width (meters)
	TidalRange float64

	// Minimum water depth for wave erosion (meters below sea level)
	MinWaveDepth float64
}

// DefaultCoastalConfig returns sensible defaults for coastal erosion
func DefaultCoastalConfig() CoastalErosionConfig {
	return CoastalErosionConfig{
		WaveEnergyScale:   1.0,
		MaxFetchDistance:  500000.0, // 500 km max fetch
		BaseErosionRate:   0.001,    // 1mm/year base rate (scaled by hardness)
		SedimentRetention: 0.8,      // 80% of eroded material becomes sediment
		TidalRange:        2.0,      // 2m default tidal range
		MinWaveDepth:      -50.0,    // Waves affect down to 50m depth
	}
}

// SimulateCoastalErosion applies wave-driven erosion to coastlines
// dt is the time step in years
// seaLevel is the current sea level in meters
func SimulateCoastalErosion(
	hm *SphereHeightmap,
	topology spatial.Topology,
	dt float64,
	seaLevel float64,
	config CoastalErosionConfig,
) {
	if hm == nil || topology == nil {
		return
	}

	res := topology.Resolution()

	// First pass: Identify coastal cells and calculate fetch
	coastalCells := findCoastalCells(hm, topology, seaLevel, config.MinWaveDepth)

	// Second pass: Apply erosion based on wave energy
	for _, cell := range coastalCells {
		applyWaveErosion(hm, topology, cell, dt, seaLevel, config, res)
	}

	// Third pass: Transport sediment alongshore (longshore drift)
	applyLongshoreDrift(hm, topology, coastalCells, seaLevel)
}

// CoastalCell represents a cell at the coastline with erosion metadata
type CoastalCell struct {
	Coord      spatial.Coordinate
	Elevation  float64
	Fetch      float64           // Distance to open ocean (meters)
	WaveEnergy float64           // Normalized wave energy (0-1)
	IsCliff    bool              // True if steep slope to water
	SeawardDir spatial.Direction // Direction toward deepest water
}

// findCoastalCells identifies cells at the land-water interface
func findCoastalCells(
	hm *SphereHeightmap,
	topology spatial.Topology,
	seaLevel float64,
	minWaveDepth float64,
) []CoastalCell {
	res := topology.Resolution()
	coastal := make([]CoastalCell, 0, res*res/10) // Estimate 10% are coastal

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				elev := hm.Get(coord)

				// Skip cells that are too deep (below wave influence)
				if elev < seaLevel+minWaveDepth {
					continue
				}

				// Check if this is a coastal cell (has water neighbor)
				seawardDir, hasWaterNeighbor := findSeawardDirection(hm, topology, coord, seaLevel)
				if !hasWaterNeighbor {
					continue
				}

				// Calculate fetch (distance to open ocean)
				fetch := calculateFetch(hm, topology, coord, seawardDir, seaLevel)

				// Calculate wave energy from fetch (diminishing returns past MaxFetch)
				waveEnergy := math.Sqrt(fetch / 500000.0) // Normalize to 500km
				if waveEnergy > 1.0 {
					waveEnergy = 1.0
				}

				// Determine if this is a cliff (steep slope)
				isCliff := isCliffCell(hm, topology, coord, seaLevel)

				coastal = append(coastal, CoastalCell{
					Coord:      coord,
					Elevation:  elev,
					Fetch:      fetch,
					WaveEnergy: waveEnergy,
					IsCliff:    isCliff,
					SeawardDir: seawardDir,
				})
			}
		}
	}

	return coastal
}

// findSeawardDirection returns the direction toward deepest water neighbor
func findSeawardDirection(
	hm *SphereHeightmap,
	topology spatial.Topology,
	coord spatial.Coordinate,
	seaLevel float64,
) (spatial.Direction, bool) {
	dirs := []spatial.Direction{spatial.North, spatial.East, spatial.South, spatial.West}

	lowestElev := math.MaxFloat64
	var seawardDir spatial.Direction
	hasWater := false

	for _, dir := range dirs {
		neighbor := topology.GetNeighbor(coord, dir)
		neighborElev := hm.Get(neighbor)

		if neighborElev < seaLevel && neighborElev < lowestElev {
			lowestElev = neighborElev
			seawardDir = dir
			hasWater = true
		}
	}

	return seawardDir, hasWater
}

// calculateFetch traces seaward to find open ocean distance
func calculateFetch(
	hm *SphereHeightmap,
	topology spatial.Topology,
	start spatial.Coordinate,
	dir spatial.Direction,
	seaLevel float64,
) float64 {
	// Trace along direction until we hit land or max distance
	// Each cell represents ~10km at typical resolutions
	cellSizeKm := 40000.0 / float64(topology.Resolution()) // Earth circumference / resolution
	maxCells := 100                                        // Max 100 cells = ~1000km fetch

	current := start
	fetchCells := 0

	for i := 0; i < maxCells; i++ {
		current = topology.GetNeighbor(current, dir)
		elev := hm.Get(current)

		if elev >= seaLevel {
			// Hit land - fetch ends
			break
		}

		fetchCells++
	}

	return float64(fetchCells) * cellSizeKm * 1000.0 // Convert to meters
}

// isCliffCell determines if a cell has cliff-like steep slope to water
func isCliffCell(
	hm *SphereHeightmap,
	topology spatial.Topology,
	coord spatial.Coordinate,
	seaLevel float64,
) bool {
	elev := hm.Get(coord)
	dirs := []spatial.Direction{spatial.North, spatial.East, spatial.South, spatial.West}

	for _, dir := range dirs {
		neighbor := topology.GetNeighbor(coord, dir)
		neighborElev := hm.Get(neighbor)

		if neighborElev < seaLevel {
			// Drop to water - check if steep (>30m drop in one cell)
			drop := elev - seaLevel
			if drop > 30.0 {
				return true
			}
		}
	}

	return false
}

// applyWaveErosion erodes a coastal cell based on wave energy and rock hardness
func applyWaveErosion(
	hm *SphereHeightmap,
	topology spatial.Topology,
	cell CoastalCell,
	dt float64,
	seaLevel float64,
	config CoastalErosionConfig,
	resolution int,
) {
	// Get rock hardness (0=soft, 1=hard)
	hardness := hm.GetRockHardness(cell.Coord)

	// Erosion rate: baseRate * waveEnergy * (1 - hardness) * dt
	// Soft rock erodes fast, hard rock slow
	erosionMultiplier := config.WaveEnergyScale * cell.WaveEnergy * (1.0 - hardness*0.9)
	erosion := config.BaseErosionRate * erosionMultiplier * dt

	// Cliffs erode faster (undercutting)
	if cell.IsCliff {
		erosion *= 2.0
	}

	// Apply erosion
	if erosion > 0 {
		actualErosion := hm.Erode(cell.Coord, erosion)

		// Convert eroded material to sediment (deposit offshore or alongshore)
		sedimentGenerated := actualErosion * config.SedimentRetention

		// Deposit sediment in seaward direction (forms beaches/shelves)
		if sedimentGenerated > 0 {
			depositCoord := topology.GetNeighbor(cell.Coord, cell.SeawardDir)
			depositElev := hm.Get(depositCoord)

			// Only deposit if underwater (forms beach at waterline)
			if depositElev < seaLevel {
				hm.AddSediment(depositCoord, sedimentGenerated)
			}
		}
	}
}

// applyLongshoreDrift moves sediment along the coastline
// Simulates wave-driven sediment transport parallel to shore
func applyLongshoreDrift(
	hm *SphereHeightmap,
	topology spatial.Topology,
	coastalCells []CoastalCell,
	seaLevel float64,
) {
	// Sort coastal cells by position to form chains
	// Then move sediment downslope along the chain

	// For each coastal cell with sediment, move some to downstream neighbor
	driftRate := 0.1 // 10% of sediment moves per iteration

	for _, cell := range coastalCells {
		cellData := hm.GetCellData(cell.Coord)
		if cellData.Sediment <= 0 {
			continue
		}

		// Sediment moves perpendicular to wave direction (alongshore)
		// Find the alongshore direction with lowest elevation
		alongshoreDir := findAlongshoreDirection(hm, topology, cell.Coord, cell.SeawardDir, seaLevel)
		if alongshoreDir == "" {
			continue
		}

		// Move sediment
		transferAmount := cellData.Sediment * driftRate

		// Remove from current cell
		cellData.Sediment -= transferAmount
		hm.SetCellData(cell.Coord, cellData)
		hm.Set(cell.Coord, hm.Get(cell.Coord)-transferAmount)

		// Add to neighbor
		neighbor := topology.GetNeighbor(cell.Coord, alongshoreDir)
		hm.AddSediment(neighbor, transferAmount)
	}
}

// findAlongshoreDirection returns the perpendicular direction with lower elevation
func findAlongshoreDirection(
	hm *SphereHeightmap,
	topology spatial.Topology,
	coord spatial.Coordinate,
	seawardDir spatial.Direction,
	seaLevel float64,
) spatial.Direction {
	// Alongshore is perpendicular to seaward
	var alongDirs []spatial.Direction
	switch seawardDir {
	case spatial.North, spatial.South:
		alongDirs = []spatial.Direction{spatial.East, spatial.West}
	case spatial.East, spatial.West:
		alongDirs = []spatial.Direction{spatial.North, spatial.South}
	default:
		return ""
	}

	// Find direction with lower coastal cell
	lowestElev := hm.Get(coord)
	var bestDir spatial.Direction = ""

	for _, dir := range alongDirs {
		neighbor := topology.GetNeighbor(coord, dir)
		neighborElev := hm.Get(neighbor)

		// Only drift to cells that are also coastal (near sea level)
		if neighborElev < lowestElev && neighborElev >= seaLevel-10 && neighborElev < seaLevel+50 {
			lowestElev = neighborElev
			bestDir = dir
		}
	}

	return bestDir
}

// =============================================================================
// Beach Formation (Natural Sediment Accumulation)
// =============================================================================

// FormBeaches builds up beaches in sheltered areas where sediment accumulates
// Called periodically to consolidate drift deposits into beach features
func FormBeaches(
	hm *SphereHeightmap,
	topology spatial.Topology,
	seaLevel float64,
) {
	res := topology.Resolution()

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				elev := hm.Get(coord)

				// Look for cells just below sea level with sediment
				if elev >= seaLevel-5 && elev < seaLevel+2 {
					cellData := hm.GetCellData(coord)

					// If sediment has accumulated, raise to form beach
					if cellData.Sediment > 1.0 {
						// Cap beach height at 3m above sea level
						targetHeight := seaLevel + math.Min(cellData.Sediment*0.5, 3.0)
						if elev < targetHeight {
							hm.Set(coord, targetHeight)
						}
					}
				}
			}
		}
	}
}
