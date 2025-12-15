package ecosystem

import (
	"math/rand"
	"time"
)

// GeologicalEventType represents different geological/climate events
type GeologicalEventType string

const (
	EventVolcanicWinter   GeologicalEventType = "volcanic_winter"
	EventAsteroidImpact   GeologicalEventType = "asteroid_impact"
	EventIceAge           GeologicalEventType = "ice_age"
	EventOceanAnoxia      GeologicalEventType = "ocean_anoxia"
	EventContinentalDrift GeologicalEventType = "continental_drift"
	EventFloodBasalt      GeologicalEventType = "flood_basalt"
)

// GeologicalEvent represents an active environmental event
type GeologicalEvent struct {
	Type           GeologicalEventType
	StartTick      int64
	DurationTicks  int64
	Severity       float64 // 0.0-1.0
	TemperatureMod float64 // degrees offset
	SunlightMod    float64 // multiplier (0.0-1.0)
	OxygenMod      float64 // multiplier (0.0-1.0)
}

// GeologicalEventManager handles long-term geological events
type GeologicalEventManager struct {
	ActiveEvents     []GeologicalEvent
	TectonicActivity float64 // 0.0-1.0: represents geological instability (volcanism, earthquakes)
	rng              *rand.Rand
}

func NewGeologicalEventManager() *GeologicalEventManager {
	return &GeologicalEventManager{
		ActiveEvents:     make([]GeologicalEvent, 0),
		TectonicActivity: 0.1, // Start with low baseline activity
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CheckForNewEvents probabilistically triggers new geological events based on time scale
// ticksElapsed is the number of ticks since last check
// At 100 ticks = 1 year, major events happen roughly every 100k-1M years
func (g *GeologicalEventManager) CheckForNewEvents(currentTick, ticksElapsed int64) {
	// Tectonic activity decays slowly over time (half-life ~50k years)
	g.TectonicActivity *= 0.9999

	// Minimum baseline activity
	if g.TectonicActivity < 0.05 {
		g.TectonicActivity = 0.05
	}

	// Scale probability by ticks elapsed
	for i := int64(0); i < ticksElapsed; i += 10000 {
		chunkSize := int64(10000)
		if i+chunkSize > ticksElapsed {
			chunkSize = ticksElapsed - i
		}

		// Volcanic winter: now tied to tectonic activity
		// Low base chance (0.2%) but scales up to 3% with high tectonic activity
		volcanicChance := (0.002 + g.TectonicActivity*0.028) * (float64(chunkSize) / 10000.0)
		if g.rng.Float64() < volcanicChance {
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventVolcanicWinter,
				StartTick:      currentTick,
				DurationTicks:  1000 + g.rng.Int63n(2000), // 10-30 years
				Severity:       0.3 + g.rng.Float64()*0.5,
				TemperatureMod: -5 - g.rng.Float64()*10, // -5 to -15 degrees
				SunlightMod:    0.4 + g.rng.Float64()*0.3,
				OxygenMod:      1.0,
			})
		}

		// Asteroid impact: 0.1% per 10k ticks
		if g.rng.Float64() < 0.001*(float64(chunkSize)/10000.0) {
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventAsteroidImpact,
				StartTick:      currentTick,
				DurationTicks:  5000 + g.rng.Int63n(10000), // 50-150 years
				Severity:       0.7 + g.rng.Float64()*0.3,
				TemperatureMod: -15 - g.rng.Float64()*20, // -15 to -35 degrees
				SunlightMod:    0.1 + g.rng.Float64()*0.3,
				OxygenMod:      0.8,
			})
		}

		// Ice age: 0.05% per 10k ticks
		if g.rng.Float64() < 0.0005*(float64(chunkSize)/10000.0) {
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventIceAge,
				StartTick:      currentTick,
				DurationTicks:  100000 + g.rng.Int63n(200000), // 1M-3M years
				Severity:       0.4 + g.rng.Float64()*0.4,
				TemperatureMod: -8 - g.rng.Float64()*12, // -8 to -20 degrees
				SunlightMod:    0.9,
				OxygenMod:      1.0,
			})
		}

		// Ocean anoxia: 0.02% per 10k ticks
		if g.rng.Float64() < 0.0002*(float64(chunkSize)/10000.0) {
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventOceanAnoxia,
				StartTick:      currentTick,
				DurationTicks:  50000 + g.rng.Int63n(100000), // 500k-1.5M years
				Severity:       0.5 + g.rng.Float64()*0.5,
				TemperatureMod: 5 + g.rng.Float64()*10, // Warmer
				SunlightMod:    1.0,
				OxygenMod:      0.3 + g.rng.Float64()*0.4, // 30-70% oxygen
			})
		}

		// Continental drift: 0.03% per 10k ticks (happens over long timescales)
		// Drift events increase tectonic activity as plates move and create volcanic zones
		if g.rng.Float64() < 0.0003*(float64(chunkSize)/10000.0) {
			severity := 0.3 + g.rng.Float64()*0.5
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventContinentalDrift,
				StartTick:      currentTick,
				DurationTicks:  500000 + g.rng.Int63n(500000), // 5-10M years
				Severity:       severity,
				TemperatureMod: 0, // No direct temperature effect
				SunlightMod:    1.0,
				OxygenMod:      1.0,
			})
			// Continental drift creates subduction zones and volcanic arcs
			g.TectonicActivity += severity * 0.2
			if g.TectonicActivity > 1.0 {
				g.TectonicActivity = 1.0
			}
		}

		// Flood basalt: 0.01% per 10k ticks (rare but impactful)
		// Major volcanic event that significantly increases tectonic activity
		if g.rng.Float64() < 0.0001*(float64(chunkSize)/10000.0) {
			severity := 0.6 + g.rng.Float64()*0.4
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventFloodBasalt,
				StartTick:      currentTick,
				DurationTicks:  10000 + g.rng.Int63n(20000), // 100-300k years
				Severity:       severity,
				TemperatureMod: -3 - g.rng.Float64()*7, // -3 to -10 degrees (volcanic gases)
				SunlightMod:    0.7 + g.rng.Float64()*0.2,
				OxygenMod:      0.9,
			})
			// Flood basalts are massive volcanic events - major activity spike
			g.TectonicActivity += severity * 0.5
			if g.TectonicActivity > 1.0 {
				g.TectonicActivity = 1.0
			}
		}
	}
}

// UpdateActiveEvents removes expired events
func (g *GeologicalEventManager) UpdateActiveEvents(currentTick int64) {
	active := make([]GeologicalEvent, 0)
	for _, e := range g.ActiveEvents {
		if currentTick < e.StartTick+e.DurationTicks {
			active = append(active, e)
		}
	}
	g.ActiveEvents = active
}

// GetEnvironmentModifiers returns combined modifiers from all active events
func (g *GeologicalEventManager) GetEnvironmentModifiers() (tempMod, sunlightMod, oxygenMod float64) {
	tempMod = 0
	sunlightMod = 1.0
	oxygenMod = 1.0

	for _, e := range g.ActiveEvents {
		tempMod += e.TemperatureMod
		sunlightMod *= e.SunlightMod
		oxygenMod *= e.OxygenMod
	}

	return tempMod, sunlightMod, oxygenMod
}
