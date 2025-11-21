package world

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSunPosition(t *testing.T) {
	dayLength := 24 * time.Hour

	tests := []struct {
		name     string
		gameTime time.Duration
		expected float64
	}{
		{"Midnight (Start)", 0, 0.0},
		{"6 AM", 6 * time.Hour, 0.25},
		{"Noon", 12 * time.Hour, 0.5},
		{"6 PM", 18 * time.Hour, 0.75},
		{"Midnight (End)", 24 * time.Hour, 0.0}, // Modulo should wrap to 0
		{"Next Day Noon", 36 * time.Hour, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := CalculateSunPosition(tt.gameTime, dayLength)
			assert.InDelta(t, tt.expected, pos, 0.001)
		})
	}
}

func TestTimeOfDayDescriptors(t *testing.T) {
	tests := []struct {
		name        string
		sunPosition float64
		expected    TimeOfDay
	}{
		{"Night Start", 0.0, TimeOfDayNight},
		{"Night End", 0.24, TimeOfDayNight},
		{"Dawn Start", 0.25, TimeOfDayDawn},
		{"Dawn End", 0.29, TimeOfDayDawn},
		{"Morning Start", 0.30, TimeOfDayMorning},
		{"Morning End", 0.44, TimeOfDayMorning},
		{"Noon Start", 0.45, TimeOfDayNoon},
		{"Noon End", 0.54, TimeOfDayNoon},
		{"Afternoon Start", 0.55, TimeOfDayAfternoon},
		{"Afternoon End", 0.69, TimeOfDayAfternoon},
		{"Dusk Start", 0.70, TimeOfDayDusk},
		{"Dusk End", 0.74, TimeOfDayDusk},
		{"Evening Start", 0.75, TimeOfDayEvening},
		{"Evening End", 0.89, TimeOfDayEvening},
		{"Night Late Start", 0.90, TimeOfDayNight},
		{"Night Late End", 0.99, TimeOfDayNight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := GetTimeOfDay(tt.sunPosition)
			assert.Equal(t, tt.expected, desc)
		})
	}
}

func TestSeasonCalculation(t *testing.T) {
	seasonLength := 90 * 24 * time.Hour // 90 days
	yearLength := seasonLength * 4

	tests := []struct {
		name             string
		gameTime         time.Duration
		expectedSeason   Season
		expectedProgress float64
	}{
		{"Spring Start", 0, SeasonSpring, 0.0},
		{"Spring Middle", seasonLength / 2, SeasonSpring, 0.5},
		{"Spring End", seasonLength - time.Hour, SeasonSpring, 0.999},

		{"Summer Start", seasonLength, SeasonSummer, 0.0},
		{"Summer Middle", seasonLength + (seasonLength / 2), SeasonSummer, 0.5},

		{"Autumn Start", seasonLength * 2, SeasonAutumn, 0.0},
		{"Winter Start", seasonLength * 3, SeasonWinter, 0.0},

		{"Next Year Spring", yearLength, SeasonSpring, 0.0},
		{"Next Year Summer", yearLength + seasonLength, SeasonSummer, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			season, progress := CalculateSeason(tt.gameTime, seasonLength)
			assert.Equal(t, tt.expectedSeason, season)
			assert.InDelta(t, tt.expectedProgress, progress, 0.001)
		})
	}
}
