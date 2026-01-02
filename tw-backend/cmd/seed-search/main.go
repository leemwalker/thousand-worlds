// Package main provides the seed-search CLI tool for finding optimal seeds.
// It runs multiple simulations in parallel and ranks them by Earth calibration score.
//
// Usage:
//
//	seed-search [flags]
//	  -start int      Starting seed (default: 1)
//	  -count int      Number of seeds to test (default: 100)
//	  -workers int    Parallel workers (default: 4)
//	  -years int      Years to simulate each seed (default: 500,000,000 for Hadean)
//	  -resolution int Map resolution (default: 128 for speed)
//	  -top int        Show top N results (default: 10)
//	  -json           Output results as JSON
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/calibration"

	"github.com/google/uuid"
)

// SeedResult holds the calibration result for a single seed.
type SeedResult struct {
	Seed            int64                          `json:"seed"`
	PassCount       int                            `json:"pass_count"`
	FailCount       int                            `json:"fail_count"`
	WarnCount       int                            `json:"warn_count"`
	TotalMetrics    int                            `json:"total_metrics"`
	Score           float64                        `json:"score"` // 0-100 composite score
	OceanCoverage   float64                        `json:"ocean_coverage"`
	MeanOceanDepth  float64                        `json:"mean_ocean_depth"`
	MeanLandHeight  float64                        `json:"mean_land_height"`
	GlobalTemp      float64                        `json:"global_temp"`
	PlateCount      int                            `json:"plate_count"`
	ContinentCount  int                            `json:"continent_count"`
	CratonCount     int                            `json:"craton_count"` // Ancient stable cores
	OrogenCount     int                            `json:"orogen_count"` // Mineral-rich fold belts
	BasinCount      int                            `json:"basin_count"`  // Sedimentary basins
	BimodalDetected bool                           `json:"bimodal_detected"`
	SimulationTime  time.Duration                  `json:"simulation_time"`
	Report          *calibration.CalibrationReport `json:"-"` // Full report (not serialized)
}

// CalculateScore computes a weighted composite score for ranking seeds.
// Higher is better, max 100.
func (r *SeedResult) CalculateScore(bench calibration.EarthBenchmarks) {
	score := 0.0
	maxScore := 0.0

	// Ocean coverage (weight: 25)
	// Target: 71%, tolerance ~15 percentage points
	oceanDiff := abs(r.OceanCoverage - bench.OceanCoveragePercent)
	oceanScore := max(0, 25-oceanDiff)
	score += oceanScore
	maxScore += 25

	// Ocean depth (weight: 15)
	// Target: -3700m, tolerance ~1000m
	depthDiff := abs(r.MeanOceanDepth-bench.MeanOceanDepthM) / 100 // Normalize
	depthScore := max(0, 15-depthDiff)
	score += depthScore
	maxScore += 15

	// Land height (weight: 10)
	// Target: 840m, tolerance ~300m
	landDiff := abs(r.MeanLandHeight-bench.MeanLandHeightM) / 50
	landScore := max(0, 10-landDiff)
	score += landScore
	maxScore += 10

	// Temperature (weight: 15)
	// For Hadean era, we expect higher temps (~50-100°C)
	// Modern Earth target: 15°C
	tempScore := 15.0
	if r.GlobalTemp < -50 || r.GlobalTemp > 150 {
		tempScore = 0 // Way out of range
	} else if r.GlobalTemp < 0 || r.GlobalTemp > 50 {
		tempScore = 7.5 // Partially in range
	}
	score += tempScore
	maxScore += 15

	// Plate count (weight: 10)
	// Target: 6-10 plates
	if r.PlateCount >= 5 && r.PlateCount <= 12 {
		score += 10
	} else if r.PlateCount >= 3 && r.PlateCount <= 15 {
		score += 5
	}
	maxScore += 10

	// Continent count (weight: 10)
	// Target: 2-8 continents for early Earth
	if r.ContinentCount >= 2 && r.ContinentCount <= 8 {
		score += 10
	} else if r.ContinentCount >= 1 && r.ContinentCount <= 10 {
		score += 5
	}
	maxScore += 10

	// Bimodal distribution (weight: 15)
	// Critical for Earth-like hypsometry
	if r.BimodalDetected {
		score += 15
	}
	maxScore += 15

	r.Score = (score / maxScore) * 100
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Parse flags
	startSeed := flag.Int64("start", 1, "Starting seed")
	seedCount := flag.Int("count", 100, "Number of seeds to test")
	workers := flag.Int("workers", 0, "Parallel workers (0 = NumCPU)")
	years := flag.Int64("years", 500_000_000, "Years to simulate (500M = Hadean era). Set to 0 to skip simulation.")
	resolution := flag.Int("resolution", 128, "Map resolution (lower = faster)")
	topN := flag.Int("top", 10, "Show top N results")
	minScore := flag.Float64("min-score", 80.0, "Minimum score to log to stderr immediately")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	profile := flag.String("profile", "modern", "Benchmark profile: 'modern' or 'hadean'")
	strategyName := flag.String("strategy", "linear", "Search strategy: 'linear' or 'halftree'")
	flag.Parse()

	if *workers == 0 {
		*workers = runtime.NumCPU()
	}

	fmt.Printf("🔍 Golden Seed Search\n")
	fmt.Printf("====================\n\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Strategy: %s\n", *strategyName)
	fmt.Printf("  Target Count: %d\n", *seedCount)
	fmt.Printf("  Years: %d (%.1f billion years)\n", *years, float64(*years)/1e9)
	fmt.Printf("  Resolution: %d\n", *resolution)
	fmt.Printf("  Workers: %d\n", *workers)
	fmt.Printf("  Profile: %s\n\n", *profile)

	var benchmarks calibration.EarthBenchmarks
	if *profile == "hadean" {
		benchmarks = calibration.HadeanEarthBenchmarks()
		fmt.Println("🌊 Using HADEAN benchmarks (92% Ocean, Hotter, Faster Plates)")
	} else {
		benchmarks = calibration.DefaultEarthBenchmarks()
		fmt.Println("🌍 Using MODERN Earth benchmarks (71% Ocean, Cooler, Stable Plates)")
	}

	// Select Strategy
	var strategy SearchStrategy
	switch *strategyName {
	case "halftree":
		strategy = &HalfTreeStrategy{}
	case "linear":
		strategy = &LinearStrategy{}
	default:
		log.Fatalf("Unknown strategy: %s", *strategyName)
	}

	// Create channels
	// Seeds: passed TO workers (from strategy)
	seeds := make(chan int64, *workers*2)
	// Worker Results: passed FROM workers (to strategy)
	workerResults := make(chan SeedResult, *workers*2)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for seed := range seeds {
				result := testSeed(seed, *years, *resolution, benchmarks)
				workerResults <- result
			}
		}(i)
	}

	// Close workerResults when all workers are done
	go func() {
		wg.Wait()
		close(workerResults)
	}()

	// Run Strategy
	config := SearchConfig{
		StartSeed:  *startSeed,
		Count:      *seedCount,
		Years:      *years,
		Resolution: *resolution,
		Workers:    *workers,
		Benchmarks: benchmarks,
	}

	// Strategy returns a channel of results to display.
	// It manages feeding 'seeds' and consuming 'workerResults'.
	displayResults := strategy.Run(seeds, workerResults, config)

	// Collect results for display
	allResults := make([]SeedResult, 0, *seedCount)
	completed := 0

	// Read from the channel returned by strategy
	for result := range displayResults {
		allResults = append(allResults, result)
		completed++

		// Real-time logging of high scores to Stderr (visible even if stdout is redirected)
		if result.Score >= *minScore {
			// Format similar to final output but succinct
			msg := fmt.Sprintf("\n🔥 FOUND CANDIDATE: Seed %d | Score: %.1f\n", result.Seed, result.Score)
			if result.SimulationTime < 100*time.Millisecond {
				msg += fmt.Sprintf("   [FAST] Plates: %d | Cratons: %d | Orogens: %d | Temp: %.1f°C\n",
					result.PlateCount, result.CratonCount, result.OrogenCount, result.GlobalTemp)
			} else {
				msg += fmt.Sprintf("   Ocean: %.1f%% | Depth: %.0fm | Land: %.0fm | Temp: %.1f°C\n",
					result.OceanCoverage, result.MeanOceanDepth, result.MeanLandHeight, result.GlobalTemp)
			}
			fmt.Fprintln(os.Stderr, msg)
		}

		if !*jsonOutput && completed%1000 == 0 {
			fmt.Printf("Progress: %d seeds tested\n", completed)
		}
	}

	// Sort by score (highest first)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// Output results
	if *jsonOutput {
		topResults := allResults
		if len(topResults) > *topN {
			topResults = topResults[:*topN]
		}
		data, _ := json.MarshalIndent(topResults, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("\n🏆 TOP %d SEEDS\n", *topN)
		fmt.Printf("================\n\n")

		showCount := *topN
		if showCount > len(allResults) {
			showCount = len(allResults)
		}

		for i := 0; i < showCount; i++ {
			r := allResults[i]
			fmt.Printf("#%d: Seed %d (Score: %.1f/100)\n", i+1, r.Seed, r.Score)
			// Only show metrics relevant to the simulation type
			if r.SimulationTime < 100*time.Millisecond {
				// Fast pass / Zero year
				fmt.Printf("    [FAST] Plates: %d | Continents: %d | Temp: %.1f°C\n",
					r.PlateCount, r.ContinentCount, r.GlobalTemp)
			} else {
				fmt.Printf("    Ocean: %.1f%% | Depth: %.0fm | Land: %.0fm | Temp: %.1f°C\n",
					r.OceanCoverage, r.MeanOceanDepth, r.MeanLandHeight, r.GlobalTemp)
				fmt.Printf("    Plates: %d | Continents: %d | Bimodal: %v\n",
					r.PlateCount, r.ContinentCount, r.BimodalDetected)
			}
			fmt.Printf("    Pass/Fail/Warn: %d/%d/%d | Time: %v\n\n",
				r.PassCount, r.FailCount, r.WarnCount, r.SimulationTime.Round(time.Millisecond))
		}

		// Show best seed prominently
		if len(allResults) > 0 {
			best := allResults[0]
			fmt.Printf("🌟 GOLDEN SEED: %d (Score: %.1f/100)\n", best.Seed, best.Score)
			fmt.Printf("\nTo use this seed:\n")
			// Suggest reasonable simulation parameters
			simYears := *years
			if simYears == 0 {
				simYears = 500_000_000
			}
			fmt.Printf("  world simulate %d --seed %d --geology\n", simYears, best.Seed)
		}
	}

	os.Exit(0)
}

func testSeed(seed, years int64, resolution int, bench calibration.EarthBenchmarks) SeedResult {
	start := time.Now()

	// Create world at Earth size
	worldID := uuid.New()
	circumference := 40_000_000.0 // 40,000 km (Earth)

	// === FAST PASS: Static Checks ===
	// Check properties that are deterministic from seed without full simulation
	// 1. Moons (Super fast)
	geo := ecosystem.NewWorldGeology(worldID, seed, circumference)

	// Check moons
	// === FAST PASS: Static Checks ===
	// Check properties that are deterministic from seed without full simulation
	// 1. Moons (Super fast)
	// geo is already initialized above

	// 3. FULL INITIALIZATION (at low res for speed)
	// We run the full init because it provides the Ocean Coverage baseline.
	geo.InitializeGeology(resolution)

	// Simulate in chunks for long durations
	if years > 0 {
		chunkSize := int64(10_000_000) // 10M years per chunk
		remaining := years

		for remaining > 0 {
			step := chunkSize
			if step > remaining {
				step = remaining
			}
			geo.SimulateGeology(step, 0.0)
			remaining -= step
		}
	} else {
		// Fast Pass: Just ensure basic systems are ready
		// InitializeGeology already does this, so nothing to do here.
	}

	// Collect statistics
	stats := calibration.CollectStats(geo)
	tolerances := calibration.DefaultTolerances()
	report := calibration.Score(stats, bench, tolerances)

	// Detect bimodal distribution
	_, _, bimodal := stats.DetectBimodalPeaks()

	result := SeedResult{
		Seed:            seed,
		PassCount:       report.PassCount,
		FailCount:       report.FailCount,
		WarnCount:       report.WarnCount,
		TotalMetrics:    len(report.Results),
		OceanCoverage:   stats.OceanCoveragePercent,
		MeanOceanDepth:  stats.MeanOceanDepthM,
		MeanLandHeight:  stats.MeanLandHeightM,
		GlobalTemp:      stats.GlobalMeanTempC,
		PlateCount:      stats.PlateCount,
		ContinentCount:  stats.ContinentCount,
		CratonCount:     stats.CratonCount,
		OrogenCount:     stats.OrogenCount,
		BasinCount:      stats.BasinCount,
		BimodalDetected: bimodal,
		SimulationTime:  time.Since(start),
		Report:          &report,
	}

	result.CalculateScore(bench)

	log.Printf("[Seed %d] Score: %.1f | Ocean: %.1f%% | Temp: %.1f°C | Time: %v",
		seed, result.Score, result.OceanCoverage, result.GlobalTemp, result.SimulationTime.Round(time.Millisecond))

	return result
}
