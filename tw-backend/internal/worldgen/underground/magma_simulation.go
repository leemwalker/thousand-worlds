package underground

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// MagmaChamber represents an underground magma reservoir
type MagmaChamber struct {
	ID           uuid.UUID
	Center       Vector3     // (x, y, z) center position
	Volume       float64     // cubic meters
	Temperature  float64     // Kelvin
	Pressure     float64     // Relative pressure (0-100)
	Viscosity    float64     // 0-1 (0=fluid, 1=viscous)
	Age          int64       // Simulation years since formation
	LastEruption int64       // Year of last eruption
	Connected    []uuid.UUID // Lava tubes connecting to surface
}

// TectonicBoundary represents a plate boundary for magma generation
type TectonicBoundary struct {
	X, Y         int
	BoundaryType string  // "convergent", "divergent", "transform"
	Intensity    float64 // 0-1, how active the boundary is
}

// MagmaSimulationConfig controls magma behavior
type MagmaSimulationConfig struct {
	CoolingRatePerYear    float64 // Degrees K per year
	EruptionThreshold     float64 // Pressure at which eruption occurs
	MagmaChamberRadius    float64 // Default chamber radius
	LavaTubeFormationProb float64 // Probability of lava tube when drained
}

// DefaultMagmaConfig returns reasonable default parameters
func DefaultMagmaConfig() MagmaSimulationConfig {
	return MagmaSimulationConfig{
		CoolingRatePerYear:    0.01, // Very slow cooling
		EruptionThreshold:     80.0, // Erupts at 80+ pressure
		MagmaChamberRadius:    20.0, // 20m radius
		LavaTubeFormationProb: 0.7,  // 70% chance of lava tube
	}
}

// SimulateMagmaChambers processes magma chamber evolution
// Returns: erupted chambers, new lava tubes
func SimulateMagmaChambers(
	columns *ColumnGrid,
	chambers []*MagmaChamber,
	boundaries []TectonicBoundary,
	years int64,
	seed int64,
	config MagmaSimulationConfig,
) (erupted []*MagmaChamber, newTubes []*Cave, collapsedCaves []*Cave) {
	rng := rand.New(rand.NewSource(seed))

	erupted = []*MagmaChamber{}
	newTubes = []*Cave{}
	collapsedCaves = []*Cave{}

	// 1. Process existing chambers
	for _, chamber := range chambers {
		// Age the chamber
		chamber.Age += years

		// Cool the magma
		chamber.Temperature -= config.CoolingRatePerYear * float64(years)
		if chamber.Temperature < 0 {
			chamber.Temperature = 0
		}

		// Check if solidified
		if chamber.Temperature < 1000 {
			// Magma has solidified - determine if cave forms or collapses
			resultCave := processSolidifiedChamber(columns, chamber, rng, config)
			if resultCave != nil {
				if resultCave.CaveType == "collapsed" {
					collapsedCaves = append(collapsedCaves, resultCave)
				} else {
					newTubes = append(newTubes, resultCave)
				}
			}
			continue
		}

		// Increase pressure from mantle heat
		chamber.Pressure += 0.001 * float64(years) // Slow pressure buildup

		// Check for eruption
		if chamber.Pressure >= config.EruptionThreshold {
			erupted = append(erupted, chamber)

			// Eruption relieves pressure and drains magma
			chamber.Pressure *= 0.3 // 70% pressure release
			chamber.Volume *= 0.5   // 50% volume loss
			chamber.LastEruption = chamber.Age

			// Possible lava tube formation
			if rng.Float64() < config.LavaTubeFormationProb {
				tube := createLavaTube(columns, chamber, rng)
				if tube != nil {
					newTubes = append(newTubes, tube)
					chamber.Connected = append(chamber.Connected, tube.ID)
				}
			}
		}
	}

	// 2. Generate new magma chambers at active boundaries
	for _, boundary := range boundaries {
		// Only active boundaries generate new chambers
		if boundary.Intensity < 0.5 {
			continue
		}

		// Probability based on boundary type and intensity
		prob := 0.0
		switch boundary.BoundaryType {
		case "divergent":
			prob = 0.02 * boundary.Intensity // Mid-ocean ridges
		case "convergent":
			prob = 0.03 * boundary.Intensity // Subduction zones
		case "transform":
			prob = 0.005 * boundary.Intensity // Rare at transform
		}

		// Scale probability with time, but cap at 1.0 to prevent runaway with large timesteps
		scaledProb := math.Min(1.0, prob*float64(years)/1000)
		if rng.Float64() < scaledProb {
			newChamber := createMagmaChamber(columns, boundary, rng)
			if newChamber != nil {
				chambers = append(chambers, newChamber)
				// Register in column
				col := columns.Get(boundary.X, boundary.Y)
				if col != nil {
					col.Magma = &MagmaInfo{
						TopZ:        newChamber.Center.Z + config.MagmaChamberRadius,
						BottomZ:     newChamber.Center.Z - config.MagmaChamberRadius,
						Temperature: newChamber.Temperature,
						Pressure:    newChamber.Pressure,
						Viscosity:   newChamber.Viscosity,
					}
				}
			}
		}
	}

	return erupted, newTubes, collapsedCaves
}

// processSolidifiedChamber handles a cooled magma chamber
func processSolidifiedChamber(columns *ColumnGrid, chamber *MagmaChamber, rng *rand.Rand, config MagmaSimulationConfig) *Cave {
	col := columns.Get(int(chamber.Center.X), int(chamber.Center.Y))
	if col == nil {
		return nil
	}

	// Check surrounding rock hardness
	// Harder rock = chamber becomes permanent cave
	// Softer rock = chamber collapses
	avgHardness := 5.0 // Default
	for _, stratum := range col.Strata {
		if stratum.ContainsDepth(chamber.Center.Z) {
			avgHardness = stratum.Hardness
			break
		}
	}

	if avgHardness > 6 {
		// Strong rock - creates a permanent cave from drained chamber
		cave := NewCave("magma_chamber", chamber.Age)
		cave.AddNode(chamber.Center, config.MagmaChamberRadius, config.MagmaChamberRadius*0.8)
		RegisterCaveInGrid(columns, cave)

		// Clear magma from column
		col.Magma = nil

		return cave
	} else {
		// Weak rock - chamber collapses
		cave := NewCave("collapsed", chamber.Age)
		cave.AddNode(chamber.Center, config.MagmaChamberRadius*0.5, config.MagmaChamberRadius*0.3)

		// Clear magma from column
		col.Magma = nil

		return cave
	}
}

// createLavaTube generates a lava tube from chamber to surface
func createLavaTube(columns *ColumnGrid, chamber *MagmaChamber, rng *rand.Rand) *Cave {
	col := columns.Get(int(chamber.Center.X), int(chamber.Center.Y))
	if col == nil {
		return nil
	}

	tube := NewCave("lava_tube", chamber.Age)

	// Create nodes from chamber up to surface
	chamberZ := chamber.Center.Z
	surfaceZ := col.Surface

	// Tube meanders upward
	currentX := chamber.Center.X
	currentY := chamber.Center.Y
	currentZ := chamberZ

	nodeCount := int((surfaceZ - chamberZ) / 100) // One node per 100m
	if nodeCount < 2 {
		nodeCount = 2
	}
	if nodeCount > 10 {
		nodeCount = 10
	}

	zStep := (surfaceZ - chamberZ) / float64(nodeCount)
	var prevNodeID uuid.UUID

	for i := 0; i < nodeCount; i++ {
		// Slight horizontal drift
		currentX += (rng.Float64()*2 - 1) * 5
		currentY += (rng.Float64()*2 - 1) * 5
		currentZ += zStep

		// Radius decreases toward surface
		radius := 5.0 * (1.0 - float64(i)/float64(nodeCount)*0.5)

		pos := Vector3{X: currentX, Y: currentY, Z: currentZ}
		nodeID := tube.AddNode(pos, radius, radius*1.5)

		if i > 0 {
			tube.Connect(prevNodeID, nodeID, radius*0.5)
		}
		prevNodeID = nodeID
	}

	// Surface vent opening
	ventPos := Vector3{X: currentX, Y: currentY, Z: surfaceZ}
	ventID := tube.AddNode(ventPos, 3.0, 5.0)
	tube.Connect(prevNodeID, ventID, 2.0)

	RegisterCaveInGrid(columns, tube)
	return tube
}

// createMagmaChamber generates a new chamber at a boundary
func createMagmaChamber(columns *ColumnGrid, boundary TectonicBoundary, rng *rand.Rand) *MagmaChamber {
	col := columns.Get(boundary.X, boundary.Y)
	if col == nil {
		return nil
	}

	// Chamber forms deep underground
	depth := col.Surface - 2000 - rng.Float64()*3000 // 2-5km deep

	return &MagmaChamber{
		ID: uuid.New(),
		Center: Vector3{
			X: float64(boundary.X),
			Y: float64(boundary.Y),
			Z: depth,
		},
		Volume:      1000000 + rng.Float64()*9000000, // 1-10 million m³
		Temperature: 1200 + rng.Float64()*300,        // 1200-1500K
		Pressure:    20 + rng.Float64()*40,           // Start at 20-60
		Viscosity:   0.3 + rng.Float64()*0.4,         // 0.3-0.7
		Age:         0,
	}
}

// GetTectonicBoundaries extracts active boundaries from plate data
// plateCentroids: slice of (x, y) plate center positions
// plateMovements: corresponding movement vectors
func GetTectonicBoundaries(width, height int, plateCentroids []Vector3, plateMovements []Vector3) []TectonicBoundary {
	boundaries := []TectonicBoundary{}

	// Simple Voronoi-based boundary detection
	// For each cell, check if neighbors belong to different plates
	// This is a simplified version - real implementation would use plate IDs

	plateMap := make([]int, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Find closest plate centroid
			minDist := math.MaxFloat64
			closestPlate := 0
			for i, centroid := range plateCentroids {
				dist := math.Sqrt(math.Pow(float64(x)-centroid.X, 2) + math.Pow(float64(y)-centroid.Y, 2))
				if dist < minDist {
					minDist = dist
					closestPlate = i
				}
			}
			plateMap[y*width+x] = closestPlate
		}
	}

	// Find boundary cells
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			currentPlate := plateMap[y*width+x]

			// Check neighbors
			neighbors := []int{
				plateMap[y*width+x+1],
				plateMap[y*width+x-1],
				plateMap[(y+1)*width+x],
				plateMap[(y-1)*width+x],
			}

			for _, neighborPlate := range neighbors {
				if neighborPlate != currentPlate && neighborPlate < len(plateCentroids) && currentPlate < len(plateCentroids) {
					// This is a boundary cell
					// Determine boundary type from relative movement
					mv1 := plateMovements[currentPlate]
					mv2 := plateMovements[neighborPlate]

					// Relative movement toward/away from each other
					dx := plateCentroids[neighborPlate].X - plateCentroids[currentPlate].X
					dy := plateCentroids[neighborPlate].Y - plateCentroids[currentPlate].Y
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist == 0 {
						continue
					}
					dx, dy = dx/dist, dy/dist

					// Relative approach velocity
					relVel := (mv2.X-mv1.X)*dx + (mv2.Y-mv1.Y)*dy

					boundaryType := "transform"
					intensity := 0.5

					if relVel < -0.2 {
						boundaryType = "convergent"
						intensity = math.Min(1.0, math.Abs(relVel))
					} else if relVel > 0.2 {
						boundaryType = "divergent"
						intensity = math.Min(1.0, relVel)
					}

					boundaries = append(boundaries, TectonicBoundary{
						X:            x,
						Y:            y,
						BoundaryType: boundaryType,
						Intensity:    intensity,
					})
					break // Only add once per cell
				}
			}
		}
	}

	return boundaries
}
