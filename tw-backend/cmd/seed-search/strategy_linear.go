package main

import (
	"fmt"
)

// LinearStrategy implements the standard sequential search
type LinearStrategy struct{}

func (s *LinearStrategy) Name() string {
	return "linear"
}

func (s *LinearStrategy) Explanation() string {
	return "Sequential Scan: Checks every seed in order from Start to Start+Count."
}

func (s *LinearStrategy) Run(seeds chan<- int64, results <-chan SeedResult, config SearchConfig) <-chan SeedResult {
	fmt.Printf("➡️ Linear Strategy: scanning %d seeds starting from %d\n", config.Count, config.StartSeed)

	output := make(chan SeedResult)

	// Start goroutine to forward results
	go func() {
		defer close(output)
		for r := range results {
			output <- r
		}
	}()

	// Start goroutine to feed seeds
	go func() {
		defer close(seeds)
		for i := int64(0); i < int64(config.Count); i++ {
			seeds <- config.StartSeed + i
		}
	}()

	return output
}
