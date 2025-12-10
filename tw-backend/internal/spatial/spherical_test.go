package spatial

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// EarthRadius is the approximate radius of Earth in meters
	EarthRadius = 6371000.0
	// Epsilon is the tolerance for floating point comparisons
	Epsilon = 0.001
)

func TestSphericalProjections(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64 // degrees
		lon      float64 // degrees
		radius   float64
		expected [3]float64 // x, y, z
	}{
		{
			name:     "Equator / Prime Meridian",
			lat:      0,
			lon:      0,
			radius:   EarthRadius,
			expected: [3]float64{EarthRadius, 0, 0},
		},
		{
			name:     "North Pole",
			lat:      90,
			lon:      0,
			radius:   EarthRadius,
			expected: [3]float64{0, 0, EarthRadius},
		},
		{
			name:     "South Pole",
			lat:      -90,
			lon:      0,
			radius:   EarthRadius,
			expected: [3]float64{0, 0, -EarthRadius},
		},
		{
			name:     "Equator / 90E",
			lat:      0,
			lon:      90,
			radius:   EarthRadius,
			expected: [3]float64{0, EarthRadius, 0},
		},
		{
			name:     "Equator / 180E",
			lat:      0,
			lon:      180,
			radius:   EarthRadius,
			expected: [3]float64{-EarthRadius, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Lat/Lon -> Cartesian
			x, y, z := ToCartesian(tt.lat, tt.lon, tt.radius)
			assert.InDelta(t, tt.expected[0], x, Epsilon, "X coordinate mismatch")
			assert.InDelta(t, tt.expected[1], y, Epsilon, "Y coordinate mismatch")
			assert.InDelta(t, tt.expected[2], z, Epsilon, "Z coordinate mismatch")

			// Test Cartesian -> Lat/Lon
			lat, lon := ToLatLon(x, y, z, tt.radius)

			// Handle pole edge cases for longitude (undefined at poles)
			if math.Abs(tt.lat) != 90 {
				// Normalize expected longitude for comparison (-180 to 180)
				expectedLon := tt.lon
				if expectedLon > 180 {
					expectedLon -= 360
				} else if expectedLon <= -180 {
					expectedLon += 360
				}
				assert.InDelta(t, expectedLon, lon, Epsilon, "Longitude mismatch")
			}
			assert.InDelta(t, tt.lat, lat, Epsilon, "Latitude mismatch")
		})
	}
}

func TestGreatCircleDistance(t *testing.T) {
	tests := []struct {
		name     string
		p1       [2]float64 // lat, lon
		p2       [2]float64 // lat, lon
		radius   float64
		expected float64
	}{
		{
			name:     "Same Point",
			p1:       [2]float64{0, 0},
			p2:       [2]float64{0, 0},
			radius:   EarthRadius,
			expected: 0,
		},
		{
			name:     "Equator: 0 to 90E (1/4 circumference)",
			p1:       [2]float64{0, 0},
			p2:       [2]float64{0, 90},
			radius:   EarthRadius,
			expected: (2 * math.Pi * EarthRadius) / 4,
		},
		{
			name:     "North Pole to South Pole (1/2 circumference)",
			p1:       [2]float64{90, 0},
			p2:       [2]float64{-90, 0},
			radius:   EarthRadius,
			expected: (2 * math.Pi * EarthRadius) / 2,
		},
		{
			name:     "Antipodal Points (Equator)",
			p1:       [2]float64{0, 0},
			p2:       [2]float64{0, 180},
			radius:   EarthRadius,
			expected: (2 * math.Pi * EarthRadius) / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := GreatCircleDistance(tt.p1[0], tt.p1[1], tt.p2[0], tt.p2[1], tt.radius)
			assert.InDelta(t, tt.expected, dist, 1.0, "Distance mismatch (1m tolerance)")
		})
	}
}

func TestCoordinateNormalization(t *testing.T) {
	tests := []struct {
		name        string
		inputLat    float64
		inputLon    float64
		expectedLat float64
		expectedLon float64
	}{
		{
			name:        "Normal Coordinates",
			inputLat:    45,
			inputLon:    90,
			expectedLat: 45,
			expectedLon: 90,
		},
		{
			name:        "Longitude Wrap Positive",
			inputLat:    0,
			inputLon:    190,
			expectedLat: 0,
			expectedLon: -170,
		},
		{
			name:        "Longitude Wrap Negative",
			inputLat:    0,
			inputLon:    -190,
			expectedLat: 0,
			expectedLon: 170,
		},
		{
			name:        "North Pole Crossing",
			inputLat:    95,
			inputLon:    0,
			expectedLat: 85,
			expectedLon: 180, // Flips to opposite side
		},
		{
			name:        "South Pole Crossing",
			inputLat:    -95,
			inputLon:    0,
			expectedLat: -85,
			expectedLon: 180, // Flips to opposite side
		},
		{
			name:        "Complex Wrap (Pole + Date Line)",
			inputLat:    95,
			inputLon:    190, // Effectively -170
			expectedLat: 85,
			expectedLon: 10, // Opposite of -170 is 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon := NormalizeCoordinates(tt.inputLat, tt.inputLon)
			assert.InDelta(t, tt.expectedLat, lat, Epsilon, "Latitude mismatch")
			assert.InDelta(t, tt.expectedLon, lon, Epsilon, "Longitude mismatch")
		})
	}
}
