package atmosphere

import (
	"math"
	"testing"
)

// TestNewAtmosphere verifies initial atmospheric composition
func TestNewAtmosphere(t *testing.T) {
	t.Run("early Earth has high CO2", func(t *testing.T) {
		atm := NewAtmosphere(0) // Hadean

		if atm.CO2Mass < 10.0 {
			t.Errorf("Early Earth should have high CO2: got %f, want >= 10.0", atm.CO2Mass)
		}

		if atm.GreenhouseFactor <= 0 {
			t.Errorf("Early Earth should have strong greenhouse effect: got %f", atm.GreenhouseFactor)
		}
	})

	t.Run("modern Earth has low CO2", func(t *testing.T) {
		atm := NewAtmosphere(4_500_000_000) // Modern

		// Modern Earth: ~400 ppm = 0.0006 atm
		if atm.CO2Mass > 0.001 {
			t.Errorf("Modern Earth should have trace CO2: got %f, want <= 0.001", atm.CO2Mass)
		}

		if atm.O2Mass <= 0 {
			t.Errorf("Modern Earth should have O2: got %f", atm.O2Mass)
		}
	})
}

// TestGreenhouseFactor verifies CO2 greenhouse calculation
func TestGreenhouseFactor(t *testing.T) {
	tests := []struct {
		name           string
		co2Mass        float64
		expectedOffset float64 // °C
		tolerance      float64
	}{
		{
			name:           "modern baseline (400 ppm)",
			co2Mass:        0.0006,
			expectedOffset: 0.0, // Reference level
			tolerance:      0.5,
		},
		{
			name:           "double CO2 (800 ppm)",
			co2Mass:        0.0012,
			expectedOffset: 3.0, // 3°C per doubling
			tolerance:      0.5,
		},
		{
			name:           "quadruple CO2 (1600 ppm)",
			co2Mass:        0.0024,
			expectedOffset: 6.0, // Two doublings
			tolerance:      0.5,
		},
		{
			name:           "10x CO2 (4000 ppm)",
			co2Mass:        0.006,
			expectedOffset: 10.0, // log2(10) ≈ 3.32 doublings
			tolerance:      1.0,
		},
		{
			name:           "100x CO2 (Hadean-like)",
			co2Mass:        0.06,
			expectedOffset: 20.0, // log2(100) ≈ 6.64 doublings
			tolerance:      2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atm := &Atmosphere{
				CO2Mass: tt.co2Mass,
				N2Mass:  0.78,
				O2Mass:  0.21,
			}
			atm.updateDerivedProperties()

			if math.Abs(atm.GreenhouseFactor-tt.expectedOffset) > tt.tolerance {
				t.Errorf("GreenhouseFactor = %f, want %f (±%f)",
					atm.GreenhouseFactor, tt.expectedOffset, tt.tolerance)
			}
		})
	}
}

// TestCalculateVolcanicOutgassing verifies scaling with planetary heat
func TestCalculateVolcanicOutgassing(t *testing.T) {
	tests := []struct {
		name          string
		planetaryHeat float64
		expectedRate  float64
		minRate       float64
	}{
		{
			name:          "Hadean (heat=10)",
			planetaryHeat: 10.0,
			expectedRate:  0.000005, // 10× modern
			minRate:       0.000004,
		},
		{
			name:          "Archean (heat=3)",
			planetaryHeat: 3.0,
			expectedRate:  0.0000015, // 3× modern
			minRate:       0.0000010,
		},
		{
			name:          "Modern (heat=1)",
			planetaryHeat: 1.0,
			expectedRate:  0.0000005, // Baseline
			minRate:       0.0000004,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := CalculateVolcanicOutgassing(tt.planetaryHeat)

			if rate < tt.minRate {
				t.Errorf("Volcanic rate too low: got %e, want >= %e", rate, tt.minRate)
			}
		})
	}
}

// TestCalculateWeatheringRate verifies temperature and CO2 dependence
func TestCalculateWeatheringRate(t *testing.T) {
	t.Run("doubles with 10°C warming (Q10=2)", func(t *testing.T) {
		modernRate := CalculateWeatheringRate(15.0, 1000.0, 1.0, 0.0006)
		hotRate := CalculateWeatheringRate(25.0, 1000.0, 1.0, 0.0006)

		ratio := hotRate / modernRate
		if math.Abs(ratio-2.0) > 0.2 {
			t.Errorf("Q10 ratio = %f, want ~2.0", ratio)
		}
	})

	t.Run("increases with higher CO2", func(t *testing.T) {
		lowCO2Rate := CalculateWeatheringRate(15.0, 1000.0, 1.0, 0.0006)
		highCO2Rate := CalculateWeatheringRate(15.0, 1000.0, 1.0, 0.006)

		if highCO2Rate <= lowCO2Rate {
			t.Errorf("Higher CO2 should increase weathering: low=%e, high=%e",
				lowCO2Rate, highCO2Rate)
		}
	})

	t.Run("increases with precipitation", func(t *testing.T) {
		dryRate := CalculateWeatheringRate(15.0, 500.0, 1.0, 0.0006)
		wetRate := CalculateWeatheringRate(15.0, 2000.0, 1.0, 0.0006)

		if wetRate <= 2*dryRate {
			t.Errorf("Wetter should have higher weathering: dry=%e, wet=%e",
				dryRate, wetRate)
		}
	})
}

// TestSimulateCarbonCycle verifies mass balance
func TestSimulateCarbonCycle(t *testing.T) {
	atm := NewAtmosphere(0)
	initialCO2 := atm.CO2Mass

	// High volcanism, low weathering → CO2 increases
	volcanicRate := 0.000005
	weatheringRate := 0.0000001
	dt := int64(1_000_000) // 1 million years

	atm.SimulateCarbonCycle(dt, volcanicRate, weatheringRate)

	expectedChange := (volcanicRate - weatheringRate) * float64(dt)
	actualChange := atm.CO2Mass - initialCO2

	if math.Abs(actualChange-expectedChange) > 0.01 {
		t.Errorf("CO2 change = %f, want %f", actualChange, expectedChange)
	}
}

// TestCarbonCycle_NegativeFeedback verifies climate regulation
func TestCarbonCycle_NegativeFeedback(t *testing.T) {
	// Scenario: Planet warms → weathering increases → CO2 decreases → cooling

	coldTemp := 10.0
	hotTemp := 30.0
	co2 := 0.01 // High CO2 state

	coldWeathering := CalculateWeatheringRate(coldTemp, 1000.0, 1.0, co2)
	hotWeathering := CalculateWeatheringRate(hotTemp, 1000.0, 1.0, co2)

	// Hot planet should weather faster (negative feedback)
	if hotWeathering <= coldWeathering {
		t.Errorf("Negative feedback broken: hot=%e should be > cold=%e",
			hotWeathering, coldWeathering)
	}

	// Ratio should be substantial (Q10=2 for 20°C difference)
	ratio := hotWeathering / coldWeathering
	if ratio < 3.0 {
		t.Errorf("Negative feedback too weak: ratio=%f, want >3", ratio)
	}
}
