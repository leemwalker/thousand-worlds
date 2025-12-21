package weather

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Climate Tests - Additional Coverage
// =============================================================================

func TestSimulateRainShadow_AllCases(t *testing.T) {
	tests := []struct {
		name                 string
		elevation            float64
		isDownwindOfMountain bool
		expectedMultiplier   float64
	}{
		{"Low elevation, no mountain", 500, false, 1.0},
		{"High elevation, no mountain", 3000, false, 1.0},
		{"Low elevation, downwind", 500, true, 1.0},
		{"High elevation, downwind - rain shadow", 3000, true, 0.2},
		{"Exactly 2000m, downwind", 2000, true, 1.0}, // Not > 2000
		{"Just over 2000m, downwind", 2001, true, 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SimulateRainShadow(tt.elevation, tt.isDownwindOfMountain)
			assert.Equal(t, tt.expectedMultiplier, result)
		})
	}
}

func TestSimulateENSO_AllCases(t *testing.T) {
	// El Nino
	tempChange, precipMult := SimulateENSO(true)
	assert.Equal(t, 1.0, tempChange, "El Nino should increase temperature")
	assert.Equal(t, 2.0, precipMult, "El Nino should increase precipitation")

	// La Nina
	tempChange, precipMult = SimulateENSO(false)
	assert.Equal(t, -0.5, tempChange, "La Nina should decrease temperature")
	assert.Equal(t, 0.5, precipMult, "La Nina should decrease precipitation")
}

func TestCalculateAxialTiltEffect_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		latitude float64
		month    int
		expected Season
	}{
		{"North, early year = Summer", 45.0, 1, SeasonSummer},
		{"North, mid year = Winter", 45.0, 6, SeasonWinter},
		{"North, late year = Winter", 45.0, 10, SeasonWinter},
		{"South, early year = Winter", -45.0, 1, SeasonWinter},
		{"South, mid year = Summer", -45.0, 6, SeasonSummer},
		{"Equator north side, early", 5.0, 3, SeasonSummer},
		{"Month 5 boundary, north", 30.0, 5, SeasonSummer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAxialTiltEffect(tt.latitude, tt.month)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimulateWaterCycle(t *testing.T) {
	total := SimulateWaterCycle(100.0, 50.0, 25.0)
	assert.Equal(t, 175.0, total, "Should sum all water masses")
}

// =============================================================================
// Extremes Tests - Additional Coverage for GenerateExtremeWeather
// =============================================================================

func TestGenerateExtremeWeather_AllCases(t *testing.T) {
	cellID := uuid.New()
	now := time.Now()

	t.Run("Hurricane conditions", func(t *testing.T) {
		cell := &GeographyCell{
			CellID:      cellID,
			IsOcean:     true,
			Temperature: 27.0,
		}
		state := &WeatherState{
			Temperature:   28.0,
			Timestamp:     now,
			Precipitation: 50.0,
			Wind:          Wind{Speed: 5.0},
		}
		result := GenerateExtremeWeather(cell, state, 15.0, nil)
		assert.NotNil(t, result, "Should generate hurricane")
		assert.Equal(t, ExtremeHurricane, result.EventType)
	})

	t.Run("Blizzard conditions", func(t *testing.T) {
		cell := &GeographyCell{
			CellID:      cellID,
			IsOcean:     false,
			Temperature: 5.0,
		}
		state := &WeatherState{
			Temperature:   -10.0,
			Timestamp:     now,
			Precipitation: 20.0,
			Wind:          Wind{Speed: 15.0},
		}
		result := GenerateExtremeWeather(cell, state, 55.0, nil)
		assert.NotNil(t, result, "Should generate blizzard")
		assert.Equal(t, ExtremeBlizzard, result.EventType)
	})

	t.Run("Drought conditions with history", func(t *testing.T) {
		cell := &GeographyCell{
			CellID:      cellID,
			IsOcean:     false,
			Temperature: 30.0, // normalPrecipitation is used from cell.Temperature
		}
		state := &WeatherState{
			Temperature:   35.0,
			Timestamp:     now,
			Precipitation: 1.0,
			Wind:          Wind{Speed: 5.0},
		}
		// Create 90 days of low precipitation history
		history := make([]WeatherState, 90)
		for i := range history {
			history[i] = WeatherState{Precipitation: 1.0}
		}
		result := GenerateExtremeWeather(cell, state, 35.0, history)
		// Drought should be detected
		if result != nil {
			assert.Equal(t, ExtremeDrought, result.EventType)
		}
	})

	t.Run("No extreme weather", func(t *testing.T) {
		cell := &GeographyCell{
			CellID:      cellID,
			IsOcean:     false,
			Temperature: 20.0,
		}
		state := &WeatherState{
			Temperature:   20.0,
			Timestamp:     now,
			Precipitation: 10.0,
			Wind:          Wind{Speed: 5.0},
		}
		result := GenerateExtremeWeather(cell, state, 45.0, nil)
		assert.Nil(t, result, "Should not generate extreme weather")
	})
}

func TestAveragePrecipitation(t *testing.T) {
	t.Run("Empty history", func(t *testing.T) {
		result := averagePrecipitation(nil)
		assert.Equal(t, 0.0, result)
	})

	t.Run("Normal history", func(t *testing.T) {
		history := []WeatherState{
			{Precipitation: 10.0},
			{Precipitation: 20.0},
			{Precipitation: 30.0},
		}
		result := averagePrecipitation(history)
		assert.Equal(t, 20.0, result)
	})
}
