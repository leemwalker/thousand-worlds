package underground

import (
	"sync"
)

// ColumnGrid manages all world columns for underground data.
// Thread-safe for concurrent access during simulation.
type ColumnGrid struct {
	Width   int
	Height  int
	columns [][]*WorldColumn // [y][x] matching heightmap orientation
	mu      sync.RWMutex
}

// NewColumnGrid creates a new column grid with the given dimensions.
func NewColumnGrid(width, height int) *ColumnGrid {
	grid := &ColumnGrid{
		Width:   width,
		Height:  height,
		columns: make([][]*WorldColumn, height),
	}

	for y := 0; y < height; y++ {
		grid.columns[y] = make([]*WorldColumn, width)
		for x := 0; x < width; x++ {
			grid.columns[y][x] = &WorldColumn{
				X:         x,
				Y:         y,
				Surface:   0,
				Bedrock:   -10000, // Default 10km depth
				Strata:    []StrataLayer{},
				Voids:     []VoidSpace{},
				Resources: []Deposit{},
			}
		}
	}

	return grid
}

// Get returns the column at (x, y). Returns nil if out of bounds.
func (g *ColumnGrid) Get(x, y int) *WorldColumn {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return nil
	}
	return g.columns[y][x]
}

// Set updates the column at (x, y). No-op if out of bounds.
func (g *ColumnGrid) Set(x, y int, col *WorldColumn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.columns[y][x] = col
	}
}

// InitFromSurface initializes column surface elevations from a 2D elevation grid.
// The elevations slice should be [y*width + x] indexed.
func (g *ColumnGrid) InitFromSurface(elevations []float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			idx := y*g.Width + x
			if idx < len(elevations) {
				g.columns[y][x].Surface = elevations[idx]
			}
		}
	}
}

// AllColumns returns all columns as a flat slice for iteration.
func (g *ColumnGrid) AllColumns() []*WorldColumn {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*WorldColumn, 0, g.Width*g.Height)
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			result = append(result, g.columns[y][x])
		}
	}
	return result
}

// GetResourcesAt returns all resources at position (x,y) within the given depth range.
func (g *ColumnGrid) GetResourcesAt(x, y int, minDepth, maxDepth float64) []Deposit {
	col := g.Get(x, y)
	if col == nil {
		return nil
	}

	var result []Deposit
	for _, d := range col.Resources {
		if d.DepthZ >= minDepth && d.DepthZ <= maxDepth {
			result = append(result, d)
		}
	}
	return result
}

// GetStratumAt returns the stratum containing the given depth at (x,y).
func (g *ColumnGrid) GetStratumAt(x, y int, depth float64) *StrataLayer {
	col := g.Get(x, y)
	if col == nil {
		return nil
	}

	for i := range col.Strata {
		if col.Strata[i].ContainsDepth(depth) {
			return &col.Strata[i]
		}
	}
	return nil
}

// GetVoidsAt returns all voids at (x,y) that contain the given depth.
func (g *ColumnGrid) GetVoidsAt(x, y int, depth float64) []VoidSpace {
	col := g.Get(x, y)
	if col == nil {
		return nil
	}

	var result []VoidSpace
	for _, v := range col.Voids {
		if depth >= v.MinZ && depth <= v.MaxZ {
			result = append(result, v)
		}
	}
	return result
}

// AddStratum adds a new stratum to the column at (x,y).
func (c *WorldColumn) AddStratum(material string, topZ, bottomZ, hardness float64, age int64, porosity float64) {
	c.Strata = append(c.Strata, StrataLayer{
		TopZ:     topZ,
		BottomZ:  bottomZ,
		Material: material,
		Hardness: hardness,
		Age:      age,
		Porosity: porosity,
	})
}

// AddVoid registers a void space in this column.
func (c *WorldColumn) AddVoid(voidID interface{ String() string }, minZ, maxZ float64, voidType string) {
	// Accept uuid.UUID via its String() method to avoid import cycle
	c.Voids = append(c.Voids, VoidSpace{
		MinZ:     minZ,
		MaxZ:     maxZ,
		VoidType: voidType,
	})
}

// AddResource adds a resource deposit to this column.
func (c *WorldColumn) AddResource(resourceType string, depthZ, quantity float64) {
	c.Resources = append(c.Resources, Deposit{
		Type:     resourceType,
		DepthZ:   depthZ,
		Quantity: quantity,
	})
}
