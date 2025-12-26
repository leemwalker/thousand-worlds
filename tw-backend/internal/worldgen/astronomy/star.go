package astronomy

// GetSolarLuminosity returns solar luminosity relative to modern Earth (1.0)
// Based on Gough (1981) stellar evolution model
//
// The Sun has been gradually brightening over its lifetime due to hydrogen fusion
// increasing helium concentration in the core, raising core temperature and pressure.
//
// Timeline:
//   - Year 0 (4.5B years ago, Hadean): L ≈ 0.714 (~71% modern brightness)
//   - Year 2.25B (mid-history): L ≈ 0.833
//   - Year 4.5B (today, Modern): L = 1.0 (100% modern brightness)
//
// Formula (Gough 1981): L(t) = 1.0 / (1.0 + 0.4 * (1.0 - t/t_now))
// Where:
//   - t = time since solar system formation
//   - t_now = current age of solar system (4.5 billion years)
//   - 0.4 = empirical constant from stellar evolution models
//
// This is the foundation for solving the "Faint Young Sun Paradox":
// Earth needed a strong greenhouse effect (high CO2) in early history
// to maintain liquid water despite reduced solar output.
func GetSolarLuminosity(year int64) float64 {
	const (
		solarAge      = 4_500_000_000 // years - current age of solar system
		dimFactor     = 0.4           // Gough constant from stellar models
		minBrightness = 0.7           // Physical minimum (~70% at formation)
		maxBrightness = 1.0           // Current brightness
	)

	// Clamp to valid range
	if year < 0 {
		year = 0
	}
	if year > solarAge {
		year = solarAge
	}

	// Normalized time (0.0 at formation, 1.0 at present)
	t_norm := float64(year) / float64(solarAge)

	// Gough (1981) formula
	// L(t) = 1 / (1 + 0.4 * (1 - t/t_now))
	luminosity := 1.0 / (1.0 + dimFactor*(1.0-t_norm))

	// Safety clamps (shouldn't be needed but ensures physical bounds)
	if luminosity < minBrightness {
		luminosity = minBrightness
	}
	if luminosity > maxBrightness {
		luminosity = maxBrightness
	}

	return luminosity
}
