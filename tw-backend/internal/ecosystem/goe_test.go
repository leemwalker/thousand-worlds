package ecosystem

import (
	"testing"
)

func TestGOE_OxygenProduction(t *testing.T) {
	cc := NewCarbonCycle()

	// Initial state: Anoxic (0% O2)
	if cc.State.OxygenLevel != 0 {
		t.Fatalf("Initial OxygenLevel = %.4f, want 0", cc.State.OxygenLevel)
	}

	// Simulate conditions for O2 production:
	// - Phosphorous flux available (from weathering)
	// - Temperature > 0 (not frozen)
	cc.State.Temperature = 15.0      // Warm
	cc.Flux.Phosphorous = 0.01       // Weathering active
	cc.Reservoir.Atmosphere = 1000.0 // CO2 present

	// Run for 100 Million Years
	for i := 0; i < 100; i++ {
		cc.Update(1.0, 1.0, 0.3, 1.0) // 1 My steps
	}

	if cc.State.OxygenLevel <= 0 {
		t.Error("Oxygen should have increased with phosphorous flux and warm temps")
	}
	t.Logf("After 100 My: OxygenLevel = %.4f (%.2f%%)", cc.State.OxygenLevel, cc.State.OxygenLevel*100)
}

func TestGOE_IronSinkAbsorbsOxygen(t *testing.T) {
	cc := NewCarbonCycle()

	// Set oxygen to just below threshold where iron sink is active
	cc.State.OxygenLevel = 0.01  // 1%
	cc.State.Temperature = -10.0 // FROZEN - prevents O2 production
	cc.Flux.Phosphorous = 0.0    // No production

	initialO2 := cc.State.OxygenLevel

	// Run for 10 Million Years
	cc.Update(10.0, 1.0, 0.3, 1.0)

	if cc.State.OxygenLevel >= initialO2 {
		t.Error("Iron sink should have reduced oxygen level")
	}
	t.Logf("Iron sink: O2 %.4f -> %.4f", initialO2, cc.State.OxygenLevel)
}

func TestGOE_MethaneCollapse(t *testing.T) {
	cc := NewCarbonCycle()

	// Set conditions for methane collapse:
	// - O2 > 1%
	// - Methane present
	cc.State.OxygenLevel = 0.05 // 5%
	cc.State.Methaneppm = 100.0 // High methane
	cc.State.Temperature = 15.0
	cc.Flux.Phosphorous = 0.0

	initialCH4 := cc.State.Methaneppm
	initialWarming := cc.GetGreenhouseWarming()

	// Run for 100 Million Years (2 half-lives of CH4 decay)
	for i := 0; i < 100; i++ {
		cc.Update(1.0, 1.0, 0.3, 1.0)
	}

	if cc.State.Methaneppm >= initialCH4*0.5 {
		t.Error("Methane should have decayed significantly with O2 > 1%")
	}

	finalWarming := cc.GetGreenhouseWarming()
	if finalWarming >= initialWarming {
		t.Error("Greenhouse warming should decrease as methane collapses")
	}

	t.Logf("Methane collapse: CH4 %.2f -> %.2f ppm, Warming %.2f -> %.2f°C",
		initialCH4, cc.State.Methaneppm, initialWarming, finalWarming)
}

func TestGOE_FullTimeline(t *testing.T) {
	// Simulate ~2.5 billion years of Earth history
	// Starting from Hadean (anoxic) through GOE to modern O2 levels

	cc := NewCarbonCycle()

	// Hadean init: 0% O2, high CH4, high CO2
	cc.State.OxygenLevel = 0.0
	cc.State.Methaneppm = 1000.0      // High Archean methane
	cc.Reservoir.Atmosphere = 10000.0 // High CO2
	cc.State.Temperature = 30.0       // Warm Archean

	// Simulate 2500 Million Years in 10My steps
	for i := 0; i < 250; i++ {
		// Weathering provides phosphorous
		cc.Flux.Phosphorous = 0.01
		cc.Update(10.0, 1.0, 0.3, 1.0)
	}

	t.Logf("After 2.5 Billion Years:")
	t.Logf("  O2: %.2f%% (Earth: ~21%%)", cc.State.OxygenLevel*100)
	t.Logf("  CH4: %.2f ppm (Earth: ~2 ppm)", cc.State.Methaneppm)
	t.Logf("  CO2: %.0f ppm", cc.State.CO2ppm)

	// Verify we approach modern levels
	if cc.State.OxygenLevel < 0.10 {
		t.Error("Oxygen should have risen significantly after 2.5B years")
	}
	if cc.State.Methaneppm > 10 {
		t.Error("Methane should have collapsed after GOE")
	}
}
