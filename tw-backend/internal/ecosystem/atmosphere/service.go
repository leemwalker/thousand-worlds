package atmosphere

import (
	"math"
	"sync"
)

// Atmosphere tracks planetary atmospheric composition and greenhouse effect over geological time
type Atmosphere struct {
	mu sync.RWMutex

	// Atmospheric masses (in Earth atmospheres - 1.0 = 1 atm)
	CO2Mass float64 // Carbon dioxide (primary greenhouse gas)
	N2Mass  float64 // Nitrogen (inert filler gas)
	O2Mass  float64 // Oxygen (from photosynthesis, not yet modeled)

	// Derived properties
	TotalMass        float64 // Total atmospheric mass (atm)
	GreenhouseFactor float64 // Temperature offset from greenhouse effect (°C)

	// Planet Physics
	PlanetMassKg    float64
	PlanetRadiusM   float64
	RetentionFactor float64 // 0.0 to 1.0 (how well the planet holds atmosphere)

	// Simulation state
	TotalYearsSimulated int64
}

// AtmosphereStats contains summary statistics for display
type AtmosphereStats struct {
	CO2_ppm          float64 // CO2 concentration in parts per million
	TotalPressure    float64 // Total atmospheric pressure (atm)
	GreenhouseOffset float64 // Temperature contribution from greenhouse effect (°C)
}

// NewAtmosphere creates initial atmospheric composition based on planetary age
//
// Early Earth (Hadean/Archean): Volcanic CO2-rich reducing atmosphere
// - High CO2 (~50× modern) to compensate for faint young Sun
// - Minimal N2, no O2 (pre-oxygenic)
//
// Modern Earth: N2-O2 atmosphere with trace CO2
// - Trace CO2 (~400 ppm)
// - N2-dominated (78%)
// - O2 from photosynthesis (21%)
// NewAtmosphere creates initial atmospheric composition based on planetary age
func NewAtmosphere(startYear int64, massKg, radiusM float64) *Atmosphere {
	atm := &Atmosphere{
		TotalYearsSimulated: startYear,
		PlanetMassKg:        massKg,
		PlanetRadiusM:       radiusM,
	}

	// Calculate atmospheric retention based on escape velocity
	atm.calculateRetentionFactor()

	// Initialize based on starting year
	if startYear < 2_000_000_000 {
		// Early Earth (before 2B years ago): High CO2
		// Need strong greenhouse to compensate for 70-80% solar luminosity
		atm.CO2Mass = 50.0 // 50 atm CO2 (50× total modern atmosphere!)
		atm.N2Mass = 0.5   // Minimal N2
		atm.O2Mass = 0.0   // Pre-oxygenic photosynthesis
	} else {
		// Modern-like Earth: Post-Great Oxidation Event
		atm.CO2Mass = 0.0006 // ~400 ppm (0.06% of 1 atm)
		atm.N2Mass = 0.78    // 78% of 1 atm
		atm.O2Mass = 0.21    // 21% of 1 atm
	}

	atm.updateDerivedProperties()
	return atm
}

// updateDerivedProperties calculates greenhouse factor and total mass from composition
//
// Greenhouse Effect Formula:
// Using logarithmic CO2 sensitivity: ΔT = α × ln(C/C₀)
// Simplified to: ΔT ≈ 3°C per doubling of CO2
//
// Physics: CO2 absorbs infrared radiation, trapping heat
// More CO2 → stronger absorption → higher surface temperature
// calculateRetentionFactor determines how well the planet holds an atmosphere
// Based on escape velocity relative to terrestrial (Earth) baseline.
// - Mars (0.1 mass): Poor retention
// - Earth (1.0 mass): Good retention
// - Super-Earth (2.0 mass): Excellent retention (thick atmosphere)
func (a *Atmosphere) calculateRetentionFactor() {
	const G = 6.67430e-11
	const EarthEscapeVelocity = 11186.0 // m/s

	if a.PlanetMassKg <= 0 || a.PlanetRadiusM <= 0 {
		a.RetentionFactor = 0.0
		return
	}

	escapeVelocity := math.Sqrt(2 * G * a.PlanetMassKg / a.PlanetRadiusM)
	ratio := escapeVelocity / EarthEscapeVelocity

	// Non-linear retention curve
	// < 0.5 (Mars-like): Rapid loss
	// 0.5 - 1.0: Increasing retention
	// > 1.0 (Super-Earth): Accumulates thick atmosphere
	if ratio < 0.5 {
		a.RetentionFactor = ratio * 0.2 // Very poor
	} else {
		a.RetentionFactor = math.Pow(ratio, 2.0) // Quadratic scaling
	}
}

func (a *Atmosphere) updateDerivedProperties() {
	// Base mass from composition
	baseMass := a.CO2Mass + a.N2Mass + a.O2Mass

	// Apply planet retention capability
	// A massive planet holds 2x atmosphere for the same "geological output"
	// A small planet loses most of it
	a.TotalMass = baseMass * a.RetentionFactor

	// Calculate greenhouse warming from CO2
	// Reference: Modern CO2 = 400 ppm = 0.0006 atm
	const modernCO2 = 0.0006
	const degreesPerDoubling = 3.0 // IPCC consensus value

	// Update greenhouse effect based on Total Mass (pressure broadening)
	// Thicker atmosphere = stronger greenhouse effect
	// Thin atmosphere = weaker greenhouse
	pressureEffect := math.Sqrt(a.TotalMass) // Square root scaling for pressure broadening

	if a.CO2Mass <= 0 {
		a.GreenhouseFactor = 0.0
		return
	}

	// Calculate number of doublings relative to modern baseline
	// log₂(C/C₀) gives doublings
	// Use effective CO2 (scaled by total pressure)
	effectiveCO2 := a.CO2Mass * pressureEffect
	doublings := math.Log2(effectiveCO2 / modernCO2)

	//Greenhouse temperature offset (°C)
	a.GreenhouseFactor = doublings * degreesPerDoubling
}

// SimulateCarbonCycle updates atmospheric CO2 based on volcanic sources and weathering sinks
//
// The carbon-silicate cycle provides long-term climate stability:
// - Source: Volcanic outgassing (proportional to planetary heat/tectonic activity)
// - Sink: Silicate weathering (proportional to temperature × precipitation × CO2)
//
// Negative Feedback Loop:
// Planet warms → More rain → Faster weathering → Less CO2 → Cooling
// Planet cools → Less rain → Slower weathering → More CO2 → Warming
//
// This thermostat keeps Earth habitable despite the Sun brightening 30% over 4.5B years!
func (a *Atmosphere) SimulateCarbonCycle(dt int64, volcanicRate, weatheringRate float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	dtYears := float64(dt)

	// Source: Volcanic outgassing adds CO2
	CO2_added := volcanicRate * dtYears

	// Sink: Silicate weathering removes CO2
	CO2_removed := weatheringRate * dtYears

	// Update CO2 mass (mass balance)
	a.CO2Mass += CO2_added - CO2_removed

	// Floor at trace levels (volcanic outgassing prevents complete removal)
	const minCO2 = 0.0001 // Trace CO2 always present
	if a.CO2Mass < minCO2 {
		a.CO2Mass = minCO2
	}

	// Update derived properties (greenhouse effect)
	a.updateDerivedProperties()

	a.TotalYearsSimulated += dt
}

// GetStats returns current atmospheric state for display/logging
func (a *Atmosphere) GetStats() AtmosphereStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Convert CO2 mass to ppm (parts per million by volume)
	co2_ppm := 0.0
	if a.TotalMass > 0 {
		co2_ppm = (a.CO2Mass / a.TotalMass) * 1_000_000
	}

	return AtmosphereStats{
		CO2_ppm:          co2_ppm,
		TotalPressure:    a.TotalMass,
		GreenhouseOffset: a.GreenhouseFactor,
	}
}

// CalculateVolcanicOutgassing estimates CO2 emissions from volcanic activity
//
// Proportional to planetary heat (tectonic activity drives volcanism):
// - Early Earth (heat=10): ~10× modern outgassing rate
// - Modern Earth (heat=1): Baseline rate
//
// Modern Earth: ~0.26 Gt C/yr ≈ 0.0000005 atm CO2/yr
func CalculateVolcanicOutgassing(planetaryHeat float64) float64 {
	const modernRate = 0.0000005 // atm CO2/yr (based on modern volcanic emissions)
	return modernRate * planetaryHeat
}

// CalculateWeatheringRate estimates CO2 consumption from silicate weathering
//
// Weathering reaction: CaSiO₃ + CO₂ + H₂O → CaCO₃ + SiO₂(aq)
//
// Driven by:
// - Temperature (Arrhenius kinetics, Q10 ≈ 2)
// - Precipitation (transport/hydrolysis)
// - Land area (exposed rock surface)
// - CO2 partial pressure (carbonic acid concentration)
//
// This is the key negative feedback that stabilizes climate!
func CalculateWeatheringRate(avgTemp, precipitation, landArea, currentCO2 float64) float64 {
	// Modern baseline parameters
	const (
		modernTemp   = 15.0      // °C
		modernPrecip = 1000.0    // mm/yr
		modernArea   = 1.0       // normalized (100% = modern land coverage)
		modernCO2    = 0.0006    // atm
		modernRate   = 0.0000005 // atm CO2/yr consumed by weathering
		q10          = 2.0       // Rate doubles every 10°C
	)

	// Temperature dependence (Arrhenius-like)
	// Q10 = 2 means rate doubles every 10°C
	tempDelta := avgTemp - modernTemp
	tempFactor := math.Pow(q10, tempDelta/10.0)

	// Precipitation dependence (linear approximation)
	// More rain → more water → more weathering
	precipFactor := precipitation / modernPrecip

	// Land area dependence
	// More exposed rock → more weathering surface
	areaFactor := landArea / modernArea

	// CO2 dependence (square root approximation)
	// More CO2 → more carbonic acid → faster weathering
	// Using sqrt to prevent runaway (diminishing returns)
	co2Factor := math.Sqrt(currentCO2 / modernCO2)

	// Combined weathering rate
	rate := modernRate * tempFactor * precipFactor * areaFactor * co2Factor

	return rate
}
