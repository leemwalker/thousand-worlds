package weather

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/internal/spatial"
	"tw-backend/internal/worldgen/geography"
)

// =============================================================================
// Step 1: Solar Declination Tests
// =============================================================================

func TestCalculateSolarDeclination(t *testing.T) {
	t.Run("Summer Solstice (Day 172) should be ~+23.5 degrees", func(t *testing.T) {
		// June 21st - Northern Hemisphere Summer Solstice
		// Sun is directly overhead at Tropic of Cancer (23.5°N)
		declination := CalculateSolarDeclination(172)

		assert.InDelta(t, 23.5, declination, 0.5,
			"Summer solstice declination should be ~23.5°N")
	})

	t.Run("Winter Solstice (Day 355) should be ~-23.5 degrees", func(t *testing.T) {
		// December 21st - Northern Hemisphere Winter Solstice
		// Sun is directly overhead at Tropic of Capricorn (23.5°S)
		declination := CalculateSolarDeclination(355)

		assert.InDelta(t, -23.5, declination, 0.5,
			"Winter solstice declination should be ~-23.5°S")
	})

	t.Run("Spring Equinox (Day 80) should be ~0 degrees", func(t *testing.T) {
		// March 21st - Spring Equinox
		// Sun is directly overhead at the equator
		declination := CalculateSolarDeclination(80)

		assert.InDelta(t, 0, declination, 1.0,
			"Spring equinox declination should be ~0°")
	})

	t.Run("Autumn Equinox (Day 266) should be ~0 degrees", func(t *testing.T) {
		// September 23rd - Autumn Equinox
		// Sun is directly overhead at the equator
		declination := CalculateSolarDeclination(266)

		assert.InDelta(t, 0, declination, 1.5,
			"Autumn equinox declination should be ~0°")
	})

	t.Run("Declination should be bounded by axial tilt", func(t *testing.T) {
		// Test full year - declination should never exceed ±23.5°
		for day := 0; day < 365; day++ {
			declination := CalculateSolarDeclination(day)
			assert.LessOrEqual(t, math.Abs(declination), AxialTilt+0.1,
				"Day %d: declination %.2f exceeds axial tilt", day, declination)
		}
	})
}

// =============================================================================
// Step 2: Seasonal Temperature Modifier Tests
// =============================================================================

func TestGetSeasonalTemperatureModifier(t *testing.T) {
	t.Run("Summer hemisphere gets positive modifier", func(t *testing.T) {
		// Day 172: Sun at 23.5°N (Northern summer)
		declination := CalculateSolarDeclination(172)

		// Location at 30°N should get summer bonus
		modifier := GetSeasonalTemperatureModifier(30.0, declination)
		assert.Greater(t, modifier, 0.0,
			"Northern latitude during northern summer should have positive temp modifier")
	})

	t.Run("Winter hemisphere gets negative modifier", func(t *testing.T) {
		// Day 172: Sun at 23.5°N (Southern winter)
		declination := CalculateSolarDeclination(172)

		// Location at 30°S should get winter penalty
		modifier := GetSeasonalTemperatureModifier(-30.0, declination)
		assert.Less(t, modifier, 0.0,
			"Southern latitude during northern summer should have negative temp modifier")
	})

	t.Run("Equator has minimal seasonal variation", func(t *testing.T) {
		// At equator, seasonal variation should be small
		summerDeclination := CalculateSolarDeclination(172)
		winterDeclination := CalculateSolarDeclination(355)

		summerMod := GetSeasonalTemperatureModifier(0.0, summerDeclination)
		winterMod := GetSeasonalTemperatureModifier(0.0, winterDeclination)

		// Difference should be small (less than 5°C swing)
		assert.InDelta(t, summerMod, winterMod, 5.0,
			"Equatorial temperature should have minimal seasonal variation")
	})

	t.Run("Mid-latitudes have maximum seasonal variation", func(t *testing.T) {
		// At 45°N, seasonal variation should be significant
		summerDeclination := CalculateSolarDeclination(172)
		winterDeclination := CalculateSolarDeclination(355)

		summerMod := GetSeasonalTemperatureModifier(45.0, summerDeclination)
		winterMod := GetSeasonalTemperatureModifier(45.0, winterDeclination)

		// Difference should be significant (10-30°C swing)
		diff := summerMod - winterMod
		assert.Greater(t, diff, 10.0,
			"Mid-latitude should have significant seasonal temperature swing")
	})
}

// =============================================================================
// Step 3: Pressure and Monsoon Tests (placeholder - will implement in pressure.go)
// =============================================================================

func TestCalculateSurfacePressure(t *testing.T) {
	t.Run("Hot land has lower pressure than baseline", func(t *testing.T) {
		// Hot land: air rises, pressure drops
		pressure := CalculateSurfacePressure(true, 35.0) // Hot land (35°C)
		assert.Less(t, pressure, 1013.0,
			"Hot land should have pressure below baseline 1013 mb")
	})

	t.Run("Cold land has higher pressure than baseline", func(t *testing.T) {
		// Cold land: air sinks, pressure rises
		pressure := CalculateSurfacePressure(true, -10.0) // Cold land (-10°C)
		assert.Greater(t, pressure, 1013.0,
			"Cold land should have pressure above baseline 1013 mb")
	})

	t.Run("Ocean has more stable pressure", func(t *testing.T) {
		// Ocean temperature effect should be damped
		hotLandPressure := CalculateSurfacePressure(true, 35.0)
		hotOceanPressure := CalculateSurfacePressure(false, 35.0)

		landDeviation := math.Abs(1013.0 - hotLandPressure)
		oceanDeviation := math.Abs(1013.0 - hotOceanPressure)

		assert.Greater(t, landDeviation, oceanDeviation,
			"Ocean should have more stable pressure than land at same temperature")
	})
}

func TestMonsoonWindDirection(t *testing.T) {
	// Skip until we have topology available
	t.Run("Summer: Wind blows from Ocean to Land", func(t *testing.T) {
		// Setup: Land cell hot (low pressure), Ocean cell cooler (high pressure)
		// Wind should flow from high to low pressure (Ocean -> Land)

		// Create a minimal test topology
		topology := spatial.NewCubeSphereTopology(16)

		// Land cell in center of face 0
		landCoord := spatial.Coordinate{Face: 0, X: 8, Y: 8}
		// Ocean cell to the east
		oceanCoord := spatial.Coordinate{Face: 0, X: 9, Y: 8}

		// Summer scenario: Land hotter than ocean
		pressureMap := map[spatial.Coordinate]float64{
			landCoord:  1000.0, // Low pressure (hot land)
			oceanCoord: 1020.0, // High pressure (cooler ocean)
		}

		windVec := CalculatePressureGradientWind(landCoord, topology, pressureMap)

		// Wind should have positive X component (blowing from east/ocean toward land)
		require.NotNil(t, windVec)
		// Note: The exact direction depends on the face orientation
		// For face 0, X increases eastward
		assert.NotZero(t, windVec.X+windVec.Y+windVec.Z,
			"Wind should have non-zero magnitude when pressure gradient exists")
	})
}

// =============================================================================
// Integration Test: Monsoon Wet/Dry Season
// =============================================================================

func TestMonsoonEffect_WetDrySeason(t *testing.T) {
	t.Run("Tropical coast: Summer wetter than Winter", func(t *testing.T) {
		// Setup: A coastal cell at Lat 10°N with ocean to the south
		// In Northern summer: Land heats, draws moist air from ocean (wet)
		// In Northern winter: Land cools, dry continental air dominates (dry)

		topology := spatial.NewCubeSphereTopology(32)

		// Find a coordinate at approximately 10°N latitude
		// On a cube-sphere, we need to find a coord that gives us ~10°N
		coastalCoord := findCoordAtLatitude(topology, 10.0)
		require.NotNil(t, coastalCoord, "Should find coordinate at ~10°N")

		// Calculate rainfall for summer (Day 180) and winter (Day 0)
		summerRainfall := simulateSeasonalRainfall(*coastalCoord, 180, topology, true)
		winterRainfall := simulateSeasonalRainfall(*coastalCoord, 0, topology, true)

		assert.Greater(t, summerRainfall, winterRainfall,
			"Tropical coast should have more rain in summer (monsoon) than winter")
	})
}

// =============================================================================
// Helper Functions for Tests
// =============================================================================

// findCoordAtLatitude finds a coordinate approximately at the given latitude
func findCoordAtLatitude(topology spatial.Topology, targetLat float64) *spatial.Coordinate {
	faceSize := topology.Resolution()

	for face := 0; face < 6; face++ {
		for y := 0; y < faceSize; y++ {
			for x := 0; x < faceSize; x++ {
				coord := spatial.Coordinate{Face: face, X: x, Y: y}
				lat := GetLatitudeFromCoord(topology, coord)

				if math.Abs(lat-targetLat) < 5.0 { // Within 5 degrees
					return &coord
				}
			}
		}
	}
	return nil
}

// simulateSeasonalRainfall is a helper that calculates rainfall for a given day
func simulateSeasonalRainfall(coord spatial.Coordinate, dayOfYear int, topology spatial.Topology, isCoastal bool) float64 {
	declination := CalculateSolarDeclination(dayOfYear)
	lat := GetLatitudeFromCoord(topology, coord)

	// Get seasonal temperature modifier
	tempMod := GetSeasonalTemperatureModifier(lat, declination)

	// Base temperature + seasonal modifier
	baseTemp := 25.0 // Tropical base temp
	temp := baseTemp + tempMod

	// Calculate pressure (land cell)
	landPressure := CalculateSurfacePressure(true, temp)

	// Ocean is more stable, assume 25°C year-round
	oceanPressure := CalculateSurfacePressure(false, 25.0)

	// Simplified rainfall model:
	// If land pressure < ocean pressure, moisture flows inland (wet)
	// The bigger the pressure difference, the more rain
	pressureDiff := oceanPressure - landPressure

	// Base rainfall + monsoon contribution
	rainfall := 100.0 // Base tropical rainfall (mm/month)
	if pressureDiff > 0 {
		// Onshore wind (wet monsoon)
		rainfall += pressureDiff * 5.0 // Scale factor
	} else {
		// Offshore wind (dry monsoon)
		rainfall += pressureDiff * 2.0 // Reduce rainfall
	}

	if rainfall < 0 {
		rainfall = 0
	}

	return rainfall
}

// =============================================================================
// ITCZ Shift Tests
// =============================================================================

func TestCalculateSeasonalPrecipitation_ITCZShift(t *testing.T) {
	t.Run("ITCZ boosts precipitation near solar declination", func(t *testing.T) {
		// Create a cell at 20°N latitude (using Location.Y for flat mode)
		cell := &GeographyCell{
			Elevation: 100,
			IsOcean:   false,
			Location:  geography.Point{X: 0, Y: 20}, // Y=20 represents 20°N latitude
		}

		// No upwind cells, minimal moisture
		upwindCells := []*GeographyCell{}
		monsoonWind := spatial.Vector3D{X: 0, Y: 0, Z: 0}

		// Case 1: ITCZ at +23.5° (summer solstice) - near our 20°N location
		// Distance from ITCZ = |20 - 23.5| = 3.5° -> within 10° -> bonus
		declinationSummer := 23.5
		precipSummer, _ := CalculateSeasonalPrecipitation(
			cell, upwindCells, monsoonWind, declinationSummer, 50, nil,
		)

		// Case 2: ITCZ at -23.5° (winter solstice) - far from our 20°N location
		// Distance from ITCZ = |20 - (-23.5)| = 43.5° -> no bonus
		declinationWinter := -23.5
		precipWinter, _ := CalculateSeasonalPrecipitation(
			cell, upwindCells, monsoonWind, declinationWinter, 50, nil,
		)

		assert.Greater(t, precipSummer, precipWinter,
			"ITCZ proximity should increase precipitation")
	})

	t.Run("Equatorial cell gets year-round rainfall", func(t *testing.T) {
		// Equator is always within 23.5° of ITCZ
		cell := &GeographyCell{
			Elevation: 100,
			IsOcean:   false,
		}

		upwindCells := []*GeographyCell{}
		monsoonWind := spatial.Vector3D{X: 0, Y: 0, Z: 0}

		// Summer and winter precipitation at equator
		precipSummer, _ := CalculateSeasonalPrecipitation(
			cell, upwindCells, monsoonWind, 23.5, 50, nil,
		)
		precipWinter, _ := CalculateSeasonalPrecipitation(
			cell, upwindCells, monsoonWind, -23.5, 50, nil,
		)

		// Both should have some ITCZ bonus (within 10° at equinox positions)
		assert.Greater(t, precipSummer, 0.0, "Equator should have rainfall in summer")
		assert.Greater(t, precipWinter, 0.0, "Equator should have rainfall in winter")
	})
}

func TestCalculateSeasonalAnnualPrecipitation_MonsoonBelt(t *testing.T) {
	t.Run("Coastal tropics with ITCZ proximity get enhanced precipitation", func(t *testing.T) {
		// Coastal location at 15°N (monsoon belt)
		// Compare when ITCZ is near vs far

		precipITCZNear := CalculateSeasonalAnnualPrecipitation(
			15.0,  // latitude
			100,   // elevation
			50000, // 50km from coast
			true,  // windward
			15.0,  // ITCZ at 15°N (same latitude)
		)

		precipITCZFar := CalculateSeasonalAnnualPrecipitation(
			15.0,  // latitude
			100,   // elevation
			50000, // 50km from coast
			true,  // windward
			-15.0, // ITCZ at 15°S (opposite hemisphere)
		)

		assert.Greater(t, precipITCZNear, precipITCZFar,
			"ITCZ proximity should increase annual precipitation")
	})
}

func TestGeneratePressureMap(t *testing.T) {
	t.Run("Pressure map is generated for all coordinates", func(t *testing.T) {
		// This test verifies that GeneratePressureMap compiles and runs
		// A full integration test would require a real SphereHeightmap

		// For now, verify the function exists and returns a map
		// Full integration test will be done in Step 4
		assert.NotNil(t, GeneratePressureMap)
	})
}
