package spatial

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateNewPosition(t *testing.T) {
	tests := []struct {
		name      string
		x, y, z   float64
		direction Direction
		distance  float64
		wantX     float64
		wantY     float64
		wantZ     float64
		wantErr   bool
	}{
		{"North", 0, 0, 0, North, 1.0, 0, 1.0, 0, false},
		{"South", 0, 0, 0, South, 1.0, 0, -1.0, 0, false},
		{"East", 0, 0, 0, East, 1.0, 1.0, 0, 0, false},
		{"West", 0, 0, 0, West, 1.0, -1.0, 0, 0, false},
		{"Up", 0, 0, 0, Up, 1.0, 0, 0, 1.0, false},
		{"Down", 0, 0, 0, Down, 1.0, 0, 0, -1.0, false},
		{"NorthEast", 0, 0, 0, NorthEast, 1.0, 0.707, 0.707, 0, false},
		{"Invalid", 0, 0, 0, "INVALID", 1.0, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY, gotZ, err := CalculateNewPosition(tt.x, tt.y, tt.z, tt.direction, tt.distance)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.wantX, gotX, 0.001)
				assert.InDelta(t, tt.wantY, gotY, 0.001)
				assert.InDelta(t, tt.wantZ, gotZ, 0.001)
			}
		})
	}
}

func TestSimpleCollisionDetector_IsPositionValid(t *testing.T) {
	d := NewSimpleCollisionDetector()
	ctx := context.Background()
	worldID := uuid.New()

	valid, err := d.IsPositionValid(ctx, worldID, 0, 0, 0)
	assert.True(t, valid)
	assert.NoError(t, err)

	valid, err = d.IsPositionValid(ctx, worldID, 2000000, 0, 0)
	assert.False(t, valid)
	assert.Error(t, err)
}
