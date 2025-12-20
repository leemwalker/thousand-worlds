package underground

import "github.com/google/uuid"

// WorldColumn stores underground data for a single (x,y) position.
// Resolution is 1:1 with the heightmap.
type WorldColumn struct {
	X, Y      int
	Surface   float64       // Current terrain elevation (from heightmap)
	Bedrock   float64       // Deepest mineable depth (extends with better tools)
	Strata    []StrataLayer // Geological layers from surface down
	Voids     []VoidSpace   // Caves/tunnels intersecting this column
	Resources []Deposit     // Minerals, fossils, oil at various depths
	Magma     *MagmaInfo    // Active magma (nil if none)
}

// StrataLayer represents a geological layer at a specific depth range.
type StrataLayer struct {
	TopZ     float64 // Top of this layer (higher = closer to surface)
	BottomZ  float64 // Bottom of this layer (lower = deeper)
	Material string  // "soil", "limestone", "granite", "basalt", etc.
	Hardness float64 // 1-10, affects mining speed and tool requirements
	Age      int64   // Simulation years since formation
	Porosity float64 // 0-1, affects water flow and cave formation potential
}

// Thickness returns the vertical extent of this stratum.
func (s *StrataLayer) Thickness() float64 {
	return s.TopZ - s.BottomZ
}

// ContainsDepth returns true if the given depth falls within this stratum.
func (s *StrataLayer) ContainsDepth(z float64) bool {
	return z <= s.TopZ && z >= s.BottomZ
}

// VoidSpace represents an empty space (cave/tunnel) at this column.
type VoidSpace struct {
	VoidID   uuid.UUID // Reference to Cave or Tunnel entity
	MinZ     float64   // Bottom of void at this column
	MaxZ     float64   // Top of void at this column
	VoidType string    // "cave", "lava_tube", "burrow", "mine"
}

// Height returns the vertical extent of this void.
func (v *VoidSpace) Height() float64 {
	return v.MaxZ - v.MinZ
}

// Deposit represents a buried resource at a specific depth.
type Deposit struct {
	ID         uuid.UUID
	Type       string         // "iron", "gold", "coal", "fossil", "oil"
	DepthZ     float64        // Center depth
	Quantity   float64        // Remaining amount (units vary by type)
	Discovered bool           // Player has found this
	Source     *OrganicSource // For fossils/oil, tracks origin
}

// OrganicSource tracks the origin of organic deposits (fossils, oil).
type OrganicSource struct {
	OriginalEntityID uuid.UUID
	Species          string
	DeathYear        int64
	BurialYear       int64
}

// Age returns years since the organism died.
func (o *OrganicSource) Age(currentYear int64) int64 {
	return currentYear - o.DeathYear
}

// BurialDuration returns years since burial.
func (o *OrganicSource) BurialDuration(currentYear int64) int64 {
	return currentYear - o.BurialYear
}

// MagmaInfo stores active magma data for volcanic columns.
type MagmaInfo struct {
	TopZ        float64 // Top of magma layer
	BottomZ     float64 // Bottom of magma layer
	Temperature float64 // Kelvin (affects surrounding rock, eruption potential)
	Pressure    float64 // Affects eruption potential
	Viscosity   float64 // Affects lava tube formation (low = flows easily)
}

// IsSolidified returns true if the magma has cooled below 1000K.
func (m *MagmaInfo) IsSolidified() bool {
	return m.Temperature < 1000
}

// Vector3 represents a 3D position.
type Vector3 struct {
	X, Y, Z float64
}
