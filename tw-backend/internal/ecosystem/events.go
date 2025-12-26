package ecosystem

import (
	"math"
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
	EventWarming          GeologicalEventType = "warming"          // Post-glacial warming
	EventGreenhouseSpike  GeologicalEventType = "greenhouse_spike" // CO2-driven warming
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
	ActiveEvents            []GeologicalEvent
	TectonicActivity        float64 // 0.0-1.0: represents geological instability (volcanism, earthquakes)
	GlobalTemperatureOffset float64 // Cumulative temperature offset from baseline
	RecentCoolingYears      int64   // Track how long world has been cooled
	ImpactShielding         float64 // From satellites (0.0-0.2): reduces asteroid impact probability
	rng                     *rand.Rand
}

func NewGeologicalEventManager() *GeologicalEventManager {
	return &GeologicalEventManager{
		ActiveEvents:            make([]GeologicalEvent, 0),
		TectonicActivity:        0.1, // Start with low baseline activity
		GlobalTemperatureOffset: 0,
		RecentCoolingYears:      0,
		ImpactShielding:         0.0, // Default to no moon shielding
		rng:                     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CheckForNewEvents probabilistically triggers new geological events based on time scale
// dt is the passed time in years
func (g *GeologicalEventManager) CheckForNewEvents(currentTick, dt int64) {
	// Tectonic activity decays slowly over time (half-life ~50k years)
	// Decay factor for dt years: (0.9999)^dt roughly
	// But let's keep it simple: approximate linear decay or use Pow
	// Old: g.TectonicActivity *= 0.9999 per CHECK (every 1000 years roughly)
	// So per year: 0.9999^(1/1000) ?? No, it was running every 365000 ticks = 1000 years.
	// So decay factor per 1000 years is 0.9999.
	// Decay per year is roughly 1.0 (very slow).
	// Let's just apply it proportionally to dt/1000.

	if dt >= 1000 {
		decaySteps := float64(dt) / 1000.0
		g.TectonicActivity *= math.Pow(0.9999, decaySteps)
	}

	// Minimum baseline activity
	if g.TectonicActivity < 0.05 {
		g.TectonicActivity = 0.05
	}

	dtFloat := float64(dt)

	// Helper to calculate P(at least one event) over dt years
	// baseProb is probability PER YEAR
	probabilityOverTime := func(baseProbPerYear float64) float64 {
		// P(event) = 1 - (1 - p)^t
		return 1.0 - math.Pow(1.0-baseProbPerYear, dtFloat)
	}

	// Volcanic winter: now tied to tectonic activity
	// Base chance ~1 per 10M years?
	// Previous code: 0.001% per 1000 years. = 0.00001 per 1000 years.
	// That's 1e-8 per year. Very rare.
	// Let's standardize to annual probabilities.
	// Old: (0.00001 + Tect*0.00014) per 1000 years.
	// Annual: Divide by 1000.
	baseVolcanic := (0.00001 + g.TectonicActivity*0.00014) / 1000.0
	if g.rng.Float64() < probabilityOverTime(baseVolcanic) {
		g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
			Type:           EventVolcanicWinter,
			StartTick:      currentTick,
			DurationTicks:  (1000 + g.rng.Int63n(2000)) * 365, // 10-30 years * 365 ticks/year
			Severity:       0.3 + g.rng.Float64()*0.5,
			TemperatureMod: -5 - g.rng.Float64()*10, // -5 to -15 degrees
			SunlightMod:    0.4 + g.rng.Float64()*0.3,
			OxygenMod:      1.0,
		})
	}

	// Asteroid impact: 0.005% per 1000 years. -> 5e-8 per year.
	baseAsteroid := 0.00005 / 1000.0
	// Apply shielding
	effectiveAsteroid := baseAsteroid * (1.0 - g.ImpactShielding)
	if g.rng.Float64() < probabilityOverTime(effectiveAsteroid) {
		g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
			Type:          EventAsteroidImpact,
			StartTick:     currentTick,
			DurationTicks: (5000 + g.rng.Int63n(10000)) * 365 / 365, // Old duration was in ticks?
			// The old code had: DurationTicks: 5000 + rand(10000).
			// Wait, the old code assumed ticks=days. 5000 ticks = 13 years.
			// My new ticks are standard.
			// I should keep the duration values consistent with "ticks".
			// If I pass currentTick which IS standardized, I should produce standardized duration.
			// 5000 ticks (old) @ 365/year -> 13 years.
			// 13 years * 365 = 4745 ticks.
			// So "5000" is roughly correct for ticks.
			Severity:       0.7 + g.rng.Float64()*0.3,
			TemperatureMod: -15 - g.rng.Float64()*20,
			SunlightMod:    0.1 + g.rng.Float64()*0.3,
			OxygenMod:      0.8,
		})
	}

	// Ocean anoxia: 0.005% per 1000 years
	baseAnoxia := 0.00005 / 1000.0
	if g.rng.Float64() < probabilityOverTime(baseAnoxia) {
		g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
			Type:           EventOceanAnoxia,
			StartTick:      currentTick,
			DurationTicks:  50000 + g.rng.Int63n(100000), // ~130-270 years
			Severity:       0.5 + g.rng.Float64()*0.5,
			TemperatureMod: 5 + g.rng.Float64()*10,
			SunlightMod:    1.0,
			OxygenMod:      0.3 + g.rng.Float64()*0.4,
		})
	}

	// Continental drift: 0.02% per 1000 years
	baseDrift := 0.0002 / 1000.0
	if g.rng.Float64() < probabilityOverTime(baseDrift) {
		severity := 0.3 + g.rng.Float64()*0.5
		g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
			Type:           EventContinentalDrift,
			StartTick:      currentTick,
			DurationTicks:  500000 + g.rng.Int63n(500000), // ~1300-2600 years
			Severity:       severity,
			TemperatureMod: 0,
			SunlightMod:    1.0,
			OxygenMod:      1.0,
		})
		g.TectonicActivity += severity * 0.2
		if g.TectonicActivity > 1.0 {
			g.TectonicActivity = 1.0
		}
	}

	// Flood basalt: 0.002% per 1000 years
	baseFlood := 0.00002 / 1000.0
	if g.rng.Float64() < probabilityOverTime(baseFlood) {
		severity := 0.6 + g.rng.Float64()*0.4
		g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
			Type:           EventFloodBasalt,
			StartTick:      currentTick,
			DurationTicks:  5000 + g.rng.Int63n(10000),
			Severity:       severity,
			TemperatureMod: -3 - g.rng.Float64()*5,
			SunlightMod:    0.7 + g.rng.Float64()*0.2,
			OxygenMod:      0.9,
		})
		// Greenhouse spike follows
		g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
			Type:           EventGreenhouseSpike,
			StartTick:      currentTick + 5000 + g.rng.Int63n(10000),
			DurationTicks:  50000 + g.rng.Int63n(100000),
			Severity:       severity * 0.8,
			TemperatureMod: 5 + g.rng.Float64()*10,
			SunlightMod:    1.0,
			OxygenMod:      0.95,
		})
		g.TectonicActivity += severity * 0.5
		if g.TectonicActivity > 1.0 {
			g.TectonicActivity = 1.0
		}
	}

	// Climate recovery
	g.updateClimateRecovery(currentTick, dt)
}

// updateClimateRecovery adds warming events to prevent permanent ice worlds
func (g *GeologicalEventManager) updateClimateRecovery(currentTick, chunkSize int64) {
	// Calculate current temperature offset from active events
	currentCooling := 0.0
	for _, e := range g.ActiveEvents {
		if currentTick >= e.StartTick && currentTick < e.StartTick+e.DurationTicks {
			currentCooling += e.TemperatureMod
		}
	}

	// Track cumulative cooling
	if currentCooling < -5 {
		g.RecentCoolingYears += chunkSize / 365 // Convert ticks to years
	} else if currentCooling > 0 {
		g.RecentCoolingYears = 0 // Reset if warming
	} else {
		// Gradual decay of cooling memory
		g.RecentCoolingYears -= chunkSize / 200
		if g.RecentCoolingYears < 0 {
			g.RecentCoolingYears = 0
		}
	}

	// If world has been cooling for >50k years, increase chance of warming event
	if g.RecentCoolingYears > 50000 {
		// Warming chance scales with how long they've been cold
		// Increased from ~10% max to ~30% max for faster climate recovery
		warmingChance := float64(g.RecentCoolingYears-50000) / 200000.0 // Up to ~30% after 200k years
		if warmingChance > 0.3 {
			warmingChance = 0.3
		}

		if g.rng.Float64() < warmingChance*(float64(chunkSize)/10000.0) {
			g.ActiveEvents = append(g.ActiveEvents, GeologicalEvent{
				Type:           EventWarming,
				StartTick:      currentTick,
				DurationTicks:  50000 + g.rng.Int63n(100000), // 500k-1.5M year warming period
				Severity:       0.4 + g.rng.Float64()*0.4,
				TemperatureMod: 8 + g.rng.Float64()*12, // +8 to +20 degrees
				SunlightMod:    1.0,
				OxygenMod:      1.0,
			})
			g.RecentCoolingYears = 0 // Reset counter
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

	// Cap temperature modifiers to prevent runaway cooling/heating
	// Earth-like worlds should stay within ±50°C of baseline
	if tempMod < -50 {
		tempMod = -50
	}
	if tempMod > 50 {
		tempMod = 50
	}

	return tempMod, sunlightMod, oxygenMod
}
