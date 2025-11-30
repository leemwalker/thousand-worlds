package minerals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDiscoveryChance(t *testing.T) {
	deposit := &MineralDeposit{
		SurfaceVisible: false,
		Depth:          500,
		VeinSize:       VeinSizeMedium,
	}

	t.Run("Surface Visible High Chance", func(t *testing.T) {
		visibleDeposit := &MineralDeposit{
			SurfaceVisible: true,
		}
		chance := CalculateDiscoveryChance(visibleDeposit, 10, 10, 4)
		assert.True(t, chance > 0.8, "Surface deposits should have high discovery chance")
	})

	t.Run("Skill Impact", func(t *testing.T) {
		lowSkillChance := CalculateDiscoveryChance(deposit, 0, 0, 4)
		highSkillChance := CalculateDiscoveryChance(deposit, 100, 100, 4)
		assert.True(t, highSkillChance > lowSkillChance)
	})

	t.Run("Time Impact", func(t *testing.T) {
		quickChance := CalculateDiscoveryChance(deposit, 50, 50, 1)
		thoroughChance := CalculateDiscoveryChance(deposit, 50, 50, 8)
		assert.True(t, thoroughChance > quickChance)
	})

	t.Run("Depth Impact", func(t *testing.T) {
		shallowDeposit := &MineralDeposit{
			SurfaceVisible: false,
			Depth:          100,
			VeinSize:       VeinSizeMedium,
		}
		deepDeposit := &MineralDeposit{
			SurfaceVisible: false,
			Depth:          3000,
			VeinSize:       VeinSizeMedium,
		}
		shallowChance := CalculateDiscoveryChance(shallowDeposit, 50, 50, 4)
		deepChance := CalculateDiscoveryChance(deepDeposit, 50, 50, 4)
		assert.True(t, shallowChance > deepChance)
	})

	t.Run("Size Impact", func(t *testing.T) {
		smallDeposit := &MineralDeposit{
			SurfaceVisible: false,
			Depth:          500,
			VeinSize:       VeinSizeSmall,
		}
		massiveDeposit := &MineralDeposit{
			SurfaceVisible: false,
			Depth:          500,
			VeinSize:       VeinSizeMassive,
		}
		smallChance := CalculateDiscoveryChance(smallDeposit, 50, 50, 4)
		massiveChance := CalculateDiscoveryChance(massiveDeposit, 50, 50, 4)
		assert.True(t, massiveChance > smallChance)
	})
}

func TestIsDiscovered(t *testing.T) {
	// Test probabilistic nature
	highChanceDiscovered := 0
	for i := 0; i < 100; i++ {
		if IsDiscovered(0.9) {
			highChanceDiscovered++
		}
	}
	assert.True(t, highChanceDiscovered > 70, "90% chance should discover most of the time")

	lowChanceDiscovered := 0
	for i := 0; i < 100; i++ {
		if IsDiscovered(0.1) {
			lowChanceDiscovered++
		}
	}
	assert.True(t, lowChanceDiscovered < 30, "10% chance should rarely discover")
}
