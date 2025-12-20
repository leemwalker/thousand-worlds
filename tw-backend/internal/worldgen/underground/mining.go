package underground

import (
	"fmt"

	"github.com/google/uuid"
)

// MiningTool represents a tool used for mining
type MiningTool struct {
	Name        string
	MaxHardness float64 // Maximum rock hardness this tool can mine
	Speed       float64 // Mining speed multiplier
	DepthLimit  float64 // Maximum depth (0 = unlimited)
	Durability  int     // Uses before breaking (0 = infinite)
}

// Standard mining tools with hardness requirements
var StandardTools = map[string]MiningTool{
	"hands":        {Name: "Bare Hands", MaxHardness: 1, Speed: 0.1, DepthLimit: 5, Durability: 0},
	"wooden_pick":  {Name: "Wooden Pickaxe", MaxHardness: 2, Speed: 0.5, DepthLimit: 50, Durability: 100},
	"stone_pick":   {Name: "Stone Pickaxe", MaxHardness: 4, Speed: 0.8, DepthLimit: 200, Durability: 200},
	"iron_pick":    {Name: "Iron Pickaxe", MaxHardness: 6, Speed: 1.0, DepthLimit: 0, Durability: 500},
	"steel_pick":   {Name: "Steel Pickaxe", MaxHardness: 8, Speed: 1.5, DepthLimit: 0, Durability: 1000},
	"diamond_pick": {Name: "Diamond Pickaxe", MaxHardness: 10, Speed: 2.0, DepthLimit: 0, Durability: 2000},
}

// MiningResult contains the outcome of a mining action
type MiningResult struct {
	Success       bool
	Reason        string     // Failure reason if not successful
	ResourceFound *Deposit   // Resource extracted (if any)
	ToolDamage    int        // Durability lost
	TimeRequired  float64    // Time in seconds to complete
	VoidCreated   *VoidSpace // If mining created a tunnel/burrow
}

// CanMine checks if a tool can mine at a specific location
func CanMine(tool MiningTool, stratum *StrataLayer, depth float64) (bool, string) {
	if stratum == nil {
		return false, "no rock at this location"
	}

	if tool.MaxHardness < stratum.Hardness {
		return false, fmt.Sprintf("tool too weak (hardness %.1f vs rock %.1f)", tool.MaxHardness, stratum.Hardness)
	}

	if tool.DepthLimit > 0 && depth > tool.DepthLimit {
		return false, fmt.Sprintf("tool cannot reach this depth (limit %.0fm, depth %.0fm)", tool.DepthLimit, depth)
	}

	return true, ""
}

// Mine attempts to mine at a specific location
func Mine(
	col *WorldColumn,
	depth float64,
	tool MiningTool,
	createTunnel bool,
) MiningResult {
	// Find stratum at depth
	var stratum *StrataLayer
	for i := range col.Strata {
		if col.Strata[i].ContainsDepth(depth) {
			stratum = &col.Strata[i]
			break
		}
	}

	// Check if can mine
	canMine, reason := CanMine(tool, stratum, col.Surface-depth)
	if !canMine {
		return MiningResult{
			Success: false,
			Reason:  reason,
		}
	}

	// Check for void at this depth (can't mine in empty space)
	for _, v := range col.Voids {
		if depth >= v.MinZ && depth <= v.MaxZ {
			return MiningResult{
				Success: false,
				Reason:  "already a void at this location",
			}
		}
	}

	// Calculate mining time based on hardness and tool speed
	baseTime := 5.0 // 5 seconds base
	timeRequired := baseTime * (stratum.Hardness / tool.Speed)

	// Tool damage based on hardness
	toolDamage := int(stratum.Hardness)

	result := MiningResult{
		Success:      true,
		ToolDamage:   toolDamage,
		TimeRequired: timeRequired,
	}

	// Check for resource at this depth
	for i := range col.Resources {
		res := &col.Resources[i]
		if res.DepthZ >= depth-1 && res.DepthZ <= depth+1 && res.Quantity > 0 {
			// Found a resource!
			result.ResourceFound = res
			res.Discovered = true
			break
		}
	}

	// Create tunnel/burrow if requested
	if createTunnel {
		void := VoidSpace{
			VoidID:   uuid.New(),
			MinZ:     depth - 1, // 2m tall tunnel
			MaxZ:     depth + 1,
			VoidType: "mine",
		}
		col.Voids = append(col.Voids, void)
		result.VoidCreated = &void
	}

	return result
}

// ExtractResource extracts a quantity of resource from a deposit
func ExtractResource(deposit *Deposit, quantity float64) (float64, bool) {
	if deposit == nil || deposit.Quantity <= 0 {
		return 0, false
	}

	extracted := quantity
	if extracted > deposit.Quantity {
		extracted = deposit.Quantity
	}

	deposit.Quantity -= extracted
	return extracted, true
}

// Burrow represents a creature-dug underground passage
type Burrow struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID // Creature that dug it
	Entrance   Vector3   // Entry point
	Chambers   []BurrowChamber
	Tunnels    []BurrowTunnel
	TotalDepth float64
}

// BurrowChamber is a room in a burrow
type BurrowChamber struct {
	ID       uuid.UUID
	Position Vector3
	Radius   float64
	Purpose  string // "nest", "storage", "den"
}

// BurrowTunnel connects chambers or entrance to chamber
type BurrowTunnel struct {
	FromID uuid.UUID // Chamber or entrance ID
	ToID   uuid.UUID // Chamber ID
	Length float64
	Radius float64
}

// CreateBurrow creates a new burrow for a creature
func CreateBurrow(
	col *WorldColumn,
	ownerID uuid.UUID,
	entranceZ float64,
	depth float64,
	chamberCount int,
) (*Burrow, error) {
	if col == nil {
		return nil, fmt.Errorf("invalid column")
	}

	// Check if depth is achievable (soil layer must extend that deep)
	canDig := false
	for _, stratum := range col.Strata {
		if stratum.Hardness <= 3 && stratum.ContainsDepth(entranceZ-depth) {
			canDig = true
			break
		}
	}
	if !canDig {
		return nil, fmt.Errorf("ground too hard for burrowing")
	}

	burrow := &Burrow{
		ID:         uuid.New(),
		OwnerID:    ownerID,
		Entrance:   Vector3{X: float64(col.X), Y: float64(col.Y), Z: entranceZ},
		Chambers:   []BurrowChamber{},
		Tunnels:    []BurrowTunnel{},
		TotalDepth: depth,
	}

	// Create chambers at increasing depths
	chamberDepth := depth / float64(chamberCount+1)
	var prevID uuid.UUID = burrow.ID // Use burrow ID as entrance ID

	for i := 0; i < chamberCount; i++ {
		currentDepth := entranceZ - chamberDepth*float64(i+1)

		chamber := BurrowChamber{
			ID:       uuid.New(),
			Position: Vector3{X: float64(col.X), Y: float64(col.Y), Z: currentDepth},
			Radius:   0.5 + float64(i)*0.2, // Deeper chambers slightly larger
			Purpose:  "den",
		}
		if i == chamberCount-1 {
			chamber.Purpose = "nest" // Deepest is the nest
		}
		burrow.Chambers = append(burrow.Chambers, chamber)

		// Connect to previous
		burrow.Tunnels = append(burrow.Tunnels, BurrowTunnel{
			FromID: prevID,
			ToID:   chamber.ID,
			Length: chamberDepth,
			Radius: 0.3,
		})
		prevID = chamber.ID
	}

	// Register burrow voids in column
	for _, chamber := range burrow.Chambers {
		col.Voids = append(col.Voids, VoidSpace{
			VoidID:   chamber.ID,
			MinZ:     chamber.Position.Z - chamber.Radius,
			MaxZ:     chamber.Position.Z + chamber.Radius,
			VoidType: "burrow",
		})
	}

	return burrow, nil
}

// DigTunnel creates a player-dug tunnel between two points
func DigTunnel(
	grid *ColumnGrid,
	startX, startY int,
	startZ float64,
	endX, endY int,
	endZ float64,
	tool MiningTool,
) ([]MiningResult, error) {
	results := []MiningResult{}

	// Simple line interpolation between start and end
	steps := max(abs(endX-startX), abs(endY-startY))
	if steps == 0 {
		steps = 1
	}

	dx := float64(endX-startX) / float64(steps)
	dy := float64(endY-startY) / float64(steps)
	dz := (endZ - startZ) / float64(steps)

	for i := 0; i <= steps; i++ {
		x := startX + int(float64(i)*dx)
		y := startY + int(float64(i)*dy)
		z := startZ + float64(i)*dz

		col := grid.Get(x, y)
		if col == nil {
			continue
		}

		result := Mine(col, z, tool, true)
		results = append(results, result)

		if !result.Success {
			return results, fmt.Errorf("mining failed at (%d,%d,%.0f): %s", x, y, z, result.Reason)
		}
	}

	return results, nil
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CalculateMiningSpeed returns the mining speed multiplier based on tool vs rock hardness.
// Returns 0 if tool cannot mine the rock (hardness too high).
// Returns speed from 0 to tool.Speed, scaled by hardness difference.
func CalculateMiningSpeed(tool MiningTool, rockHardness float64) float64 {
	if tool.MaxHardness < rockHardness {
		return 0 // Cannot mine
	}

	// Speed scales with how much harder the tool is vs rock
	hardnessAdvantage := tool.MaxHardness - rockHardness
	baseFactor := hardnessAdvantage / tool.MaxHardness

	// Minimum 10% speed if tool can mine at all
	if baseFactor < 0.1 {
		baseFactor = 0.1
	}

	return tool.Speed * baseFactor
}

// StrataContext provides parameters for generating appropriate strata
type StrataContext struct {
	IsVolcanic bool
	IsAncient  bool
	IsOceanic  bool
	Elevation  float64
}

// GenerateStrataForContext creates geologically appropriate strata layers
func GenerateStrataForContext(ctx StrataContext) []StrataLayer {
	strata := []StrataLayer{}

	if ctx.IsVolcanic {
		// Volcanic regions: basalt and volcanic rock
		strata = append(strata, StrataLayer{
			TopZ: 0, BottomZ: -50, Material: "basalt", Hardness: 6.0, Porosity: 0.1,
		})
		strata = append(strata, StrataLayer{
			TopZ: -50, BottomZ: -500, Material: "gabbro", Hardness: 7.0, Porosity: 0.05,
		})
	} else if ctx.IsAncient {
		// Ancient cratons: granite and metamorphic basement
		strata = append(strata, StrataLayer{
			TopZ: 0, BottomZ: -30, Material: "soil", Hardness: 1.5, Porosity: 0.4,
		})
		strata = append(strata, StrataLayer{
			TopZ: -30, BottomZ: -200, Material: "granite", Hardness: 7.5, Porosity: 0.01,
		})
		strata = append(strata, StrataLayer{
			TopZ: -200, BottomZ: -5000, Material: "gneiss", Hardness: 8.0, Porosity: 0.005,
		})
	} else {
		// Default: sedimentary basin
		strata = append(strata, StrataLayer{
			TopZ: 0, BottomZ: -20, Material: "soil", Hardness: 1.0, Porosity: 0.5,
		})
		strata = append(strata, StrataLayer{
			TopZ: -20, BottomZ: -200, Material: "sandstone", Hardness: 4.0, Porosity: 0.2,
		})
		strata = append(strata, StrataLayer{
			TopZ: -200, BottomZ: -800, Material: "limestone", Hardness: 4.0, Porosity: 0.25,
		})
	}

	return strata
}
