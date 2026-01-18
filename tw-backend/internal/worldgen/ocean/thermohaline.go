package ocean

import (
	"math"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/weather"
)

// InitializeTemperature sets baseline water temperature from latitude.
func (s *System) InitializeTemperature() {
	res := s.topology.Resolution()

	for face := 0; face < 6; face++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				idx := y*res + x

				// Populate static isOcean membership
				isOcean := s.geo.Get(coord) <= s.seaLevel
				s.isOcean[face][idx] = isOcean

				if !isOcean {
					s.WaterTemperature[face][idx] = 0 // Default for land
					continue
				}

				latitude := weather.GetLatitudeFromCoord(s.topology, coord)
				latRad := latitude * math.Pi / 180.0
				temp := 28.0*math.Cos(latRad) - 2.0*math.Pow(math.Sin(latRad), 2)

				if temp < -2.0 {
					temp = -2.0
				}
				if temp > 30.0 {
					temp = 30.0
				}

				s.WaterTemperature[face][idx] = temp
			}
		}
	}
}

// SimulateThermodynamics runs heat advection simulation.
func (s *System) SimulateThermodynamics(iterations int) {
	res := s.topology.Resolution()
	totalCells := res * res
	dt := 0.15

	// Double buffer using pre-allocated slices to avoid map allocations
	current := s.WaterTemperature
	next := [6][]float64{}
	for i := 0; i < 6; i++ {
		next[i] = make([]float64, totalCells)
	}

	// Pre-identify ocean cells and their advection targets once per simulation run
	type oceanRef struct {
		face, idx int
		speedFact float64
		targetF   int
		targetIdx int
	}
	oceanRefs := make([]oceanRef, 0)

	for face := 0; face < 6; face++ {
		for idx := 0; idx < totalCells; idx++ {
			if !s.isOcean[face][idx] {
				continue
			}

			coord := spatial.Coordinate{Face: face, X: idx % res, Y: idx / res}
			currentVec := s.CurrentMap[face][idx]
			speed := currentVec.Length()
			if speed < 0.01 {
				continue
			}

			speedFactor := math.Min(speed/10.0, 1.0) * dt
			dir := weather.WindToLocalDirection(s.topology, coord, currentVec)
			targetCoord := s.topology.GetNeighbor(coord, dir)

			// Target must also be a pre-initialized ocean cell
			targetIdx := targetCoord.Y*res + targetCoord.X
			if !s.isOcean[targetCoord.Face][targetIdx] {
				continue
			}

			oceanRefs = append(oceanRefs, oceanRef{
				face:      face,
				idx:       idx,
				speedFact: speedFactor,
				targetF:   targetCoord.Face,
				targetIdx: targetIdx,
			})
		}
	}

	for iter := 0; iter < iterations; iter++ {
		// 1. Copy current to next
		for f := 0; f < 6; f++ {
			copy(next[f], current[f])
		}

		// 2. Apply advection from pre-computed refs
		for _, ref := range oceanRefs {
			sourceTemp := current[ref.face][ref.idx]
			targetTemp := current[ref.targetF][ref.targetIdx]

			newTargetTemp := targetTemp + (sourceTemp-targetTemp)*ref.speedFact

			// Accumulate contribs (weighted average handles multiple sources)
			next[ref.targetF][ref.targetIdx] = next[ref.targetF][ref.targetIdx]*0.5 + newTargetTemp*0.5
		}

		// 3. Swap buffers for next iteration
		for f := 0; f < 6; f++ {
			current[f], next[f] = next[f], current[f]
		}
	}

	s.WaterTemperature = current
}

// GetAverageOceanTemp returns the average temperature of neighboring ocean cells.
func (s *System) GetAverageOceanTemp(coord spatial.Coordinate) (float64, bool) {
	res := s.topology.Resolution()
	directions := []spatial.Direction{
		spatial.North, spatial.South, spatial.East, spatial.West,
	}

	sum := 0.0
	count := 0

	for _, dir := range directions {
		neighbor := s.topology.GetNeighbor(coord, dir)
		if s.IsOcean(neighbor) {
			idx := neighbor.Y*res + neighbor.X
			sum += s.WaterTemperature[neighbor.Face][idx]
			count++
		}
	}

	if count == 0 {
		return 0, false
	}

	return sum / float64(count), true
}
