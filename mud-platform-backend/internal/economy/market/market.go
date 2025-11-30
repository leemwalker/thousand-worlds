package market

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MarketManager handles market data aggregation and analysis
type MarketManager struct {
	repo Repository
}

// Repository defines data access for market data
type Repository interface {
	GetMarketData(locationID, itemID uuid.UUID) (*MarketData, error)
	UpdateMarketData(data *MarketData) error
	RecordPriceHistory(point *PriceHistory) error
}

// PriceHistory represents a historical price record
type PriceHistory struct {
	LocationID uuid.UUID
	ItemID     uuid.UUID
	Price      float64
	RecordedAt time.Time
}

func NewMarketManager(repo Repository) *MarketManager {
	return &MarketManager{repo: repo}
}

// UpdateMarketData aggregates merchant data to update local market stats
func (m *MarketManager) UpdateMarketData(ctx context.Context, locationID, itemID uuid.UUID, merchants []MerchantSnapshot) error {
	totalPrice := 0.0
	totalSupply := 0
	merchantCount := 0

	for _, merch := range merchants {
		if merch.LocationID != locationID {
			continue
		}

		// Check if merchant sells this item
		if itemData, ok := merch.Inventory[itemID]; ok {
			totalPrice += float64(itemData.Price)
			totalSupply += itemData.Quantity
			merchantCount++
		}
	}

	avgPrice := 0.0
	if merchantCount > 0 {
		avgPrice = totalPrice / float64(merchantCount)
	}

	// Get existing data to preserve demand and history
	data, err := m.repo.GetMarketData(locationID, itemID)
	if err != nil {
		// Create new if not exists
		data = &MarketData{
			LocationID: locationID,
			ItemID:     itemID,
		}
	}

	data.LocalSupply = totalSupply
	data.AveragePrice = avgPrice
	data.LastUpdated = time.Now()

	// Record history
	history := &PriceHistory{
		LocationID: locationID,
		ItemID:     itemID,
		Price:      avgPrice,
		RecordedAt: time.Now(),
	}
	_ = m.repo.RecordPriceHistory(history)

	// Calculate indicators
	// Shortage: if supply < demand
	if data.LocalDemand > 0 {
		ratio := float64(data.LocalSupply) / float64(data.LocalDemand)
		if ratio < 1.0 {
			data.ShortageLevel = 1.0 - ratio
		} else {
			data.ShortageLevel = 0.0
		}
	} else {
		data.ShortageLevel = 0.0
	}

	return m.repo.UpdateMarketData(data)
}

// RecordDemand increments the demand counter for an item
func (m *MarketManager) RecordDemand(ctx context.Context, locationID, itemID uuid.UUID, quantity int) error {
	data, err := m.repo.GetMarketData(locationID, itemID)
	if err != nil {
		data = &MarketData{
			LocationID: locationID,
			ItemID:     itemID,
		}
	}

	data.LocalDemand += quantity
	data.LastUpdated = time.Now()

	return m.repo.UpdateMarketData(data)
}

// MerchantSnapshot represents the relevant state of a merchant for market analysis
type MerchantSnapshot struct {
	LocationID uuid.UUID
	Inventory  map[uuid.UUID]struct {
		Price    int
		Quantity int
	}
}

// Helper method to convert inventory map for snapshot
func CreateMerchantSnapshot(locationID uuid.UUID, inventory map[uuid.UUID]int, prices map[uuid.UUID]int) MerchantSnapshot {
	inv := make(map[uuid.UUID]struct {
		Price    int
		Quantity int
	})

	for itemID, qty := range inventory {
		if price, ok := prices[itemID]; ok {
			inv[itemID] = struct {
				Price    int
				Quantity int
			}{Price: price, Quantity: qty}
		}
	}

	return MerchantSnapshot{
		LocationID: locationID,
		Inventory:  inv,
	}
}
