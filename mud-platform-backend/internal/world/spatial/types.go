package spatial

import (
	"math"

	"github.com/google/uuid"
)

// Portal represents a doorway to another world
type Portal struct {
	ID          uuid.UUID `json:"portal_id"`
	WorldID     uuid.UUID `json:"world_id"` // Destination world
	LocationX   float64   `json:"location_x"`
	LocationY   float64   `json:"location_y"`
	Side        string    `json:"side"` // "east" or "west"
	Description string    `json:"description"`
}

// WorldDimensions defines the size and shape of a world
type WorldDimensions struct {
	CircumferenceM   float64
	RadiusM          float64
	SurfaceAreaM2    float64
	MetersPerDegreeX float64
	MetersPerDegreeY float64
}

// NewWorldDimensions creates dimensions from circumference
func NewWorldDimensions(circumference float64) WorldDimensions {
	radius := circumference / (2 * math.Pi)
	surfaceArea := 4 * math.Pi * math.Pow(radius, 2)

	return WorldDimensions{
		CircumferenceM:   circumference,
		RadiusM:          radius,
		SurfaceAreaM2:    surfaceArea,
		MetersPerDegreeX: circumference / 360.0,
		MetersPerDegreeY: circumference / 360.0,
	}
}

// Coordinate represents a 2D location on the map
type Coordinate struct {
	X float64
	Y float64
}

// Location represents a full position including world
type Location struct {
	WorldID uuid.UUID
	X       float64
	Y       float64
	Z       float64
}
