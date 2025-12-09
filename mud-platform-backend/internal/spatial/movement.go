package spatial

import (
	"errors"
)

// Direction represents a movement direction.
type Direction string

const (
	North     Direction = "N"
	South     Direction = "S"
	East      Direction = "E"
	West      Direction = "W"
	NorthEast Direction = "NE"
	SouthEast Direction = "SE"
	SouthWest Direction = "SW"
	NorthWest Direction = "NW"
	Up        Direction = "UP"
	Down      Direction = "DOWN"
)

// CalculateNewPosition calculates the new position based on direction and distance.
// Default distance is 1.0 meter.
func CalculateNewPosition(x, y, z float64, direction Direction, distance float64) (float64, float64, float64, error) {
	if distance <= 0 {
		return x, y, z, errors.New("distance must be positive")
	}

	// Diagonal movement: As per requirements, moving NE moves 1 unit North AND 1 unit East.
	// This results in a Euclidean distance of sqrt(2), but logic dictates grid-like steps.

	switch direction {
	case North:
		return x, y + distance, z, nil
	case South:
		return x, y - distance, z, nil
	case East:
		return x + distance, y, z, nil
	case West:
		return x - distance, y, z, nil
	case NorthEast:
		return x + distance, y + distance, z, nil
	case SouthEast:
		return x + distance, y - distance, z, nil
	case SouthWest:
		return x - distance, y - distance, z, nil
	case NorthWest:
		return x - distance, y + distance, z, nil
	case Up:
		return x, y, z + distance, nil
	case Down:
		return x, y, z - distance, nil
	default:
		return x, y, z, errors.New("invalid direction")
	}
}
