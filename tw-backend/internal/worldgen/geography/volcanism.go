package geography

import (
	"math"
	"math/rand"
)

// ApplyHotspots creates volcanic chains over moving plates
func ApplyHotspots(hm *Heightmap, plates []TectonicPlate, seed int64) {
	r := rand.New(rand.NewSource(seed))

	width, height := hm.Width, hm.Height

	// Number of hotspots based on map size
	// e.g., 2-5 hotspots for a standard map
	numHotspots := 2 + r.Intn(4)

	for i := 0; i < numHotspots; i++ {
		// Pick a random location for the hotspot (mantle plume)
		hotspotX := float64(r.Intn(width))
		hotspotY := float64(r.Intn(height))

		// Intensity and size
		intensity := 2000.0 + r.Float64()*3000.0 // Height of volcanoes
		radius := 2.0 + r.Float64()*3.0          // Width of cones

		// Map hotspot to the tectonic plate above it
		// We simulate the plate moving OVER the hotspot for millions of years.
		// Or simpler: We just trace a line backwards/forwards along the plate's movement vector.

		// Find which plate is currently over this spot
		closestPlateIdx := 0
		minDist := math.MaxFloat64
		for idx, p := range plates {
			d := distance(hotspotX, hotspotY, p.Centroid.X, p.Centroid.Y)
			if d < minDist {
				minDist = d
				closestPlateIdx = idx
			}
		}
		plate := plates[closestPlateIdx]

		// Trace the "chain"
		// The volcanoes form along a line opposite to plate movement
		// Plate moves V, so old volcanoes are at Position - t*V

		steps := 5                            // Number of volcanoes in the chain
		stepSize := 5.0 * (1.0 + r.Float64()) // Distance between volcanoes

		for s := 0; s < steps; s++ {
			// Current age of this volcano (0 is newest, active one)
			age := float64(s)

			// Calculate position offset
			// Older volcanoes are further along the vector
			// Wait: if plate moves RIGHT, the hotspot stays still.
			// The crust moves RIGHT. The mark made by the hotspot starts at X,
			// but the crust at X moves to X+V.
			// So the OLD volcano (created t ago) is now at Hotspot + t*V.
			// The NEW volcano is at Hotspot.

			vx := hotspotX + (plate.MovementVector.X * stepSize * age)
			vy := hotspotY + (plate.MovementVector.Y * stepSize * age)

			// Erode older volcanoes?
			currentIntensity := intensity * (1.0 - (age * 0.15)) // Older ones shrink
			if currentIntensity <= 0 {
				break
			}

			// Apply volcano capability
			applyVolcano(hm, vx, vy, radius, currentIntensity)
		}
	}
}

func applyVolcano(hm *Heightmap, x, y, radius, height float64) {
	// Simple cone or bell curve
	radCeil := int(math.Ceil(radius * 3))

	ix, iy := int(x), int(y)

	for dy := -radCeil; dy <= radCeil; dy++ {
		for dx := -radCeil; dx <= radCeil; dx++ {
			px, py := ix+dx, iy+dy

			if px >= 0 && px < hm.Width && py >= 0 && py < hm.Height {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < radius*3 {
					// Bell curve shape: e^(-dist^2 / 2sigma^2)
					// sigma approx radius
					val := height * math.Exp(-(dist*dist)/(2*radius*radius))

					current := hm.Get(px, py)
					// Additive? Or max?
					// Additive usually builds nicely on existing terrain
					hm.Set(px, py, current+val)
				}
			}
		}
	}
}
