// Package calibration provides geophysical validation tools for world simulation.
// It compares simulated worlds against Earth benchmarks to ensure realistic outputs.
package calibration

import (
	"fmt"
	"math"
)

// =============================================================================
// Earth Benchmark Constants
// =============================================================================

// EarthBenchmarks contains target metrics based on real Earth data.
// Used for validating simulation outputs against known geophysical constants.
//
// Sources:
//   - Hypsometry: NOAA, USGS topographic data
//   - Climate: IPCC, NASA GISS
//   - Geology: Plate tectonics research, geological surveys
type EarthBenchmarks struct {
	// Hypsometry (Elevation Distribution)
	OceanCoveragePercent float64 // ~71% of Earth's surface
	MeanOceanDepthM      float64 // ~-3,700m below sea level
	MeanLandHeightM      float64 // ~+840m above sea level
	BimodalPeakOceanM    float64 // Ocean histogram peak: ~-4,500m
	BimodalPeakLandM     float64 // Land histogram peak: ~+500m

	// Climate
	GlobalMeanTempC    float64 // ~15°C annual average
	EquatorToPoleGradC float64 // ~45°C difference (equator ~27°C, poles ~-18°C)

	// Hydrology
	RiverDensityPercent float64 // % of land cells with significant flux

	// Geology
	PlateCount       int // Major tectonic plates: ~7-8
	SubplateCount    int // Minor plates + microplates: ~50+
	ContinentCount   int // Major continents: ~7
	HotspotCount     int // Active volcanic hotspots: ~40-50
	OceanTrenchCount int // Major ocean trenches: ~20+

	// Astronomy
	MoonCount int // Natural satellites: 1 (Luna)
}

// DefaultEarthBenchmarks returns the standard Modern Earth target values.
func DefaultEarthBenchmarks() EarthBenchmarks {
	return EarthBenchmarks{
		// Hypsometry
		OceanCoveragePercent: 71.0,
		MeanOceanDepthM:      -3700.0,
		MeanLandHeightM:      840.0,
		BimodalPeakOceanM:    -4500.0,
		BimodalPeakLandM:     500.0,

		// Climate
		GlobalMeanTempC:    15.0,
		EquatorToPoleGradC: 45.0,

		// Hydrology
		RiverDensityPercent: 10.0, // ~10% of land has significant drainage

		// Geology
		PlateCount:       8,  // Major plates
		SubplateCount:    50, // Microplates and sub-regions
		ContinentCount:   7,  // Major continents
		HotspotCount:     40, // Volcanic hotspots
		OceanTrenchCount: 20, // Subduction trenches

		// Astronomy
		MoonCount: 1,
	}
}

// HadeanEarthBenchmarks returns target values for the Hadean Eon (4.6-4.0 Ga).
// Based on "Water World" hypothesis and hotter early Earth conditions.
func HadeanEarthBenchmarks() EarthBenchmarks {
	return EarthBenchmarks{
		// Hypsometry: "Water World" with island arcs
		OceanCoveragePercent: 92.0,    // Very high, mostly ocean
		MeanOceanDepthM:      -2900.0, // Global layer depth ~2.6km, distributed over 92% area
		MeanLandHeightM:      500.0,   // Lower land (mostly volcanic islands)
		BimodalPeakOceanM:    -4200.0,
		BimodalPeakLandM:     200.0,

		// Climate: Hot House / Steam Atmosphere
		GlobalMeanTempC:    85.0, // Surface water hot but liquid (high pressure)
		EquatorToPoleGradC: 20.0, // Lower gradient due to efficient heat transport

		// Hydrology
		RiverDensityPercent: 5.0, // Less land = fewer developed river systems

		// Geology
		PlateCount:       12, // Faster, smaller plates (heat dissipation)
		SubplateCount:    30, // Less formed provinces
		ContinentCount:   2,  // Proto-continents only (Vaalbara, etc.)
		HotspotCount:     60, // Very active volcanism
		OceanTrenchCount: 30,

		// Astronomy
		MoonCount: 1, // Moon formed very early
	}
}

// ArcheanEarthBenchmarks returns target values for the Archean Eon (4.0-2.5 Ga).
// First stable continents form, still mostly water world, anoxic atmosphere.
func ArcheanEarthBenchmarks() EarthBenchmarks {
	return EarthBenchmarks{
		// Hypsometry: Transition from water world to first continents
		OceanCoveragePercent: 85.0,    // Still mostly ocean
		MeanOceanDepthM:      -3200.0, // Slightly deeper oceans
		MeanLandHeightM:      600.0,   // Building cratons
		BimodalPeakOceanM:    -4300.0,
		BimodalPeakLandM:     300.0,

		// Climate: Warm but cooler than Hadean
		GlobalMeanTempC:    40.0, // Hot greenhouse
		EquatorToPoleGradC: 25.0, // Still relatively uniform

		// Hydrology
		RiverDensityPercent: 8.0, // Developing drainage systems

		// Geology
		PlateCount:       10, // Still fast-cycling plates
		SubplateCount:    40,
		ContinentCount:   3,  // Vaalbara, Ur, Kenorland forming
		HotspotCount:     50, // High volcanism
		OceanTrenchCount: 25,

		// Astronomy
		MoonCount: 1,
	}
}

// ProterozoicEarthBenchmarks returns target values for the Proterozoic Eon (2.5-0.5 Ga).
// Stable continents, supercontinent cycles, Great Oxidation Event.
func ProterozoicEarthBenchmarks() EarthBenchmarks {
	return EarthBenchmarks{
		// Hypsometry: Recognizable continents forming
		OceanCoveragePercent: 78.0,    // Approaching modern
		MeanOceanDepthM:      -3500.0, // Approaching modern
		MeanLandHeightM:      750.0,   // Higher mountains forming
		BimodalPeakOceanM:    -4400.0,
		BimodalPeakLandM:     400.0,

		// Climate: Variable (Snowball Earth events in Neoproterozoic)
		GlobalMeanTempC:    20.0, // Cooler than Archean, variable
		EquatorToPoleGradC: 35.0, // Developing gradient

		// Hydrology
		RiverDensityPercent: 12.0, // Mature drainage systems

		// Geology
		PlateCount:       9, // Slowing plate motion
		SubplateCount:    45,
		ContinentCount:   5,  // Multiple continents (Rodinia, Columbia)
		HotspotCount:     45, // Moderate volcanism
		OceanTrenchCount: 22,

		// Astronomy
		MoonCount: 1,
	}
}

// =============================================================================
// Tolerance Configuration
// =============================================================================

// Tolerances defines acceptable deviation ranges for each metric.
// Values are expressed as percentages (0.0-1.0) or absolute values.
type Tolerances struct {
	// Hypsometry (percentage tolerance)
	OceanCoverage float64 // ±15% → 56-86%
	OceanDepth    float64 // ±30% → -2590 to -4810m
	LandHeight    float64 // ±30% → 588-1092m

	// Climate (absolute tolerance in °C)
	GlobalMeanTemp  float64 // ±5°C → 10-20°C
	EquatorPoleGrad float64 // ±15°C → 30-60°C

	// Geology (percentage tolerance)
	PlateCount     float64 // ±50% → 4-12 plates
	SubplateCount  float64 // ±50%
	ContinentCount float64 // ±50% → 4-10 continents

	// Hydrology (percentage tolerance)
	RiverDensity float64 // ±50%
}

// DefaultTolerances returns sanity-check tolerances for CI/CD.
// These are intentionally loose to prevent false failures while
// still catching major regressions.
func DefaultTolerances() Tolerances {
	return Tolerances{
		// Hypsometry
		OceanCoverage: 0.20, // ±20%
		OceanDepth:    0.30, // ±30%
		LandHeight:    0.30, // ±30%

		// Climate
		GlobalMeanTemp:  5.0,  // ±5°C absolute
		EquatorPoleGrad: 15.0, // ±15°C absolute

		// Geology
		PlateCount:     0.50, // ±50%
		SubplateCount:  0.50, // ±50%
		ContinentCount: 0.50, // ±50%

		// Hydrology
		RiverDensity: 0.50, // ±50%
	}
}

// StrictTolerances returns tighter tolerances for fine-tuning.
func StrictTolerances() Tolerances {
	return Tolerances{
		OceanCoverage:   0.10, // ±10%
		OceanDepth:      0.20, // ±20%
		LandHeight:      0.20, // ±20%
		GlobalMeanTemp:  3.0,  // ±3°C
		EquatorPoleGrad: 10.0, // ±10°C
		PlateCount:      0.30, // ±30%
		SubplateCount:   0.40, // ±40%
		ContinentCount:  0.30, // ±30%
		RiverDensity:    0.30, // ±30%
	}
}

// =============================================================================
// Calibration Results
// =============================================================================

// MetricStatus represents the pass/fail/warn status of a metric.
type MetricStatus int

const (
	StatusPass MetricStatus = iota
	StatusWarn
	StatusFail
)

func (s MetricStatus) String() string {
	switch s {
	case StatusPass:
		return "PASS"
	case StatusWarn:
		return "WARN"
	case StatusFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// MetricResult holds the comparison result for a single metric.
type MetricResult struct {
	Name       string
	Actual     float64
	Target     float64
	Tolerance  float64
	Status     MetricStatus
	Adjustment string // Suggested adjustment if failed
}

// IsWithinTolerance checks if actual is within tolerance of target.
// For percentage tolerances, tolerance is 0.0-1.0 (e.g., 0.20 = ±20%).
// For absolute tolerances, tolerance is the absolute deviation allowed.
func IsWithinTolerance(actual, target, tolerance float64, isAbsolute bool) bool {
	if isAbsolute {
		return math.Abs(actual-target) <= tolerance
	}
	// Percentage tolerance
	if target == 0 {
		return actual == 0
	}
	deviation := math.Abs(actual-target) / math.Abs(target)
	return deviation <= tolerance
}

// FormatMetricResult formats a metric result for display.
func FormatMetricResult(r MetricResult) string {
	statusStr := fmt.Sprintf("[%s]", r.Status)
	valueStr := fmt.Sprintf("%.1f", r.Actual)
	targetStr := fmt.Sprintf("%.1f", r.Target)

	line := fmt.Sprintf("  %-6s %s: %s (Target: %s)", statusStr, r.Name, valueStr, targetStr)

	if r.Status != StatusPass && r.Adjustment != "" {
		line += fmt.Sprintf("\n         → Adjustment: %s", r.Adjustment)
	}

	return line
}

// =============================================================================
// Simulation Statistics
// =============================================================================

// SimulationStats holds collected statistics from a simulation run.
type SimulationStats struct {
	// Metadata
	Seed       int64
	Resolution int
	Years      int64

	// Hypsometry
	OceanCoveragePercent float64
	MeanOceanDepthM      float64
	MeanLandHeightM      float64
	MinElevationM        float64
	MaxElevationM        float64
	ElevationHistogram   []int   // Binned elevation counts
	HistogramBinSize     float64 // Meters per bin

	// Climate
	GlobalMeanTempC  float64
	MinTempC         float64
	MaxTempC         float64
	EquatorMeanTempC float64
	PoleMeanTempC    float64
	MeanRainfallMM   float64
	MaxRainfallMM    float64

	// Hydrology
	RiverDensityPercent float64
	RiverCount          int
	LakeCount           int

	// Geology
	PlateCount     int
	ProvinceCount  int // Sub-regions within plates
	ContinentCount int
	HotspotCount   int
	CaveCount      int

	// Astronomy
	MoonCount int
}

// CalculateEquatorPoleGradient returns the temperature difference.
func (s SimulationStats) CalculateEquatorPoleGradient() float64 {
	return s.EquatorMeanTempC - s.PoleMeanTempC
}

// DetectBimodalPeaks analyzes the elevation histogram for bimodal distribution.
// Returns (oceanPeak, landPeak, isBimodal).
// A bimodal distribution indicates proper crustal differentiation.
func (s SimulationStats) DetectBimodalPeaks() (float64, float64, bool) {
	if len(s.ElevationHistogram) == 0 {
		return 0, 0, false
	}

	// Find peaks: local maxima with significant count
	// Ocean peak should be in negative elevation range
	// Land peak should be in positive elevation range

	binSize := s.HistogramBinSize
	minElev := s.MinElevationM

	var oceanPeak, landPeak float64
	var oceanMax, landMax int

	for i, count := range s.ElevationHistogram {
		elevation := minElev + float64(i)*binSize + binSize/2

		if elevation < 0 {
			// Ocean range
			if count > oceanMax {
				oceanMax = count
				oceanPeak = elevation
			}
		} else {
			// Land range
			if count > landMax {
				landMax = count
				landPeak = elevation
			}
		}
	}

	// Bimodal if both peaks have significant counts
	// and there's a valley between them
	totalCount := 0
	for _, c := range s.ElevationHistogram {
		totalCount += c
	}

	threshold := totalCount / 20 // 5% of total
	isBimodal := oceanMax > threshold && landMax > threshold

	return oceanPeak, landPeak, isBimodal
}
