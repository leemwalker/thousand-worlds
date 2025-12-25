// Package astronomy provides orbital mechanics and satellite calculations
// for world generation, including natural satellite generation and their
// physical effects on planetary systems.
package astronomy

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// Physical constants
const (
	// GravitationalConstant is Newton's gravitational constant (m³/(kg·s²))
	GravitationalConstant = 6.674e-11

	// EarthRadiusMeters is Earth's mean radius in meters
	EarthRadiusMeters = 6.371e6

	// EarthMassKg is Earth's mass in kilograms
	EarthMassKg = 5.972e24

	// MoonMassKg is Earth's Moon mass in kilograms
	MoonMassKg = 7.342e22

	// MoonDistanceMeters is Earth-Moon distance in meters
	MoonDistanceMeters = 384400e3

	// MoonRadiusMeters is Earth's Moon radius in meters
	MoonRadiusMeters = 1.7374e6

	// RocheLimitFactor is the multiplier for planet radius to get Roche limit
	// Below this distance, tidal forces tear satellites apart
	RocheLimitFactor = 2.5

	// HillSphereLimit is the maximum stable orbit distance in meters
	// Approximately 1.5 million km for Earth-like planets
	HillSphereLimit = 1.5e9
)

// SatelliteConfig controls natural satellite generation
type SatelliteConfig struct {
	// Override indicates whether to use Count instead of random generation
	Override bool
	// Count is the exact number of moons when Override is true
	Count int
}

// Satellite represents a natural satellite (moon) orbiting a planet
type Satellite struct {
	// ID is a unique identifier for this satellite
	ID uuid.UUID
	// Name is the display name for this satellite
	Name string
	// Mass in kilograms (Earth's Moon: 7.342e22 kg)
	Mass float64
	// Distance from planet center in meters (Moon: 384,400 km)
	Distance float64
	// Period is the orbital period in seconds (Moon: ~27.3 days)
	Period float64
	// Radius in meters (Moon: 1,737.4 km)
	Radius float64
}

// PlanetarySystem holds the planet and its natural satellites
type PlanetarySystem struct {
	// PlanetMass in kilograms (Earth: 5.972e24 kg)
	PlanetMass float64
	// PlanetRadius in meters (Earth: 6.371e6 m)
	PlanetRadius float64
	// Satellites is the list of natural satellites
	Satellites []Satellite
	// Seed used for generation
	Seed int64
}

// GenerateMoons creates natural satellites based on configuration.
//
// If config.Override is true, exactly config.Count moons are generated.
// Otherwise, the count is determined randomly:
//   - 10% chance: 0 moons
//   - 60% chance: 1 moon
//   - 20% chance: 2 moons
//   - 10% chance: 3+ moons (up to 5)
//
// All generated satellites satisfy orbital constraints:
//   - Distance > Roche Limit (2.5 × planet radius)
//   - Distance < Hill Sphere (~1.5 billion meters)
func GenerateMoons(seed int64, planetMass float64, config SatelliteConfig) []Satellite {
	rng := rand.New(rand.NewSource(seed))

	// Determine moon count
	var count int
	if config.Override {
		count = config.Count
	} else {
		count = randomMoonCount(rng)
	}

	if count == 0 {
		return []Satellite{}
	}

	// Calculate orbital constraints
	rocheLimit := RocheLimitFactor * EarthRadiusMeters // Use Earth radius as baseline
	maxDistance := HillSphereLimit
	orbitalRange := maxDistance - rocheLimit

	satellites := make([]Satellite, count)
	for i := 0; i < count; i++ {
		// Distribute moons across the orbital range
		// Use stratified sampling to avoid overlapping orbits
		segmentStart := rocheLimit + (orbitalRange*float64(i))/float64(count)
		segmentEnd := rocheLimit + (orbitalRange*float64(i+1))/float64(count)
		distance := segmentStart + rng.Float64()*(segmentEnd-segmentStart)

		// Generate mass relative to Earth's Moon (0.1x to 2.0x)
		massMultiplier := 0.1 + rng.Float64()*1.9
		mass := MoonMassKg * massMultiplier

		// Radius scales with cube root of mass (assuming constant density)
		// radius ∝ mass^(1/3)
		radiusMultiplier := math.Pow(massMultiplier, 1.0/3.0)
		radius := MoonRadiusMeters * radiusMultiplier

		// Calculate orbital period using Kepler's 3rd Law: T = 2π√(a³/GM)
		period := 2 * math.Pi * math.Sqrt(math.Pow(distance, 3)/(GravitationalConstant*planetMass))

		satellites[i] = Satellite{
			ID:       uuid.New(),
			Name:     generateMoonName(i, seed),
			Mass:     mass,
			Distance: distance,
			Period:   period,
			Radius:   radius,
		}
	}

	return satellites
}

// randomMoonCount determines moon count based on probability distribution:
// 10% → 0, 60% → 1, 20% → 2, 10% → 3+
func randomMoonCount(rng *rand.Rand) int {
	roll := rng.Float64()

	switch {
	case roll < 0.10:
		return 0
	case roll < 0.70: // 0.10 + 0.60
		return 1
	case roll < 0.90: // 0.70 + 0.20
		return 2
	default:
		// 3-5 moons for the remaining 10%
		return 3 + rng.Intn(3)
	}
}

// generateMoonName creates a name for a moon based on its index
func generateMoonName(index int, seed int64) string {
	// Roman numeral-style naming (I, II, III, IV, V...)
	numerals := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X"}
	if index < len(numerals) {
		return "Moon " + numerals[index]
	}
	return "Moon " + string(rune('A'+index))
}

// NewPlanetarySystem creates a new planetary system with the given parameters
func NewPlanetarySystem(planetMass, planetRadius float64, seed int64) *PlanetarySystem {
	return &PlanetarySystem{
		PlanetMass:   planetMass,
		PlanetRadius: planetRadius,
		Satellites:   []Satellite{},
		Seed:         seed,
	}
}

// GenerateSatellites populates the planetary system with moons
func (ps *PlanetarySystem) GenerateSatellites(config SatelliteConfig) {
	ps.Satellites = GenerateMoons(ps.Seed, ps.PlanetMass, config)
}

// TotalMoonMass returns the combined mass of all satellites
func (ps *PlanetarySystem) TotalMoonMass() float64 {
	var total float64
	for _, sat := range ps.Satellites {
		total += sat.Mass
	}
	return total
}

// MoonCount returns the number of satellites
func (ps *PlanetarySystem) MoonCount() int {
	return len(ps.Satellites)
}
