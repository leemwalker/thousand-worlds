package astronomy

import (
	"math"
	"math/rand"
)

type GiantImpactOutcome string

const (
	OutcomeMoonFormation GiantImpactOutcome = "moon"    // Debris coalesces into orbit
	OutcomeDirectImpact  GiantImpactOutcome = "impact"  // Object hits planet directly
	OutcomeCaptured      GiantImpactOutcome = "capture" // Object captured as-is (rare)
	OutcomeEscape        GiantImpactOutcome = "escape"  // Object misses or escapes
)

// SimulateGiantImpact determines outcome based on physics
// impactorMass: in kg
// planetMass: in kg
// velocityRatio: ratio of impact velocity to escape velocity (typically 1.0-1.5)
// impactAngle: degrees from head-on (0 = head-on, 90 = grazing/miss)
// rng: random source for variability
func SimulateGiantImpact(
	impactorMass float64,
	planetMass float64,
	velocityRatio float64,
	impactAngle float64,
	rng *rand.Rand,
) (GiantImpactOutcome, float64) {
	// 1. Check for miss
	if impactAngle > 85.0 {
		return OutcomeEscape, 0
	}

	massRatio := impactorMass / planetMass

	// Giant Impact Hypothesis (Theia):
	// Needs Mars-sized impactor (~0.1 Earth mass)
	// Grazing angle (~45 degrees)
	// Low velocity (near escape velocity)

	// Thresholds
	const (
		MoonFormationMassRatioMin = 0.05
		MoonFormationMassRatioMax = 0.20
		MoonFormationAngleMin     = 30.0
		MoonFormationAngleMax     = 60.0
		CaptureProbability        = 0.05
	)

	// Determine Outcome
	if massRatio >= MoonFormationMassRatioMin && massRatio <= MoonFormationMassRatioMax {
		if impactAngle >= MoonFormationAngleMin && impactAngle <= MoonFormationAngleMax {
			if velocityRatio < 1.2 {
				// Perfect conditions for Moon formation
				// Debris efficiency: how much of impactor becomes moon?
				// Typically 1-4% of impactor mass ends up as moon
				efficiency := 0.01 + rng.Float64()*0.03
				moonMass := impactorMass * efficiency
				return OutcomeMoonFormation, moonMass
			}
		}
	}

	// Capture logic (very rare for large bodies, but possible)
	if velocityRatio < 1.05 && impactAngle > 70 {
		if rng.Float64() < CaptureProbability {
			return OutcomeCaptured, impactorMass
		}
	}

	// Default: Direct Impact / Merger
	// The impactor merges with the planet
	return OutcomeDirectImpact, 0
}

// GenerateImpactor creates a random impactor for Hadean bombardment
func GenerateImpactor(planetMass float64, rng *rand.Rand) (mass, velocityRatio, angle float64) {
	// Power law distribution for mass
	// Most are small, few are large
	// Log-uniform distribution for simplicity in this scale
	// range: 1e-6 to 0.2 planet mass
	logMin := math.Log(1e-6 * planetMass)
	logMax := math.Log(0.2 * planetMass)
	mass = math.Exp(logMin + rng.Float64()*(logMax-logMin))

	// Velocity: 1.0 to 1.5 escape velocity
	velocityRatio = 1.0 + rng.Float64()*0.5

	// Angle: Area-weighted distribution (sin(angle))
	// Probability of impact at angle theta is proportional to sin(2*theta)
	// Simple approximation: 0-90 degrees
	angle = rng.Float64() * 90.0

	return
}
