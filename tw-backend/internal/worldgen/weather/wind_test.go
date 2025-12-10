package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateWind(t *testing.T) {
	t.Run("Equatorial Trade Winds", func(t *testing.T) {
		wind := CalculateWind(10, 0, SeasonSummer)
		// Should be easterly (negative direction before Coriolis)
		// With Coriolis: NH deflects right, so less negative
		assert.True(t, wind.Speed >= 5 && wind.Speed <= 10)
	})

	t.Run("Mid-Latitude Westerlies", func(t *testing.T) {
		wind := CalculateWind(45, 0, SeasonSummer)
		// Should be westerly with higher speed
		assert.True(t, wind.Speed >= 8 && wind.Speed <= 13)
	})

	t.Run("Polar Easterlies", func(t *testing.T) {
		wind := CalculateWind(75, 0, SeasonSummer)
		// Should be easterly with lower speed
		assert.True(t, wind.Speed >= 3 && wind.Speed <= 6)
	})

	t.Run("Northern vs Southern Hemisphere Coriolis", func(t *testing.T) {
		northWind := CalculateWind(30, 0, SeasonSummer)
		southWind := CalculateWind(-30, 0, SeasonSummer)

		// Both should have similar speeds (same latitude magnitude)
		assert.InDelta(t, northWind.Speed, southWind.Speed, 0.1)

		// Directions should differ due to Coriolis (opposite deflection)
		// This is a general check - exact values depend on implementation
		assert.NotEqual(t, northWind.Direction, southWind.Direction)
	})
}

func TestGetAtmosphericCell(t *testing.T) {
	t.Run("Hadley Cell", func(t *testing.T) {
		cell := GetAtmosphericCell(15)
		assert.Equal(t, CellHadley, cell)

		cell = GetAtmosphericCell(-25)
		assert.Equal(t, CellHadley, cell)
	})

	t.Run("Ferrel Cell", func(t *testing.T) {
		cell := GetAtmosphericCell(45)
		assert.Equal(t, CellFerrel, cell)

		cell = GetAtmosphericCell(-50)
		assert.Equal(t, CellFerrel, cell)
	})

	t.Run("Polar Cell", func(t *testing.T) {
		cell := GetAtmosphericCell(75)
		assert.Equal(t, CellPolar, cell)

		cell = GetAtmosphericCell(-85)
		assert.Equal(t, CellPolar, cell)
	})
}

func TestGetPressureAtLatitude(t *testing.T) {
	t.Run("Equatorial Low", func(t *testing.T) {
		pressure := GetPressureAtLatitude(5)
		assert.Equal(t, PressureLow, pressure)
	})

	t.Run("Subtropical High", func(t *testing.T) {
		pressure := GetPressureAtLatitude(30)
		assert.Equal(t, PressureHigh, pressure)
	})

	t.Run("Subpolar Low", func(t *testing.T) {
		pressure := GetPressureAtLatitude(60)
		assert.Equal(t, PressureLow, pressure)
	})

	t.Run("Polar High", func(t *testing.T) {
		pressure := GetPressureAtLatitude(85)
		assert.Equal(t, PressureHigh, pressure)
	})
}
