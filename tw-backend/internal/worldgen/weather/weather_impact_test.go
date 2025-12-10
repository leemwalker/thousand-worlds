package weather

import (
	"testing"
	"time"

	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestWeatherDeterminism(t *testing.T) {
	// Setup cells: High/Dry vs Low/Wet
	highDry := &GeographyCell{
		CellID:      uuid.New(),
		Location:    geography.Point{X: 0, Y: 45}, // Mid-latitude
		Elevation:   2000.0,
		Temperature: 20.0, // Base temp
		IsOcean:     false,
		RiverWidth:  0,
	}

	lowWet := &GeographyCell{
		CellID:      uuid.New(),
		Location:    geography.Point{X: 10, Y: 45},
		Elevation:   10.0,
		Temperature: 20.0,
		IsOcean:     true, // Ocean!
		RiverWidth:  0,
	}

	cells := []*GeographyCell{highDry, lowWet}
	now := time.Now()

	// Run update
	states := UpdateWeather(cells, now, SeasonSpring)

	if len(states) != 2 {
		t.Fatalf("Expected 2 states, got %d", len(states))
	}

	highState := states[0]
	lowState := states[1]

	// 1. Check Temperature (Height Map impact)
	// High elevation should be colder
	if highState.Temperature >= lowState.Temperature {
		t.Errorf("Expected high elevation to be colder. High: %.2f, Low: %.2f", highState.Temperature, lowState.Temperature)
	}

	// 2. Check Humidity/Precipitation (Water Level impact)
	// Ocean cell should have higher humidity potential (though precipitation depends on other factors too)
	// In current simple model, IsOcean -> High Water Proximity -> High Evap -> High Moisture
	if lowState.Humidity <= highState.Humidity {
		// Note: UpdateWeather simplification might need checking.
		// evap := CalculateEvaporation...
		// precip, moisture := CalculatePrecipitation(cell, ..., evap*10)
		t.Errorf("Expected ocean cell to have higher humidity. High: %.2f, Low: %.2f", highState.Humidity, lowState.Humidity)
	}
}
