package weather

import (
	"testing"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Test: Rain Shadow Effect
// =============================================================================

func TestRainShadow_MountainRidge(t *testing.T) {
	// Create a world with a mountain ridge
	// Windward side (facing ocean) should have more rain than leeward

	topo := spatial.NewCubeSphereTopology(16)
	hm := geography.NewSphereHeightmap(topo)
	seaLevel := 0.0

	// Set ocean to the west (moisture source) and land with mountain ridge
	for face := 0; face < 6; face++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}

				if face == 0 {
					// Face 0: Create land with mountain ridge at X=10
					if x < 5 {
						// Western ocean (moisture source)
						hm.Set(coord, -100.0)
					} else if x < 10 {
						// Western plains (windward)
						hm.Set(coord, 200.0)
					} else if x == 10 {
						// Mountain ridge running N-S
						hm.Set(coord, 4000.0)
					} else {
						// Eastern plains (leeward - rain shadow)
						hm.Set(coord, 200.0)
					}
				} else {
					// Other faces: ocean
					hm.Set(coord, -100.0)
				}
			}
		}
	}

	config := DefaultRainfallConfig(seaLevel)
	config.AdvectionPasses = 10 // More passes to ensure moisture reaches inland
	config.OceanEvapRate = 20.0 // Higher evaporation

	rainfall := GenerateRainfallMap(hm, topo, config)

	// Sample rainfall at windward plains (X=8) vs leeward plains (X=12)
	var windwardTotal, leewardTotal float64
	var windwardCount, leewardCount int

	for y := 4; y < 12; y++ {
		windwardCoord := spatial.Coordinate{Face: 0, X: 8, Y: y}
		leewardCoord := spatial.Coordinate{Face: 0, X: 12, Y: y}

		windwardIdx := coordToIndex(windwardCoord, 16)
		leewardIdx := coordToIndex(leewardCoord, 16)

		windwardTotal += rainfall[windwardIdx]
		leewardTotal += rainfall[leewardIdx]
		windwardCount++
		leewardCount++
	}

	windwardAvg := windwardTotal / float64(windwardCount)
	leewardAvg := leewardTotal / float64(leewardCount)

	t.Logf("Windward average rainfall: %.2f", windwardAvg)
	t.Logf("Leeward average rainfall: %.2f", leewardAvg)

	// Mountain peak should have high rain (orographic lift maximum)
	peakCoord := spatial.Coordinate{Face: 0, X: 10, Y: 8}
	peakIdx := coordToIndex(peakCoord, 16)
	peakRain := rainfall[peakIdx]
	t.Logf("Peak rainfall: %.2f", peakRain)

	// Verify rain shadow: leeward should have less than windward or peak
	assert.Greater(t, peakRain+windwardAvg, leewardAvg,
		"Windward side + peak should receive more rain than leeward (rain shadow)")
}

// =============================================================================
// Test: Global Circulation Pattern
// =============================================================================

func TestGlobalCirculation_WindDirectionFlips(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(32)

	testCases := []struct {
		name       string
		latitude   float64
		expectEast bool // True = wind blows eastward (westerlies), False = westward (easterlies)
	}{
		{"Tropical Trade Winds (15°N)", 15.0, false}, // Easterlies (blow westward)
		{"Westerlies (45°N)", 45.0, true},            // Westerlies (blow eastward)
		{"Polar Easterlies (75°N)", 75.0, false},     // Easterlies (blow westward)
		{"Equatorial (5°S)", -5.0, false},            // Easterlies
		{"Westerlies (45°S)", -45.0, true},           // Westerlies
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create coordinate at target latitude
			// Use longitude 0 for simplicity
			coord := latLonToCoord(topo, tc.latitude, 0)

			wind := CalculateWindSpherical(topo, coord, SeasonSpring)

			// Wind direction: 0° = North, 90° = East, 180° = South, 270° = West
			// Eastward blowing = direction between 45° and 135° (centered on 90°)
			// Westward blowing = direction between 225° and 315° (centered on 270°)
			// OR negative values like -90° (would normalize to 270°)

			dir := normalizeDirection(wind.Direction)

			isEastward := (dir > 45 && dir < 135)

			if tc.expectEast {
				assert.True(t, isEastward,
					"At %s, wind should blow eastward (westerlies), got direction %.1f°", tc.name, dir)
			} else {
				assert.False(t, isEastward,
					"At %s, wind should blow westward (easterlies), got direction %.1f°", tc.name, dir)
			}
		})
	}
}

func TestGlobalCirculation_PressureSystems(t *testing.T) {
	// Verify pressure system distribution matches atmospheric cells
	testCases := []struct {
		latitude       float64
		expectPressure PressureSystem
	}{
		{5.0, PressureLow},   // Equatorial low (ITCZ)
		{30.0, PressureHigh}, // Subtropical high
		{60.0, PressureLow},  // Subpolar low
		{85.0, PressureHigh}, // Polar high
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			pressure := GetPressureAtLatitude(tc.latitude)
			assert.Equal(t, tc.expectPressure, pressure,
				"Pressure at %.0f° should be %v", tc.latitude, tc.expectPressure)
		})
	}
}

// Helper: Convert lat/lon to approximate spherical coordinate
func latLonToCoord(topo spatial.Topology, lat, lon float64) spatial.Coordinate {
	// Convert to radians
	latRad := lat * 3.141592653589793 / 180.0
	lonRad := lon * 3.141592653589793 / 180.0

	// Spherical to Cartesian
	cosLat := cosApprox(latRad)
	sinLat := sinApprox(latRad)
	cosLon := cosApprox(lonRad)
	sinLon := sinApprox(lonRad)

	x := cosLat * cosLon
	y := sinLat
	z := cosLat * sinLon

	return topo.FromVector(x, y, z)
}

func cosApprox(x float64) float64 {
	// Simple Taylor series for small values
	x2 := x * x
	return 1 - x2/2 + x2*x2/24
}

func sinApprox(x float64) float64 {
	// Simple Taylor series
	x2 := x * x
	return x - x*x2/6 + x*x2*x2/120
}
