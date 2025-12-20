package underground

import "github.com/google/uuid"

// Cave represents a cave system with connected chambers and passages.
type Cave struct {
	ID           uuid.UUID
	CaveType     string     // "karst", "lava_tube", "sea_cave", "magma_chamber"
	Nodes        []CaveNode // Connected chambers
	Passages     []CaveEdge // Tunnels between nodes
	WaterLevel   float64    // Water table intersection (if flooded)
	FormationAge int64      // Simulation year when cave formed
}

// CaveNode represents a single chamber within a cave system.
type CaveNode struct {
	ID       uuid.UUID
	Position Vector3 // (x, y, z) center of chamber
	Radius   float64 // Horizontal extent
	Height   float64 // Vertical extent
}

// CaveEdge represents a passage connecting two cave nodes.
type CaveEdge struct {
	FromNodeID uuid.UUID
	ToNodeID   uuid.UUID
	Radius     float64 // Passage width/height
}

// NewCave creates a new cave with the given type.
func NewCave(caveType string, formationAge int64) *Cave {
	return &Cave{
		ID:           uuid.New(),
		CaveType:     caveType,
		Nodes:        []CaveNode{},
		Passages:     []CaveEdge{},
		FormationAge: formationAge,
	}
}

// AddNode adds a chamber to the cave system.
func (c *Cave) AddNode(pos Vector3, radius, height float64) uuid.UUID {
	node := CaveNode{
		ID:       uuid.New(),
		Position: pos,
		Radius:   radius,
		Height:   height,
	}
	c.Nodes = append(c.Nodes, node)
	return node.ID
}

// Connect creates a passage between two nodes.
func (c *Cave) Connect(fromID, toID uuid.UUID, passageRadius float64) {
	c.Passages = append(c.Passages, CaveEdge{
		FromNodeID: fromID,
		ToNodeID:   toID,
		Radius:     passageRadius,
	})
}

// GetNode returns a node by ID, or nil if not found.
func (c *Cave) GetNode(id uuid.UUID) *CaveNode {
	for i := range c.Nodes {
		if c.Nodes[i].ID == id {
			return &c.Nodes[i]
		}
	}
	return nil
}

// Bounds returns the approximate bounding box of the cave system.
func (c *Cave) Bounds() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	if len(c.Nodes) == 0 {
		return 0, 0, 0, 0, 0, 0
	}

	minX, maxX = c.Nodes[0].Position.X, c.Nodes[0].Position.X
	minY, maxY = c.Nodes[0].Position.Y, c.Nodes[0].Position.Y
	minZ, maxZ = c.Nodes[0].Position.Z, c.Nodes[0].Position.Z

	for _, node := range c.Nodes {
		if node.Position.X-node.Radius < minX {
			minX = node.Position.X - node.Radius
		}
		if node.Position.X+node.Radius > maxX {
			maxX = node.Position.X + node.Radius
		}
		if node.Position.Y-node.Radius < minY {
			minY = node.Position.Y - node.Radius
		}
		if node.Position.Y+node.Radius > maxY {
			maxY = node.Position.Y + node.Radius
		}
		if node.Position.Z-node.Height/2 < minZ {
			minZ = node.Position.Z - node.Height/2
		}
		if node.Position.Z+node.Height/2 > maxZ {
			maxZ = node.Position.Z + node.Height/2
		}
	}

	return
}

// GetAffectedColumns returns the (x,y) grid positions that this cave intersects.
func (c *Cave) GetAffectedColumns() [][2]int {
	minX, minY, _, maxX, maxY, _ := c.Bounds()

	var columns [][2]int
	for x := int(minX); x <= int(maxX); x++ {
		for y := int(minY); y <= int(maxY); y++ {
			columns = append(columns, [2]int{x, y})
		}
	}
	return columns
}
