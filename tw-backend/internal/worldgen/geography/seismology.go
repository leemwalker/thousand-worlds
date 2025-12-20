package geography

import "math"

// CalculateSeismicActivity returns the expected seismic profile for a boundary type
// This is a deterministic simulation based on geological principles
func CalculateSeismicActivity(boundaryType BoundaryType) SeismicEvent {
	event := SeismicEvent{
		BoundaryType: boundaryType,
	}

	switch string(boundaryType) {
	case "divergent_ridge":
		// Shallow earthquakes, moderate magnitude
		event.Depth = "Shallow"
		event.Magnitude = 6.5
	case "transform_fault":
		// Shallow earthquakes, high magnitude
		event.Depth = "Shallow"
		event.Magnitude = 8.0
	case "subduction_zone":
		// Deep earthquakes (Benioff zone), very high magnitude
		event.Depth = "Deep"
		event.Magnitude = 9.5
	case "continental_collision":
		// Intermediate depth, high magnitude
		event.Depth = "Intermediate"
		event.Magnitude = 8.5
	default:
		event.Depth = "Shallow"
		event.Magnitude = 5.0
	}

	return event
}

// GenerateTsunami checks if an event generates a tsunami and calculates its properties
func GenerateTsunami(event SeismicEvent, waterDepth float64) *Tsunami {
	// Tsunamis require:
	// 1. High magnitude (> 7.5)
	// 2. Underwater location (waterDepth > 0)
	// 3. Vertical displacement (usually subduction/convergent)

	if event.Magnitude < 7.5 || waterDepth <= 0 {
		return nil
	}

	// Calculate initial wave height based on magnitude
	// Simplified formula: (Mag - 6) * 1.5
	waveHeight := (event.Magnitude - 6.0) * 1.5

	// Calculate velocity: sqrt(g * depth) * 3.6 for km/h
	// g ~ 9.8 m/s^2
	velocity := math.Sqrt(9.8*waterDepth) * 3.6

	return &Tsunami{
		OriginLocation:    event.Epicenter,
		InitialWaveHeight: waveHeight,
		TravelVelocity:    velocity,
		AffectedCoasts:    []Point{{X: 0, Y: 0}}, // Stub for now, would need map context to find coasts
	}
}
