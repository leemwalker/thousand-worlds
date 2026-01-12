package geography

import (
	"testing"
	"tw-backend/internal/spatial"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateVolcanicEmissions(t *testing.T) {
	tests := []struct {
		name      string
		magnitude float64
		magmaType string
		check     func(*testing.T, VolcanicGasComposition)
	}{
		{
			name:      "Basaltic Low Magnitude",
			magnitude: 0.1,
			magmaType: "basaltic",
			check: func(t *testing.T, c VolcanicGasComposition) {
				assert.Greater(t, c.CO2, 0.0)
				assert.Less(t, c.SO2, c.CO2) // Basalt has more CO2 than SO2
			},
		},
		{
			name:      "Andesitic High Magnitude",
			magnitude: 1.0,
			magmaType: "andesitic",
			check: func(t *testing.T, c VolcanicGasComposition) {
				assert.Greater(t, c.SO2, c.CO2) // Andesitic has more SO2
			},
		},
		{
			name:      "Default (Unknown Type)",
			magnitude: 0.5,
			magmaType: "unknown",
			check: func(t *testing.T, c VolcanicGasComposition) {
				assert.Greater(t, c.H2O, 0.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateVolcanicEmissions(tt.magnitude, tt.magmaType)
			tt.check(t, got)
		})
	}
}

func TestSphereHeightmap_DeltaTracking(t *testing.T) {
	topo := spatial.NewCubeSphereTopology(4)
	hm := NewSphereHeightmap(topo)

	// Initially disabled
	assert.False(t, hm.IsDeltaTrackingEnabled())
	assert.Equal(t, 0, hm.DeltaCount())

	// Enable
	hm.EnableDeltaTracking()
	assert.True(t, hm.IsDeltaTrackingEnabled())

	// Perform action that should record delta
	coord := spatial.Coordinate{Face: 0, X: 1, Y: 1}
	hm.Set(coord, 100.0)

	assert.Equal(t, 1, hm.DeltaCount())

	// Flush
	deltas := hm.FlushDeltas()
	assert.Len(t, deltas, 1)
	assert.Equal(t, 0, hm.DeltaCount()) // Reset after flush
	assert.Equal(t, 100.0, deltas[0].ElevationDelta)

	// Disable
	hm.DisableDeltaTracking()
	assert.False(t, hm.IsDeltaTrackingEnabled())

	// Should not record when disabled
	hm.Set(coord, 200.0)
	assert.Equal(t, 0, hm.DeltaCount())
}

func TestGetTargetElevation(t *testing.T) {
	pOcean := TectonicPlate{ID: uuid.New(), Type: PlateOceanic}
	pCont := TectonicPlate{ID: uuid.New(), Type: PlateContinental}

	tests := []struct {
		name     string
		p1, p2   TectonicPlate
		bType    BoundaryType
		expected float64
	}{
		// Divergent
		{"Div Ocean-Ocean", pOcean, pOcean, BoundaryDivergent, -2000},
		{"Div Cont-Cont", pCont, pCont, BoundaryDivergent, -200},
		{"Div Mixed", pOcean, pCont, BoundaryDivergent, 100},

		// Convergent
		{"Conv Ocean-Ocean", pOcean, pOcean, BoundaryConvergent, -9000},
		{"Conv Cont-Cont", pCont, pCont, BoundaryConvergent, 8000},
		{"Conv Mixed", pOcean, pCont, BoundaryConvergent, 6000},

		// Transform
		{"Transform", pOcean, pCont, BoundaryTransform, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTargetElevation(tt.p1, tt.p2, tt.bType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// Simple test for GenerateSubCellSeed coverage (zoom.go)
// It was 0% covered.
func TestGenerateSubCellSeed(t *testing.T) {
	seed := GenerateSubCellSeed(12345, 2, 3)
	assert.NotZero(t, seed)

	seed2 := GenerateSubCellSeed(12345, 2, 3)
	assert.Equal(t, seed, seed2, "Deterministic seed expected")

	seed3 := GenerateSubCellSeed(12345, 2, 4)
	assert.NotEqual(t, seed, seed3, "Different coordinates should yield different seeds")
}

func TestNewBoundaryCache(t *testing.T) {
	bc := NewBoundaryCache()
	assert.NotNil(t, bc)
	assert.Empty(t, bc.Cells)
	assert.False(t, bc.Valid)
}

func TestCalculateElevationChange_Deprecated(t *testing.T) {
	// Wrapper around GetTargetElevation
	pOcean := TectonicPlate{ID: uuid.New(), Type: PlateOceanic}
	change := calculateElevationChange(pOcean, pOcean, BoundaryDivergent)
	assert.Equal(t, -2000.0, change)
}
