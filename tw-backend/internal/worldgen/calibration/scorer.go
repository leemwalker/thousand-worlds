package calibration

import (
	"fmt"
	"math"
	"strings"
)

// =============================================================================
// Calibration Scorer
// =============================================================================

// CalibrationReport contains the full comparison results.
type CalibrationReport struct {
	Stats     SimulationStats
	Benchmark EarthBenchmarks
	Results   []MetricResult
	PassCount int
	WarnCount int
	FailCount int
}

// Score compares simulation stats against Earth benchmarks.
func Score(stats SimulationStats, bench EarthBenchmarks, tol Tolerances) CalibrationReport {
	report := CalibrationReport{
		Stats:     stats,
		Benchmark: bench,
		Results:   make([]MetricResult, 0, 15),
	}

	// Hypsometry metrics
	report.addResult(checkPercentage(
		"Ocean Coverage",
		stats.OceanCoveragePercent,
		bench.OceanCoveragePercent,
		tol.OceanCoverage,
		"Adjust continental vs oceanic crust ratio in plate initialization",
	))

	report.addResult(checkPercentage(
		"Mean Ocean Depth",
		stats.MeanOceanDepthM,
		bench.MeanOceanDepthM,
		tol.OceanDepth,
		"Increase oceanic crust density difference for deeper basins",
	))

	report.addResult(checkPercentage(
		"Mean Land Height",
		stats.MeanLandHeightM,
		bench.MeanLandHeightM,
		tol.LandHeight,
		"Tune tectonic collision uplift factor",
	))

	// Climate metrics
	report.addResult(checkAbsolute(
		"Global Mean Temp",
		stats.GlobalMeanTempC,
		bench.GlobalMeanTempC,
		tol.GlobalMeanTemp,
		"°C",
		"Adjust baseline temperature in climate generator",
	))

	gradient := stats.CalculateEquatorPoleGradient()
	report.addResult(checkAbsolute(
		"Equator-Pole Gradient",
		gradient,
		bench.EquatorToPoleGradC,
		tol.EquatorPoleGrad,
		"°C",
		"Modify latitude temperature formula",
	))

	// Geology metrics
	report.addResult(checkPercentage(
		"Plate Count",
		float64(stats.PlateCount),
		float64(bench.PlateCount),
		tol.PlateCount,
		"Adjust initial plate generation count",
	))

	report.addResult(checkPercentage(
		"Province Count",
		float64(stats.ProvinceCount),
		float64(bench.SubplateCount),
		tol.SubplateCount,
		"Tune geological province generation",
	))

	report.addResult(checkPercentage(
		"Continent Count",
		float64(stats.ContinentCount),
		float64(bench.ContinentCount),
		tol.ContinentCount,
		"Adjust continental plate assignment ratio",
	))

	// Hydrology metrics
	report.addResult(checkPercentage(
		"River Density",
		stats.RiverDensityPercent,
		bench.RiverDensityPercent,
		tol.RiverDensity,
		"Tune flux threshold or erosion parameters",
	))

	// Astronomy metrics (informational, no tolerance)
	report.addResult(MetricResult{
		Name:   "Moon Count",
		Actual: float64(stats.MoonCount),
		Target: float64(bench.MoonCount),
		Status: StatusPass, // Always pass, just informational
	})

	// Bimodal distribution check
	oceanPeak, landPeak, isBimodal := stats.DetectBimodalPeaks()
	bimodalResult := MetricResult{
		Name:   "Bimodal Distribution",
		Actual: 1.0,
		Target: 1.0,
	}
	if isBimodal {
		bimodalResult.Status = StatusPass
		bimodalResult.Adjustment = fmt.Sprintf("Detected peaks: Ocean %.0fm, Land %.0fm", oceanPeak, landPeak)
	} else {
		bimodalResult.Status = StatusFail
		bimodalResult.Adjustment = "Crustal differentiation not detected - check isostasy implementation"
	}
	report.addResult(bimodalResult)

	return report
}

func (r *CalibrationReport) addResult(result MetricResult) {
	r.Results = append(r.Results, result)
	switch result.Status {
	case StatusPass:
		r.PassCount++
	case StatusWarn:
		r.WarnCount++
	case StatusFail:
		r.FailCount++
	}
}

// checkPercentage compares values using percentage tolerance.
func checkPercentage(name string, actual, target, tolerance float64, adjustment string) MetricResult {
	result := MetricResult{
		Name:       name,
		Actual:     actual,
		Target:     target,
		Tolerance:  tolerance * 100, // Convert to display percentage
		Adjustment: adjustment,
	}

	if target == 0 {
		if actual == 0 {
			result.Status = StatusPass
		} else {
			result.Status = StatusFail
		}
		return result
	}

	deviation := math.Abs(actual-target) / math.Abs(target)

	if deviation <= tolerance {
		result.Status = StatusPass
	} else if deviation <= tolerance*1.5 {
		result.Status = StatusWarn
	} else {
		result.Status = StatusFail
	}

	return result
}

// checkAbsolute compares values using absolute tolerance.
func checkAbsolute(name string, actual, target, tolerance float64, unit, adjustment string) MetricResult {
	result := MetricResult{
		Name:       name + " (" + unit + ")",
		Actual:     actual,
		Target:     target,
		Tolerance:  tolerance,
		Adjustment: adjustment,
	}

	deviation := math.Abs(actual - target)

	if deviation <= tolerance {
		result.Status = StatusPass
	} else if deviation <= tolerance*1.5 {
		result.Status = StatusWarn
	} else {
		result.Status = StatusFail
	}

	return result
}

// =============================================================================
// Report Formatting
// =============================================================================

// FormatScorecard returns a formatted text scorecard.
func (r CalibrationReport) FormatScorecard() string {
	var sb strings.Builder

	sb.WriteString("============ EARTH CALIBRATION SCORECARD ============\n")
	sb.WriteString(fmt.Sprintf("Seed: %d | Resolution: %d | Years: %d\n\n",
		r.Stats.Seed, r.Stats.Resolution, r.Stats.Years))

	// Group results by category
	sb.WriteString("HYPSOMETRY\n")
	for _, result := range r.Results {
		if strings.Contains(result.Name, "Ocean") || strings.Contains(result.Name, "Land") ||
			strings.Contains(result.Name, "Bimodal") {
			sb.WriteString(FormatMetricResult(result))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nCLIMATE\n")
	for _, result := range r.Results {
		if strings.Contains(result.Name, "Temp") || strings.Contains(result.Name, "Gradient") {
			sb.WriteString(FormatMetricResult(result))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nGEOLOGY\n")
	for _, result := range r.Results {
		if strings.Contains(result.Name, "Plate") || strings.Contains(result.Name, "Province") ||
			strings.Contains(result.Name, "Continent") {
			sb.WriteString(FormatMetricResult(result))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nHYDROLOGY\n")
	for _, result := range r.Results {
		if strings.Contains(result.Name, "River") {
			sb.WriteString(FormatMetricResult(result))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nASTRONOMY\n")
	for _, result := range r.Results {
		if strings.Contains(result.Name, "Moon") {
			sb.WriteString(FormatMetricResult(result))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("\nOVERALL: %d/%d PASS", r.PassCount, len(r.Results)))
	if r.FailCount > 0 {
		sb.WriteString(fmt.Sprintf(" | %d FAIL", r.FailCount))
	}
	if r.WarnCount > 0 {
		sb.WriteString(fmt.Sprintf(" | %d WARN", r.WarnCount))
	}
	sb.WriteString("\n")

	return sb.String()
}

// IsCalibrated returns true if all critical metrics pass.
func (r CalibrationReport) IsCalibrated() bool {
	return r.FailCount == 0
}

// =============================================================================
// Habitability Scoring
// =============================================================================

// CalculateHabitabilityScore computes a 0-100 Earth-like score.
// Higher scores indicate more Earth-like habitability conditions.
// Scoring factors and weights match the seed-search tool for consistency.
func CalculateHabitabilityScore(stats SimulationStats, bench EarthBenchmarks) HabitabilityScore {
	result := HabitabilityScore{
		ContinentCount: stats.ContinentCount,
		PlateCount:     stats.PlateCount,
	}

	// --- Ocean Score (max 25) ---
	// Ocean coverage (weight: 15) - Target: 71%, tolerance ~15 percentage points
	oceanDiff := math.Abs(stats.OceanCoveragePercent - bench.OceanCoveragePercent)
	oceanCoverageScore := math.Max(0, 15-oceanDiff)

	// Ocean depth (weight: 10) - Target: -3700m, tolerance ~1000m
	depthDiff := math.Abs(stats.MeanOceanDepthM-bench.MeanOceanDepthM) / 100 // Normalize
	oceanDepthScore := math.Max(0, 10-depthDiff)

	result.OceanScore = oceanCoverageScore + oceanDepthScore

	// --- Land Score (max 20) ---
	// Land height (weight: 10) - Target: 840m, tolerance ~300m
	landDiff := math.Abs(stats.MeanLandHeightM-bench.MeanLandHeightM) / 50
	landHeightScore := math.Max(0, 10-landDiff)

	// Continent count (weight: 10) - Target: 2-8 continents
	var continentScore float64
	if stats.ContinentCount >= 2 && stats.ContinentCount <= 8 {
		continentScore = 10
	} else if stats.ContinentCount >= 1 && stats.ContinentCount <= 10 {
		continentScore = 5
	}

	result.LandScore = landHeightScore + continentScore

	// --- Climate Score (max 25) ---
	// Temperature (weight: 15) - Modern Earth target: 0-30°C optimal
	var tempScore float64 = 15.0
	if stats.GlobalMeanTempC < -50 || stats.GlobalMeanTempC > 150 {
		tempScore = 0 // Way out of range
	} else if stats.GlobalMeanTempC < 0 || stats.GlobalMeanTempC > 50 {
		tempScore = 7.5 // Partially in range
	}

	// Equator-pole gradient (weight: 10) - Target: ~45°C
	gradient := stats.CalculateEquatorPoleGradient()
	gradDiff := math.Abs(gradient-bench.EquatorToPoleGradC) / 5
	gradScore := math.Max(0, 10-gradDiff)

	result.ClimateScore = tempScore + gradScore

	// --- Tectonic Score (max 30) ---
	// Plate count (weight: 10) - Target: 6-10 plates
	var plateScore float64
	if stats.PlateCount >= 5 && stats.PlateCount <= 12 {
		plateScore = 10
	} else if stats.PlateCount >= 3 && stats.PlateCount <= 15 {
		plateScore = 5
	}

	// Bimodal distribution (weight: 20) - Critical for Earth-like hypsometry
	_, _, bimodal := stats.DetectBimodalPeaks()
	result.BimodalOK = bimodal
	var bimodalScore float64
	if bimodal {
		bimodalScore = 20
	}

	result.TectonicScore = plateScore + bimodalScore

	// Total score (max 100)
	result.Score = result.OceanScore + result.LandScore + result.ClimateScore + result.TectonicScore

	return result
}
