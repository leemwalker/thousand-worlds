// Package main provides the verify-earth CLI tool for geophysical calibration.
// It runs a full world simulation and compares results against Earth benchmarks.
//
// Usage:
//
//	verify-earth [flags]
//	  -seed int       Random seed (default: time-based)
//	  -resolution int Map resolution (default: 256)
//	  -years int      Years to simulate (default: 10,000,000)
//	  -strict         Use strict tolerances for fine-tuning
//	  -verbose        Print detailed statistics
//	  -json           Output results as JSON
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/calibration"

	"github.com/google/uuid"
)

func main() {
	// Parse flags
	seed := flag.Int64("seed", 0, "Random seed (0 = time-based)")
	resolution := flag.Int("resolution", 256, "Map resolution (width)")
	years := flag.Int64("years", 10_000_000, "Years to simulate")
	strict := flag.Bool("strict", false, "Use strict tolerances")
	verbose := flag.Bool("verbose", false, "Print detailed statistics")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	// Use time-based seed if not specified
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	fmt.Printf("🌍 Earth Calibration Tool\n")
	fmt.Printf("========================\n\n")

	// Step 1: Generate World
	fmt.Printf("Step 1: Generating world (seed: %d, resolution: %d)...\n", *seed, *resolution)
	startTime := time.Now()

	worldID := uuid.New()
	// Calculate circumference to achieve target resolution
	// Resolution 256 at 10km/pixel = 2560km circumference
	// But we need a reasonable Earth-like size: 40,000km
	circumference := 40_000_000.0 // 40,000 km in meters

	geo := ecosystem.NewWorldGeology(worldID, *seed, circumference)
	geo.InitializeGeology()

	initTime := time.Since(startTime)
	fmt.Printf("   World initialized in %v\n", initTime)

	// Step 2: Simulate
	fmt.Printf("Step 2: Simulating %d years...\n", *years)
	simStart := time.Now()

	// Simulate in chunks to show progress
	chunkSize := int64(1_000_000) // 1M years per chunk
	remaining := *years

	for remaining > 0 {
		step := chunkSize
		if step > remaining {
			step = remaining
		}
		geo.SimulateGeology(step, 0.0)
		remaining -= step

		if *verbose {
			fmt.Printf("   Simulated to year %d...\n", *years-remaining)
		}
	}

	simTime := time.Since(simStart)
	fmt.Printf("   Simulation completed in %v\n", simTime)

	// Step 3: Collect Statistics
	fmt.Printf("Step 3: Collecting statistics...\n")
	stats := calibration.CollectStats(geo)

	// Step 4: Score against benchmarks
	fmt.Printf("Step 4: Scoring against Earth benchmarks...\n\n")

	benchmarks := calibration.DefaultEarthBenchmarks()
	var tolerances calibration.Tolerances
	if *strict {
		tolerances = calibration.StrictTolerances()
		fmt.Println("Using STRICT tolerances for fine-tuning\n")
	} else {
		tolerances = calibration.DefaultTolerances()
	}

	report := calibration.Score(stats, benchmarks, tolerances)

	// Output results
	if *jsonOutput {
		outputJSON(report)
	} else {
		fmt.Println(report.FormatScorecard())

		if *verbose {
			printDetailedStats(stats)
		}
	}

	// Exit with appropriate code
	if report.IsCalibrated() {
		fmt.Println("\n✅ Calibration PASSED")
		os.Exit(0)
	} else {
		fmt.Println("\n❌ Calibration FAILED - adjustments needed")
		os.Exit(1)
	}
}

func printDetailedStats(stats calibration.SimulationStats) {
	fmt.Println("\n========== DETAILED STATISTICS ==========")
	fmt.Println("\nHYPSOMETRY:")
	fmt.Printf("  Min Elevation:    %.1fm\n", stats.MinElevationM)
	fmt.Printf("  Max Elevation:    %.1fm\n", stats.MaxElevationM)
	fmt.Printf("  Ocean Coverage:   %.1f%%\n", stats.OceanCoveragePercent)
	fmt.Printf("  Mean Ocean Depth: %.1fm\n", stats.MeanOceanDepthM)
	fmt.Printf("  Mean Land Height: %.1fm\n", stats.MeanLandHeightM)

	fmt.Println("\nCLIMATE:")
	fmt.Printf("  Global Mean Temp: %.1f°C\n", stats.GlobalMeanTempC)
	fmt.Printf("  Min Temperature:  %.1f°C\n", stats.MinTempC)
	fmt.Printf("  Max Temperature:  %.1f°C\n", stats.MaxTempC)
	fmt.Printf("  Equator Mean:     %.1f°C\n", stats.EquatorMeanTempC)
	fmt.Printf("  Pole Mean:        %.1f°C\n", stats.PoleMeanTempC)
	fmt.Printf("  Mean Rainfall:    %.1fmm/year\n", stats.MeanRainfallMM)

	fmt.Println("\nGEOLOGY:")
	fmt.Printf("  Plate Count:      %d\n", stats.PlateCount)
	fmt.Printf("  Province Count:   %d\n", stats.ProvinceCount)
	fmt.Printf("  Continent Count:  %d\n", stats.ContinentCount)
	fmt.Printf("  Hotspot Count:    %d\n", stats.HotspotCount)
	fmt.Printf("  Cave Count:       %d\n", stats.CaveCount)

	fmt.Println("\nHYDROLOGY:")
	fmt.Printf("  River Count:      %d\n", stats.RiverCount)
	fmt.Printf("  River Density:    %.1f%%\n", stats.RiverDensityPercent)
	fmt.Printf("  Lake Count:       %d\n", stats.LakeCount)

	fmt.Println("\nASTRONOMY:")
	fmt.Printf("  Moon Count:       %d\n", stats.MoonCount)

	// Print histogram summary
	oceanPeak, landPeak, isBimodal := stats.DetectBimodalPeaks()
	fmt.Println("\nELEVATION DISTRIBUTION:")
	fmt.Printf("  Bimodal:          %v\n", isBimodal)
	if isBimodal {
		fmt.Printf("  Ocean Peak:       %.0fm\n", oceanPeak)
		fmt.Printf("  Land Peak:        %.0fm\n", landPeak)
	}

	// Print simplified histogram
	if len(stats.ElevationHistogram) > 0 {
		fmt.Println("\n  Histogram (simplified):")
		printHistogramBar(stats)
	}
}

func printHistogramBar(stats calibration.SimulationStats) {
	// Compress to ~20 bins for display
	displayBins := 20
	histogram := stats.ElevationHistogram
	binSize := len(histogram) / displayBins
	if binSize < 1 {
		binSize = 1
	}

	maxCount := 0
	compressed := make([]int, displayBins)
	for i := 0; i < displayBins; i++ {
		start := i * binSize
		end := start + binSize
		if end > len(histogram) {
			end = len(histogram)
		}
		sum := 0
		for j := start; j < end; j++ {
			sum += histogram[j]
		}
		compressed[i] = sum
		if sum > maxCount {
			maxCount = sum
		}
	}

	// Print bars
	width := 40
	for i, count := range compressed {
		barLen := count * width / maxCount
		if barLen < 0 {
			barLen = 0
		}
		elev := stats.MinElevationM + float64(i*binSize)*stats.HistogramBinSize
		bar := ""
		for j := 0; j < barLen; j++ {
			bar += "█"
		}
		fmt.Printf("  %6.0fm |%s\n", elev, bar)
	}
}

func outputJSON(report calibration.CalibrationReport) {
	output := map[string]interface{}{
		"stats":      report.Stats,
		"benchmark":  report.Benchmark,
		"results":    report.Results,
		"pass_count": report.PassCount,
		"warn_count": report.WarnCount,
		"fail_count": report.FailCount,
		"calibrated": report.IsCalibrated(),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(data))
}
