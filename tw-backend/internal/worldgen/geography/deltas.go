package geography

import (
	"math"
	"math/rand"

	"tw-backend/internal/spatial"
)

// =============================================================================
// River Delta Formation (Photorealistic Coastlines - Phase 2)
// =============================================================================
//
// Physics-based delta formation at river mouths.
// High-flux rivers deposit sediment as they enter the sea, creating:
// 1. Distributary Channels: Main river splits into multiple branches
// 2. Sediment Fans: Arc-shaped deposition patterns
// 3. Marshlands: Low-lying areas between distributaries

// DeltaConfig controls delta formation parameters
type DeltaConfig struct {
	// Minimum flux required to form a delta (else just terminates)
	MinFluxForDelta float64

	// Number of distributary channels (scales with flux)
	MinChannels int
	MaxChannels int

	// Sediment deposition rate (meters per unit flux)
	SedimentPerFlux float64

	// Delta fan radius (cells outward from mouth)
	MaxDeltaRadius int
}

// DefaultDeltaConfig returns sensible defaults
func DefaultDeltaConfig() DeltaConfig {
	return DeltaConfig{
		MinFluxForDelta: 200.0, // High flux rivers only
		MinChannels:     2,
		MaxChannels:     5,
		SedimentPerFlux: 0.01, // Meters of sediment per unit flux
		MaxDeltaRadius:  5,    // 5 cells outward
	}
}

// FormDeltasAtRiverMouths identifies high-flux river terminations and creates deltas
// Returns the list of new distributary river paths created
func FormDeltasAtRiverMouths(
	hm *SphereHeightmap,
	topology spatial.Topology,
	seaLevel float64,
	seed int64,
	config DeltaConfig,
) []SphericalRiverPath {
	var newRivers []SphericalRiverPath
	if hm == nil || topology == nil {
		return newRivers
	}

	rng := rand.New(rand.NewSource(seed))
	res := topology.Resolution()

	// Find river mouths (coastal cells with high flux)
	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				elev := hm.Get(coord)
				data := hm.GetCellData(coord)

				// Skip if not at sea level or low flux
				if elev > seaLevel+5 || elev < seaLevel-10 {
					continue
				}
				if data.Flux < config.MinFluxForDelta {
					continue
				}

				// Check if this is a river mouth (has water neighbor)
				if !hasOceanNeighbor(hm, topology, coord, seaLevel) {
					continue
				}

				// Form delta at this location
				distributaries := formDelta(hm, topology, coord, data.Flux, seaLevel, rng, config)
				newRivers = append(newRivers, distributaries...)
			}
		}
	}
	return newRivers
}

// hasOceanNeighbor checks if any neighbor is below sea level
func hasOceanNeighbor(
	hm *SphereHeightmap,
	topology spatial.Topology,
	coord spatial.Coordinate,
	seaLevel float64,
) bool {
	dirs := []spatial.Direction{spatial.North, spatial.East, spatial.South, spatial.West}

	for _, dir := range dirs {
		neighbor := topology.GetNeighbor(coord, dir)
		if hm.Get(neighbor) < seaLevel {
			return true
		}
	}
	return false
}

// formDelta creates distributary channels and deposits sediment
func formDelta(
	hm *SphereHeightmap,
	topology spatial.Topology,
	mouth spatial.Coordinate,
	flux float64,
	seaLevel float64,
	rng *rand.Rand,
	config DeltaConfig,
) []SphericalRiverPath {
	var paths []SphericalRiverPath
	// Calculate number of distributaries based on flux
	// More flux = more channels
	fluxScale := flux / config.MinFluxForDelta
	numChannels := config.MinChannels + int(fluxScale)
	if numChannels > config.MaxChannels {
		numChannels = config.MaxChannels
	}

	// Find primary direction toward ocean
	primaryDir := findOceanDirection(hm, topology, mouth, seaLevel)
	if primaryDir == "" {
		return paths
	}

	// Get base angles for fan spread
	// Distributaries spread in an arc centered on the primary ocean direction
	baseAngle := directionToAngle(primaryDir)
	spreadAngle := math.Pi / 3.0 // 60 degrees total spread

	// Create each distributary channel
	for i := 0; i < numChannels; i++ {
		// Calculate angle for this channel
		t := float64(i) / float64(numChannels-1) // 0 to 1
		if numChannels == 1 {
			t = 0.5
		}
		angle := baseAngle - spreadAngle/2 + spreadAngle*t

		// Add some randomness
		angle += (rng.Float64() - 0.5) * 0.2

		// Trace distributary and deposit sediment
		path := traceDistributary(hm, topology, mouth, angle, flux/float64(numChannels), seaLevel, config)
		if len(path) > 1 {
			paths = append(paths, SphericalRiverPath{Points: path})
		}
	}

	// Deposit sediment in fan pattern around the delta
	depositSedimentFan(hm, topology, mouth, baseAngle, flux, seaLevel, config)

	return paths
}

// findOceanDirection returns the direction toward deepest water
func findOceanDirection(
	hm *SphereHeightmap,
	topology spatial.Topology,
	coord spatial.Coordinate,
	seaLevel float64,
) spatial.Direction {
	dirs := []spatial.Direction{spatial.North, spatial.East, spatial.South, spatial.West}

	lowestElev := seaLevel
	var bestDir spatial.Direction = ""

	for _, dir := range dirs {
		neighbor := topology.GetNeighbor(coord, dir)
		elev := hm.Get(neighbor)
		if elev < lowestElev {
			lowestElev = elev
			bestDir = dir
		}
	}

	return bestDir
}

// directionToAngle converts cardinal direction to radians
func directionToAngle(dir spatial.Direction) float64 {
	switch dir {
	case spatial.North:
		return math.Pi / 2 // 90 degrees (up)
	case spatial.South:
		return -math.Pi / 2 // -90 degrees (down)
	case spatial.East:
		return 0 // 0 degrees (right)
	case spatial.West:
		return math.Pi // 180 degrees (left)
	default:
		return 0
	}
}

// traceDistributary traces a single distributary channel, depositing sediment
func traceDistributary(
	hm *SphereHeightmap,
	topology spatial.Topology,
	start spatial.Coordinate,
	angle float64,
	flux float64,
	seaLevel float64,
	config DeltaConfig,
) []spatial.Coordinate {
	path := []spatial.Coordinate{start}
	current := start
	sedimentRemaining := flux * config.SedimentPerFlux

	for step := 0; step < config.MaxDeltaRadius*2; step++ {
		// Calculate next position based on angle
		// Convert angle to direction preference
		nextDir := angleToDirection(angle)

		// Move to next cell
		next := topology.GetNeighbor(current, nextDir)
		nextElev := hm.Get(next)

		// Only continue into shallow water
		if nextElev > seaLevel+2 {
			break
		}

		// Deposit sediment (more at start, less as we go out)
		depositAmount := sedimentRemaining * 0.3 // 30% per step
		if depositAmount > 0.1 {
			hm.AddSediment(next, depositAmount)
			sedimentRemaining -= depositAmount
		}

		if sedimentRemaining < 0.1 {
			break
		}

		current = next
		path = append(path, current)

		// Add some wandering to the channel
		angle += (rand.Float64() - 0.5) * 0.3
	}
	return path
}

// angleToDirection converts angle to nearest cardinal direction
func angleToDirection(angle float64) spatial.Direction {
	// Normalize angle to 0-2π
	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}

	// Map to quadrants
	if angle < math.Pi/4 || angle >= 7*math.Pi/4 {
		return spatial.East
	} else if angle < 3*math.Pi/4 {
		return spatial.North
	} else if angle < 5*math.Pi/4 {
		return spatial.West
	} else {
		return spatial.South
	}
}

// depositSedimentFan creates an arc-shaped sediment deposit around the delta
func depositSedimentFan(
	hm *SphereHeightmap,
	topology spatial.Topology,
	center spatial.Coordinate,
	baseAngle float64,
	flux float64,
	seaLevel float64,
	config DeltaConfig,
) {
	// Fill in between distributaries with sediment
	// This creates the characteristic fan shape

	totalSediment := flux * config.SedimentPerFlux * 0.5 // 50% of total for fan fill

	// Radial deposit pattern
	for radius := 1; radius <= config.MaxDeltaRadius; radius++ {
		// Deposit less as we go outward
		radiusFactor := 1.0 - float64(radius)/float64(config.MaxDeltaRadius+1)
		depositPerCell := totalSediment * radiusFactor / float64(radius*4)

		// Sweep arc at this radius
		for i := 0; i < radius*4; i++ {
			t := float64(i) / float64(radius*4)
			angle := baseAngle - math.Pi/3 + (2*math.Pi/3)*t

			// Calculate approximate cell position
			dx := int(math.Round(float64(radius) * math.Cos(angle)))
			dy := int(math.Round(float64(radius) * math.Sin(angle)))

			targetCoord := spatial.Coordinate{
				Face: center.Face,
				X:    center.X + dx,
				Y:    center.Y + dy,
			}

			// Bounds check
			res := topology.Resolution()
			if targetCoord.X < 0 || targetCoord.X >= res || targetCoord.Y < 0 || targetCoord.Y >= res {
				continue
			}

			// Only deposit in water
			if hm.Get(targetCoord) < seaLevel {
				hm.AddSediment(targetCoord, depositPerCell)
			}
		}
	}
}
