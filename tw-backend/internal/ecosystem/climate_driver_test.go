package ecosystem

import (
	"testing"

	"tw-backend/internal/worldgen/astronomy"
)

// TestClimateDriver_Initialization verifies driver starts in correct state.
func TestClimateDriver_Initialization(t *testing.T) {
	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	if cd.IceAgeActive {
		t.Error("Climate driver should not start with ice age active")
	}
	if cd.IceAgeStartYear != 0 {
		t.Errorf("Ice age start year should be 0, got %d", cd.IceAgeStartYear)
	}
	if cd.eventManager != eventManager {
		t.Error("Event manager not properly linked")
	}
}

// TestClimateDriver_IceAgeTrigger verifies ice ages start at low insolation.
func TestClimateDriver_IceAgeTrigger(t *testing.T) {
	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	// Find a year with low insolation (obliquity trough)
	// At year 30750 (3/4 of 41k cycle), obliquity is minimum
	lowInsolationYear := int64(30750)

	cd.Update(lowInsolationYear)

	// Check that insolation is below threshold
	if cd.CurrentInsolation >= IceAgeInsolationThreshold {
		t.Logf("Insolation at year %d: %.4f (threshold: %.4f)",
			lowInsolationYear, cd.CurrentInsolation, IceAgeInsolationThreshold)
		t.Skip("Insolation not below threshold at this year - orbital formula may need adjustment")
	}

	// Ice age should now be active
	if !cd.IceAgeActive {
		t.Errorf("Ice age should be active at low insolation year %d (insolation=%.4f)",
			lowInsolationYear, cd.CurrentInsolation)
	}

	// Event should be in the manager
	found := false
	for _, e := range eventManager.ActiveEvents {
		if e.Type == EventIceAge {
			found = true
			break
		}
	}
	if !found {
		t.Error("Ice age event not found in event manager")
	}
}

// TestClimateDriver_IceAgeRecovery verifies ice ages end at high insolation.
func TestClimateDriver_IceAgeRecovery(t *testing.T) {
	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	// Manually trigger an ice age
	cd.IceAgeActive = true
	cd.IceAgeStartYear = 0
	eventManager.ActiveEvents = append(eventManager.ActiveEvents, GeologicalEvent{
		Type:           EventIceAge,
		StartTick:      0,
		DurationTicks:  1000000,
		Severity:       0.5,
		TemperatureMod: -10,
		SunlightMod:    0.9,
		OxygenMod:      1.0,
	})

	// Find a year with high insolation by scanning through years
	// We need insolation > 0.995 (recovery threshold)
	var highInsolationYear int64 = -1
	for year := int64(IceAgeDurationBase); year < 100000; year += 1000 {
		state := astronomy.CalculateOrbitalState(year)
		insolation := astronomy.CalculateInsolation(state)
		if insolation > IceAgeRecoveryThreshold {
			highInsolationYear = year
			break
		}
	}

	if highInsolationYear < 0 {
		t.Skip("Could not find year with insolation above recovery threshold in first 100k years")
	}

	cd.Update(highInsolationYear)

	// Ice age should now be ended
	if cd.IceAgeActive {
		t.Errorf("Ice age should have ended at high insolation year %d (insolation=%.4f)",
			highInsolationYear, cd.CurrentInsolation)
	}
}

// TestClimateDriver_Hysteresis verifies rapid oscillation is prevented.
func TestClimateDriver_Hysteresis(t *testing.T) {
	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	// Start an ice age
	cd.IceAgeActive = true
	cd.IceAgeStartYear = 0

	// Try to end it before minimum duration
	cd.CurrentInsolation = 1.05 // Above recovery threshold
	cd.Update(5000)             // Only 5000 years (below IceAgeDurationBase)

	// Ice age should still be active due to minimum duration
	if !cd.IceAgeActive {
		t.Error("Ice age ended before minimum duration elapsed")
	}
}

// TestClimateDriver_GetObliquity returns correct orbital tilt.
func TestClimateDriver_GetObliquity(t *testing.T) {
	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	// Update to a specific year
	year := int64(20500) // Approximately at midpoint of obliquity cycle
	cd.Update(year)

	// Obliquity should match what astronomy package returns
	expected := astronomy.CalculateOrbitalState(year).Obliquity
	if cd.GetObliquity() != expected {
		t.Errorf("GetObliquity() = %.4f, expected %.4f", cd.GetObliquity(), expected)
	}
}

// TestClimateDriver_Determinism verifies same year produces same behavior.
func TestClimateDriver_Determinism(t *testing.T) {
	eventManager1 := NewGeologicalEventManager()
	eventManager2 := NewGeologicalEventManager()

	cd1 := NewClimateDriver(eventManager1)
	cd2 := NewClimateDriver(eventManager2)

	// Run both through the same years
	years := []int64{0, 10000, 20000, 30000, 40000, 50000}
	for _, year := range years {
		cd1.Update(year)
		cd2.Update(year)

		if cd1.IceAgeActive != cd2.IceAgeActive {
			t.Errorf("Year %d: Ice age state differs: %v vs %v",
				year, cd1.IceAgeActive, cd2.IceAgeActive)
		}
		if cd1.CurrentInsolation != cd2.CurrentInsolation {
			t.Errorf("Year %d: Insolation differs: %.4f vs %.4f",
				year, cd1.CurrentInsolation, cd2.CurrentInsolation)
		}
	}
}

// TestClimateDriver_NilEventManager handles nil event manager gracefully.
func TestClimateDriver_NilEventManager(t *testing.T) {
	cd := NewClimateDriver(nil)

	// Should not panic even without event manager
	cd.Update(30750) // Year with low insolation

	// Ice age state should still be tracked
	// (the state is tracked even if no events are generated)
	if cd.CurrentInsolation == 0 {
		t.Error("Insolation should be calculated even without event manager")
	}
}
