package weather

import (
	"testing"
	"time"

	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCheckForHurricane(t *testing.T) {
	t.Run("Warm Ocean Tropical", func(t *testing.T) {
		oceanCell := &GeographyCell{IsOcean: true}
		isHurricane := CheckForHurricane(oceanCell, 28, 15)
		assert.True(t, isHurricane)
	})

	t.Run("Cold Ocean No Hurricane", func(t *testing.T) {
		oceanCell := &GeographyCell{IsOcean: true}
		isHurricane := CheckForHurricane(oceanCell, 20, 15)
		assert.False(t, isHurricane)
	})

	t.Run("Land No Hurricane", func(t *testing.T) {
		landCell := &GeographyCell{IsOcean: false}
		isHurricane := CheckForHurricane(landCell, 30, 15)
		assert.False(t, isHurricane)
	})

	t.Run("High Latitude No Hurricane", func(t *testing.T) {
		oceanCell := &GeographyCell{IsOcean: true}
		isHurricane := CheckForHurricane(oceanCell, 28, 40)
		assert.False(t, isHurricane)
	})
}

func TestCheckForBlizzard(t *testing.T) {
	t.Run("Cold With Precipitation And Wind", func(t *testing.T) {
		isBlizzard := CheckForBlizzard(-10, 15, 12)
		assert.True(t, isBlizzard)
	})

	t.Run("Warm No Blizzard", func(t *testing.T) {
		isBlizzard := CheckForBlizzard(5, 15, 12)
		assert.False(t, isBlizzard)
	})

	t.Run("Cold But No Precipitation", func(t *testing.T) {
		isBlizzard := CheckForBlizzard(-10, 2, 12)
		assert.False(t, isBlizzard)
	})

	t.Run("Cold And Precipitation But Low Wind", func(t *testing.T) {
		isBlizzard := CheckForBlizzard(-10, 15, 5)
		assert.False(t, isBlizzard)
	})
}

func TestCheckForDrought(t *testing.T) {
	t.Run("Low Precipitation Extended Period", func(t *testing.T) {
		isDrought := CheckForDrought(50, 200, 95)
		assert.True(t, isDrought)
	})

	t.Run("Normal Precipitation No Drought", func(t *testing.T) {
		isDrought := CheckForDrought(150, 200, 95)
		assert.False(t, isDrought)
	})

	t.Run("Low Precipitation But Short Period", func(t *testing.T) {
		isDrought := CheckForDrought(50, 200, 60)
		assert.False(t, isDrought)
	})
}

func TestCheckForHeatWave(t *testing.T) {
	t.Run("High Temperature Extended Period", func(t *testing.T) {
		isHeatWave := CheckForHeatWave(35, 20, 7)
		assert.True(t, isHeatWave)
	})

	t.Run("Normal Temperature", func(t *testing.T) {
		isHeatWave := CheckForHeatWave(25, 20, 7)
		assert.False(t, isHeatWave)
	})

	t.Run("High Temperature But Short Period", func(t *testing.T) {
		isHeatWave := CheckForHeatWave(35, 20, 5)
		assert.False(t, isHeatWave)
	})
}

func TestGenerateExtremeWeather(t *testing.T) {
	t.Run("Hurricane Generation", func(t *testing.T) {
		cell := &GeographyCell{
			CellID:      uuid.New(),
			Location:    geography.Point{X: 0, Y: 10},
			Elevation:   0,
			IsOcean:     true,
			Temperature: 28,
		}

		state := &WeatherState{
			CellID:        cell.CellID,
			Timestamp:     time.Now(),
			Temperature:   28,
			Precipitation: 5,
			Wind:          Wind{Speed: 12},
		}

		event := GenerateExtremeWeather(cell, state, 10, []WeatherState{})
		assert.NotNil(t, event)
		assert.Equal(t, ExtremeHurricane, event.EventType)
	})

	t.Run("Blizzard Generation", func(t *testing.T) {
		cell := &GeographyCell{
			CellID:      uuid.New(),
			Location:    geography.Point{X: 0, Y: 60},
			Elevation:   500,
			IsOcean:     false,
			Temperature: -10,
		}

		state := &WeatherState{
			CellID:        cell.CellID,
			Timestamp:     time.Now(),
			Temperature:   -10,
			Precipitation: 20,
			Wind:          Wind{Speed: 15},
		}

		event := GenerateExtremeWeather(cell, state, 60, []WeatherState{})
		assert.NotNil(t, event)
		assert.Equal(t, ExtremeBlizzard, event.EventType)
	})
}
