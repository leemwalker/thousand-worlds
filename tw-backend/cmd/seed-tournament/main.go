// Package main provides the seed-tournament CLI tool for evolutionary seed discovery.
// It runs seeds through progressive geological eras, filtering the top 10% at each stage.
//
// Usage:
//
//	seed-tournament [flags]
//	  -start int      Starting seed (default: 1)
//	  -count int      Initial seed pool (default: 1000)
//	  -workers int    Parallel workers (default: NumCPU)
//	  -resolution int Map resolution (default: 128)
//	  -survival float Survival rate per era (default: 0.1 = top 10%)
//	  -output string  Output file for results (default: tournament_results.json)
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

// Era represents a geological time period with its benchmark targets.
type Era struct {
	Name       string
	StartYear  int64
	EndYear    int64
	Benchmarks calibration.EarthBenchmarks
}

// SeedResult holds the calibration result for a single seed at a specific era.
type SeedResult struct {
	Seed           int64         `json:"seed"`
	Era            string        `json:"era"`
	Score          float64       `json:"score"`
	OceanCoverage  float64       `json:"ocean_coverage"`
	MeanOceanDepth float64       `json:"mean_ocean_depth"`
	MeanLandHeight float64       `json:"mean_land_height"`
	GlobalTemp     float64       `json:"global_temp"`
	PlateCount     int           `json:"plate_count"`
	ContinentCount int           `json:"continent_count"`
	SimDuration    time.Duration `json:"sim_duration"`
}

// TournamentResult holds the final lineage of seeds that survived all eras.
type TournamentResult struct {
	GoldenSeeds []SeedLineage `json:"golden_seeds"`
	TotalTime   time.Duration `json:"total_time"`
	EraStats    []EraStats    `json:"era_stats"`
}

// SeedLineage tracks a seed's performance through all eras.
type SeedLineage struct {
	Seed       int64      `json:"seed"`
	FinalScore float64    `json:"final_score"`
	Scores     []EraScore `json:"era_scores"`
}

// EraScore holds the score for a specific era.
type EraScore struct {
	Era   string  `json:"era"`
	Score float64 `json:"score"`
}

// EraStats holds statistics for a single era round.
type EraStats struct {
	Era         string        `json:"era"`
	SeedsIn     int           `json:"seeds_in"`
	SeedsOut    int           `json:"seeds_out"`
	TopScore    float64       `json:"top_score"`
	BottomScore float64       `json:"bottom_score"`
	Duration    time.Duration `json:"duration"`
}

// Define the geological eras
func getEras() []Era {
	return []Era{
		{
			Name:       "Hadean",
			StartYear:  0,
			EndYear:    500_000_000, // 500 million years
			Benchmarks: calibration.HadeanEarthBenchmarks(),
		},
		{
			Name:       "Archean",
			StartYear:  500_000_000,
			EndYear:    2_500_000_000, // 2.5 billion years
			Benchmarks: calibration.ArcheanEarthBenchmarks(),
		},
		{
			Name:       "Proterozoic",
			StartYear:  2_500_000_000,
			EndYear:    4_000_000_000, // 4 billion years
			Benchmarks: calibration.ProterozoicEarthBenchmarks(),
		},
		{
			Name:       "Modern",
			StartYear:  4_000_000_000,
			EndYear:    4_500_000_000, // 4.5 billion years
			Benchmarks: calibration.DefaultEarthBenchmarks(),
		},
	}
}

func main() {
	// Parse flags
	startSeed := flag.Int64("start", 1, "Starting seed")
	seedCount := flag.Int("count", 1000, "Initial seed pool size")
	workers := flag.Int("workers", 0, "Parallel workers (0 = NumCPU)")
	resolution := flag.Int("resolution", 128, "Map resolution (lower = faster)")
	survivalRate := flag.Float64("survival", 0.10, "Survival rate per era (0.10 = top 10%)")
	outputFile := flag.String("output", "tournament_results.json", "Output file")
	flag.Parse()

	if *workers == 0 {
		*workers = runtime.NumCPU()
	}

	fmt.Println("🏆 GOLDEN SEED TOURNAMENT")
	fmt.Println("=========================")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Initial Pool: %d seeds\n", *seedCount)
	fmt.Printf("  Survival Rate: %.0f%%\n", *survivalRate*100)
	fmt.Printf("  Resolution: %d\n", *resolution)
	fmt.Printf("  Workers: %d\n\n", *workers)

	eras := getEras()
	tournamentStart := time.Now()

	// Initialize seed pool
	currentSeeds := make([]int64, *seedCount)
	for i := 0; i < *seedCount; i++ {
		currentSeeds[i] = *startSeed + int64(i)
	}

	// Track lineage for seeds that survive
	lineage := make(map[int64][]EraScore)
	for _, seed := range currentSeeds {
		lineage[seed] = []EraScore{}
	}

	eraStats := []EraStats{}

	// Run each era
	for eraIdx, era := range eras {
		eraStart := time.Now()
		fmt.Printf("\n📅 ERA %d: %s (%d → %d years)\n", eraIdx+1, era.Name, era.StartYear/1_000_000, era.EndYear/1_000_000)
		fmt.Printf("   Seeds entering: %d\n", len(currentSeeds))

		// Calculate years to simulate for this era
		yearsToSim := era.EndYear - era.StartYear

		// Run seeds in parallel
		results := runEraParallel(currentSeeds, era, yearsToSim, *resolution, *workers)

		// Sort by score
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})

		// Calculate survival count
		survivalCount := int(float64(len(results)) * *survivalRate)
		if survivalCount < 1 {
			survivalCount = 1
		}
		if survivalCount > len(results) {
			survivalCount = len(results)
		}

		// Take top survivors
		survivors := results[:survivalCount]

		// Update lineage
		for _, r := range results {
			if scores, exists := lineage[r.Seed]; exists {
				lineage[r.Seed] = append(scores, EraScore{Era: era.Name, Score: r.Score})
			}
		}

		// Prune non-survivors from lineage
		newLineage := make(map[int64][]EraScore)
		for _, r := range survivors {
			newLineage[r.Seed] = lineage[r.Seed]
		}
		lineage = newLineage

		// Update current seeds for next era
		currentSeeds = make([]int64, len(survivors))
		for i, r := range survivors {
			currentSeeds[i] = r.Seed
		}

		// Record era stats
		stats := EraStats{
			Era:      era.Name,
			SeedsIn:  len(results),
			SeedsOut: len(survivors),
			Duration: time.Since(eraStart),
		}
		if len(results) > 0 {
			stats.TopScore = results[0].Score
			stats.BottomScore = results[len(results)-1].Score
		}
		eraStats = append(eraStats, stats)

		fmt.Printf("   Survivors: %d (Top: %.1f, Bottom: %.1f)\n", len(survivors), stats.TopScore, stats.BottomScore)
		fmt.Printf("   Duration: %v\n", stats.Duration.Round(time.Second))

		// Show top 3 for this era
		for i := 0; i < 3 && i < len(survivors); i++ {
			r := survivors[i]
			fmt.Printf("   #%d: Seed %d (Score: %.1f) Ocean: %.1f%%, Temp: %.1f°C\n",
				i+1, r.Seed, r.Score, r.OceanCoverage, r.GlobalTemp)
		}
	}

	// Build final results
	goldenSeeds := []SeedLineage{}
	for seed, scores := range lineage {
		finalScore := 0.0
		for _, s := range scores {
			finalScore += s.Score
		}
		finalScore /= float64(len(scores)) // Average across eras

		goldenSeeds = append(goldenSeeds, SeedLineage{
			Seed:       seed,
			FinalScore: finalScore,
			Scores:     scores,
		})
	}

	// Sort by final score
	sort.Slice(goldenSeeds, func(i, j int) bool {
		return goldenSeeds[i].FinalScore > goldenSeeds[j].FinalScore
	})

	result := TournamentResult{
		GoldenSeeds: goldenSeeds,
		TotalTime:   time.Since(tournamentStart),
		EraStats:    eraStats,
	}

	// Write results to file
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}
	if err := os.WriteFile(*outputFile, data, 0644); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	// Print final results
	fmt.Println("\n🌟 GOLDEN SEEDS 🌟")
	fmt.Println("==================")
	for i, gs := range goldenSeeds {
		fmt.Printf("#%d: Seed %d (Average Score: %.1f)\n", i+1, gs.Seed, gs.FinalScore)
		for _, s := range gs.Scores {
			fmt.Printf("    %s: %.1f\n", s.Era, s.Score)
		}
	}

	fmt.Printf("\n✅ Tournament complete in %v\n", result.TotalTime.Round(time.Second))
	fmt.Printf("📁 Results saved to: %s\n", *outputFile)
}

// runEraParallel runs all seeds for an era in parallel.
func runEraParallel(seeds []int64, era Era, yearsToSim int64, resolution, workers int) []SeedResult {
	seedChan := make(chan int64, len(seeds))
	resultChan := make(chan SeedResult, len(seeds))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for seed := range seedChan {
				result := simulateSeedForEra(seed, era, yearsToSim, resolution)
				resultChan <- result
			}
		}()
	}

	// Feed seeds
	go func() {
		for _, seed := range seeds {
			seedChan <- seed
		}
		close(seedChan)
	}()

	// Wait and close
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]SeedResult, 0, len(seeds))
	completed := 0
	for r := range resultChan {
		results = append(results, r)
		completed++
		if completed%100 == 0 {
			fmt.Printf("   Progress: %d/%d\n", completed, len(seeds))
		}
	}

	return results
}

// simulateSeedForEra runs a single seed through an era.
func simulateSeedForEra(seed int64, era Era, yearsToSim int64, resolution int) SeedResult {
	start := time.Now()

	worldID := uuid.New()
	circumference := 40_000_000.0

	geo := ecosystem.NewWorldGeology(worldID, seed, circumference)
	geo.InitializeGeology(0)

	// Simulate to the start of this era first (if not Hadean)
	if era.StartYear > 0 {
		chunkSize := int64(50_000_000) // 50M year chunks
		remaining := era.StartYear
		for remaining > 0 {
			step := chunkSize
			if step > remaining {
				step = remaining
			}
			geo.SimulateGeology(step, 0.0)
			remaining -= step
		}
	}

	// Now simulate this era
	chunkSize := int64(50_000_000)
	remaining := yearsToSim
	for remaining > 0 {
		step := chunkSize
		if step > remaining {
			step = remaining
		}
		geo.SimulateGeology(step, 0.0)
		remaining -= step
	}

	// Collect stats and score
	stats := calibration.CollectStats(geo)
	score := calculateEraScore(stats, era.Benchmarks)

	return SeedResult{
		Seed:           seed,
		Era:            era.Name,
		Score:          score,
		OceanCoverage:  stats.OceanCoveragePercent,
		MeanOceanDepth: stats.MeanOceanDepthM,
		MeanLandHeight: stats.MeanLandHeightM,
		GlobalTemp:     stats.GlobalMeanTempC,
		PlateCount:     stats.PlateCount,
		ContinentCount: stats.ContinentCount,
		SimDuration:    time.Since(start),
	}
}

// calculateEraScore computes a weighted score against era benchmarks.
func calculateEraScore(stats calibration.SimulationStats, bench calibration.EarthBenchmarks) float64 {
	score := 0.0
	maxScore := 100.0

	// Ocean coverage (25 points)
	oceanDiff := abs(stats.OceanCoveragePercent - bench.OceanCoveragePercent)
	score += max(0, 25-oceanDiff)

	// Ocean depth (15 points)
	depthDiff := abs(stats.MeanOceanDepthM-bench.MeanOceanDepthM) / 100
	score += max(0, 15-depthDiff)

	// Land height (10 points)
	landDiff := abs(stats.MeanLandHeightM-bench.MeanLandHeightM) / 50
	score += max(0, 10-landDiff)

	// Temperature (20 points)
	tempDiff := abs(stats.GlobalMeanTempC - bench.GlobalMeanTempC)
	score += max(0, 20-tempDiff/2)

	// Plate count (15 points)
	plateDiff := abs(float64(stats.PlateCount - bench.PlateCount))
	score += max(0, 15-plateDiff*2)

	// Bimodal detection (15 points)
	_, _, bimodal := stats.DetectBimodalPeaks()
	if bimodal {
		score += 15
	}

	return (score / maxScore) * 100
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
