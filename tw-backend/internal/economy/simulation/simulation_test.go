package simulation

import (
	"context"
	"math/rand"
	"testing"

	"mud-platform-backend/internal/economy/npc"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStability(t *testing.T) {
	sim := NewSimulationManager()

	// Initialize 100 merchants with varying starting wealth
	for i := 0; i < 100; i++ {
		wealth := 1000 + rand.Intn(9000) // 1000 - 10000
		sim.AddMerchant(&npc.Merchant{
			NPCID:  uuid.New(),
			Wealth: wealth,
		})
	}

	initialMetrics := sim.GetMetrics()
	t.Logf("Initial Gini: %.4f, Total Wealth: %d", initialMetrics.GiniCoefficient, initialMetrics.TotalWealth)

	// Run 365 days
	for day := 0; day < 365; day++ {
		err := sim.RunDay(context.Background())
		assert.NoError(t, err)

		// Inject some random wealth generation (trading profit)
		for _, m := range sim.merchants {
			profit := rand.Intn(200) // 0-200 daily profit
			m.Wealth += profit
		}
	}

	finalMetrics := sim.GetMetrics()
	t.Logf("Final Gini: %.4f, Total Wealth: %d", finalMetrics.GiniCoefficient, finalMetrics.TotalWealth)

	// Assertions
	// 1. Wealth should grow but not explode (due to sinks)
	assert.Greater(t, finalMetrics.TotalWealth, initialMetrics.TotalWealth)

	// 2. Inequality should not become extreme (Gini < 0.6)
	// Starting random distribution is usually around 0.3-0.4
	assert.Less(t, finalMetrics.GiniCoefficient, 0.6, "Inequality too high")

	// 3. No hyperinflation (implied by wealth check, but explicit price check would be better if we simulated prices)
}

func TestGiniCalculation(t *testing.T) {
	// Perfect equality
	assert.Equal(t, 0.0, CalculateGiniCoefficient([]int{100, 100, 100}))

	// Extreme inequality
	// 1 person has almost everything
	// [0, 0, 100] -> Gini should be close to 0.66 (for n=3, max is (n-1)/n)
	// Formula: (2*300)/(3*100) - 4/3 = 2 - 1.33 = 0.66
	gini := CalculateGiniCoefficient([]int{0, 0, 100})
	assert.InDelta(t, 0.66, gini, 0.01)

	// Standard distribution
	// [10, 20, 30, 40, 50]
	// Mean = 30
	// Sum = 150
	// DiffSum = 1*10 + 2*20 + 3*30 + 4*40 + 5*50 = 10+40+90+160+250 = 550
	// G = (2/5) * (550/150) - 6/5 = 0.4 * 3.66 - 1.2 = 1.466 - 1.2 = 0.266
	gini2 := CalculateGiniCoefficient([]int{10, 20, 30, 40, 50})
	assert.InDelta(t, 0.266, gini2, 0.01)
}
