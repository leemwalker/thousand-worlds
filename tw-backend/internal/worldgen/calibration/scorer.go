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
