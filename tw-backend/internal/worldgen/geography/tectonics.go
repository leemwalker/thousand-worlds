package geography

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// GeneratePlates creates a set of tectonic plates
func GeneratePlates(count int, width, height int, seed int64) []TectonicPlate {
	r := rand.New(rand.NewSource(seed))
	plates := make([]TectonicPlate, count)

	// 1. Initialize plates with random centroids
	for i := 0; i < count; i++ {
		plates[i] = TectonicPlate{
			PlateID:  uuid.New(),
			Centroid: Point{X: float64(r.Intn(width)), Y: float64(r.Intn(height))},
			// Random movement vector (-1 to 1)
			MovementVector: Vector{
				X: r.Float64()*2 - 1,
				Y: r.Float64()*2 - 1,
			},
			Age: r.Float64() * 100, // 0-100 million years
		}

		// Normalize movement vector
		mag := math.Sqrt(plates[i].MovementVector.X*plates[i].MovementVector.X + plates[i].MovementVector.Y*plates[i].MovementVector.Y)
		if mag > 0 {
			plates[i].MovementVector.X /= mag
			plates[i].MovementVector.Y /= mag
		}

		// Assign type (30% continental, 70% oceanic)
		if i < int(float64(count)*0.3) {
			plates[i].Type = PlateContinental
			plates[i].Thickness = 30 + r.Float64()*20 // 30-50km
		} else {
			plates[i].Type = PlateOceanic
			plates[i].Thickness = 5 + r.Float64()*5 // 5-10km
		}
	}

	return plates
}

// SimulateTectonics calculates elevation modifiers based on plate interactions
// Returns a map of (x,y) -> elevation modifier
func SimulateTectonics(plates []TectonicPlate, width, height int) *Heightmap {
	modifiers := NewHeightmap(width, height)

	// Map each pixel to a plate (Voronoi)
	plateMap := make([]int, width*height) // Index of plate
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			minDist := math.MaxFloat64
			closestPlate := 0

			for i, p := range plates {
				dist := distance(float64(x), float64(y), p.Centroid.X, p.Centroid.Y)
				if dist < minDist {
					minDist = dist
					closestPlate = i
				}
			}
			plateMap[y*width+x] = closestPlate
		}
	}

	// Detect boundaries and apply modifiers
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			currentPlateIdx := plateMap[y*width+x]
			currentPlate := plates[currentPlateIdx]

			// Check neighbors to find boundary
			neighbors := []Point{{X: 1, Y: 0}, {X: 0, Y: 1}}

			for _, n := range neighbors {
				nx, ny := x+int(n.X), y+int(n.Y)
				if nx >= width || ny >= height {
					continue
				}

				neighborPlateIdx := plateMap[ny*width+nx]
				if currentPlateIdx != neighborPlateIdx {
					neighborPlate := plates[neighborPlateIdx]

					// Calculate interaction
					// Vector from current to neighbor
					dx := neighborPlate.Centroid.X - currentPlate.Centroid.X
					dy := neighborPlate.Centroid.Y - currentPlate.Centroid.Y

					// Normalize direction
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist == 0 {
						continue
					}
					dirX, dirY := dx/dist, dy/dist

					// Relative movement: neighbor - current
					mvX := neighborPlate.MovementVector.X - currentPlate.MovementVector.X
					mvY := neighborPlate.MovementVector.Y - currentPlate.MovementVector.Y

					// Interaction factor: positive = divergent, negative = convergent
					// Dot product of direction and relative movement
					interaction := dirX*mvX + dirY*mvY

					applyBoundaryEffect(modifiers, x, y, currentPlate, neighborPlate, interaction)
				}
			}
		}
	}

	return modifiers
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func applyBoundaryEffect(hm *Heightmap, x, y int, p1, p2 TectonicPlate, interaction float64) {
	var elevationChange float64
	// Assuming 1 pixel = 10km for now, so radius = 5 pixels
	// TODO: Make scale configurable
	pixelRadius := 5

	if interaction > 0.2 {
		// Divergent
		if p1.Type == PlateOceanic && p2.Type == PlateOceanic {
			elevationChange = 500 // Mid-ocean ridge
		} else if p1.Type == PlateContinental && p2.Type == PlateContinental {
			elevationChange = -200 // Rift valley
		}
	} else if interaction < -0.2 {
		// Convergent
		if p1.Type == PlateOceanic && p2.Type == PlateOceanic {
			elevationChange = -8000 // Trench
		} else if p1.Type == PlateContinental && p2.Type == PlateContinental {
			elevationChange = 6000 // Mountains
		} else {
			// Oceanic-Continental
			elevationChange = 4000 // Coastal mountains
		}
	} else {
		// Transform
		elevationChange = 50 // Minimal
	}

	// Apply with falloff
	for dy := -pixelRadius; dy <= pixelRadius; dy++ {
		for dx := -pixelRadius; dx <= pixelRadius; dx++ {
			px, py := x+dx, y+dy
			if px >= 0 && px < hm.Width && py >= 0 && py < hm.Height {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= float64(pixelRadius) {
					factor := (1.0 - dist/float64(pixelRadius))
					factor = factor * factor // Square for smoother falloff

					current := hm.Get(px, py)
					// Additive blending
					hm.Set(px, py, current+elevationChange*factor)
				}
			}
		}
	}
}
