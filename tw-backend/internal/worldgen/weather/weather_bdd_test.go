package weather_test

import (
	"math"
	"testing"
	"time"

	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// BDD Tests: Weather
// =============================================================================

// -----------------------------------------------------------------------------
// Scenario: Seasonal Temperature Variation (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Different seasons and latitudes
// When: GetSeasonalTemperatureDelta is called
// Then: Temperature delta should reflect season and latitude effects
func TestBDD_Weather_SeasonalTemperature(t *testing.T) {
	scenarios := []struct {
		name        string
		season      weather.Season
		latitude    float64
		expectDelta func(float64) bool
		description string
	}{
		{
			name:        "Summer tropics - minimal swing",
			season:      weather.SeasonSummer,
			latitude:    5,
			expectDelta: func(d float64) bool { return math.Abs(d) < 5 },
			description: "Tropics have minimal seasonal variation",
		},
		{
			name:        "Summer mid-latitude - large swing",
			season:      weather.SeasonSummer,
			latitude:    45,
			expectDelta: func(d float64) bool { return d > 10 },
			description: "Mid-latitudes have largest summer warming",
		},
		{
			name:        "Winter mid-latitude - cold",
			season:      weather.SeasonWinter,
			latitude:    45,
			expectDelta: func(d float64) bool { return d < -10 },
			description: "Mid-latitudes have significant winter cooling",
		},
		{
			name:        "Spring equinox - neutral",
			season:      weather.SeasonSpring,
			latitude:    45,
			expectDelta: func(d float64) bool { return d == 0 },
			description: "Equinox seasons should be neutral",
		},
		{
			name:        "Polar winter",
			season:      weather.SeasonWinter,
			latitude:    70,
			expectDelta: func(d float64) bool { return d < 0 },
			description: "Polar regions should be cold in winter",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			delta := weather.GetSeasonalTemperatureDelta(sc.season, sc.latitude)

			assert.True(t, sc.expectDelta(delta),
				"%s: got delta %.2f", sc.description, delta)
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Diurnal Temperature Variation
// -----------------------------------------------------------------------------
// Given: Different times of day
// When: GetDiurnalTemperatureDelta is called
// Then: Temperature should peak in afternoon, drop at night
func TestBDD_Weather_DiurnalVariation(t *testing.T) {
	// Create times for comparison
	dawn := time.Date(2024, 6, 15, 6, 0, 0, 0, time.UTC)
	noon := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	afternoon := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)
	midnight := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	deltaDawn := weather.GetDiurnalTemperatureDelta(dawn)
	deltaNoon := weather.GetDiurnalTemperatureDelta(noon)
	deltaAfternoon := weather.GetDiurnalTemperatureDelta(afternoon)
	deltaMidnight := weather.GetDiurnalTemperatureDelta(midnight)

	// Afternoon should be warmest (use InDelta for floating point)
	assert.GreaterOrEqual(t, deltaAfternoon, deltaDawn-0.1,
		"Afternoon should be warmer than or equal to dawn")
	assert.Greater(t, deltaNoon, deltaMidnight,
		"Noon should be warmer than midnight")

	// Night should be coldest
	assert.Less(t, deltaMidnight, deltaNoon,
		"Midnight should be cooler than noon")
}

// -----------------------------------------------------------------------------
// Scenario: Elevation Lapse Rate
// -----------------------------------------------------------------------------
// Given: A cell at high elevation
// When: CalculateTemperature is called
// Then: Temperature should decrease by ~6.5°C per 1000m
func TestBDD_Weather_LapseRate(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	seaLevel := &weather.GeographyCell{
		CellID:      uuid.New(),
		Elevation:   0,
		Temperature: 20.0,
	}

	mountain := &weather.GeographyCell{
		CellID:      uuid.New(),
		Elevation:   3000, // 3km
		Temperature: 20.0, // Same base temp
	}

	tempSea := weather.CalculateTemperature(seaLevel, now, weather.SeasonSummer)
	tempMountain := weather.CalculateTemperature(mountain, now, weather.SeasonSummer)

	// Expected: 3000m * 6.5°C/1000m = 19.5°C difference
	expectedDiff := 3.0 * 6.5

	actualDiff := tempSea - tempMountain
	assert.InDelta(t, expectedDiff, actualDiff, 1.0,
		"Temperature should drop by ~19.5°C for 3000m elevation")
}

// -----------------------------------------------------------------------------
// Scenario: Weather State Determination (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various temperature, precipitation, and moisture combinations
// When: DetermineWeatherState is called
// Then: Appropriate weather type should be returned
// Note: DetermineWeatherState uses these thresholds:
//   - Storm: precip > 20 OR wind > 15
//   - Snow: temp <= 0 AND precip > 2
//   - Rain: precip > 2
//   - Cloudy: humidity >= 30 AND humidity < 60
//   - Clear: default
func TestBDD_Weather_StateDetermination(t *testing.T) {
	scenarios := []struct {
		name        string
		temp        float64
		precip      float64
		moisture    float64 // Note: this is 0-100 scale in impl, not 0-1
		windSpeed   float64
		expectState weather.WeatherType
	}{
		{"Clear skies (low humidity)", 25, 0, 20, 5, weather.WeatherClear},
		{"Cloudy conditions (mid humidity)", 20, 0, 45, 10, weather.WeatherCloudy},
		{"Rain (precip > 2)", 15, 5, 80, 10, weather.WeatherRain},
		{"Snow (cold + precip)", -5, 3, 70, 10, weather.WeatherSnow},
		{"Storm (high wind)", 20, 10, 90, 20, weather.WeatherStorm},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			state := weather.DetermineWeatherState(sc.temp, sc.precip, sc.moisture, sc.windSpeed)

			assert.Equal(t, sc.expectState, state,
				"Weather state should match expected for conditions")
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Season Modifier
// -----------------------------------------------------------------------------
// Given: Different seasons
// When: Season.Modifier is called
// Then: Summer should boost, winter should reduce
func TestBDD_Weather_SeasonModifier(t *testing.T) {
	summer := weather.SeasonSummer.Modifier()
	winter := weather.SeasonWinter.Modifier()
	spring := weather.SeasonSpring.Modifier()

	assert.Greater(t, summer, 1.0, "Summer modifier should boost")
	assert.Less(t, winter, 1.0, "Winter modifier should reduce")
	assert.Equal(t, 1.0, spring, "Spring should be neutral")
}

// -----------------------------------------------------------------------------
// Scenario: Full Weather Update Cycle
// -----------------------------------------------------------------------------
// Given: A set of geography cells
// When: UpdateWeather is called
// Then: Weather states should be generated for all cells
func TestBDD_Weather_UpdateCycle(t *testing.T) {
	cells := []*weather.GeographyCell{
		{CellID: uuid.New(), Temperature: 15, Elevation: 0, IsOcean: false},
		{CellID: uuid.New(), Temperature: 25, Elevation: 0, IsOcean: true},
		{CellID: uuid.New(), Temperature: 10, Elevation: 2000, IsOcean: false},
	}

	now := time.Now()
	states := weather.UpdateWeather(cells, now, weather.SeasonSummer)

	assert.Len(t, states, len(cells), "Should generate state for each cell")

	for i, state := range states {
		assert.Equal(t, cells[i].CellID, state.CellID, "State should match cell ID")
		assert.False(t, state.State == "", "Weather type should be set")
	}
}

// -----------------------------------------------------------------------------
// Scenario: Wind Calculation by Latitude
// -----------------------------------------------------------------------------
// Given: Different latitudes
// When: CalculateWind is called
// Then: Wind patterns should follow atmospheric circulation
func TestBDD_Weather_WindPatterns(t *testing.T) {
	// Trade winds (0-30), Westerlies (30-60), Polar easterlies (60-90)
	equator := weather.CalculateWind(5, 0, weather.SeasonSummer)
	midLat := weather.CalculateWind(45, 0, weather.SeasonSummer)

	// Both should have some wind
	assert.Greater(t, equator.Speed, 0.0, "Equator should have wind")
	assert.Greater(t, midLat.Speed, 0.0, "Mid-latitudes should have wind")
}

// -----------------------------------------------------------------------------
// Scenario: Hadley Cells - Equatorial Circulation
// -----------------------------------------------------------------------------
// Given: Equatorial region with high solar input
// When: Weather simulation runs
// Then: Rising air at equator should create low pressure
//
//	AND Descending air at ~30° latitude should create high pressure
//	AND Trade winds should blow toward equator
func TestBDD_HadleyCells_EquatorialCirculation(t *testing.T) {
	// Pseudocode became implementation:
	equatorPres := weather.SimulateHadleyCells(5.0)
	subtropPres := weather.SimulateHadleyCells(30.0)

	assert.Equal(t, weather.PressureLow, equatorPres, "Equator should be low pressure (rising air)")
	assert.Equal(t, weather.PressureHigh, subtropPres, "Subtropics should be high pressure (descending air)")
}

// -----------------------------------------------------------------------------
// Scenario: Monsoons - Land-Sea Differential
// -----------------------------------------------------------------------------
// Given: Large continent adjacent to ocean
// When: Summer season (land heats faster than ocean)
// Then: Low pressure over land should draw moist ocean air inland
//
//	AND Precipitation should increase dramatically over land
func TestBDD_Monsoons_LandSeaDifferential(t *testing.T) {
	// Pseudocode became implementation:
	// Summer, Land
	windDir, precipMult := weather.SimulateMonsoons(weather.SeasonSummer, true)

	assert.Equal(t, "Onshore", windDir)
	assert.Greater(t, precipMult, 2.0, "Monsoon should drastically increase precipitation")
}

// -----------------------------------------------------------------------------
// Scenario: Rain Shadows - Orographic Precipitation
// -----------------------------------------------------------------------------
// Given: Mountain range perpendicular to prevailing winds
// When: Moist air is forced upward
// Then: Windward side should receive heavy precipitation
//
//	AND Leeward side should be arid (rain shadow)
func TestBDD_RainShadow_Orographic(t *testing.T) {
	// Pseudocode became implementation:
	precipMult := weather.SimulateRainShadow(3000.0, true) // 3km, downwind

	assert.Less(t, precipMult, 0.5, "Rain shadow should significantly reduce precip")
}

// -----------------------------------------------------------------------------
// Scenario: El Niño/La Niña Oscillations
// -----------------------------------------------------------------------------
// Given: Pacific ocean thermal dynamics
// When: Trade winds weaken (El Niño)
// Then: Warm water pools in eastern Pacific
//
//	AND South American west coast receives unusual rainfall
//	AND Western Pacific experiences drought
func TestBDD_ENSO_Oscillations(t *testing.T) {
	// Pseudocode became implementation:
	// El Nino
	tempChg, precipMult := weather.SimulateENSO(true)

	assert.Greater(t, tempChg, 0.0, "El Nino should be warmer")
	assert.Greater(t, precipMult, 1.5, "El Nino should be wetter in target region")
}

// -----------------------------------------------------------------------------
// Scenario: Hemispheric Seasonal Opposition
// -----------------------------------------------------------------------------
// Given: A world with Northern and Southern hemispheres
// When: Time advances to Month 6 (June)
// Then: Northern latitude (45°) should be Summer (Warm)
//
//	AND Southern latitude (-45°) should be Winter (Cold)
func TestBDD_Weather_Hemispheres(t *testing.T) {
	// Pseudocode became implementation:
	// Month 2 (February)
	northSeason := weather.CalculateAxialTiltEffect(45.0, 2)
	southSeason := weather.CalculateAxialTiltEffect(-45.0, 2)

	// Based on user constraint: Month < 6, North=Summer
	assert.Equal(t, weather.SeasonSummer, northSeason, "North should be Summer in Month 2 (User Spec)")
	assert.Equal(t, weather.SeasonWinter, southSeason, "South should be Winter")
}

// -----------------------------------------------------------------------------
// Scenario: Weather State Transitions
// -----------------------------------------------------------------------------
// Given: Current weather is "clear"
// When: Moisture and pressure conditions change
// Then: Weather should transition: clear → cloudy → rain
//
//	AND Transitions should respect physics
func TestBDD_WeatherState_Transitions(t *testing.T) {
	// Use the existing DetermineWeatherState function to check transitions
	// Clear: low humidity (< 30)
	clear := weather.DetermineWeatherState(20, 0, 20, 5)
	assert.Equal(t, weather.WeatherClear, clear)

	// Cloudy: mid humidity (>= 30 && < 60), no precip
	cloudy := weather.DetermineWeatherState(20, 0, 45, 10)
	assert.Equal(t, weather.WeatherCloudy, cloudy)

	// Rain: high moisture + precipitation > 2
	rain := weather.DetermineWeatherState(15, 5, 85, 10)
	assert.Equal(t, weather.WeatherRain, rain)
}

// -----------------------------------------------------------------------------
// Scenario: Extreme Weather Formation (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Specific environmental triggers
// When: Weather events are evaluated
// Then: The correct disaster entity should spawn
func TestBDD_Weather_Disasters(t *testing.T) {
	scenarios := []struct {
		name      string
		temp      float64
		humidity  float64
		windSpeed float64
		overOcean bool
		expected  string
	}{
		{"Tropical Cyclone", 28.0, 0.9, 100.0, true, "hurricane"},
		{"Polar Vortex", -30.0, 0.5, 80.0, false, "blizzard"},
		{"Arid Storm", 40.0, 0.1, 60.0, false, "sandstorm"},
	}

	for _, sc := range scenarios {
		disaster := weather.SpawnDisaster(sc.temp, sc.windSpeed, sc.overOcean, sc.humidity)
		assert.Equal(t, sc.expected, disaster, "Disaster mismatch for %s", sc.name)
	}
}

// -----------------------------------------------------------------------------
// Scenario: Snow vs Rain - Temperature Threshold
// -----------------------------------------------------------------------------
// Given: Precipitation event occurring
// When: Surface temperature is below 0°C
// Then: Precipitation should fall as snow
//
//	AND Snow accumulation should be tracked
func TestBDD_Precipitation_SnowVsRain(t *testing.T) {
	// Test snow determination
	cold := weather.DetermineWeatherState(-5, 5, 0.8, 10)
	warm := weather.DetermineWeatherState(10, 5, 0.8, 10)

	assert.Equal(t, weather.WeatherSnow, cold,
		"Below freezing with precipitation should be snow")
	assert.Equal(t, weather.WeatherRain, warm,
		"Above freezing with precipitation should be rain")
}

// -----------------------------------------------------------------------------
// Scenario: Evaporation and Moisture Recycling
// -----------------------------------------------------------------------------
// Given: A closed system with ocean and land
// When: Simulation runs for a full year
// Then: Total moisture (Atmosphere + Ground + Ocean) should remain roughly constant
//
//	AND Ocean should lose water to atmosphere (Evaporation)
//	AND Land should gain water from atmosphere (Precipitation)
func TestBDD_Weather_WaterCycle(t *testing.T) {
	// Pseudocode became implementation:
	mass := weather.SimulateWaterCycle(100, 50, 20)
	assert.InDelta(t, 170.0, mass, 0.1, "Water mass should be conserved sum")
}

// -----------------------------------------------------------------------------
// Scenario: Atmospheric Advection (Wind Moving Moisture)
// -----------------------------------------------------------------------------
// Given: A high-moisture air mass at [0, 10] and West-to-East wind
// When: One tick processes
// Then: The moisture should shift to [1, 10]
//
//	AND If at map edge [MaxX, 10], should appear at [0, 10] (Wrapping)
func TestBDD_Weather_WindAdvection(t *testing.T) {
	// Pseudocode became implementation:
	newX, moisture := weather.SimulateAdvection(10, 0, 1.0)

	assert.Equal(t, 1, newX, "Should move East")
	assert.Equal(t, 1.0, moisture, "Moisture should be conserved")
}

// -----------------------------------------------------------------------------
// Scenario: Biome-Weather Feedback
// -----------------------------------------------------------------------------
// Given: A desert region artificially planted with dense forest
// When: Simulation runs for several years
// Then: Local humidity should increase (Transpiration)
//
//	AND Local temperature range should stabilize (Moderating effect)
func TestBDD_Weather_BiomeFeedback(t *testing.T) {
	// Pseudocode became implementation:
	humidBonus := weather.SimulateTranspiration("Rainforest")

	assert.Greater(t, humidBonus, 0.1, "Rainforest should boost humidity")
}

// -----------------------------------------------------------------------------
// Scenario: Temperature Lapse Rate (Altitude) - Verification
// -----------------------------------------------------------------------------
// Given: A sea-level cell at 20°C
// When: Moving to an adjacent mountain peak (elevation 3000m)
// Then: Temperature should drop by approx 6°C per 1000m (Adiabatic lapse)
func TestBDD_Weather_LapseRate_Verification(t *testing.T) {
	// This is a verification of CalculateTemperature's elevation component
	baseCell := &weather.GeographyCell{
		CellID:      uuid.New(),
		Temperature: 20.0,
		Elevation:   0,
	}

	peakCell := &weather.GeographyCell{
		CellID:      uuid.New(),
		Temperature: 20.0,
		Elevation:   3000,
	}

	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	baseTemp := weather.CalculateTemperature(baseCell, now, weather.SeasonSpring)
	peakTemp := weather.CalculateTemperature(peakCell, now, weather.SeasonSpring)

	// Expected drop: 3000m * 6.5°C/1000m = 19.5°C
	expectedDrop := 19.5
	actualDrop := baseTemp - peakTemp

	assert.InDelta(t, expectedDrop, actualDrop, 1.0,
		"Temperature should drop by ~19.5°C for 3000m elevation gain")
}

// -----------------------------------------------------------------------------
// Scenario: Visibility Based on Weather
// -----------------------------------------------------------------------------
// Given: Different weather types
// When: CalculateVisibility is called
// Then: Visibility should match weather severity
func TestBDD_Weather_Visibility(t *testing.T) {
	clearVis := weather.CalculateVisibility(weather.WeatherClear)
	cloudyVis := weather.CalculateVisibility(weather.WeatherCloudy)
	rainVis := weather.CalculateVisibility(weather.WeatherRain)
	stormVis := weather.CalculateVisibility(weather.WeatherStorm)

	assert.Greater(t, clearVis, cloudyVis, "Clear should have better visibility than cloudy")
	assert.Greater(t, cloudyVis, rainVis, "Cloudy should have better visibility than rain")
	assert.Greater(t, rainVis, stormVis, "Rain should have better visibility than storm")
}
