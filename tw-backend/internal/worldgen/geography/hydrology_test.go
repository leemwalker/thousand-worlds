package geography

import (
	"testing"
	"tw-backend/internal/spatial"
)

func TestCalculateFlowAccumulation_Slope(t *testing.T) {
	// Create a 10x10 simple slope (Face 0)
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	// Slope down from Y=0 to Y=9
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			hm.Set(spatial.Coordinate{Face: 0, X: x, Y: y}, float64(10-y)*10.0)
		}
	}

	// Calculate flux
	CalculateGlobalFlux(hm)

	// Check flux at bottom (Y=9) should be higher than top (Y=0)
	topFlux := hm.GetCellData(spatial.Coordinate{Face: 0, X: 5, Y: 0}).Flux
	bottomFlux := hm.GetCellData(spatial.Coordinate{Face: 0, X: 5, Y: 9}).Flux

	if bottomFlux <= topFlux {
		t.Errorf("Flux did not accumulate downhill. Top: %f, Bottom: %f", topFlux, bottomFlux)
	}

	// Flux at Y=9 should be roughly Rain * 10 (accumulating from 10 cells above)
	// Assuming base rain is 1.0
	// 9.0 is acceptable if one cell leaked or edge case
	expectedFlux := 9.0
	if bottomFlux < expectedFlux {
		t.Errorf("Bottom flux too low. Got %f, want >= %f", bottomFlux, expectedFlux)
	}
}

func TestCalculateFlowAccumulation_Valley(t *testing.T) {
	// V-shape valley in X direction, sloping down in Y
	// Water should concentrate in the center (X=5) and flow to bottom Y=9
	topo := spatial.NewCubeSphereTopology(10)
	hm := NewSphereHeightmap(topo)

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			// Y slope + X V-shape
			elev := float64(10-y)*10.0 + float64(abs(x-5))*5.0
			hm.Set(spatial.Coordinate{Face: 0, X: x, Y: y}, elev)
		}
	}

	CalculateGlobalFlux(hm)

	centerFlux := hm.GetCellData(spatial.Coordinate{Face: 0, X: 5, Y: 9}).Flux
	sideFlux := hm.GetCellData(spatial.Coordinate{Face: 0, X: 0, Y: 9}).Flux

	if centerFlux <= sideFlux {
		t.Errorf("Flux did not concentrate in valley. Center: %f, Side: %f", centerFlux, sideFlux)
	}
}
