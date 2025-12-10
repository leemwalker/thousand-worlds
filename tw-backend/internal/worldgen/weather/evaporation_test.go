package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTemperatureFactor(t *testing.T) {
	t.Run("Below 5C No Evaporation", func(t *testing.T) {
		factor := CalculateTemperatureFactor(0)
		assert.Equal(t, 0.0, factor)

		factor = CalculateTemperatureFactor(-10)
		assert.Equal(t, 0.0, factor)
	})

	t.Run("Linear Increase 5-35C", func(t *testing.T) {
		factor5 := CalculateTemperatureFactor(5)
		assert.Equal(t, 0.0, factor5)

		factor20 := CalculateTemperatureFactor(20)
		assert.InDelta(t, 0.5, factor20, 0.01)

		factor35 := CalculateTemperatureFactor(35)
		assert.InDelta(t, 1.0, factor35, 0.01)
	})

	t.Run("Capped Above 35C", func(t *testing.T) {
		factor := CalculateTemperatureFactor(40)
		assert.Equal(t, 1.0, factor)

		factor = CalculateTemperatureFactor(50)
		assert.Equal(t, 1.0, factor)
	})
}

func TestCalculateWaterProximity(t *testing.T) {
	t.Run("Ocean Cell", func(t *testing.T) {
		cell := &GeographyCell{IsOcean: true}
		proximity := CalculateWaterProximity(cell, 0)
		assert.Equal(t, 1.0, proximity)
	})

	t.Run("River Cell", func(t *testing.T) {
		cell := &GeographyCell{RiverWidth: 50}
		proximity := CalculateWaterProximity(cell, 0)
		assert.Equal(t, 0.5, proximity)

		wideRiver := &GeographyCell{RiverWidth: 150}
		proximity = CalculateWaterProximity(wideRiver, 0)
		assert.Equal(t, 1.0, proximity) // Capped at 1.0
	})

	t.Run("Adjacent To Water", func(t *testing.T) {
		cell := &GeographyCell{IsOcean: false}
		proximity := CalculateWaterProximity(cell, 500) // 500m from water
		assert.Equal(t, 0.1, proximity)
	})

	t.Run("Distant From Water", func(t *testing.T) {
		cell := &GeographyCell{IsOcean: false}
		proximity := CalculateWaterProximity(cell, 15000) // 15km from water
		assert.Equal(t, 0.01, proximity)
	})
}

func TestCalculateSunlight(t *testing.T) {
	t.Run("Equator Maximum", func(t *testing.T) {
		sunlight := CalculateSunlight(0, SeasonSummer)
		assert.True(t, sunlight > 1.0) // cos(0) = 1.0 * 1.3 = 1.3
	})

	t.Run("Polar Minimum", func(t *testing.T) {
		sunlight := CalculateSunlight(85, SeasonWinter)
		assert.True(t, sunlight < 0.1) // Close to 0
	})

	t.Run("Seasonal Variation", func(t *testing.T) {
		summer := CalculateSunlight(45, SeasonSummer)
		winter := CalculateSunlight(45, SeasonWinter)
		assert.True(t, summer > winter)
	})
}

func TestCalculateEvaporation(t *testing.T) {
	t.Run("Optimal Conditions", func(t *testing.T) {
		// Warm ocean at equator in summer
		evap := CalculateEvaporation(30, 1.0, 0, SeasonSummer)
		assert.True(t, evap > 5.0) // Should be near maximum
	})

	t.Run("Cold Conditions No Evaporation", func(t *testing.T) {
		evap := CalculateEvaporation(0, 1.0, 45, SeasonWinter)
		assert.Equal(t, 0.0, evap)
	})

	t.Run("Land vs Ocean", func(t *testing.T) {
		ocean := CalculateEvaporation(25, 1.0, 30, SeasonSummer)
		land := CalculateEvaporation(25, 0.01, 30, SeasonSummer)
		assert.True(t, ocean > land*50) // Ocean much higher
	})
}
