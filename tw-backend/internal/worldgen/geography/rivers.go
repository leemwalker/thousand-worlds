package geography

import (
	"math/rand"
)

// GenerateRivers creates river paths based on heightmap
func GenerateRivers(hm *Heightmap, seaLevel float64, seed int64) [][]Point {
	var rivers [][]Point
	r := rand.New(rand.NewSource(seed))

	width, height := hm.Width, hm.Height
	visited := make(map[int]bool) // Avoid merging rivers too often or loops

	// Try to spawn N rivers
	// Density: 1 river per 100km^2 approx.
	// If 50x50 grid = 2500 pixels. If 1 pixel = 10km, then 100km^2 = 1 pixel.
	// So 2500 rivers? That's too many for this scale.
	// Let's aim for ~50 rivers for a 50x50 map.
	numRivers := (width * height) / 50

	for i := 0; i < numRivers; i++ {
		// Pick candidate source
		sx, sy := r.Intn(width), r.Intn(height)
		elev := hm.Get(sx, sy)

		// Must be high elevation and not already visited
		if elev > seaLevel+500 && !visited[sy*width+sx] {
			path := traceRiver(hm, sx, sy, seaLevel, visited)
			if len(path) > 5 { // Min length
				rivers = append(rivers, path)

				// Mark path as visited/eroded
				for _, p := range path {
					idx := int(p.Y)*width + int(p.X)
					visited[idx] = true

					// Erosion: Carve valley
					current := hm.Get(int(p.X), int(p.Y))
					hm.Set(int(p.X), int(p.Y), current-20)
				}
			}
		}
	}

	return rivers
}

func traceRiver(hm *Heightmap, sx, sy int, seaLevel float64, visited map[int]bool) []Point {
	path := []Point{{X: float64(sx), Y: float64(sy)}}
	currX, currY := sx, sy

	for {
		// Find lowest neighbor
		bestX, bestY := -1, -1
		minElev := hm.Get(currX, currY)

		neighbors := [][2]int{
			{0, 1}, {0, -1}, {1, 0}, {-1, 0},
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
		}

		foundDownhill := false

		for _, n := range neighbors {
			nx, ny := currX+n[0], currY+n[1]
			if nx >= 0 && nx < hm.Width && ny >= 0 && ny < hm.Height {
				elev := hm.Get(nx, ny)
				if elev < minElev {
					minElev = elev
					bestX, bestY = nx, ny
					foundDownhill = true
				}
			}
		}

		if !foundDownhill {
			// Local minimum (lake) or ocean
			break
		}

		// Move
		currX, currY = bestX, bestY
		path = append(path, Point{X: float64(currX), Y: float64(currY)})

		// Check if reached ocean
		if minElev <= seaLevel {
			break
		}

		// Loop detection or max length
		if len(path) > 500 {
			break
		}

		// If we hit an existing river, merge and stop
		if visited[currY*hm.Width+currX] {
			break
		}
	}

	return path
}
