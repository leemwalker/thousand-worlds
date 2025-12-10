package npc

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrItemNotFound      = errors.New("item not found in inventory")
)

// MerchantManager handles merchant operations
type MerchantManager struct {
	repo MerchantRepository
}

// MerchantRepository defines data access for merchants
type MerchantRepository interface {
	GetMerchant(npcID uuid.UUID) (*Merchant, error)
	UpdateMerchant(merchant *Merchant) error
	RecordSale(sale *SaleRecord) error
	RecordPurchase(purchase *PurchaseRecord) error
}

// NewMerchantManager creates a new manager
func NewMerchantManager(repo MerchantRepository) *MerchantManager {
	return &MerchantManager{repo: repo}
}

// ProcessTransaction handles a player buying from a merchant
func (m *MerchantManager) ProcessTransaction(ctx context.Context, merchantID, customerID, itemID uuid.UUID, quantity int) error {
	merchant, err := m.repo.GetMerchant(merchantID)
	if err != nil {
		return err
	}

	// 1. Check availability (simplified: assume merchant inventory is tracked elsewhere or part of Merchant struct)
	// For this implementation, we'll assume the Merchant struct has an Inventory map
	// But the current Merchant struct in types.go doesn't have it explicitly.
	// Let's assume we use the same InventoryManager as NPCs.

	// 2. Calculate price
	basePrice := 100 // Placeholder, would come from item definition
	price := int(float64(basePrice) * merchant.PriceModifier)
	totalCost := price * quantity

	// 3. Record sale
	sale := &SaleRecord{
		Timestamp:  time.Now(),
		ItemID:     itemID,
		Quantity:   quantity,
		PricePer:   price,
		CustomerID: customerID,
		Profit:     totalCost / 2, // Simplified profit calculation
	}

	merchant.SalesHistory = append(merchant.SalesHistory, *sale)
	merchant.Wealth += totalCost

	if err := m.repo.UpdateMerchant(merchant); err != nil {
		return err
	}

	return m.repo.RecordSale(sale)
}

// Restock analyzes sales history to determine what to buy
func (m *MerchantManager) Restock(ctx context.Context, merchantID uuid.UUID, marketData MarketDataProvider) error {
	merchant, err := m.repo.GetMerchant(merchantID)
	if err != nil {
		return err
	}

	// 1. Analyze sales history (last 7 days)
	cutoff := time.Now().AddDate(0, 0, -7)
	demandMap := make(map[uuid.UUID]int)

	for _, sale := range merchant.SalesHistory {
		if sale.Timestamp.After(cutoff) {
			demandMap[sale.ItemID] += sale.Quantity
		}
	}

	// 2. Determine restock needs
	// Target stock = 1.5 * weekly demand
	for itemID, demand := range demandMap {
		targetStock := int(float64(demand) * 1.5)
		currentStock := 0 // Placeholder: get from inventory

		needed := targetStock - currentStock
		if needed > 0 {
			// 3. Purchase logic (simplified)
			cost := 50 // Placeholder base cost
			totalCost := cost * needed

			if merchant.Wealth >= totalCost {
				merchant.Wealth -= totalCost
				// Add to inventory...

				purchase := &PurchaseRecord{
					Timestamp: time.Now(),
					ItemID:    itemID,
					Quantity:  needed,
					PricePer:  cost,
				}
				merchant.PurchaseHistory = append(merchant.PurchaseHistory, *purchase)
				_ = m.repo.RecordPurchase(purchase)
			}
		}
	}

	return m.repo.UpdateMerchant(merchant)
}

// UpdatePrices adjusts price modifiers based on supply and demand
func (m *MerchantManager) UpdatePrices(ctx context.Context, merchantID uuid.UUID, marketData MarketDataProvider) error {
	merchant, err := m.repo.GetMerchant(merchantID)
	if err != nil {
		return err
	}

	// 1. Get local market conditions
	// Simplified: just adjust global modifier for now
	// Real implementation would have per-item pricing

	// If wealth is low, lower prices to drive volume? Or raise to increase margin?
	// Standard logic:
	// High demand -> Raise prices
	// Low demand -> Lower prices

	// Let's use a simple heuristic based on recent sales volume vs inventory
	// If sales > expected, increase price modifier

	recentSales := 0
	cutoff := time.Now().AddDate(0, 0, -3)
	for _, sale := range merchant.SalesHistory {
		if sale.Timestamp.After(cutoff) {
			recentSales += sale.Quantity
		}
	}

	// Arbitrary threshold for "high volume"
	if recentSales > 50 {
		merchant.PriceModifier = math.Min(merchant.PriceModifier+0.05, 2.0)
	} else if recentSales < 10 {
		merchant.PriceModifier = math.Max(merchant.PriceModifier-0.05, 0.5)
	}

	return m.repo.UpdateMerchant(merchant)
}
