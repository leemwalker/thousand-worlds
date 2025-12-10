package simulation

import (
	"sort"
)

// EconomicMetrics holds key indicators of economic health
type EconomicMetrics struct {
	TotalWealth     int
	MoneySupply     int
	GiniCoefficient float64
	InflationRate   float64
	TradeVolume     int
	ActiveMerchants int
}

// CalculateGiniCoefficient measures wealth inequality (0.0 = perfect equality, 1.0 = perfect inequality)
func CalculateGiniCoefficient(wealths []int) float64 {
	if len(wealths) == 0 {
		return 0.0
	}

	// Sort wealths
	sortedWealths := make([]int, len(wealths))
	copy(sortedWealths, wealths)
	sort.Ints(sortedWealths)

	n := float64(len(sortedWealths))
	sum := 0.0
	for _, w := range sortedWealths {
		sum += float64(w)
	}

	if sum == 0 {
		return 0.0
	}

	diffSum := 0.0

	for i, w := range sortedWealths {
		// Gini formula: (2 * sum(i * w_i)) / (n * sum(w_i)) - (n + 1) / n
		// Using simpler form: sum(|xi - xj|) / (2 * n^2 * mean)
		// Or the sorted form:
		// G = (2 / n) * (sum(i * xi) / sum(xi)) - (n + 1) / n
		// where i is 1-based index

		diffSum += float64(i+1) * float64(w)
	}

	gini := (2.0/n)*(diffSum/sum) - (n+1.0)/n
	return gini
}

// CalculateInflationRate measures price changes over a period
func CalculateInflationRate(currentPrices, previousPrices map[string]float64) float64 {
	if len(currentPrices) == 0 || len(previousPrices) == 0 {
		return 0.0
	}

	totalChange := 0.0
	count := 0

	for item, currPrice := range currentPrices {
		if prevPrice, ok := previousPrices[item]; ok && prevPrice > 0 {
			change := (currPrice - prevPrice) / prevPrice
			totalChange += change
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalChange / float64(count)
}
