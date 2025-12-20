package underground

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// CaveFormationConfig controls cave generation parameters
type CaveFormationConfig struct {
	DissolutionRate   float64 // Base dissolution rate per year
	MinLimestoneDepth float64 // Minimum limestone thickness for caves
	WaterFlowFactor   float64 // Multiplier for water-based dissolution
	CO2Factor         float64 // Atmospheric CO2 multiplier (carbonic acid)
}

// DefaultCaveConfig returns reasonable default cave formation parameters
func DefaultCaveConfig() CaveFormationConfig {
	return CaveFormationConfig{
		DissolutionRate:   0.0001, // 0.1mm per 1000 years
		MinLimestoneDepth: 50,     // At least 50m of limestone needed
		WaterFlowFactor:   1.0,
		CO2Factor:         1.0,
	}
}

// SimulateCaveFormation processes cave formation over time
// Returns newly formed caves
func SimulateCaveFormation(
	columns *ColumnGrid,
	rainfall []float64, // Rainfall per column (0-1 normalized)
	years int64,
	seed int64,
	config CaveFormationConfig,
) []*Cave {
	rng := rand.New(rand.NewSource(seed))
	newCaves := []*Cave{}

	// Track dissolution progress per column
	for _, col := range columns.AllColumns() {
		// Find limestone strata
		for i := range col.Strata {
			stratum := &col.Strata[i]
			if stratum.Material != "limestone" && stratum.Material != "chalk" {
				continue
			}

			// Skip thin limestone layers
			if stratum.Thickness() < config.MinLimestoneDepth {
				continue
			}

			// Calculate dissolution factors
			rainfallIdx := col.Y*columns.Width + col.X
			waterFlow := 0.5 // Default
			if rainfallIdx < len(rainfall) {
				waterFlow = rainfall[rainfallIdx]
			}

			// Dissolution rate depends on:
			// 1. Water flow (rainfall)
			// 2. Porosity of rock
			// 3. CO2 levels (carbonic acid)
			// 4. Time elapsed
			effectiveRate := config.DissolutionRate *
				waterFlow * config.WaterFlowFactor *
				stratum.Porosity *
				config.CO2Factor *
				float64(years)

			// Threshold for cave formation (accumulated dissolution)
			// Higher porosity = caves form faster
			threshold := stratum.Hardness * 10 // Harder rock needs more dissolution

			// Random factor for natural variation
			variation := 0.5 + rng.Float64() // 0.5 to 1.5

			if effectiveRate*variation > threshold {
				// Cave forms!
				cave := createCaveInStratum(col, stratum, rng, years)
				if cave != nil {
					newCaves = append(newCaves, cave)

					// Register void in column
					RegisterCaveInColumn(col, cave)
				}
			}
		}
	}

	// Connect nearby caves into networks
	ConnectAdjacentCaves(newCaves, 30.0) // 30m max connection distance

	return newCaves
}

// createCaveInStratum generates a cave within a limestone stratum
func createCaveInStratum(col *WorldColumn, stratum *StrataLayer, rng *rand.Rand, formationYear int64) *Cave {
	// Cave forms in the middle of the stratum
	centerZ := (stratum.TopZ + stratum.BottomZ) / 2

	// Cave type based on rock
	caveType := "karst"
	if stratum.Material == "chalk" {
		caveType = "sea_cave"
	}

	cave := NewCave(caveType, formationYear)

	// Initial chamber size based on stratum thickness
	thickness := stratum.Thickness()
	radius := math.Min(thickness*0.2, 20) // Max 20m radius
	height := math.Min(thickness*0.3, 15) // Max 15m height

	// Create main chamber
	mainPos := Vector3{
		X: float64(col.X),
		Y: float64(col.Y),
		Z: centerZ,
	}
	mainID := cave.AddNode(mainPos, radius, height)

	// Possibly add secondary chambers (50% chance)
	if rng.Float64() < 0.5 {
		// Secondary chamber offset
		offsetX := (rng.Float64()*2 - 1) * 15 // -15 to +15
		offsetY := (rng.Float64()*2 - 1) * 15
		offsetZ := (rng.Float64()*2 - 1) * 5 // Smaller vertical offset

		secondaryPos := Vector3{
			X: float64(col.X) + offsetX,
			Y: float64(col.Y) + offsetY,
			Z: centerZ + offsetZ,
		}
		secondaryID := cave.AddNode(secondaryPos, radius*0.6, height*0.7)

		// Connect chambers
		passageRadius := math.Min(radius*0.3, 3) // Max 3m passage
		cave.Connect(mainID, secondaryID, passageRadius)
	}

	return cave
}

// RegisterCaveInColumn adds a cave's void space to the column
func RegisterCaveInColumn(col *WorldColumn, cave *Cave) {
	for _, node := range cave.Nodes {
		// Check if this node is in this column (within radius)
		dx := node.Position.X - float64(col.X)
		dy := node.Position.Y - float64(col.Y)
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist <= node.Radius {
			// This column contains part of the cave
			col.Voids = append(col.Voids, VoidSpace{
				VoidID:   cave.ID,
				MinZ:     node.Position.Z - node.Height/2,
				MaxZ:     node.Position.Z + node.Height/2,
				VoidType: cave.CaveType,
			})
		}
	}
}

// RegisterCaveInGrid registers a cave's voids in all affected columns
func RegisterCaveInGrid(grid *ColumnGrid, cave *Cave) {
	affectedCols := cave.GetAffectedColumns()
	for _, coords := range affectedCols {
		col := grid.Get(coords[0], coords[1])
		if col != nil {
			RegisterCaveInColumn(col, cave)
		}
	}
}

// ConnectAdjacentCaves links nearby caves into networks
func ConnectAdjacentCaves(caves []*Cave, maxDistance float64) {
	for i := 0; i < len(caves); i++ {
		for j := i + 1; j < len(caves); j++ {
			cave1, cave2 := caves[i], caves[j]

			// Find closest nodes between caves
			minDist := math.MaxFloat64
			var closest1, closest2 *CaveNode

			for ni := range cave1.Nodes {
				for nj := range cave2.Nodes {
					n1, n2 := &cave1.Nodes[ni], &cave2.Nodes[nj]
					dist := distance3D(n1.Position, n2.Position)
					if dist < minDist {
						minDist = dist
						closest1, closest2 = n1, n2
					}
				}
			}

			// Connect if within range
			if minDist <= maxDistance && closest1 != nil && closest2 != nil {
				// Merge cave2 into cave1
				mergeCaves(cave1, cave2, closest1.ID, closest2.ID)
			}
		}
	}
}

// mergeCaves combines two caves by adding cave2's nodes to cave1
func mergeCaves(cave1, cave2 *Cave, connectFrom, connectTo uuid.UUID) {
	// Add all nodes from cave2 to cave1
	nodeIDMap := make(map[uuid.UUID]uuid.UUID) // old ID -> new ID
	for _, node := range cave2.Nodes {
		newID := cave1.AddNode(node.Position, node.Radius, node.Height)
		nodeIDMap[node.ID] = newID
	}

	// Add passages from cave2
	for _, edge := range cave2.Passages {
		newFrom := nodeIDMap[edge.FromNodeID]
		newTo := nodeIDMap[edge.ToNodeID]
		cave1.Connect(newFrom, newTo, edge.Radius)
	}

	// Connect the two cave systems
	newConnectTo := nodeIDMap[connectTo]
	passageRadius := 2.0 // Default narrow passage
	cave1.Connect(connectFrom, newConnectTo, passageRadius)
}

func distance3D(a, b Vector3) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
