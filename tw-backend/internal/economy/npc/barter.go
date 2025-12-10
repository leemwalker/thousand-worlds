package npc

import (
	"context"
	"errors"
	"time"

	"tw-backend/internal/economy/market"

	"github.com/google/uuid"
)

var (
	ErrInvalidOffer = errors.New("invalid offer")
	ErrOfferExpired = errors.New("offer expired")
)

// BarterManager handles non-currency trade negotiations
type BarterManager struct {
	marketData MarketDataProvider
}

func NewBarterManager(marketData MarketDataProvider) *BarterManager {
	return &BarterManager{marketData: marketData}
}

// EvaluateOffer determines if an NPC accepts a barter proposal
func (m *BarterManager) EvaluateOffer(ctx context.Context, offer *market.BarterOffer, npcID uuid.UUID) (market.BarterStatus, *market.BarterOffer, error) {
	if offer.ExpiresAt.Before(time.Now()) {
		return market.BarterExpired, nil, ErrOfferExpired
	}

	// 1. Calculate value of offered items (what NPC receives)
	offeredValue := m.calculateTotalValue(offer.OfferedItems, npcID)

	// 2. Calculate value of requested items (what NPC gives)
	requestedValue := m.calculateTotalValue(offer.RequestedItems, npcID)

	// 3. Determine acceptance threshold
	// Base: Offered value must be >= Requested value
	// Modifiers: Relationship, Need

	// Simplified relationship modifier (1.0 = neutral, 0.9 = friend, 1.1 = enemy)
	relationshipMod := 1.0
	// In real system, fetch relationship from RelationshipManager

	requiredValue := float64(requestedValue) * relationshipMod

	// 4. Decision logic
	if float64(offeredValue) >= requiredValue {
		return market.BarterAccepted, nil, nil
	}

	// 5. Counter-offer logic
	// If value is within 20%, propose a counter-offer
	if float64(offeredValue) >= requiredValue*0.8 {
		// Simple counter: ask for more of the first offered item
		// Real logic: ask for specific missing value

		counter := *offer
		counter.OfferID = uuid.New()
		counter.Status = market.BarterCountered
		counter.CounterOffer = offer // Link back to original

		// Naive adjustment: increase quantity of first item by 25%
		if len(counter.OfferedItems) > 0 {
			counter.OfferedItems[0].Quantity = int(float64(counter.OfferedItems[0].Quantity) * 1.25)
			if counter.OfferedItems[0].Quantity == offer.OfferedItems[0].Quantity {
				counter.OfferedItems[0].Quantity++ // Ensure at least +1
			}
		}

		return market.BarterCountered, &counter, nil
	}

	return market.BarterRejected, nil, nil
}

func (m *BarterManager) calculateTotalValue(items []market.ItemStack, locationID uuid.UUID) int {
	total := 0.0
	for _, item := range items {
		price := m.marketData.GetAveragePrice(locationID, item.ItemID)
		// Adjust for quality?
		total += price * float64(item.Quantity)
	}
	return int(total)
}
