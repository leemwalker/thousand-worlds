package geography

import (
	"math"
	"math/rand"
)

// ApplyThermalErosion improves slope stability by moving material from steep slopes to lower neighbors
func ApplyThermalErosion(hm *Heightmap, iterations int, seed int64) {
	// Talus angle approximation (max difference allowed)
	threshold := 40.0
	width, height := hm.Width, hm.Height

	for iter := 0; iter < iterations; iter++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				currentElev := hm.Get(x, y)
				maxDiff := 0.0
				var bestNeighX, bestNeighY int

				// Check neighbors
				neighbors := [][2]int{
					{0, 1}, {0, -1}, {1, 0}, {-1, 0},
					{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
				}

				// Find lowest neighbor
				for _, n := range neighbors {
					nx, ny := x+n[0], y+n[1]
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						diff := currentElev - hm.Get(nx, ny)
						if diff > maxDiff {
							maxDiff = diff
							bestNeighX, bestNeighY = nx, ny
						}
					}
				}

				// If slope is too steep, erode
				if maxDiff > threshold {
					transfer := maxDiff * 0.1 // Move 10% of excess
					hm.Set(x, y, currentElev-transfer)
					hm.Set(bestNeighX, bestNeighY, hm.Get(bestNeighX, bestNeighY)+transfer)
				}
			}
		}
	}
}

// ApplyHydraulicErosion simulates rain and water flow to carve valleys
func ApplyHydraulicErosion(hm *Heightmap, drops int, seed int64) {
	r := rand.New(rand.NewSource(seed))
	width, height := hm.Width, hm.Height

	// Constants
	dt := 1.2
	density := 1.0 // Density of water
	evapRate := 0.001
	depositionRate := 0.3
	minVol := 0.01
	friction := 0.1

	for i := 0; i < drops; i++ {
		// Spawn drop
		x := float64(r.Intn(width))
		y := float64(r.Intn(height))

		// Drop properties
		speedX, speedY := 0.0, 0.0
		volume := 1.0
		sediment := 0.0

		for volume > minVol {
			ix, iy := int(x), int(y)
			if ix < 0 || ix >= width-1 || iy < 0 || iy >= height-1 {
				break
			}

			// Get surface normal / gradient
			n00 := hm.Get(ix, iy)
			n10 := hm.Get(ix+1, iy)
			n01 := hm.Get(ix, iy+1)
			n11 := hm.Get(ix+1, iy+1)

			gx := (n10 + n11) - (n00 + n01)
			gy := (n01 + n11) - (n00 + n10)

			// Update Position
			// F = ma, but here just assume F ~ gradient
			speedX = (speedX * (1 - friction)) - (gx * 0.5)
			speedY = (speedY * (1 - friction)) - (gy * 0.5)

			x += speedX * dt
			y += speedY * dt

			if x < 0 || x >= float64(width-1) || y < 0 || y >= float64(height-1) {
				break
			}

			// New elevation
			// Interpolate new height
			// Simplified: just use nearest integer for erosion target
			newIx, newIy := int(x), int(y)
			newElev := hm.Get(newIx, newIy)
			// oldElev := hm.Get(ix, iy) // Unused

			// Approximate height difference along trajectory
			heightDiff := newElev - hm.Get(ix, iy)

			// Sediment capacity
			// Capacity is proportional to velocity and volume
			velocity := math.Sqrt(speedX*speedX + speedY*speedY)
			capacity := math.Max(-heightDiff, minVol) * velocity * volume * density

			if heightDiff > 0 {
				// Moving uphill? Fill depression
				// Deposit sediment
				amount := math.Min(sediment, heightDiff)
				sediment -= amount
				hm.Set(ix, iy, hm.Get(ix, iy)+amount)
			} else {
				if sediment > capacity {
					// Deposit
					amount := (sediment - capacity) * depositionRate
					sediment -= amount
					hm.Set(ix, iy, hm.Get(ix, iy)+amount)
				} else {
					// Erode
					amount := math.Min((capacity-sediment)*0.3, -heightDiff)
					sediment += amount
					hm.Set(ix, iy, hm.Get(ix, iy)-amount)
				}
			}

			volume *= (1.0 - evapRate)
		}
	}
}
