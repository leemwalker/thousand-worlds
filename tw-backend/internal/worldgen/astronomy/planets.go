package astronomy

import "math"

// Standard astronomical constants
const (
	EarthMassKg           = 5.972e24
	EarthRadiusMeters     = 6.371e6
	EarthDaySeconds       = 86400.0
	EarthAxialTilt        = 23.5 // degrees
	EarthGravity          = 9.81
	GravitationalConstant = 6.67430e-11 // m³/(kg·s²)
)

// PlanetParameters defines the physical characteristics of a planet.
type PlanetParameters struct {
	MassKg       float64
	RadiusM      float64
	DayLengthSec float64
	AxialTiltDeg float64
}

// NewEarthParams returns parameters for an Earth-like planet.
func NewEarthParams() PlanetParameters {
	return PlanetParameters{
		MassKg:       EarthMassKg,
		RadiusM:      EarthRadiusMeters,
		DayLengthSec: EarthDaySeconds,
		AxialTiltDeg: EarthAxialTilt,
	}
}

// SurfaceGravity calculates the surface gravity in m/s^2.
// g = G * M / r^2
func (p PlanetParameters) SurfaceGravity() float64 {
	if p.RadiusM == 0 {
		return 0
	}
	return (GravitationalConstant * p.MassKg) / (p.RadiusM * p.RadiusM)
}

// EscapeVelocity calculates the escape velocity in m/s.
// ve = sqrt(2 * G * M / r)
func (p PlanetParameters) EscapeVelocity() float64 {
	if p.RadiusM == 0 {
		return 0
	}
	return math.Sqrt((2 * GravitationalConstant * p.MassKg) / p.RadiusM)
}
