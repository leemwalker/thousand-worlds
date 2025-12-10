package weather

import (
	"testing"
	"time"

	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetSeasonalTemperatureDelta(t *testing.T) {
	t.Run("Tropical Minimal Variation", func(t *testing.T) {
		summerDelta := GetSeasonalTemperatureDelta(SeasonSummer, 10)
		winterDelta := GetSeasonalTemperatureDelta(SeasonWinter, 10)

		assert.True(t, summerDelta >= -5 && summerDelta <= 5)
		assert.True(t, winterDelta >= -5 && winterDelta <= 5)
	})

	t.Run("Mid-Latitude Large Variation", func(t *testing.T) {
		summerDelta := GetSeasonalTemperatureDelta(SeasonSummer, 45)
		winterDelta := GetSeasonalTemperatureDelta(SeasonWinter, 45)

		// Should have roughly 30°C total swing (±15°C)
		totalSwing := summerDelta - winterDelta
		assert.True(t, totalSwing >= 25 && totalSwing <= 35)
	})

	t.Run("Spring/Fall Neutral", func(t *testing.T) {
		springDelta := GetSeasonalTemperatureDelta(SeasonSpring, 40)
		fallDelta := GetSeasonalTemperatureDelta(SeasonFall, 40)

		assert.Equal(t, 0.0, springDelta)
		assert.Equal(t, 0.0, fallDelta)
	})
}

func TestGetDiurnalTemperatureDelta(t *testing.T) {
	t.Run("Afternoon Warmest", func(t *testing.T) {
		afternoon := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
		delta := GetDiurnalTemperatureDelta(afternoon)
		// At 14:00, should be positive (warmest time)
		assert.True(t, delta > 0) // Should be warm
	})

	t.Run("Early Morning Coolest", func(t *testing.T) {
		morning := time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC)
		delta := GetDiurnalTemperatureDelta(morning)
		// At 4:00 AM, sine function at 0, so delta should be 0
		assert.InDelta(t, 0.0, delta, 1.0) // Around zero
	})
}

func TestCalculateTemperature(t *testing.T) {
	cell := &GeographyCell{
		CellID:      uuid.New(),
		Location:    geography.Point{X: 45, Y: 45},
		Elevation:   0,
		Temperature: 20, // Base 20°C
	}

	t.Run("Elevation Lapse Rate", func(t *testing.T) {
		highCell := &GeographyCell{
			CellID:      uuid.New(),
			Location:    geography.Point{X: 45, Y: 45},
			Elevation:   1000, // 1km up
			Temperature: 20,
		}

		currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		temp := CalculateTemperature(highCell, currentTime, SeasonSpring)

		// Should be cooler at 1km elevation: base 20 - 6.5 = 13.5
		// Plus seasonal (spring = 0) and diurnal (~5 at noon) = ~18.5
		assert.True(t, temp < 20, "Higher elevation should be cooler")
	})

	t.Run("Seasonal Variation Applied", func(t *testing.T) {
		currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		summerTemp := CalculateTemperature(cell, currentTime, SeasonSummer)
		winterTemp := CalculateTemperature(cell, currentTime, SeasonWinter)

		assert.True(t, summerTemp > winterTemp)
	})
}

func TestUpdateWeather(t *testing.T) {
	cells := []*GeographyCell{
		{
			CellID:      uuid.New(),
			Location:    geography.Point{X: 0, Y: 0},
			Elevation:   100,
			IsOcean:     false,
			Temperature: 25,
		},
		{
			CellID:      uuid.New(),
			Location:    geography.Point{X: 180, Y: -45},
			Elevation:   0,
			IsOcean:     true,
			Temperature: 15,
		},
	}

	currentTime := time.Now()
	states := UpdateWeather(cells, currentTime, SeasonSummer)

	assert.Len(t, states, 2)
	for _, state := range states {
		assert.NotNil(t, state)
		assert.True(t, state.Temperature > -100 && state.Temperature < 100)
		assert.True(t, state.Humidity >= 0 && state.Humidity <= 100)
		assert.True(t, state.Wind.Speed >= 0)
	}
}
