package simulation

import (
	"context"

	"tw-backend/internal/economy/npc"
)

// SimulationManager runs the economic simulation
type SimulationManager struct {
	merchants []*npc.Merchant
	// In a real system, this would interact with repositories
}

func NewSimulationManager() *SimulationManager {
	return &SimulationManager{
		merchants: make([]*npc.Merchant, 0),
	}
}

// AddMerchant adds a merchant to the simulation
func (s *SimulationManager) AddMerchant(m *npc.Merchant) {
	s.merchants = append(s.merchants, m)
}

// RunDay simulates one day of economic activity
func (s *SimulationManager) RunDay(ctx context.Context) error {
	// 1. Merchants restock
	// 2. Merchants trade
	// 3. Prices update
	// 4. Wealth changes

	// Simplified simulation logic for testing stability
	for _, m := range s.merchants {
		// Simulate random sales
		// m.Wealth += ...

		// Simulate expenses (taxes, spoilage, living costs)
		// Money sink
		tax := int(float64(m.Wealth) * 0.01) // 1% daily tax
		m.Wealth -= tax

		// Resource sink (spoilage)
		// ...
	}

	return nil
}

// GetMetrics returns current economic indicators
func (s *SimulationManager) GetMetrics() EconomicMetrics {
	wealths := make([]int, len(s.merchants))
	totalWealth := 0

	for i, m := range s.merchants {
		wealths[i] = m.Wealth
		totalWealth += m.Wealth
	}

	return EconomicMetrics{
		TotalWealth:     totalWealth,
		MoneySupply:     totalWealth, // Simplified
		GiniCoefficient: CalculateGiniCoefficient(wealths),
		ActiveMerchants: len(s.merchants),
	}
}
