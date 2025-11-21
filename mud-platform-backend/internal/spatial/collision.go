package spatial

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CollisionDetector defines methods for checking collisions.
type CollisionDetector interface {
	IsPositionValid(ctx context.Context, worldID uuid.UUID, x, y, z float64) (bool, error)
}

// SimpleCollisionDetector checks world bounds and maybe obstacles.
type SimpleCollisionDetector struct {
	// For now, just bounds.
	// In future, could inject SpatialRepository to check for overlapping entities.
	MinX, MaxX float64
	MinY, MaxY float64
	MinZ, MaxZ float64
}

// NewSimpleCollisionDetector creates a new detector with default bounds.
// Default bounds: +/- 1000000 meters (1000km)
func NewSimpleCollisionDetector() *SimpleCollisionDetector {
	return &SimpleCollisionDetector{
		MinX: -1000000, MaxX: 1000000,
		MinY: -1000000, MaxY: 1000000,
		MinZ: -1000, MaxZ: 10000, // -1km to +10km elevation
	}
}

func (d *SimpleCollisionDetector) IsPositionValid(ctx context.Context, worldID uuid.UUID, x, y, z float64) (bool, error) {
	if x < d.MinX || x > d.MaxX {
		return false, fmt.Errorf("position X out of bounds")
	}
	if y < d.MinY || y > d.MaxY {
		return false, fmt.Errorf("position Y out of bounds")
	}
	if z < d.MinZ || z > d.MaxZ {
		return false, fmt.Errorf("position Z out of bounds")
	}
	return true, nil
}
