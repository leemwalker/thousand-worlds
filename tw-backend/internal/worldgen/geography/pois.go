package geography

import (
	"math"
	"sort"

	"tw-backend/internal/spatial"

	"github.com/google/uuid"
)

// POIType represents the category of a Point Of Interest
type POIType string

const (
	POITypeMountainPeak POIType = "mountain_peak"
	POITypeVolcano      POIType = "volcano"
	POITypeCanyon       POIType = "canyon"
	POITypeRiverMouth   POIType = "river_mouth"
	POITypeLake         POIType = "lake"
	POITypeValley       POIType = "valley"
	POITypeDeepOcean    POIType = "deep_ocean"
)

// PointOfInterest represents a significant geological or geographical feature
type PointOfInterest struct {
	ID       uuid.UUID `json:"id"`
	Type     POIType   `json:"type"`
	Name     string    `json:"name,omitempty"` // Can be generated on demand
	Location struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"location"`
	Coordinates Coordinates `json:"coordinates"`
	Elevation   float64     `json:"elevation"`
	Importance  float64     `json:"importance"` // 0.0 to 1.0, for filtering
	Description string      `json:"description,omitempty"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// GeneratePOIs scans the terrain to find significant features.
// limit controls the maximum number of POIs returned.
func GeneratePOIs(hm *SphereHeightmap, seaLevel float64, limit int) []PointOfInterest {
	var candidates []PointOfInterest
	res := hm.Resolution()
	directions := []spatial.Direction{spatial.North, spatial.South, spatial.East, spatial.West}

	// 1. Find Mountain Peaks
	// A peak is a point higher than all its neighbors
	for f := 0; f < 6; f++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: f, X: x, Y: y}
				h := hm.Get(coord)

				if h <= seaLevel {
					continue
				}

				isPeak := true
				// Check 4 direct neighbors (handles face wrapping)
				for _, dir := range directions {
					if hm.GetNeighborElevation(coord, dir) >= h {
						isPeak = false
						break
					}
				}

				if isPeak {
					// Calculate prominence/importance based on elevation relative to max
					// Use 0.6 as threshold for significance
					importance := (h - seaLevel) / (hm.MaxElev - seaLevel)
					if importance > 0.6 {
						candidates = append(candidates, createPOI(coord, h, POITypeMountainPeak, importance, hm))
					}
				}
			}
		}
	}

	// 2. Find Deep Ocean Trenches
	for f := 0; f < 6; f++ {
		for y := 0; y < res; y++ {
			for x := 0; x < res; x++ {
				coord := spatial.Coordinate{Face: f, X: x, Y: y}
				h := hm.Get(coord)

				if h >= seaLevel {
					continue
				}

				isDeepest := true
				for _, dir := range directions {
					if hm.GetNeighborElevation(coord, dir) <= h {
						isDeepest = false
						break
					}
				}

				if isDeepest {
					// Importance based on depth
					depth := seaLevel - h
					maxDepth := seaLevel - hm.MinElev
					if maxDepth == 0 {
						maxDepth = 1.0 // Avoid division by zero
					}
					importance := depth / maxDepth
					if importance > 0.7 {
						candidates = append(candidates, createPOI(coord, h, POITypeDeepOcean, importance, hm))
					}
				}
			}
		}
	}

	// Sort by importance (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Importance > candidates[j].Importance
	})

	// Apply limit
	if limit > 0 && len(candidates) > limit {
		return candidates[:limit]
	}

	return candidates
}

func createPOI(coord spatial.Coordinate, elevation float64, poiType POIType, importance float64, hm *SphereHeightmap) PointOfInterest {
	// Convert grid to lat/lon using topology
	topo := hm.Topology()
	sx, sy, sz := topo.ToSphere(coord)

	// Convert 3D unit sphere to Lat/Lon
	// y is up in this coordinate system (usually, check spatial)
	// assuming standard:
	// y = sin(lat)
	// z = cos(lat) * sin(lon)
	// x = cos(lat) * cos(lon)

	latRad := math.Asin(sy)
	lonRad := math.Atan2(sz, sx)

	lat := latRad * 180 / math.Pi
	lon := lonRad * 180 / math.Pi

	return PointOfInterest{
		ID:   uuid.New(),
		Type: poiType,
		Location: struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{X: coord.X, Y: coord.Y},
		Coordinates: Coordinates{
			Lat: lat,
			Lon: lon,
		},
		Elevation:   elevation,
		Importance:  importance,
		Description: generateDescription(poiType, elevation, importance),
	}
}

func generateDescription(t POIType, elevation, importance float64) string {
	switch t {
	case POITypeMountainPeak:
		if importance > 0.9 {
			return "A majestic summit dominating the landscape."
		}
		return "A prominent mountain peak."
	case POITypeDeepOcean:
		return "Abyssal trench depth."
	default:
		return "A notable landmark."
	}
}

// FindHighestPeak is a helper to find the single highest point
func FindHighestPeak(pois []PointOfInterest) *PointOfInterest {
	var highest *PointOfInterest
	for i := range pois {
		p := &pois[i]
		if p.Type == POITypeMountainPeak {
			if highest == nil || p.Elevation > highest.Elevation {
				highest = p
			}
		}
	}
	return highest
}
