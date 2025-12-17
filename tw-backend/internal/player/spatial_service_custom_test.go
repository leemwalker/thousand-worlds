package player

import (
	"testing"
	worldspatial "tw-backend/internal/world/spatial"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSphericalPosition_MeterBased(t *testing.T) {
	// Setup world dimensions: 40,000 km circumference (Earth-like)
	circumference := 40000000.0
	dims := worldspatial.NewWorldDimensions(circumference)

	// Case 1: Standard movement (no wrapping)
	// Start at 500, 500 (meters from origin)
	// Move North 100 meters
	nx, ny, msg := calculateSphericalPosition(500, 500, 0, 100, "north", dims)
	assert.InDelta(t, 500.0, nx, 0.1, "X should stay same")
	assert.InDelta(t, 600.0, ny, 0.1, "Y should increase by 100")
	assert.Empty(t, msg, "No message should be returned")

	// Case 2: Crossing North Pole
	// Quarter Circumference = 10,000,000m
	// Start at 9,999,990m (10m from pole)
	// Move North 20m -> should overshoot by 10m
	// Result Y: 10,000,000 - 10 = 9,999,990m (same latitude, opposite side)
	// Result X: StartX + HalfCircumference
	quarterCirc := 10000000.0
	startX := 500.0
	startY := quarterCirc - 10.0
	nx, ny, msg = calculateSphericalPosition(startX, startY, 0, 20.0, "north", dims)

	expectedY := quarterCirc - 10.0
	expectedX := startX + (circumference / 2.0)

	assert.InDelta(t, expectedY, ny, 0.1, "Y should reflect back from pole")
	assert.InDelta(t, expectedX, nx, 0.1, "X should flip to opposite side")
	assert.Contains(t, msg, "cross the North Pole")

	// Case 3: Circumnavigation (East)
	// Start at Circumference - 10m
	// Move East 20m -> should wrap to 10m
	startX = circumference - 10.0
	startY = 500.0
	nx, ny, msg = calculateSphericalPosition(startX, startY, 20.0, 0, "east", dims)

	expectedX = 10.0
	assert.InDelta(t, expectedX, nx, 0.1, "X should wrap around")
	assert.InDelta(t, startY, ny, 0.1, "Y should stay same")
	assert.Contains(t, msg, "circled back around")
}
