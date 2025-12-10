package minerals

import (
	"testing"

	"mud-platform-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDepletionHistory(t *testing.T) {
	deposit := &MineralDeposit{
		DepositID: uuid.New(),
		Quantity:  1000,
		Location:  geography.Point{X: 0, Y: 0},
	}

	history := NewDepletionHistory(deposit)

	t.Run("Initial State", func(t *testing.T) {
		assert.Equal(t, deposit.DepositID, history.DepositID)
		assert.Equal(t, 1000, history.OriginalQuantity)
		assert.Equal(t, 1000, history.CurrentQuantity)
		assert.False(t, history.IsDepleted())
		assert.Nil(t, history.DepletedAt)
	})

	t.Run("Partial Extraction", func(t *testing.T) {
		miner1 := uuid.New()
		extracted := history.Extract(300, miner1)
		assert.Equal(t, 300, extracted)
		assert.Equal(t, 700, history.CurrentQuantity)
		assert.False(t, history.IsDepleted())
		assert.Contains(t, history.ExtractedBy, miner1)
	})

	t.Run("Multiple Miners", func(t *testing.T) {
		miner2 := uuid.New()
		history.Extract(200, miner2)
		assert.Equal(t, 2, len(history.ExtractedBy))
	})

	t.Run("Over Extraction", func(t *testing.T) {
		// Attempt to extract more than available
		miner3 := uuid.New()
		extracted := history.Extract(1000, miner3)
		assert.Equal(t, 500, extracted) // Only 500 left
		assert.True(t, history.IsDepleted())
		assert.NotNil(t, history.DepletedAt)
		assert.Equal(t, 0, history.CurrentQuantity)
	})

	t.Run("Cannot Extract From Depleted", func(t *testing.T) {
		miner4 := uuid.New()
		extracted := history.Extract(100, miner4)
		assert.Equal(t, 0, extracted)
		assert.True(t, history.IsDepleted())
	})
}
