// Package simulation provides stability constants for adaptive sub-stepping.
package simulation

// MaxGeologyStep is the maximum safe delta-time for geology simulation
// before tectonic movements lose precision or cause tunneling.
const MaxGeologyStep int64 = 100 // years

// MaxClimateStep is the maximum safe delta-time for climate simulation.
// Atmospheric changes need finer granularity.
const MaxClimateStep int64 = 50 // years

// MaxSubStep is the default engine sub-step size.
// This is the lowest common denominator of all subsystem step limits.
const MaxSubStep int64 = 50 // years
