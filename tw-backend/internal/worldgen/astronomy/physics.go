package astronomy

import "math"

// Physics constants for calculations
const (
	// earthMoonTidalBaseline is the tidal stress from Earth's Moon
	// Used to normalize tidal stress calculations to ~1.0 for Earth-Moon
	// Formula: Mass / Distance³ for Earth's Moon
	earthMoonTidalBaseline = MoonMassKg / (MoonDistanceMeters * MoonDistanceMeters * MoonDistanceMeters)

	// stableMoonMassRatio is the threshold for obliquity stability
	// If total moon mass / planet mass > this value, obliquity is stable
	stableMoonMassRatio = 0.01 // 1% of planet mass

	// impactShieldingPerMoon is the shielding factor per moon
	impactShieldingPerMoon = 0.05

	// maxImpactShielding caps the maximum shielding effect
	maxImpactShielding = 0.20
)

// CalculateTidalStress returns normalized tidal stress from satellites.
//
// The formula is: Σ(Mass / Distance³), normalized against Earth-Moon baseline.
// A return value of 1.0 represents tidal stress equivalent to Earth's Moon.
// Higher values indicate stronger tidal forces (useful for tectonics/volcanism).
//
// Returns 0.0 if there are no moons.
func CalculateTidalStress(moons []Satellite) float64 {
	if len(moons) == 0 {
		return 0.0
	}

	var totalStress float64
	for _, moon := range moons {
		if moon.Distance > 0 {
			// Tidal force ∝ Mass / Distance³
			stress := moon.Mass / math.Pow(moon.Distance, 3)
			totalStress += stress
		}
	}

	// Normalize against Earth-Moon baseline
	return totalStress / earthMoonTidalBaseline
}

// CalculateObliquityStability returns the axial tilt stability factor (0.0 to 1.0).
//
// A large moon stabilizes a planet's axial tilt, preventing chaotic wobble.
// This affects long-term climate stability and habitability.
//
// Rules:
//   - If total moon mass / planet mass > 1% → 1.0 (Stable, like Earth)
//   - Otherwise → 0.1 (Chaotic, like Mars)
//
// Returns 0.1 for planets with no moons or small moons.
func CalculateObliquityStability(moons []Satellite, planetMass float64) float64 {
	if len(moons) == 0 || planetMass <= 0 {
		return 0.1 // Chaotic without stabilizing moon
	}

	// Sum total moon mass
	var totalMoonMass float64
	for _, moon := range moons {
		totalMoonMass += moon.Mass
	}

	// Check if total moon mass exceeds stability threshold
	massRatio := totalMoonMass / planetMass
	if massRatio > stableMoonMassRatio {
		return 1.0 // Stable obliquity
	}

	return 0.1 // Chaotic obliquity
}

// CalculateImpactShielding returns the probability reduction factor for asteroid impacts.
//
// Moons act as gravitational shields, absorbing or deflecting some incoming objects.
// This is a simplified heuristic: 0.05 per moon, capped at 0.20.
//
// Returns:
//   - 0.0: No moons, no shielding
//   - 0.05: One moon (5% impact reduction)
//   - 0.10: Two moons (10% impact reduction)
//   - 0.15: Three moons (15% impact reduction)
//   - 0.20: Four or more moons (20% max impact reduction)
func CalculateImpactShielding(moons []Satellite) float64 {
	count := len(moons)
	if count == 0 {
		return 0.0
	}

	shielding := float64(count) * impactShieldingPerMoon
	if shielding > maxImpactShielding {
		return maxImpactShielding
	}

	return shielding
}

// CalculateTotalTidalHeating estimates tidal heating in watts.
//
// Tidal heating occurs when a moon's orbit causes tidal flexing of the planet.
// This can drive volcanic activity and potentially maintain subsurface oceans.
//
// Simplified formula based on tidal stress and orbital properties.
// Returns value relative to Earth (Earth = 1.0).
func CalculateTotalTidalHeating(moons []Satellite, planetMass float64) float64 {
	if len(moons) == 0 {
		return 0.0
	}

	var totalHeating float64
	for _, moon := range moons {
		// Heating ∝ Mass² / Distance⁶
		// Normalized against Earth-Moon baseline
		if moon.Distance > 0 {
			heating := (moon.Mass * moon.Mass) / math.Pow(moon.Distance, 6)
			earthMoonHeating := (MoonMassKg * MoonMassKg) / math.Pow(MoonDistanceMeters, 6)
			totalHeating += heating / earthMoonHeating
		}
	}

	return totalHeating
}

// CalculateMoonInfluenceIndex returns a composite score of moon influence on the planet.
//
// This combines tidal stress, obliquity stability, and impact shielding
// into a single normalized score from 0.0 to 1.0.
//
// Weights:
//   - Tidal Stress: 40% (capped at 2.0 before normalization)
//   - Obliquity Stability: 40%
//   - Impact Shielding: 20% (multiply by 5 to normalize 0.2 max to 1.0)
func CalculateMoonInfluenceIndex(moons []Satellite, planetMass float64) float64 {
	if len(moons) == 0 {
		return 0.0
	}

	// Calculate individual components
	tidalRaw := CalculateTidalStress(moons)
	tidal := math.Min(tidalRaw, 2.0) / 2.0 // Cap at 2.0, normalize to 0-1

	stability := CalculateObliquityStability(moons, planetMass)

	shieldingRaw := CalculateImpactShielding(moons)
	shielding := shieldingRaw / maxImpactShielding // Normalize to 0-1

	// Weighted combination
	return 0.4*tidal + 0.4*stability + 0.2*shielding
}
