package ecosystem

import (
	"math"
)

// CarbonReservoir tracks the global carbon inventory in Gigatons (Gt).
// Mass balance is crucial: Source (Mantle) -> Atmosphere/Ocean -> Crust (Sink).
type CarbonReservoir struct {
	Atmosphere float64 // Gt Carbon (CO2)
	Ocean      float64 // Gt Carbon (Dissolved HCO3-)
	Crust      float64 // Gt Carbon (Carbonates/Buried Organic)
	Mantle     float64 // Gt Carbon (Deep reservoir, source of outgassing)
}

// AtmosphereState represents the chemical composition of the atmosphere.
// This drives the greenhouse effect and pressure calculations.
type AtmosphereState struct {
	CO2ppm      float64 // Parts per Million (Greenhouse gas)
	Methaneppm  float64 // Parts per Million (Strong greenhouse gas, early Earth)
	Nitrogen    float64 // Partial Pressure (Atm) - Dominant gas, sets Boiling Point
	OxygenLevel float64 // Percentage (0.0 - 1.0) - Affects Methane lifetime
	Temperature float64 // Global Average Surface Temperature (°C)
}

// NutrientFlux tracks limiting nutrients released by weathering.
type NutrientFlux struct {
	Phosphorous float64 // Mt/Year (Megatons) - Limits biological carrying capacity
}

// CarbonCycle manages the long-term geochemical cycling of carbon and nutrients.
type CarbonCycle struct {
	Reservoir CarbonReservoir
	State     AtmosphereState
	Flux      NutrientFlux
}

// NewCarbonCycle initializes the system with Hadean Eon conditions.
// Hadean Earth (4.5 BYA):
// - High CO2 (10-20% atmosphere) due to differentiation and impact degassing.
// - High Methane (if pre-biotic, maybe low, but methanogens rise in Archean).
// - High Nitrogen (potentially 2-3 bar, later reduced).
// - Zero Oxygen.
// - Hot Mantle (high outgassing potential).
func NewCarbonCycle() *CarbonCycle {
	return &CarbonCycle{
		Reservoir: CarbonReservoir{
			Atmosphere: 100000.0, // Massive initial atmosphere (~100x Modern)
			Ocean:      40000.0,  // Equilibrium with high pCO2
			Crust:      0.0,      // Minimal continental crust initially
			Mantle:     1e8,      // Vast reservoir
		},
		State: AtmosphereState{
			CO2ppm:      100000.0, // 10% CO2
			Methaneppm:  10.0,     // Trace initially
			Nitrogen:    2.0,      // 2 Atm pressure
			OxygenLevel: 0.0,      // Anoxic
			Temperature: 80.0,     // Very hot (Greenhouse + Geothermal offset)
		},
		Flux: NutrientFlux{
			Phosphorous: 0.0, // No weathering yet (no continents)
		},
	}
}

// CalculateGreenhouseTemp derives global temperature from atmospheric composition
// and solar luminosity.
// luminosity: Relative to modern sun (0.7 - 1.0).
// Returns: Surface Temperature in Celsius.
func (cc *CarbonCycle) CalculateGreenhouseTemp(luminosity float64) float64 {
	// 1. Blackbody/Solar Component
	// Modern Earth (L=1.0) is ~14°C.
	// Hadean Sun (L=0.7) would be ~20°C colder without greenhouse compensation.
	// Sensitivity: dT ~ 70 * (L - 1.0)
	// (Check: 0.3 drop * 70 = 21°C drop. 14 - 21 = -7. Matches Stefan-Boltzmann approx).
	solarForcing := (luminosity - 1.0) * 70.0

	// 2. Greenhouse Forcing (dT from Pre-Industrial Baseline)
	// CO2 Forcing: 5.35 * ln(C/C0)
	// Modern Baseline (C0) = 280 ppm
	co2Ratio := cc.State.CO2ppm / 280.0
	if co2Ratio < 0.001 {
		co2Ratio = 0.001
	}
	forcingCO2 := 5.35 * math.Log(co2Ratio)

	// Methane Forcing: ~0.5 * sqrt(M) (Simplified)
	forcingCH4 := 0.5 * math.Sqrt(cc.State.Methaneppm)
	if forcingCH4 < 0 {
		forcingCH4 = 0
	}

	totalForcingWatts := forcingCO2 + forcingCH4

	// Climate Sensitivity: ~0.8 °C per W/m2
	greenhouseWarming := totalForcingWatts * 0.8

	// Base Temp = 14.0 (Modern Average)
	return 14.0 + solarForcing + greenhouseWarming
}

// GetGreenhouseWarming returns just the temperature offset from CO2 and Methane.
// usage: climateDriver.SetGreenhouseOffset(cc.GetGreenhouseWarming())
func (cc *CarbonCycle) GetGreenhouseWarming() float64 {
	// 1. CO2 Forcing: 5.35 * ln(C/C0)
	co2Ratio := cc.State.CO2ppm / 280.0
	if co2Ratio < 0.001 {
		co2Ratio = 0.001
	}
	forcingCO2 := 5.35 * math.Log(co2Ratio)

	// 2. Methane Forcing
	forcingCH4 := 0.5 * math.Sqrt(cc.State.Methaneppm)
	if forcingCH4 < 0 {
		forcingCH4 = 0
	}

	totalForcingWatts := forcingCO2 + forcingCH4
	return totalForcingWatts * 0.8 // Sensitivity
}

// Update runs the carbon cycle for a given time step.
// dt: Time step in millions of years.
// volcanicActivity: Normalized factor (1.0 = modern, 5.0 = Hadean).
// continentalArea: Fraction of surface covered by land (0.0 - 0.3).
// rainfall: Global average rainfall multiplier (1.0 = modern).
func (cc *CarbonCycle) Update(dt float64, volcanicActivity, continentalArea, rainfall float64) {
	// 1. Volcanic Outgassing (Source)
	// Flux from Mantle to Atmosphere
	// Base rate ~0.1 Gt/year (Modern) -> 100 Mt/yr
	// Adjusted by Tectonic Activity
	outgassingRate := 0.1 * volcanicActivity    // Gt/yr
	outgassingFlux := outgassingRate * dt * 1e6 // Total Gt over dt (Million Years)

	// Transfer Carbon
	if cc.Reservoir.Mantle > outgassingFlux {
		cc.Reservoir.Mantle -= outgassingFlux
		cc.Reservoir.Atmosphere += outgassingFlux
	}

	// 2. Silicate Weathering (Sink)
	// Flux from Atmosphere to Crust (via Ocean)
	// Dependent on:
	// - Rainfall (more rain = more acid scrub)
	// - Continental Area (more rock to weather)
	// - Temperature (Arrhenius equation kinetic boost)

	// Temperature Factor: Weathering doubles every ~10°C increase
	// Using simplified exponential: exp((T - 15) / 13.7)
	// Walker et al (1981) feedback
	tempFactor := math.Exp((cc.State.Temperature - 15.0) / 13.7)

	// Weathering Rate
	// Base ~0.1 Gt/yr (balances modern outgassing)
	weatheringRate := 0.1 * rainfall * (continentalArea / 0.3) * tempFactor
	weatheringFlux := weatheringRate * dt * 1e6

	// Nutrient flux linked to weathering
	// Phosphorous release ~ 1/1000th of Carbon weathering
	cc.Flux.Phosphorous = weatheringFlux / 1000.0

	// Transfer Carbon (Atmosphere -> Crust)
	// (Simplified: Skip Ocean buffer dynamics for now, assume equilibrium)
	actualSink := math.Min(cc.Reservoir.Atmosphere, weatheringFlux)
	cc.Reservoir.Atmosphere -= actualSink
	cc.Reservoir.Crust += actualSink

	// 3. Update State (PPM)
	// Atmosphere Mass ~ 5.15e18 kg
	// 1 Gt C = 1e12 kg => conversion to ppmv
	// Very rough conversion: 1 Gt C ~= 0.5 ppm CO2 (on modern Earth)
	// But early Earth had thicker atmosphere (Nitrogen), so ppm fraction might differ.
	// For simulation stability, we use a fixed conversion scaler.
	const GtToPPRatio = 0.5
	cc.State.CO2ppm = cc.Reservoir.Atmosphere * GtToPPRatio

	// =========================================================================
	// Phase 9c: GREAT OXIDATION EVENT (GOE)
	// =========================================================================
	// Models the rise of atmospheric oxygen from ~0% to ~21% over ~2 billion years.
	// Key players:
	// - Cyanobacteria: Produce O2 via photosynthesis (Source)
	// - Iron Sinks: Dissolved Fe2+ in oceans absorbs O2 -> Banded Iron Formations (Sink)
	// - Methane Collapse: O2 destroys CH4 -> Reduces greenhouse -> Cooling spike

	// Nutrient-limited oxygen production (Cyanobacteria)
	// Rate scales with Phosphorous availability (from weathering) and non-frozen conditions.
	// Base rate ~0.001%/My when nutrients available.
	if cc.Flux.Phosphorous > 0 && cc.State.Temperature > 0 {
		// O2 production rate: Phosphorous flux * nutrient efficiency
		// Capped at 0.5% per million years (realistic biological growth)
		o2ProductionRate := math.Min(cc.Flux.Phosphorous*0.1, 0.005)
		cc.State.OxygenLevel += o2ProductionRate * dt
	}

	// Iron Sink (Banded Iron Formations)
	// Early oceans were anoxic with dissolved Fe2+.
	// O2 reacts with Fe2+ -> Fe2O3 (precipitates, forms BIFs).
	// This "sponges" O2 until iron is exhausted (~2.4B years on Earth).
	// We model as: If O2 < 2% and planetAge < 2.5B, sink consumes O2.
	// Simplified: Sink rate = 0.3 * O2Level (first-order consumption)
	if cc.State.OxygenLevel > 0 && cc.State.OxygenLevel < 0.02 {
		ironSinkRate := 0.3 * cc.State.OxygenLevel * dt
		cc.State.OxygenLevel -= ironSinkRate
		if cc.State.OxygenLevel < 0 {
			cc.State.OxygenLevel = 0
		}
	}

	// Methane Collapse
	// Once O2 rises above ~1%, it rapidly oxidizes atmospheric CH4.
	// CH4 + 2 O2 -> CO2 + 2 H2O (Exothermic, but net cooling due to CH4 loss)
	// Methane is a powerful greenhouse gas; its destruction causes cooling.
	// Model: If O2 > 1%, CH4 decays exponentially with half-life of ~50My.
	if cc.State.OxygenLevel > 0.01 && cc.State.Methaneppm > 0.1 {
		// Decay constant: Half-life of 50My -> lambda = ln(2)/50 = 0.0139
		decayRate := 0.0139 * dt
		cc.State.Methaneppm *= math.Exp(-decayRate)
		if cc.State.Methaneppm < 0.1 {
			cc.State.Methaneppm = 0.1 // Trace minimum
		}
	}

	// Clamp Oxygen Level to [0, 0.21] (max 21% like modern Earth)
	if cc.State.OxygenLevel > 0.21 {
		cc.State.OxygenLevel = 0.21
	}
}
