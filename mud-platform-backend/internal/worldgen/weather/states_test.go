package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineWeatherState(t *testing.T) {
	t.Run("Clear Weather", func(t *testing.T) {
		state := DetermineWeatherState(25, 0, 20, 5)
		assert.Equal(t, WeatherClear, state)
	})

	t.Run("Cloudy Weather", func(t *testing.T) {
		state := DetermineWeatherState(20, 1, 45, 5)
		assert.Equal(t, WeatherCloudy, state)
	})

	t.Run("Rain", func(t *testing.T) {
		state := DetermineWeatherState(15, 5, 70, 8)
		assert.Equal(t, WeatherRain, state)
	})

	t.Run("Snow", func(t *testing.T) {
		state := DetermineWeatherState(-5, 5, 70, 8)
		assert.Equal(t, WeatherSnow, state)
	})

	t.Run("Storm Heavy Precipitation", func(t *testing.T) {
		state := DetermineWeatherState(20, 25, 90, 10)
		assert.Equal(t, WeatherStorm, state)
	})

	t.Run("Storm High Winds", func(t *testing.T) {
		state := DetermineWeatherState(25, 5, 60, 20)
		assert.Equal(t, WeatherStorm, state)
	})
}

func TestCalculateVisibility(t *testing.T) {
	t.Run("Clear Visibility", func(t *testing.T) {
		vis := CalculateVisibility(WeatherClear)
		assert.Equal(t, 50.0, vis)
	})

	t.Run("Cloudy Visibility", func(t *testing.T) {
		vis := CalculateVisibility(WeatherCloudy)
		assert.Equal(t, 30.0, vis)
	})

	t.Run("Rain Visibility", func(t *testing.T) {
		vis := CalculateVisibility(WeatherRain)
		assert.Equal(t, 10.0, vis)
	})

	t.Run("Snow Visibility", func(t *testing.T) {
		vis := CalculateVisibility(WeatherSnow)
		assert.Equal(t, 5.0, vis)
	})

	t.Run("Storm Visibility", func(t *testing.T) {
		vis := CalculateVisibility(WeatherStorm)
		assert.Equal(t, 2.0, vis)
	})
}
