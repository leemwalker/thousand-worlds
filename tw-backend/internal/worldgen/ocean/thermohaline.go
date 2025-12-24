package ocean

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/weather"
)

// InitializeTemperature sets baseline water temperature from latitude.
// Uses similar physics to climate_generator but for ocean surface water.
func (s *System) InitializeTemperature() {
	resolution := s.topology.Resolution()

	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				// Skip land cells
				if !s.IsOcean(coord) {
					continue
				}

				latitude := weather.GetLatitudeFromCoord(s.topology, coord)

				// Ocean surface temperature model:
				// Equator: ~28°C
				// Poles: ~0°C (can go slightly negative due to salinity)
				// Uses cosine for smooth transition
				latRad := latitude * math.Pi / 180.0
				temp := 28.0*math.Cos(latRad) - 2.0*math.Pow(math.Sin(latRad), 2)

				// Clamp to realistic ocean surface temps
				if temp < -2.0 {
					temp = -2.0 // Freezing point of seawater
				}
				if temp > 30.0 {
					temp = 30.0
				}

				s.WaterTemperature[coord] = temp
			}
		}
	}
}

// SimulateThermodynamics runs heat advection simulation.
// Uses double-buffering to prevent order-dependent artifacts.
//
// Physics: Heat moves in the direction of ocean currents.
// Algorithm: Temp[target] = lerp(Temp[target], Temp[source], currentSpeed * dt)
func (s *System) SimulateThermodynamics(iterations int) {
	// Pre-compute ocean cells for efficiency
	oceanCells := make([]spatial.Coordinate, 0)
	for coord := range s.WaterTemperature {
		oceanCells = append(oceanCells, coord)
	}

	// Double buffer for proper physics
	current := s.WaterTemperature
	next := make(map[spatial.Coordinate]float64, len(current))

	// Time step factor: larger = faster convergence, smaller = more stable
	// 0.15 is balanced for ~50 iterations to achieve Gulf Stream effect
	dt := 0.15

	for iter := 0; iter < iterations; iter++ {
		// Copy current state to next buffer
		for coord, temp := range current {
			next[coord] = temp
		}

		// Apply advection from each cell
		for _, sourceCoord := range oceanCells {
			currentVec, hasCurrent := s.CurrentMap[sourceCoord]
			if !hasCurrent {
				continue
			}

			sourceTemp := current[sourceCoord]

			// Get the direction the current is flowing
			speed := currentVec.Length()
			if speed < 0.01 {
				continue // No significant current
			}

			// Normalize speed influence (cap at reasonable value)
			speedFactor := math.Min(speed/10.0, 1.0) * dt

			// Find target cell in current direction
			dir := weather.WindToLocalDirection(s.topology, sourceCoord, currentVec)
			targetCoord := s.topology.GetNeighbor(sourceCoord, dir)

			// Only advect to ocean cells
			if _, isOcean := current[targetCoord]; !isOcean {
				continue
			}

			// Lerp: target temperature approaches source temperature
			targetTemp := current[targetCoord]
			newTargetTemp := targetTemp + (sourceTemp-targetTemp)*speedFactor

			// Write to next buffer (accumulate if multiple sources)
			// Use weighted average to handle multiple contributions
			existingTemp := next[targetCoord]
			next[targetCoord] = existingTemp*0.5 + newTargetTemp*0.5
		}

		// Swap buffers
		current, next = next, current
	}

	// Store final result
	s.WaterTemperature = current
}

// GetAverageOceanTemp returns the average temperature of neighboring ocean cells.
// Used for coastal land temperature moderation.
func (s *System) GetAverageOceanTemp(coord spatial.Coordinate) (float64, bool) {
	directions := []spatial.Direction{
		spatial.North, spatial.South, spatial.East, spatial.West,
	}

	var sum float64
	var count int

	for _, dir := range directions {
		neighbor := s.topology.GetNeighbor(coord, dir)
		if temp, isOcean := s.WaterTemperature[neighbor]; isOcean {
			sum += temp
			count++
		}
	}

	if count == 0 {
		return 0, false
	}

	return sum / float64(count), true
}
