package ecosystem

import (
	"tw-backend/internal/worldgen/astronomy"
)

// IceAgeInsolationThreshold is the insolation value below which ice ages begin.
// When insolation drops below this threshold, conditions favor ice accumulation.
// Based on orbital calculations, insolation ranges from ~0.96 to ~1.04.
// Threshold tuned to 0.975 to target ~20-25% active duty cycle.
const IceAgeInsolationThreshold = 0.975

// IceAgeRecoveryThreshold is the insolation value above which ice ages end.
// Uses hysteresis to prevent rapid oscillation between states.
const IceAgeRecoveryThreshold = 0.990

// IceAgeDurationBase is the minimum duration of an ice age event in years.
const IceAgeDurationBase = 10000

// ClimateDriver connects orbital mechanics to climate events.
// It monitors the current orbital state and triggers ice ages deterministically
// based on insolation values rather than random number generation.
type ClimateDriver struct {
	// CurrentState holds the current orbital parameters
	CurrentState astronomy.OrbitalState

	// CurrentInsolation is the calculated solar energy factor (Orbital * Luminosity)
	CurrentInsolation float64

	// OrbitalInsolation is the normalized insolation from orbital mechanics (nominally ~1.0)
	// Triggers ice ages when finding Milankovitch lows, independent of solar luminosity
	OrbitalInsolation float64

	// IceAgeActive indicates whether an ice age is currently in progress
	IceAgeActive bool

	// IceAgeStartYear is when the current ice age began (0 if none)
	IceAgeStartYear int64

	// ObliquityStability is the satellite-derived stability factor (0.0-1.0)
	// 0.0 = chaotic (no moons), 1.0 = stable (Earth-Moon)
	// Affects the amplitude of obliquity oscillations
	ObliquityStability float64

	// BaseObliquity is the planetary average axial tilt in degrees (default 23.44)
	BaseObliquity float64

	// DayLengthSec is the length of a solar day in seconds (default 86400)
	// Affects diurnal temperature swings and coriolis forces
	DayLengthSec float64

	// GeothermalOffset is the temperature increase from planetary internal heat
	// High during early Earth (Hadean), approaches zero in modern era
	// Represents geothermal flux from mantle/core cooling
	GeothermalOffset float64

	// SolarLuminosity is the Sun's brightness relative to modern (0.7-1.0)
	// Faint Young Sun: Year 0 ≈ 0.71, Year 4.5B = 1.0
	// Based on Gough (1981) stellarevolution model
	SolarLuminosity float64

	// GreenhouseOffset is the temperature increase from atmospheric CO2
	// Set by atmospheric simulation, represents greenhouse warming
	// Early Earth: High CO2 → +50°C, Modern: Low CO2 → ~0°C
	GreenhouseOffset float64

	// GlobalAlbedo is the planetary reflectivity (0.0 = black body, 1.0 = perfect reflector)
	// Modern Earth ≈ 0.30, Snowball Earth ≈ 0.70
	GlobalAlbedo float64

	// IsSnowball indicates if the planet is in a global glaciation state
	// Triggers when global temp < SnowballThreshold and recovery when CO2 builds up
	IsSnowball bool

	// eventManager is the geological event system
	eventManager *GeologicalEventManager
}

// NewClimateDriver creates a climate driver connected to the event system.
func NewClimateDriver(eventManager *GeologicalEventManager) *ClimateDriver {
	return &ClimateDriver{
		CurrentState:       astronomy.CalculateOrbitalState(0),
		eventManager:       eventManager,
		IceAgeActive:       false,
		IceAgeStartYear:    0,
		ObliquityStability: 1.0,                         // Default to Earth-like stability
		BaseObliquity:      astronomy.ObliquityBaseline, // Default to Earth-like tilt
		DayLengthSec:       86400.0,                     // Default to Earth-like day (24h)
		GeothermalOffset:   0.0,                         // Will be calculated on first Update
		SolarLuminosity:    0.71,                        // Early Earth baseline
		OrbitalInsolation:  1.0,                         // Initial value
		GreenhouseOffset:   0.0,                         // Will be set by atmosphere
		GlobalAlbedo:       0.30,                        // Modern Earth baseline
		IsSnowball:         false,
	}
}

// Update checks the orbital state and triggers/ends ice ages based on insolation.
// This should be called periodically (e.g., every 100,000 simulation years).
// Uses ObliquityStability to scale obliquity variance for chaotic/stable worlds.
// Calculates GeothermalOffset from planetary age for early Earth heating.
// Calculates SolarLuminosity from stellar evolution (Faint Young Sun).
func (cd *ClimateDriver) Update(year int64) {
	// Calculate current orbital state with stability-adjusted obliquity
	cd.CurrentState = astronomy.CalculateOrbitalStateWithStability(year, cd.ObliquityStability, cd.BaseObliquity)
	baseInsolation := astronomy.CalculateInsolation(cd.CurrentState, cd.BaseObliquity)
	cd.OrbitalInsolation = baseInsolation

	// Calculate solar luminosity evolution (Faint Young Sun)
	// Year 0: ~71% modern brightness → Year 4.5B: 100% brightness (Gough 1981)
	cd.SolarLuminosity = astronomy.GetSolarLuminosity(year)

	// Apply solar luminosity to insolation
	// Effective insolation = orbital_insolation × solar_brightness
	cd.CurrentInsolation = baseInsolation * cd.SolarLuminosity

	// Calculate geothermal contribution from planetary internal heat
	// Uses the same thermal evolution model as geology
	heat := GetPlanetaryHeat(year)
	if heat > 2.0 {
		// Early Earth (Hadean/early Archean): significant geothermal heating
		// heat=10.0 → +90°C, heat=4.0 → +30°C, heat=2.0 → +10°C
		// This is geothermal flux reaching the surface from thin/fractured crust
		cd.GeothermalOffset = (heat - 1.0) * 10.0
	} else {
		// Transitional to modern Earth: gradually declining geothermal contribution
		// heat=2.0 → +2°C, heat=1.0 → +0°C
		cd.GeothermalOffset = (heat - 1.0) * 2.0
	}

	// Check for ice age transitions using hysteresis
	// IMPORTANT: Skip ice age logic during early Earth (heat > 2.0)
	// The Faint Young Sun paradox is resolved by high geothermal + greenhouse warming,
	// which kept early Earth warm despite low solar luminosity. Ice ages only became
	// possible after the planet cooled (~2-3 billion years in).
	if heat <= 2.0 {
		// Check 'OrbitalInsolation' instead of 'CurrentInsolation' to trigger based on
		// Milankovitch cycles rather than absolute solar brightness. This assumes the
		// Carbon Cycle compensates for the long-term trend of the Faint Young Sun.
		if !cd.IceAgeActive && cd.OrbitalInsolation < IceAgeInsolationThreshold {
			cd.startIceAge(year)
		} else if cd.IceAgeActive && cd.OrbitalInsolation > IceAgeRecoveryThreshold {
			// Only end if minimum duration has passed
			if year-cd.IceAgeStartYear >= IceAgeDurationBase {
				cd.endIceAge(year)
			}
		}
	}

	// =========================================================================
	// Phase 9b: SNOWBALL EARTH (Global Glaciation / Ice-Albedo Feedback)
	// =========================================================================
	// Calculate approximate global temperature for Snowball trigger.
	// This is a simplified energy balance: Solar * (1 - Albedo) + Greenhouse + Geothermal.
	// We don't recalculate from scratch; we use offsets.
	//
	// Base temp at Solar Constant = 1.0, Albedo = 0.30 is ~14°C (Modern Earth).
	// Offset from solar dimming: (Luminosity - 1.0) * 70 (rough sensitivity).
	// Then add GreenhouseOffset and GeothermalOffset.
	solarTempDelta := (cd.SolarLuminosity - 1.0) * 70.0

	// Albedo feedback: Higher albedo = colder.
	// Modern albedo 0.30 is baseline. Snowball at 0.70.
	// Each 0.1 increase in albedo from 0.30 drops temp by ~10°C.
	albedoTempDelta := (0.30 - cd.GlobalAlbedo) * 100.0 // 0.30 -> 0, 0.70 -> -40

	globalTemp := 14.0 + solarTempDelta + albedoTempDelta + cd.GreenhouseOffset + cd.GeothermalOffset

	// Snowball Trigger: If temp drops below -10°C globally, ice covers oceans.
	// This is a runaway feedback: more ice -> higher albedo -> colder -> more ice.
	const (
		SnowballTriggerTemp  = -10.0 // °C
		SnowballRecoveryTemp = 10.0  // °C (requires massive greenhouse buildup)
		SnowballAlbedo       = 0.70  // Ice-covered planet
		ModernAlbedo         = 0.30  // Normal Earth
	)

	if !cd.IsSnowball && globalTemp < SnowballTriggerTemp {
		cd.IsSnowball = true
		cd.GlobalAlbedo = SnowballAlbedo

		if cd.eventManager != nil {
			cd.eventManager.ActiveEvents = append(cd.eventManager.ActiveEvents, GeologicalEvent{
				Type:           EventGlobalGlaciation,
				StartTick:      year * 365,
				DurationTicks:  1_000_000 * 365, // Long duration
				Severity:       1.0,
				TemperatureMod: -40,
				SunlightMod:    0.5, // Ice reflects
				OxygenMod:      0.9, // Reduced biosphere
			})
		}
	} else if cd.IsSnowball && globalTemp > SnowballRecoveryTemp {
		// Recovery: Volcanic CO2 buildup has warmed the planet enough to melt ice.
		cd.IsSnowball = false
		cd.GlobalAlbedo = ModernAlbedo

		// End the glaciation event (truncate to current year)
		if cd.eventManager != nil {
			for i := range cd.eventManager.ActiveEvents {
				if cd.eventManager.ActiveEvents[i].Type == EventGlobalGlaciation {
					cd.eventManager.ActiveEvents[i].DurationTicks = year*365 - cd.eventManager.ActiveEvents[i].StartTick
				}
			}
		}
	}
}

// startIceAge triggers a new ice age event through the event manager.
func (cd *ClimateDriver) startIceAge(year int64) {
	cd.IceAgeActive = true
	cd.IceAgeStartYear = year

	if cd.eventManager == nil {
		return
	}

	// Calculate severity based on how far below threshold we are (using OrbitalInsolation)
	// Insolation of 0.95 → severity 1.0, insolation of 0.98 → severity 0.3
	severity := (IceAgeInsolationThreshold - cd.OrbitalInsolation) / 0.03
	if severity > 1.0 {
		severity = 1.0
	}
	if severity < 0.3 {
		severity = 0.3
	}

	// Create deterministic ice age event
	// Duration based on orbital mechanics: typically 10k-50k years
	// We estimate duration based on insolation cycle
	iceAgeEvent := GeologicalEvent{
		Type:           EventIceAge,
		StartTick:      year * 365, // Convert years to ticks (365 ticks/year)
		DurationTicks:  100000,     // Initial estimate, will be extended by orbital state
		Severity:       severity,
		TemperatureMod: -8 - severity*12, // -8 to -20 degrees
		SunlightMod:    0.9,
		OxygenMod:      1.0,
	}

	cd.eventManager.ActiveEvents = append(cd.eventManager.ActiveEvents, iceAgeEvent)
}

// endIceAge terminates the current ice age event.
func (cd *ClimateDriver) endIceAge(year int64) {
	cd.IceAgeActive = false
	cd.IceAgeStartYear = 0

	if cd.eventManager == nil {
		return
	}

	// Remove ice age events from active events
	// Note: We update end time rather than removing to maintain history
	filtered := make([]GeologicalEvent, 0, len(cd.eventManager.ActiveEvents))
	for _, e := range cd.eventManager.ActiveEvents {
		if e.Type == EventIceAge {
			// Truncate this ice age event to current year
			e.DurationTicks = year*365 - e.StartTick
			if e.DurationTicks > 0 {
				filtered = append(filtered, e)
			}
			// Skip adding it back since it's now ended
			continue
		}
		filtered = append(filtered, e)
	}
	cd.eventManager.ActiveEvents = filtered
}

// GetObliquity returns the current axial tilt for weather calculations.
// This allows the weather system to use dynamic orbital parameters.
func (cd *ClimateDriver) GetObliquity() float64 {
	return cd.CurrentState.Obliquity
}

// GetInsolation returns the current normalized solar energy factor.
func (cd *ClimateDriver) GetInsolation() float64 {
	return cd.CurrentInsolation
}

// GetOrbitalState returns the full current orbital state.
func (cd *ClimateDriver) GetOrbitalState() astronomy.OrbitalState {
	return cd.CurrentState
}

// IsIceAge returns true if an ice age is currently active.
func (cd *ClimateDriver) IsIceAge() bool {
	return cd.IceAgeActive
}

// GetGeothermalOffset returns the current temperature offset from internal planetary heat.
// This value decreases from ~90°C in early Earth to ~0°C in modern Earth.
func (cd *ClimateDriver) GetGeothermalOffset() float64 {
	return cd.GeothermalOffset
}

// SetGreenhouseOffset sets the temperature offset from atmospheric greenhouse gases.
// This is updated by the atmospheric simulation based on CO2 levels.
func (cd *ClimateDriver) SetGreenhouseOffset(offset float64) {
	cd.GreenhouseOffset = offset
}

// GetGreenhouseOffset returns the current temperature offset from atmospheric CO2.
// Early Earth: +50°C from high CO2, Modern Earth: ~0°C from trace CO2.
func (cd *ClimateDriver) GetGreenhouseOffset() float64 {
	return cd.GreenhouseOffset
}

// GetSolarLuminosity returns the current solar brightness relative to modern (0.7-1.0).
// Year 0: ~0.71 (Faint Young Sun), Year 4.5B: 1.0 (Modern Sun)
func (cd *ClimateDriver) GetSolarLuminosity() float64 {
	return cd.SolarLuminosity
}
