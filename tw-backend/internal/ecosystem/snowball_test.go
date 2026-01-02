package ecosystem

import (
	"testing"
)

func TestSnowballEarth_TriggerAndRecovery(t *testing.T) {
	// Setup ClimateDriver
	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	// Verify initial state
	if cd.IsSnowball {
		t.Fatal("ClimateDriver should not start in Snowball state")
	}
	if cd.GlobalAlbedo != 0.30 {
		t.Errorf("GlobalAlbedo = %.2f, want 0.30 (Modern)", cd.GlobalAlbedo)
	}

	t.Run("Trigger Snowball Earth", func(t *testing.T) {
		// Simulate conditions that cause global temp < -10°C
		// With SolarLuminosity = 0.71 (early Earth), modern albedo, and NO greenhouse:
		// solarTempDelta = (0.71 - 1.0) * 70 = -20.3
		// albedoTempDelta = (0.30 - 0.30) * 100 = 0
		// globalTemp = 14 - 20.3 + 0 + GreenhouseOffset + GeothermalOffset
		//
		// For Hadean, Geothermal is HIGH. We need to simulate a later time.
		// Let's set to Year 2.5B where solar is ~0.83 and geothermal is ~1.
		// solarTempDelta = (0.83 - 1.0) * 70 = -11.9
		// If GreenhouseOffset is LOW (e.g., CO2 was weathered away), we get:
		// globalTemp = 14 - 11.9 + 0 - 25 + 0 = -22.9 -> Snowball!

		cd.GreenhouseOffset = -25.0 // Simulate massive CO2 crash (weathering)
		cd.GeothermalOffset = 1.0   // Late planet
		cd.SolarLuminosity = 0.83   // Mid-age sun

		cd.Update(2_500_000_000) // 2.5B years

		if !cd.IsSnowball {
			t.Error("ClimateDriver did not trigger Snowball Earth")
		}
		if cd.GlobalAlbedo != 0.70 {
			t.Errorf("GlobalAlbedo = %.2f, want 0.70 (Snowball)", cd.GlobalAlbedo)
		}

		// Check event was created
		foundEvent := false
		for _, e := range eventManager.ActiveEvents {
			if e.Type == EventGlobalGlaciation {
				foundEvent = true
				t.Logf("GlobalGlaciation event created: Severity=%.2f, TempMod=%.0f", e.Severity, e.TemperatureMod)
			}
		}
		if !foundEvent {
			t.Error("No GlobalGlaciation event was created")
		}
	})

	t.Run("Recover from Snowball", func(t *testing.T) {
		// Simulate volcanic CO2 buildup during Snowball
		// Weathering stops (frozen), volcanoes continue.
		// GreenhouseOffset rises until globalTemp > 10°C.
		//
		// With Albedo = 0.70:
		// albedoTempDelta = (0.30 - 0.70) * 100 = -40
		// globalTemp = 14 - 11.9 - 40 + Greenhouse + 1
		// To get globalTemp > 10: Greenhouse > 10 - 14 + 11.9 + 40 - 1 = 46.9

		cd.GreenhouseOffset = 50.0 // Massive volcanic CO2 buildup
		cd.Update(2_550_000_000)   // 50M years later

		if cd.IsSnowball {
			t.Error("ClimateDriver did not recover from Snowball Earth")
		}
		if cd.GlobalAlbedo != 0.30 {
			t.Errorf("GlobalAlbedo = %.2f, want 0.30 (Modern)", cd.GlobalAlbedo)
		}
		t.Log("Successfully recovered from Snowball Earth")
	})
}

func TestSnowballEarth_EarlyEarthDoesNotFreeze(t *testing.T) {
	// Hadean Earth should NOT freeze despite dim sun because:
	// 1. High Geothermal offset
	// 2. High Greenhouse offset from massive CO2

	eventManager := NewGeologicalEventManager()
	cd := NewClimateDriver(eventManager)

	// Simulate Hadean conditions
	cd.GreenhouseOffset = 50.0 // Massive CO2
	cd.GeothermalOffset = 90.0 // Hot planet
	cd.SolarLuminosity = 0.71  // Dim sun

	cd.Update(100_000_000) // 100M years

	if cd.IsSnowball {
		t.Error("Hadean Earth should NOT freeze with high geothermal + greenhouse")
	}
}
