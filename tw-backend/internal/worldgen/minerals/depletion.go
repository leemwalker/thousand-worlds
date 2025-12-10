package minerals

import (
	"time"

	"github.com/google/uuid"
)

// DepletionHistory tracks the extraction of a deposit
type DepletionHistory struct {
	DepositID        uuid.UUID
	OriginalQuantity int
	CurrentQuantity  int
	ExtractedBy      []uuid.UUID // NPCs/Players who mined
	FirstExtracted   time.Time
	DepletedAt       *time.Time // null if not depleted
	ExtractionRate   float64    // Units per day (average)
}

// NewDepletionHistory creates a new tracking record
func NewDepletionHistory(deposit *MineralDeposit) *DepletionHistory {
	return &DepletionHistory{
		DepositID:        deposit.DepositID,
		OriginalQuantity: deposit.Quantity,
		CurrentQuantity:  deposit.Quantity,
		ExtractedBy:      make([]uuid.UUID, 0),
		FirstExtracted:   time.Now(),
	}
}

// Extract removes minerals from the deposit
func (h *DepletionHistory) Extract(amount int, minerID uuid.UUID) int {
	extracted := amount
	if h.CurrentQuantity < amount {
		extracted = h.CurrentQuantity
	}

	h.CurrentQuantity -= extracted

	// Track miner if new
	known := false
	for _, id := range h.ExtractedBy {
		if id == minerID {
			known = true
			break
		}
	}
	if !known {
		h.ExtractedBy = append(h.ExtractedBy, minerID)
	}

	// Check depletion
	if h.CurrentQuantity <= 0 {
		h.CurrentQuantity = 0
		now := time.Now()
		h.DepletedAt = &now
	}

	return extracted
}

// IsDepleted returns true if the deposit is empty
func (h *DepletionHistory) IsDepleted() bool {
	return h.CurrentQuantity <= 0
}
