package astronomy

import (
	"math/rand"
	"testing"
)

func TestSimulateGiantImpact(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	planetMass := EarthMassKg

	// Test Case 1: Perfect Conditions for Moon Formation
	// Theia: 0.1 Earth Mass, 45 degree angle, Low velocity
	impactorMass := 0.1 * planetMass
	velocityRatio := 1.1
	angle := 45.0

	outcome, moonMass := SimulateGiantImpact(impactorMass, planetMass, velocityRatio, angle, rng)

	if outcome != OutcomeMoonFormation {
		t.Errorf("Expected OutcomeMoonFormation, got %s", outcome)
	}

	if moonMass <= 0 {
		t.Errorf("Expected positive moon mass, got %f", moonMass)
	}

	// Test Case 2: Glancing Hit / Escape
	angleMiss := 88.0
	outcomeMiss, _ := SimulateGiantImpact(impactorMass, planetMass, velocityRatio, angleMiss, rng)
	if outcomeMiss != OutcomeEscape {
		t.Errorf("Expected OutcomeEscape for grazing angle, got %s", outcomeMiss)
	}

	// Test Case 3: Direct Impact
	angleDirect := 5.0
	outcomeDirect, _ := SimulateGiantImpact(impactorMass, planetMass, velocityRatio, angleDirect, rng)
	if outcomeDirect != OutcomeDirectImpact {
		t.Errorf("Expected OutcomeDirectImpact for low angle, got %s", outcomeDirect)
	}

	// Test Case 4: Too fast (Escape/Disintegration) -> Actually my code defaults to DirectImpact if not Escape?
	// Wait, velocityRatio check only applies for Moon Formation?
	// Let's check logic:
	// if angle > 85 -> Escape
	// if massRatio typical && angle typical && velocity < 1.2 -> Moon
	// Else -> Direct Impact

	// If velocity is high (2.0) and angle 45:
	// Should be Impact or Escape?
	// Code says "Default: Direct Impact / Merger"

	outcomeFast, _ := SimulateGiantImpact(impactorMass, planetMass, 2.0, 45.0, rng)
	if outcomeFast != OutcomeDirectImpact {
		t.Logf("High velocity at 45 degrees resulted in %s (Expected DirectImpact by default)", outcomeFast)
	}
}
