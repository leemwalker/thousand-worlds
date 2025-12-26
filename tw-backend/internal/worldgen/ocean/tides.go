// Package ocean provides oceanic simulation including currents, thermohaline
// circulation, and tidal mechanics.
package ocean

import (
	"math"

	"tw-backend/internal/worldgen/astronomy"
)

// Physical constants for tidal calculations
const (
	// earthMoonTidalBaseline is the tidal force from Earth's Moon
	// Formula: Mass / Distance³ for Earth's Moon
	// This normalizes tidal amplitude so Earth-Moon = 1.0 meters
	earthMoonMass     = 7.342e22 // kg
	earthMoonDistance = 384400e3 // meters (384,400 km)
	earthMoonBaseline = earthMoonMass / (earthMoonDistance * earthMoonDistance * earthMoonDistance)
)

// CalculateTidalAmplitude returns the tidal range in meters.
//
// The formula is: Σ(Mass / Distance³), normalized against Earth-Moon baseline.
// A return value of 1.0 represents ~1 meter tidal range (open ocean).
// Coastal amplification can multiply this by 2-10x depending on geography.
//
// Physics: Tidal force ∝ M/d³ (mass over distance cubed)
//
// Returns 0.0 if there are no moons.
func CalculateTidalAmplitude(satellites []astronomy.Satellite) float64 {
	if len(satellites) == 0 {
		return 0.0
	}

	var totalForce float64
	for _, sat := range satellites {
		if sat.Distance > 0 {
			// Tidal force ∝ Mass / Distance³
			force := sat.Mass / math.Pow(sat.Distance, 3)
			totalForce += force
		}
	}

	// Normalize against Earth-Moon baseline
	// Result: 1.0 = Earth-like tides (~1 meter open ocean)
	return totalForce / earthMoonBaseline
}

// GetTidalCategory returns a classification based on tidal amplitude in meters.
//
// Categories:
//   - "micro": < 0.5m - Minimal tides, calm coastlines
//   - "normal": 0.5-2m - Earth-like tides, typical intertidal zones
//   - "strong": 2-5m - Large tides, extensive intertidal ecosystems
//   - "extreme": > 5m - Extreme tides, dramatic daily flooding
func GetTidalCategory(amplitude float64) string {
	switch {
	case amplitude < 0.5:
		return "micro"
	case amplitude < 2.0:
		return "normal"
	case amplitude < 5.0:
		return "strong"
	default:
		return "extreme"
	}
}

// CalculateSpringNeapRatio returns the ratio between spring and neap tides.
//
// For a single moon, spring/neap ratio is ~1.0 (no variation).
// For multiple moons, the ratio depends on their orbital periods.
// Higher ratios mean more dramatic tidal variation through lunar cycles.
//
// Returns 1.0 for 0 or 1 moons.
func CalculateSpringNeapRatio(satellites []astronomy.Satellite) float64 {
	if len(satellites) <= 1 {
		return 1.0
	}

	// Simplified model: ratio increases with number of moons
	// Real physics would involve orbital period resonances
	// For now: 1.0 (1 moon), ~1.5 (2 moons), ~2.0 (3+ moons)
	switch len(satellites) {
	case 2:
		return 1.5
	case 3:
		return 1.8
	default:
		return 2.0
	}
}

// TidalInfo aggregates tidal data for a planetary system
type TidalInfo struct {
	// Amplitude is the average tidal range in meters
	Amplitude float64
	// Category is the classification (micro/normal/strong/extreme)
	Category string
	// SpringNeapRatio is the ratio between maximum and minimum tides
	SpringNeapRatio float64
	// MoonCount is the number of satellites contributing
	MoonCount int
}

// CalculateTidalInfo returns complete tidal information for a set of satellites
func CalculateTidalInfo(satellites []astronomy.Satellite) TidalInfo {
	amp := CalculateTidalAmplitude(satellites)
	return TidalInfo{
		Amplitude:       amp,
		Category:        GetTidalCategory(amp),
		SpringNeapRatio: CalculateSpringNeapRatio(satellites),
		MoonCount:       len(satellites),
	}
}
