package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeatherConsistencyWithGeography(t *testing.T) {
	t.Run("Tropical Regions High Precipitation", func(t *testing.T) {
		// Tropical: 0-15° latitude
		precip := CalculateAnnualPrecipitation(5, 0, 50000, true)
		assert.True(t, precip > 2000, "Tropical regions should have >2000mm/year")
	})

	t.Run("Subtropical Desert Low Precipitation", func(t *testing.T) {
		// Desert: 20-30° latitude, inland, leeward
		// Calculation: 500 base * 1.0 coastal * 1.0 elev = 500
		precip := CalculateAnnualPrecipitation(25, 500, 800000, false)
		assert.True(t, precip < 600, "Desert regions should have low precipitation")
	})

	t.Run("Polar Low Precipitation", func(t *testing.T) {
		// Polar: >70° latitude
		precip := CalculateAnnualPrecipitation(75, 0, 100000, true)
		assert.True(t, precip < 400, "Polar regions should have <400mm/year")
	})

	t.Run("Coastal Higher Than Inland", func(t *testing.T) {
		coastal := CalculateAnnualPrecipitation(40, 100, 20000, true)
		inland := CalculateAnnualPrecipitation(40, 100, 600000, true)
		assert.True(t, coastal > inland, "Coastal areas should be wetter than inland")
	})

	t.Run("Windward vs Leeward Mountains", func(t *testing.T) {
		windward := CalculateAnnualPrecipitation(35, 2500, 100000, true)
		leeward := CalculateAnnualPrecipitation(35, 2500, 100000, false)
		assert.True(t, windward > leeward*2, "Windward should be much wetter than leeward")
	})
}

func TestSeasonalVariation(t *testing.T) {
	t.Run("Mid-Latitude Temperature Swing", func(t *testing.T) {
		summerDelta := GetSeasonalTemperatureDelta(SeasonSummer, 45)
		winterDelta := GetSeasonalTemperatureDelta(SeasonWinter, 45)

		swing := summerDelta - winterDelta
		assert.True(t, swing >= 20 && swing <= 35, "Mid-latitude should have 20-30°C seasonal swing")
	})

	t.Run("Tropical Minimal Swing", func(t *testing.T) {
		summerDelta := GetSeasonalTemperatureDelta(SeasonSummer, 5)
		winterDelta := GetSeasonalTemperatureDelta(SeasonWinter, 5)

		swing := summerDelta - winterDelta
		assert.True(t, swing < 10, "Tropics should have <10°C seasonal swing")
	})
}

func TestEarthPatternComparison(t *testing.T) {
	t.Run("Amazon Rainforest Pattern", func(t *testing.T) {
		// Equatorial rainforest: ~2500-3500mm/year
		precip := CalculateAnnualPrecipitation(0, 100, 50000, true)
		assert.True(t, precip >= 2000 && precip <= 4500, "Equatorial rainforest pattern")
	})

	t.Run("Sahara Desert Pattern", func(t *testing.T) {
		// Sahara: ~25° latitude, inland - similar to subtropical desert
		precip := CalculateAnnualPrecipitation(25, 300, 1000000, false)
		assert.True(t, precip < 600, "Sahara-like desert pattern")
	})

	t.Run("Seattle Coastal Pattern", func(t *testing.T) {
		// Seattle: ~47° latitude, coastal
		precip := CalculateAnnualPrecipitation(47, 100, 1000, true)
		assert.True(t, precip >= 800 && precip <= 1500, "Seattle-like coastal pattern")
	})

	t.Run("Antarctic Interior Pattern", func(t *testing.T) {
		// Antarctic interior: very low precip
		precip := CalculateAnnualPrecipitation(85, 3000, 500000, false)
		assert.True(t, precip < 300, "Antarctic-like interior pattern")
	})
}
