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

// SimulateFossilFormation buries organic matter and creates fossil deposits over time.
// Fossils form after 10,000+ years of burial in sediment.
// Returns the count of fossil deposits created.
func SimulateFossilFormation(columns *ColumnGrid, deadOrganism Deposit, years int64) int {
	// Fossils require minimum burial time
	const minFossilYears int64 = 10_000

	if years < minFossilYears {
		return 0
	}

	// Check for proper burial conditions
	if deadOrganism.Type != "organic" || deadOrganism.Source == nil {
		return 0
	}

	// Fossil forms if burial conditions met
	// More time = more fossils (partial preservation)
	fossilCount := 1
	if years > 100_000 {
		fossilCount = 2
	}
	if years > 1_000_000 {
		fossilCount = 3
	}

	// Place fossil in the grid at the organism location (center of grid for now)
	width := columns.Width
	height := columns.Height
	col := columns.Get(width/2, height/2)
	if col != nil {
		// Convert organic deposit to fossil
		fossilDeposit := Deposit{
			ID:       deadOrganism.ID,
			Type:     "fossil",
			DepthZ:   deadOrganism.DepthZ,
			Quantity: deadOrganism.Quantity * 0.1, // Only 10% preserved
			Source:   deadOrganism.Source,
		}
		col.Resources = append(col.Resources, fossilDeposit)
	}

	return fossilCount
}

// SimulateOilFormation converts ancient organic deposits into oil over geological time.
// Oil requires organic source, heat (depth > 2km), and cap rock (impermeable layer).
// Takes millions of years to form.
func SimulateOilFormation(columns *ColumnGrid, years int64) []*Deposit {
	// Oil formation requires millions of years
	const minOilYears int64 = 10_000_000 // 10 million years

	if years < minOilYears {
		return nil
	}

	oilDeposits := make([]*Deposit, 0)

	// Search grid for organic deposits with proper conditions
	for _, col := range columns.AllColumns() {
		if col == nil {
			continue
		}

		// Check for cap rock (impermeable layer at surface)
		hasCapRock := false
		for _, stratum := range col.Strata {
			if stratum.Porosity < 0.1 && stratum.TopZ >= col.Surface-100 {
				hasCapRock = true
				break
			}
		}

		if !hasCapRock {
			continue
		}

		// Look for organic deposits deep enough for thermal maturation
		for i, deposit := range col.Resources {
			if deposit.Type == "organic" && deposit.DepthZ <= -100 {
				// Convert to oil
				oilDeposit := &Deposit{
					ID:       deposit.ID,
					Type:     "oil",
					DepthZ:   deposit.DepthZ,
					Quantity: deposit.Quantity * 0.5, // 50% conversion efficiency
					Source:   deposit.Source,
				}
				oilDeposits = append(oilDeposits, oilDeposit)

				// Mark original deposit as converted
				col.Resources[i].Type = "depleted"
			}
		}
	}

	if len(oilDeposits) == 0 {
		return nil
	}
	return oilDeposits
}

// CalculateRoofStability assesses cave roof collapse probability.
// Depends on span (radius), height, and rock type.
// Returns 0.0 (unstable) to 1.0 (stable).
func CalculateRoofStability(cave *Cave, nodeID uuid.UUID) float64 {
	// Find the node in the cave
	var targetNode *CaveNode
	for i := range cave.Nodes {
		if cave.Nodes[i].ID == nodeID {
			targetNode = &cave.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		return 0.0
	}

	// Stability decreases with larger spans and heights
	// Based on rock mechanics: unsupported spans > 10m become unstable
	// Max safe span is roughly 10m for limestone/granite

	// Span factor: larger span = less stable
	maxSafeSpan := 10.0                                 // meters
	spanFactor := maxSafeSpan / (targetNode.Radius + 1) // +1 to avoid division by zero
	if spanFactor > 1.0 {
		spanFactor = 1.0
	}

	// Height factor: taller chambers are less stable
	maxSafeHeight := 10.0 // meters
	heightFactor := maxSafeHeight / (targetNode.Height + 1)
	if heightFactor > 1.0 {
		heightFactor = 1.0
	}

	// Combined stability (geometric mean)
	stability := math.Sqrt(spanFactor * heightFactor)

	// Ensure minimum positive value
	if stability < 0.01 {
		stability = 0.01
	}

	return stability
}

// SimulateRockCycle transforms rock types over geological time.
// RED STATE: Returns 0 - not yet implemented.
func SimulateRockCycle(columns *ColumnGrid, years int64, temperature, pressure float64) int {
	// TODO: Implement rock cycle
	// Sedimentary -> Metamorphic with heat/pressure
	// Returns number of transformed strata
	return 0
}

// SimulateBurrowCreation allows creatures to create tunnels.
// Checks tool strength against material hardness.
func SimulateBurrowCreation(col *WorldColumn, depth float64, toolStrength float64) *VoidSpace {
	if col == nil {
		return nil
	}

	// Find the stratum at the given depth
	for _, stratum := range col.Strata {
		if depth >= stratum.BottomZ && depth <= stratum.TopZ {
			// Check if tool is strong enough to dig
			if toolStrength >= stratum.Hardness {
				// Create burrow void
				return &VoidSpace{
					VoidID:   uuid.New(),
					MinZ:     depth - 1.0, // 2m tall burrow
					MaxZ:     depth + 1.0,
					VoidType: "burrow",
				}
			}
			// Tool not strong enough for this rock
			return nil
		}
	}

	// No stratum at this depth - assume air/void, burrow succeeds
	return &VoidSpace{
		VoidID:   uuid.New(),
		MinZ:     depth - 1.0,
		MaxZ:     depth + 1.0,
		VoidType: "burrow",
	}
}

// PunctureAquifer simulates breaching an underground water source.
// Returns water flow rate based on porosity of surrounding rock.
func PunctureAquifer(col *WorldColumn, depth float64) float64 {
	if col == nil {
		return 0.0
	}

	// Find the stratum at the given depth
	for _, stratum := range col.Strata {
		if depth >= stratum.BottomZ && depth <= stratum.TopZ {
			// Water flow rate depends on porosity
			// Porosity 0.0 = no flow, 1.0 = maximum flow
			// Base flow rate is 100 liters/minute, scaled by porosity
			baseFlowRate := 100.0 // liters per minute
			flowRate := baseFlowRate * stratum.Porosity

			return flowRate
		}
	}

	// No stratum at depth - return small default flow
	return 10.0 // Small trickle
}
