package main

import (
	"tw-backend/internal/worldgen/calibration"
)

// SearchStrategy defines how seeds are selected and processed
type SearchStrategy interface {
	// Name returns the strategy name
	Name() string

	// Explanation returns a description of the strategy
	Explanation() string

	// Run executes the search strategy
	// seeds: input channel to feed seeds to workers
	// results: input channel from workers (strategy can inspect results for feedback)
	// config: search configuration
	// Returns: output channel for main to display results
	Run(seeds chan<- int64, results <-chan SeedResult, config SearchConfig) <-chan SeedResult
}

// SearchConfig holds configuration for the search strategy
type SearchConfig struct {
	StartSeed  int64
	Count      int
	Years      int64
	Resolution int
	Workers    int
	Benchmarks calibration.EarthBenchmarks
}
