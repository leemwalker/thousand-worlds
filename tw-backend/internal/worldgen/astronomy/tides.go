package astronomy

import (
	"math"
)

// CalculateTidalRange computes the tidal range (height difference between high and low tide)
// based on the gravitational influence of the planet's satellites and star.
//
// Physics:
// Tidal potential V ∝ M / d³
// We normalize to Earth-Moon system where:
// - Moon Factor ≈ 1.0 (generates ~50-100cm open ocean tides)
// - Sun Factor ≈ 0.46 (add/subtract for spring/neap tides)
//
// The returned value is in meters, representing the maximum spring tide range
// (Sun + Moons aligned) to be used for intertidal zone sizing.
func CalculateTidalRange(satellites []Satellite, planetMass, planetRadius float64) float64 {
	// Baseline tidal range from solar tides (Earth's solar tides are ~1/2 lunar)
	// If planet has no moons, it still has solar tides.
	// Assuming Earth-like distance to star, we start with 0.5m base range.
	totalTidalForce := 0.5 // Solar contribution baseline

	// Earth-Moon reference values for normalization
	const (
		earthMoonMass     = 7.342e22 // kg
		earthMoonDistance = 3.844e8  // meters
	)

	// Calculate reference tidal potential standard (M/d³)
	referencePotential := earthMoonMass / math.Pow(earthMoonDistance, 3)

	for _, sat := range satellites {
		// Calculate tidal potential for this satellite -> M / d³
		potential := sat.Mass / math.Pow(sat.Distance, 3)

		// Normalize against Earth-Moon system
		relativeForce := potential / referencePotential

		// Add to total force (assuming alignment for maximum spring tide)
		totalTidalForce += relativeForce
	}

	// Convert force factor to range in meters
	// Earth average coastal tidal range is ~2m, driven by 1.0 (Moon) + 0.46 (Sun) ≈ 1.5 force
	// So multiplier is approx 1.33 x Force
	tidalRange := totalTidalForce * 1.5

	// Cap at reasonable limits (e.g. 20m for extreme scenarios)
	if tidalRange > 20.0 {
		tidalRange = 20.0
	}

	return tidalRange
}
