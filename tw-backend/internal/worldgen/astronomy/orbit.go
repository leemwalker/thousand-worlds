// Package astronomy provides orbital mechanics calculations for long-term
// climate cycles based on Milankovitch theory.
//
// The package models three primary orbital parameters:
//   - Eccentricity: Shape of the orbit (100,000 year cycle)
//   - Obliquity: Axial tilt (41,000 year cycle)
//   - Precession: Orbital wobble (26,000 year cycle)
//
// These cycles combine to create variations in solar radiation (insolation)
// that drive ice ages and interglacial periods.
package astronomy

import "math"

// Orbital cycle periods in years
const (
	// EccentricityCycle is the period of orbital shape variation (~100k years)
	EccentricityCycle = 100000

	// ObliquityCycle is the period of axial tilt variation (~41k years)
	ObliquityCycle = 41000

	// PrecessionCycle is the period of orbital wobble (~26k years)
	PrecessionCycle = 26000
)

// Orbital parameter constants
const (
	// EccentricityBaseline is Earth's average orbital eccentricity
	EccentricityBaseline = 0.017

	// EccentricityAmplitude is the variation range (±0.01)
	EccentricityAmplitude = 0.01

	// ObliquityBaseline is Earth's average axial tilt in degrees
	ObliquityBaseline = 23.44

	// ObliquityAmplitude is the tilt variation range (±1.2°)
	ObliquityAmplitude = 1.2
)

// OrbitalState represents planetary orbital parameters at a given time.
// These parameters determine the distribution and intensity of solar radiation.
type OrbitalState struct {
	// Eccentricity is the shape of the orbit (0.0 = circle, higher = ellipse)
	// Range: 0.007 to 0.027 (Earth-like)
	Eccentricity float64

	// Obliquity is the axial tilt in degrees
	// Range: 22.1° to 24.5° (Earth-like)
	// Lower tilt = milder seasons = ice age potential
	Obliquity float64

	// Precession is the orbital wobble factor
	// Range: -1.0 to 1.0 (normalized sine wave)
	// Affects timing of perihelion relative to seasons
	Precession float64
}

// CalculateOrbitalState computes orbital parameters for a given simulation year.
// Uses simplified sine wave superposition based on Milankovitch theory.
//
// The function is deterministic: the same year always produces identical results.
// This is equivalent to CalculateOrbitalStateWithStability(year, 1.0) for Earth-like stability.
func CalculateOrbitalState(year int64) OrbitalState {
	return CalculateOrbitalStateWithStability(year, 1.0)
}

// CalculateOrbitalStateWithStability computes orbital parameters with obliquity chaos.
//
// The stability parameter represents how well natural satellites stabilize the axial tilt:
//   - 1.0 = Earth-Moon: Stable obliquity, ~2.4° total swing
//   - 0.5 = Reduced stability: ~7° swing
//   - 0.0 = No moons (Mars-like): Chaotic obliquity, ~26° swing
//
// Physics: Large moons stabilize axial tilt through gravitational interactions.
// Without a stabilizing moon, axial tilt can swing wildly over millions of years,
// causing extreme climate variations (flip between ice ages and hothouse conditions).
//
// The chaos multiplier scales the amplitude of obliquity oscillation:
//   - At stability=1.0: variance = 1.2° (Earth normal)
//   - At stability=0.0: variance = 1.2° × 11 = 13.2° (chaos)
func CalculateOrbitalStateWithStability(year int64, stability float64) OrbitalState {
	// Clamp stability to valid range
	if stability < 0 {
		stability = 0
	}
	if stability > 1 {
		stability = 1
	}

	// Convert year to float for calculations
	y := float64(year)

	// Calculate angular positions in each cycle (radians)
	eccAngle := 2 * math.Pi * y / float64(EccentricityCycle)
	oblAngle := 2 * math.Pi * y / float64(ObliquityCycle)
	precAngle := 2 * math.Pi * y / float64(PrecessionCycle)

	// Chaos multiplier: at stability=0, obliquity variance is 11x normal
	// At stability=1.0: multiplier = 1.0 (normal)
	// At stability=0.0: multiplier = 11.0 (chaotic, 13.2° swing)
	chaosMultiplier := 1.0 + (1.0-stability)*10.0
	effectiveAmplitude := ObliquityAmplitude * chaosMultiplier

	return OrbitalState{
		// Eccentricity: baseline ± amplitude * sin(cycle)
		// Range: 0.017 ± 0.01 = [0.007, 0.027]
		Eccentricity: EccentricityBaseline + EccentricityAmplitude*math.Sin(eccAngle),

		// Obliquity: baseline ± effective amplitude * sin(cycle)
		// Stable: 23.44° ± 1.2° = [22.24°, 24.64°]
		// Chaotic: 23.44° ± 13.2° = [10.24°, 36.64°]
		Obliquity: ObliquityBaseline + effectiveAmplitude*math.Sin(oblAngle),

		// Precession: simple sine wave [-1, 1]
		// Represents the phase of orbital precession
		Precession: math.Sin(precAngle),
	}
}

// CalculateInsolation returns a normalized solar energy factor.
// A value of 1.0 represents baseline insolation.
//
// The calculation considers:
//   - Obliquity: Lower tilt = milder summers = less ice melt = ice age potential
//   - Eccentricity: Higher eccentricity affects seasonal asymmetry
//   - Precession: Modulates the effect of eccentricity on seasons
//
// Returns a value typically in the range [0.93, 1.07].
func CalculateInsolation(state OrbitalState) float64 {
	// Baseline insolation
	insolation := 1.0

	// Obliquity effect on summer insolation at high latitudes
	// Higher obliquity = more extreme seasons = more summer ice melt
	// Normalized: (current - min) / (max - min) scaled to [0.95, 1.05]
	//
	// At min obliquity (22.24°): reduced summer heating → ice age risk
	// At max obliquity (24.64°): enhanced summer heating → interglacial
	obliquityMin := ObliquityBaseline - ObliquityAmplitude                            // 22.24
	obliquityMax := ObliquityBaseline + ObliquityAmplitude                            // 24.64
	obliquityNorm := (state.Obliquity - obliquityMin) / (obliquityMax - obliquityMin) // 0 to 1

	// Map to [-0.03, +0.03] effect on insolation
	// High obliquity adds insolation, low obliquity subtracts
	obliquityEffect := (obliquityNorm - 0.5) * 0.06

	// Eccentricity effect: higher eccentricity amplifies precession effects
	// The effect is modulated by precession phase
	// When perihelion occurs in summer (precession ~1), high eccentricity helps
	// When perihelion occurs in winter (precession ~-1), high eccentricity hurts
	eccentricityEffect := state.Eccentricity * state.Precession * 0.5

	// Combine effects
	insolation += obliquityEffect
	insolation += eccentricityEffect

	return insolation
}

// IceAgePotential returns a value from 0.0 to 1.0 indicating the likelihood
// of ice age conditions based on the orbital state.
// A value > 0.5 suggests ice age prone conditions.
func IceAgePotential(state OrbitalState) float64 {
	insolation := CalculateInsolation(state)

	// Map insolation to ice age potential
	// Insolation ~0.93 → potential ~1.0 (maximum ice age risk)
	// Insolation ~1.07 → potential ~0.0 (minimum ice age risk)
	// Linear mapping with 1.0 as the neutral point
	potential := (1.0 - insolation) / 0.07

	// Clamp to [0, 1]
	if potential < 0 {
		return 0
	}
	if potential > 1 {
		return 1
	}
	return potential
}
