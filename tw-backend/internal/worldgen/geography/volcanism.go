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
			ApplyVolcano(hm, vx, vy, radius, currentIntensity)
		}
	}
}

// ApplyVolcano adds a single volcanic cone to the heightmap
func ApplyVolcano(hm *Heightmap, x, y, radius, height float64) {
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
	// Simple linear erosion: loses 1% height per million years?
	// The prompt implies a multiplier.
	// If 50 million years, erosion should be significant.
	erosion := 1.0 - (float64(years) / 100_000_000.0)
	if erosion < 0 {
		return 0
	}
	return erosion
}

// SimulateFloodBasalt calculates the impact of a Large Igneous Province event
func SimulateFloodBasalt(severity float64) (radius float64, so2 float64, cooling float64) {
	// Severity 0-1
	return 1000.0 * severity, 100.0 * severity, -2.0 * severity
}

// SimulateCalderaCollapse handles supervolcano aftermath
func SimulateCalderaCollapse(peakElevation float64) (newElevation float64, shape string) {
	// Collapse to < 50% of peak?
	return peakElevation * 0.4, "basin"
}

// SimulateAtollFormation determines the state of a sinking volcanic island
func SimulateAtollFormation(ageMillionYears float64, originalType string) (elevation float64, isReef bool) {
	// Darwin's subsidence theory
	if originalType == "volcanic" && ageMillionYears > 2.0 {
		// Island sinks, coral grows up
		return -10.0, true // Shallow water reef
	}
	return 100.0, false // Still an island
}

// SimulateVolcanicWorldFrequency returns expected eruption count
func SimulateVolcanicWorldFrequency(years int64) int {
	// Volcanic worlds are 5x more active
	baseEruptions := int(years / 10000)
	return baseEruptions * 5
}

// SimulateSoilFertility calculates fertility bonus from volcanic ash (andisols)
func SimulateSoilFertility(years int64, hasAsh bool) float64 {
	if !hasAsh {
		return 0.5
	}
	// Takes time to weather into soil
	if years > 500 {
		return 0.9 // High fertility
	}
	return 0.6 // improving
}

// SimulateClimateFeedback estimates global impact
// SimulateClimateFeedback estimates global impact
func SimulateClimateFeedback(volcanicIndex float64) (co2Increase float64, tempChange float64) {
	// Short term cooling (aerosols), long term warming (CO2)
	// Simplified model: return net effect
	return 100.0 * volcanicIndex, 1.5 * volcanicIndex
}

// SimulateTerrainBurial checks if features are buried
func SimulateTerrainBurial(depth float64) bool {
	return depth > 20.0
}
