package geography

import (
	"math"
	"math/rand"

	"tw-backend/internal/spatial"
)

// ApplyHotspots creates volcanic chains over moving plates (spherical version)
func ApplyHotspots(hm *SphereHeightmap, plates []TectonicPlate, topology spatial.Topology, seed int64) {
	r := rand.New(rand.NewSource(seed))
	resolution := topology.Resolution()

	// Number of hotspots based on total area
	numHotspots := 2 + r.Intn(4)

	for i := 0; i < numHotspots; i++ {
		// Pick a random location for the hotspot (mantle plume)
		face := r.Intn(6)
		x := r.Intn(resolution)
		y := r.Intn(resolution)
		hotspotCoord := spatial.Coordinate{Face: face, X: x, Y: y}

		// Intensity and size
		intensity := 2000.0 + r.Float64()*3000.0 // Height of volcanoes
		coneRadius := 2.0 + r.Float64()*3.0      // Width of cones

		// Find which plate is over this spot
		var closestPlate *TectonicPlate
		for idx := range plates {
			if _, exists := plates[idx].Region[hotspotCoord]; exists {
				closestPlate = &plates[idx]
				break
			}
		}

		if closestPlate == nil {
			continue // No plate found
		}

		// Trace the volcanic chain along the plate velocity
		// Older volcanoes are offset in the direction of plate movement
		steps := 5
		stepSize := 5 // Grid cells between volcanoes

		currentCoord := hotspotCoord
		for s := 0; s < steps; s++ {
			age := float64(s)
			currentIntensity := intensity * (1.0 - (age * 0.15))
			if currentIntensity <= 0 {
				break
			}

			// Apply volcano at current position
			ApplyVolcanoSpherical(hm, currentCoord, topology, coneRadius, currentIntensity)

			// Move along the velocity direction to get next position
			// Approximate by stepping in the dominant direction
			dir := velocityToDirection(closestPlate.Velocity, topology, currentCoord)
			for j := 0; j < stepSize; j++ {
				currentCoord = topology.GetNeighbor(currentCoord, dir)
			}
		}
	}
}

// velocityToDirection converts a 3D velocity vector to the best local grid direction
func velocityToDirection(velocity spatial.Vector3D, topology spatial.Topology, coord spatial.Coordinate) spatial.Direction {
	// Project velocity onto local surface tangent directions
	sx, sy, sz := topology.ToSphere(coord)
	surfaceNormal := spatial.Vector3D{X: sx, Y: sy, Z: sz}

	// Local "up" direction on sphere is the surface normal
	// Calculate local east and north directions
	worldUp := spatial.Vector3D{X: 0, Y: 0, Z: 1}
	localEast := worldUp.Cross(surfaceNormal).Normalize()
	localNorth := surfaceNormal.Cross(localEast).Normalize()

	// Project velocity onto local tangent plane
	eastComponent := velocity.Dot(localEast)
	northComponent := velocity.Dot(localNorth)

	// Pick dominant direction
	if math.Abs(eastComponent) > math.Abs(northComponent) {
		if eastComponent > 0 {
			return spatial.East
		}
		return spatial.West
	}
	if northComponent > 0 {
		return spatial.North
	}
	return spatial.South
}

// ApplyVolcanoSpherical adds a single volcanic cone to the sphere heightmap
func ApplyVolcanoSpherical(hm *SphereHeightmap, center spatial.Coordinate, topology spatial.Topology, radius, height float64) {
	// Use BFS to apply cone shape
	radCeil := int(math.Ceil(radius * 3))

	visited := make(map[spatial.Coordinate]struct{})
	type queueItem struct {
		coord    spatial.Coordinate
		distance int
	}
	queue := []queueItem{{coord: center, distance: 0}}
	visited[center] = struct{}{}

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.distance > radCeil {
			continue
		}

		// Bell curve shape: e^(-dist^2 / 2sigma^2)
		dist := float64(current.distance)
		val := height * math.Exp(-(dist*dist)/(2*radius*radius))

		currentElev := hm.Get(current.coord)
		hm.Set(current.coord, currentElev+val)

		// Expand to neighbors
		if current.distance < radCeil {
			for _, dir := range directions {
				neighbor := topology.GetNeighbor(current.coord, dir)
				if _, exists := visited[neighbor]; !exists {
					visited[neighbor] = struct{}{}
					queue = append(queue, queueItem{coord: neighbor, distance: current.distance + 1})
				}
			}
		}
	}
}

// ApplyVolcanoFlat adds a single volcanic cone to a flat heightmap (legacy support).
// This is the original ApplyVolcano function preserved for backward compatibility.
func ApplyVolcanoFlat(hm *Heightmap, x, y, radius, height float64) {
	radCeil := int(math.Ceil(radius * 3))
	ix, iy := int(x), int(y)

	for dy := -radCeil; dy <= radCeil; dy++ {
		for dx := -radCeil; dx <= radCeil; dx++ {
			px, py := ix+dx, iy+dy

			if px >= 0 && px < hm.Width && py >= 0 && py < hm.Height {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < radius*3 {
					// Bell curve shape: e^(-dist^2 / 2sigma^2)
					val := height * math.Exp(-(dist*dist)/(2*radius*radius))
					current := hm.Get(px, py)
					hm.Set(px, py, current+val)
				}
			}
		}
	}
}

// GetEruptionStyle determines volcanic behavior based on magma composition
func GetEruptionStyle(magmaType string) (viscosity string, explosivity string) {
	switch magmaType {
	case "rhyolite":
		return "explosive", "pyroclastic_flow"
	case "andesite":
		return "mixed", "lahar"
	case "basalt":
		return "effusive", "lava_flow"
	default:
		return "mixed", "lava_flow"
	}
}

// SimulateHotspotErosion calculates the erosion factor for volcanic islands over time
func SimulateHotspotErosion(years int64) float64 {
	erosion := 1.0 - (float64(years) / 100_000_000.0)
	if erosion < 0 {
		return 0
	}
	return erosion
}

// SimulateFloodBasalt calculates the impact of a Large Igneous Province event
func SimulateFloodBasalt(severity float64) (radius float64, so2 float64, cooling float64) {
	return 1000.0 * severity, 100.0 * severity, -2.0 * severity
}

// SimulateCalderaCollapse handles supervolcano aftermath
func SimulateCalderaCollapse(peakElevation float64) (newElevation float64, shape string) {
	return peakElevation * 0.4, "basin"
}

// SimulateAtollFormation determines the state of a sinking volcanic island
func SimulateAtollFormation(ageMillionYears float64, originalType string) (elevation float64, isReef bool) {
	if originalType == "volcanic" && ageMillionYears > 2.0 {
		return -10.0, true
	}
	return 100.0, false
}

// SimulateVolcanicWorldFrequency returns expected eruption count
func SimulateVolcanicWorldFrequency(years int64) int {
	baseEruptions := int(years / 10000)
	return baseEruptions * 5
}

// SimulateSoilFertility calculates fertility bonus from volcanic ash
func SimulateSoilFertility(years int64, hasAsh bool) float64 {
	if !hasAsh {
		return 0.5
	}
	if years > 500 {
		return 0.9
	}
	return 0.6
}

// SimulateClimateFeedback estimates global impact
func SimulateClimateFeedback(volcanicIndex float64) (co2Increase float64, tempChange float64) {
	return 100.0 * volcanicIndex, 1.5 * volcanicIndex
}

// SimulateTerrainBurial checks if features are buried
func SimulateTerrainBurial(depth float64) bool {
	return depth > 20.0
}
