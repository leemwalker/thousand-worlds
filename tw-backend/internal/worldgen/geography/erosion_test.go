package geography

import (
	"testing"
)

func TestApplyThermalErosion_Spike(t *testing.T) {
	// Create a 10x10 heightmap with a single spike
	hm := NewHeightmap(10, 10)
	centerX, centerY := 5, 5
	initialHeight := 1000.0
	hm.Set(centerX, centerY, initialHeight)

	// Verify initial state
	if hm.Get(centerX, centerY) != initialHeight {
		t.Fatalf("Setup failed: expected height %f, got %f", initialHeight, hm.Get(centerX, centerY))
	}

	// Apply thermal erosion
	// Uses random seed 123
	ApplyThermalErosion(hm, 50, 123)

	// Check results
	finalHeight := hm.Get(centerX, centerY)
	if finalHeight >= initialHeight {
		t.Errorf("Erosion failed: spike height did not decrease. Got %f", finalHeight)
	}

	// Check that material was moved to neighbors
	// Neighbor at 4,5 should have received something
	neighborHeight := hm.Get(centerX-1, centerY)
	if neighborHeight <= 0 {
		t.Errorf("Erosion failed: neighbor did not gain material. Got %f", neighborHeight)
	}
}

func TestApplyHydraulicErosion_Slope(t *testing.T) {
	// Create a 20x20 slope
	width, height := 20, 20
	hm := NewHeightmap(width, height)

	// Uniform slope from Y=0 (high) to Y=19 (low)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			hm.Set(x, y, float64(height-y)*10.0)
		}
	}

	// Apply hydraulic erosion
	// High number of drops to ensure visible effect
	ApplyHydraulicErosion(hm, 5000, 123)

	// Check for channel formation
	// We expect some variance in the X direction for a given Y row now
	// Ideally, we'd look for valley formation.

	// Simple check: Ensure mass conservation (approximate) or just change
	// Since hydraulic erosion takes material away (transport) or adds it (deposition),
	// the total mass might change slightly due to evaporation/sediment loss at edges,
	// but mostly we want to see change.

	changed := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			original := float64(height-y) * 10.0
			current := hm.Get(x, y)
			if current != original {
				changed = true
				break
			}
		}
	}

	if !changed {
		t.Error("Hydraulic erosion resulted in no changes to the heightmap")
	}
}
