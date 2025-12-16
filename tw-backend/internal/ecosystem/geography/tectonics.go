// Package geography provides tectonic plate simulation for continental configuration.
// This replaces the single float ContinentalFragmentation with a realistic system.
package geography

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// PlateType represents the type of tectonic plate
type PlateType string

const (
	PlateContinental PlateType = "continental" // Lighter, thicker - forms landmasses
	PlateOceanic     PlateType = "oceanic"     // Denser, thinner - forms ocean floor
	PlateMixed       PlateType = "mixed"       // Has both continental and oceanic regions
)

// TectonicPlate represents a single tectonic plate
type TectonicPlate struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"` // e.g., "Pacific", "North American"
	Type        PlateType `json:"plate_type"`
	VelocityX   float32   `json:"velocity_x"`   // Movement in X direction (cm/year scaled)
	VelocityY   float32   `json:"velocity_y"`   // Movement in Y direction (cm/year scaled)
	RotationZ   float32   `json:"rotation_z"`   // Rotation rate (degrees per million years)
	LandmassPct float32   `json:"landmass_pct"` // Percentage of plate that is land (0.0-1.0)
	CellCount   int       `json:"cell_count"`   // Number of hex cells this plate covers
	OriginYear  int64     `json:"origin_year"`  // Year plate was formed/split
}

// NewTectonicPlate creates a new tectonic plate
func NewTectonicPlate(name string, plateType PlateType, rng *rand.Rand) *TectonicPlate {
	// Random initial velocity
	velX := (rng.Float32() - 0.5) * 2.0 // -1.0 to 1.0
	velY := (rng.Float32() - 0.5) * 2.0
	rot := (rng.Float32() - 0.5) * 0.5 // Slower rotation

	// Landmass percentage based on type
	var landPct float32
	switch plateType {
	case PlateContinental:
		landPct = 0.6 + rng.Float32()*0.3 // 60-90% land
	case PlateOceanic:
		landPct = rng.Float32() * 0.1 // 0-10% land (island arcs)
	case PlateMixed:
		landPct = 0.3 + rng.Float32()*0.3 // 30-60% land
	}

	return &TectonicPlate{
		ID:          uuid.New(),
		Name:        name,
		Type:        plateType,
		VelocityX:   velX,
		VelocityY:   velY,
		RotationZ:   rot,
		LandmassPct: landPct,
	}
}

// Speed returns the absolute plate movement speed
func (p *TectonicPlate) Speed() float32 {
	return float32(math.Sqrt(float64(p.VelocityX*p.VelocityX + p.VelocityY*p.VelocityY)))
}

// BoundaryType represents the type of plate boundary
type BoundaryType string

const (
	BoundaryDivergent  BoundaryType = "divergent"  // Plates moving apart - mid-ocean ridges, rifts
	BoundaryConvergent BoundaryType = "convergent" // Plates colliding - mountains, subduction
	BoundaryTransform  BoundaryType = "transform"  // Plates sliding past - earthquakes
	BoundaryCollision  BoundaryType = "collision"  // Continental-continental collision - mega-mountains
)

// PlateBoundary represents the boundary between two plates
type PlateBoundary struct {
	Plate1ID       uuid.UUID    `json:"plate1_id"`
	Plate2ID       uuid.UUID    `json:"plate2_id"`
	Type           BoundaryType `json:"boundary_type"`
	ActivityLevel  float32      `json:"activity_level"`  // 0.0-1.0, affects volcanism/earthquakes
	MountainHeight float32      `json:"mountain_height"` // For convergent boundaries
	RiftWidth      float32      `json:"rift_width"`      // For divergent boundaries
}

// TectonicSystem manages all plates and their interactions
type TectonicSystem struct {
	WorldID            uuid.UUID                    `json:"world_id"`
	Plates             map[uuid.UUID]*TectonicPlate `json:"plates"`
	Boundaries         []*PlateBoundary             `json:"boundaries"`
	CurrentYear        int64                        `json:"current_year"`
	LastUpdateYear     int64                        `json:"last_update_year"`
	SupercontinentMode bool                         `json:"supercontinent_mode"` // True if plates are converging toward supercontinent
	rng                *rand.Rand
}

// NewTectonicSystem creates a new tectonic system with initial plates
func NewTectonicSystem(worldID uuid.UUID, seed int64) *TectonicSystem {
	rng := rand.New(rand.NewSource(seed))

	ts := &TectonicSystem{
		WorldID:    worldID,
		Plates:     make(map[uuid.UUID]*TectonicPlate),
		Boundaries: make([]*PlateBoundary, 0),
		rng:        rng,
	}

	// Create initial plates (Earth-like configuration)
	initialPlates := []struct {
		name      string
		plateType PlateType
	}{
		{"Pangaea Prime", PlateContinental}, // Main supercontinent
		{"Pacific Prime", PlateOceanic},     // Major ocean plate
		{"Atlantic Proto", PlateOceanic},    // Nascent ocean
	}

	for _, ip := range initialPlates {
		plate := NewTectonicPlate(ip.name, ip.plateType, rng)
		ts.Plates[plate.ID] = plate
	}

	return ts
}

// Update advances the tectonic system by the given number of years
// This should be called periodically (e.g., every 10,000 years)
func (ts *TectonicSystem) Update(years int64) {
	ts.CurrentYear += years
	elapsed := ts.CurrentYear - ts.LastUpdateYear

	// Only update if sufficient time has passed
	if elapsed < 10000 {
		return
	}
	ts.LastUpdateYear = ts.CurrentYear

	// Apply plate movement (very slow - continental drift takes millions of years)
	for _, plate := range ts.Plates {
		// Plates move a few cm per year, so over 10,000 years that's ~100m
		// This is abstracted here - actual effects are calculated per update

		// Random velocity perturbation (convection cell changes)
		plate.VelocityX += (ts.rng.Float32() - 0.5) * 0.01
		plate.VelocityY += (ts.rng.Float32() - 0.5) * 0.01
	}

	// Check for plate splitting (rare - once every ~100M years)
	if ts.rng.Float64() < 0.0001*float64(elapsed)/1000000 {
		ts.splitRandomPlate()
	}

	// Check for plate merging (rare - subduction)
	if ts.rng.Float64() < 0.00005*float64(elapsed)/1000000 {
		ts.mergeRandomPlates()
	}

	// Update boundaries
	ts.updateBoundaries()
}

// splitRandomPlate splits a large continental plate into two
func (ts *TectonicSystem) splitRandomPlate() {
	// Find largest continental plate
	var largest *TectonicPlate
	for _, plate := range ts.Plates {
		if plate.Type == PlateContinental || plate.Type == PlateMixed {
			if largest == nil || plate.CellCount > largest.CellCount {
				largest = plate
			}
		}
	}

	if largest == nil || largest.CellCount < 100 {
		return // Not enough to split
	}

	// Create two new plates from the split
	newPlate := NewTectonicPlate(
		largest.Name+" Fragment",
		PlateMixed,
		ts.rng,
	)
	newPlate.OriginYear = ts.CurrentYear
	newPlate.CellCount = largest.CellCount / 3
	largest.CellCount = largest.CellCount * 2 / 3

	ts.Plates[newPlate.ID] = newPlate

	// Create divergent boundary between them
	ts.Boundaries = append(ts.Boundaries, &PlateBoundary{
		Plate1ID:      largest.ID,
		Plate2ID:      newPlate.ID,
		Type:          BoundaryDivergent,
		ActivityLevel: 0.8,
		RiftWidth:     0.1,
	})
}

// mergeRandomPlates merges two small plates via subduction
func (ts *TectonicSystem) mergeRandomPlates() {
	if len(ts.Plates) < 3 {
		return // Need at least 3 plates
	}

	// Find smallest oceanic plate
	var smallest *TectonicPlate
	var smallestID uuid.UUID
	for id, plate := range ts.Plates {
		if plate.Type == PlateOceanic {
			if smallest == nil || plate.CellCount < smallest.CellCount {
				smallest = plate
				smallestID = id
			}
		}
	}

	if smallest == nil {
		return
	}

	// Find a convergent partner (largest continental)
	var partner *TectonicPlate
	for _, plate := range ts.Plates {
		if plate.ID != smallestID && (plate.Type == PlateContinental || plate.Type == PlateMixed) {
			if partner == nil || plate.CellCount > partner.CellCount {
				partner = plate
			}
		}
	}

	if partner == nil {
		return
	}

	// Subduct the oceanic plate under the continental one
	partner.CellCount += smallest.CellCount / 2 // Some crust is recycled
	delete(ts.Plates, smallestID)

	// Remove any boundaries involving the subducted plate
	newBoundaries := make([]*PlateBoundary, 0)
	for _, b := range ts.Boundaries {
		if b.Plate1ID != smallestID && b.Plate2ID != smallestID {
			newBoundaries = append(newBoundaries, b)
		}
	}
	ts.Boundaries = newBoundaries
}

// updateBoundaries recalculates boundary types based on plate movements
func (ts *TectonicSystem) updateBoundaries() {
	// For each boundary, determine type based on relative velocities
	for _, boundary := range ts.Boundaries {
		p1 := ts.Plates[boundary.Plate1ID]
		p2 := ts.Plates[boundary.Plate2ID]
		if p1 == nil || p2 == nil {
			continue
		}

		// Calculate relative velocity (simplified)
		relVelX := p1.VelocityX - p2.VelocityX
		relVelY := p1.VelocityY - p2.VelocityY
		relSpeed := float32(math.Sqrt(float64(relVelX*relVelX + relVelY*relVelY)))

		// Determine boundary type from relative motion
		// This is simplified - real determination would consider geometry
		if relSpeed < 0.2 {
			boundary.Type = BoundaryTransform
		} else if (relVelX > 0) == (p1.VelocityX > 0) {
			boundary.Type = BoundaryDivergent
			boundary.RiftWidth += 0.001
		} else {
			if p1.Type == PlateContinental && p2.Type == PlateContinental {
				boundary.Type = BoundaryCollision
				boundary.MountainHeight += 0.01
			} else {
				boundary.Type = BoundaryConvergent
			}
		}

		boundary.ActivityLevel = relSpeed / 2.0
		if boundary.ActivityLevel > 1.0 {
			boundary.ActivityLevel = 1.0
		}
	}
}

// CalculateFragmentation returns the continental fragmentation level (0.0-1.0)
// This is derived from the plate configuration for backward compatibility
// 0.0 = supercontinent (all land connected), 1.0 = maximum fragmentation
func (ts *TectonicSystem) CalculateFragmentation() float32 {
	if len(ts.Plates) == 0 {
		return 0.5
	}

	// Count continental plates
	continentalCount := 0
	var totalLandCells int
	var largestContinentCells int

	for _, plate := range ts.Plates {
		if plate.Type == PlateContinental || plate.Type == PlateMixed {
			continentalCount++
			landCells := int(float32(plate.CellCount) * plate.LandmassPct)
			totalLandCells += landCells
			if landCells > largestContinentCells {
				largestContinentCells = landCells
			}
		}
	}

	if totalLandCells == 0 {
		return 1.0 // Water world
	}

	// Fragmentation is based on:
	// 1. Number of continental plates
	// 2. How much land is on the largest continent vs total
	plateCountFactor := float32(continentalCount-1) / 10.0 // 1 plate = 0, 10+ plates = 0.9+
	if plateCountFactor > 0.9 {
		plateCountFactor = 0.9
	}

	concentrationFactor := 1.0 - float32(largestContinentCells)/float32(totalLandCells)

	return plateCountFactor*0.6 + concentrationFactor*0.4
}

// GetPlateAt returns the plate that owns the given hex cell (stub for integration)
func (ts *TectonicSystem) GetPlateAt(coord HexCoord) *TectonicPlate {
	// This would normally look up from HexGrid
	// For now, return a random plate
	for _, plate := range ts.Plates {
		return plate
	}
	return nil
}

// IsBoundaryCell returns true if the cell is on a plate boundary
func (ts *TectonicSystem) IsBoundaryCell(coord HexCoord, grid *HexGrid) bool {
	cell := grid.GetCell(coord)
	if cell == nil {
		return false
	}

	// Check if any neighbor is on a different plate
	for _, neighbor := range grid.GetNeighbors(coord) {
		if neighbor.PlateID != cell.PlateID {
			return true
		}
	}
	return false
}

// GetBoundaryActivity returns the tectonic activity level at a cell (0.0-1.0)
func (ts *TectonicSystem) GetBoundaryActivity(coord HexCoord, grid *HexGrid) float32 {
	cell := grid.GetCell(coord)
	if cell == nil {
		return 0
	}

	// Find boundaries involving this cell's plate
	for _, boundary := range ts.Boundaries {
		if boundary.Plate1ID == cell.PlateID || boundary.Plate2ID == cell.PlateID {
			// Check if we're on this boundary
			for _, neighbor := range grid.GetNeighbors(coord) {
				if (neighbor.PlateID == boundary.Plate1ID && cell.PlateID == boundary.Plate2ID) ||
					(neighbor.PlateID == boundary.Plate2ID && cell.PlateID == boundary.Plate1ID) {
					return boundary.ActivityLevel
				}
			}
		}
	}
	return 0
}
