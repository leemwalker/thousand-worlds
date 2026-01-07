package geography

import (
	"math"
)

// PlanetaryCore defines the fundamental physical properties of the planet.
// These properties influence gravity, tectonics, atmosphere, and magnetosphere.
type PlanetaryCore struct {
	Mass          float64 // In Earth Masses (M⊕)
	Radius        float64 // In Earth Radii (R⊕)
	Density       float64 // Average density in g/cm³
	Gravity       float64 // Surface gravity in G (1.0 = Earth)
	CoreHeat      float64 // Internal heat flux factor (1.0 = Earth current)
	MagneticField float64 // Magnetic field strength in Gauss (approx)
	Age           float64 // Planet age in Billion Years (Ga)
}

// NewPlanetaryCore creates a planet with physics based on mass and age.
// Uses scaling laws for rocky planets.
func NewPlanetaryCore(mass float64, age float64) *PlanetaryCore {
	// Mass-Radius relationship for rocky planets: R ~ M^0.27
	radius := math.Pow(mass, 0.27)

	// Gravity: g = M / R^2
	gravity := mass / (radius * radius)

	// Average Density (relative to Earth). D ~ M / R^3
	// Earth density ~ 5.51 g/cm3
	// We store absolute? Or relative? Struct says g/cm3.
	// Vol ~ R^3. M ~ Vol * Density.
	// If R ~ M^0.27, R^3 ~ M^0.81.
	// Density ~ M / M^0.81 ~ M^0.19.
	// Larger planets are denser (compression).
	density := 5.51 * math.Pow(mass, 0.19)

	// Core Heat (Radiogenic + Primordial)
	// Decays with time. H ~ 1/Age?
	// Simplified: Larger planets retain heat longer.
	// Heat ~ (Mass / Radius^2) * e^(-Age/HalfLife)
	// Let's use simplified factor relative to Earth (Age 4.5).
	// Earth heat flux ~ 1.0.
	// Hadean (Age 0.5) -> Heat ~ 3-4x.
	decayFactor := math.Exp(-(age - 4.5) / 3.0) // 3 billion year cooling scale
	coreHeat := mass * decayFactor              // More mass = more fuel/retention

	// Magnetic Field (Dynamo)
	// Requires Rotation + Convection (Heat) + Conductive Core.
	// Simplified: proportional to CoreHeat and Rotation (assumed Earth-like).
	magneticField := 0.5 * coreHeat // Gauss. Earth ~ 0.5.

	return &PlanetaryCore{
		Mass:          mass,
		Radius:        radius,
		Density:       density,
		Gravity:       gravity,
		CoreHeat:      coreHeat,
		MagneticField: magneticField,
		Age:           age,
	}
}

// GetMaxElevation returns the maximum stable mountain height.
// Limit is inversely proportional to gravity.
// Earth max ~10-12km. Mars (0.38g) ~25km.
func (c *PlanetaryCore) GetMaxElevation() float64 {
	baseMax := 12000.0 // meters on Earth
	return baseMax / c.Gravity
}
