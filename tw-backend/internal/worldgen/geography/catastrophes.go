package geography

import (
	"math/rand"

	"tw-backend/internal/spatial"
)

// ApplyVolcanicMountains adds volcanic features to the heightmap.
// Used for "Volcanic Winter" events.
func ApplyVolcanicMountains(hm *SphereHeightmap, topology spatial.Topology, severity float64, rng *rand.Rand) {
	// Number of volcanoes based on severity
	numVolcanoes := 1 + int(severity*3)
	resolution := topology.Resolution()

	for i := 0; i < numVolcanoes; i++ {
		// Random location on sphere
		face := rng.Intn(6)
		x := rng.Intn(resolution)
		y := rng.Intn(resolution)
		center := spatial.Coordinate{Face: face, X: x, Y: y}

		// Volcano height based on severity (200-500m per event)
		height := 200 + severity*300
		radius := 2.0 + rng.Float64()*2.0

		ApplyVolcanoSpherical(hm, center, topology, radius, height)
	}
}

// ApplyImpactCrater creates a crater from asteroid impact.
func ApplyImpactCrater(hm *SphereHeightmap, topology spatial.Topology, severity float64, rng *rand.Rand) {
	// Crater size based on severity (10-50 cells radius)
	radius := int(10 + severity*40)

	// Depth based on severity (500-3000m)
	depth := 500 + severity*2500

	// Rim height (15% of depth)
	rimHeight := depth * 0.15

	resolution := topology.Resolution()
	// Random impact location on sphere
	centerFace := rng.Intn(6)
	centerX := rng.Intn(resolution)
	centerY := rng.Intn(resolution)
	center := spatial.Coordinate{Face: centerFace, X: centerX, Y: centerY}

	// Use BFS to apply crater with proper cross-face handling
	visited := make(map[spatial.Coordinate]bool)
	queue := []struct {
		coord spatial.Coordinate
		dist  int
	}{{center, 0}}
	visited[center] = true

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		dist := float64(current.dist)
		if dist < float64(radius) {
			// Inside crater - depression
			factor := 1.0 - (dist / float64(radius))
			currentElev := hm.Get(current.coord)
			hm.Set(current.coord, currentElev-depth*factor*factor)
		} else if dist < float64(radius)*1.3 {
			// Crater rim - raised
			t := (dist - float64(radius)) / (float64(radius) * 0.3)
			factor := 1.0 - t
			currentElev := hm.Get(current.coord)
			hm.Set(current.coord, currentElev+rimHeight*factor)
		}

		// Only expand if within extended radius
		if current.dist < int(float64(radius)*1.5) {
			for _, dir := range directions {
				neighbor := topology.GetNeighbor(current.coord, dir)
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, struct {
						coord spatial.Coordinate
						dist  int
					}{neighbor, current.dist + 1})
				}
			}
		}
	}
}

// ApplyFloodBasalt creates large volcanic provinces.
func ApplyFloodBasalt(hm *SphereHeightmap, topology spatial.Topology, severity float64, rng *rand.Rand) {
	// Radius based on severity (30-100 cells)
	radius := 30 + int(severity*70)

	// Height of basalt layers (100-500m)
	height := 100 + severity*400

	resolution := topology.Resolution()
	// Random center on sphere
	centerFace := rng.Intn(6)
	centerX := rng.Intn(resolution)
	centerY := rng.Intn(resolution)
	center := spatial.Coordinate{Face: centerFace, X: centerX, Y: centerY}

	// Use BFS to apply basalt with proper cross-face handling
	visited := make(map[spatial.Coordinate]bool)
	queue := []struct {
		coord spatial.Coordinate
		dist  int
	}{{center, 0}}
	visited[center] = true

	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.dist < radius {
			dist := float64(current.dist)
			factor := 1.0 - (dist / float64(radius))
			factor = factor * factor // Smoother falloff

			currentElev := hm.Get(current.coord)
			hm.Set(current.coord, currentElev+height*factor)

			// Expand to neighbors
			for _, dir := range directions {
				neighbor := topology.GetNeighbor(current.coord, dir)
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, struct {
						coord spatial.Coordinate
						dist  int
					}{neighbor, current.dist + 1})
				}
			}
		}
	}
}

// ApplyIceAgeEffects lowers sea level (handled by caller) and applies glacial erosion.
// seaLevel is the CURRENT sea level (after drop).
func ApplyIceAgeEffects(hm *SphereHeightmap, topology spatial.Topology, severity float64) {
	// Glacial erosion - carve U-shaped valleys in high-elevation areas
	threshold := hm.MaxElev * 0.6 // Top 40% of elevation
	resolution := topology.Resolution()

	for face := 0; face < 6; face++ {
		for y := 0; y < resolution; y++ {
			for x := 0; x < resolution; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				elev := hm.Get(coord)
				if elev > threshold {
					erosion := (elev - threshold) * 0.1 * severity
					hm.Set(coord, elev-erosion)
				}
			}
		}
	}
}
