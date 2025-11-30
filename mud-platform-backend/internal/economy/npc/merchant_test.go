package npc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockMerchantRepository struct {
	mock.Mock
}

func (m *MockMerchantRepository) GetMerchant(npcID uuid.UUID) (*Merchant, error) {
	args := m.Called(npcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Merchant), args.Error(1)
}

func (m *MockMerchantRepository) UpdateMerchant(merchant *Merchant) error {
	args := m.Called(merchant)
	return args.Error(0)
}

func (m *MockMerchantRepository) RecordSale(sale *SaleRecord) error {
	args := m.Called(sale)
	return args.Error(0)
}

func (m *MockMerchantRepository) RecordPurchase(purchase *PurchaseRecord) error {
	args := m.Called(purchase)
	return args.Error(0)
}

type MockMarketData struct {
	mock.Mock
}

func (m *MockMarketData) GetAveragePrice(locationID, itemID uuid.UUID) float64 {
	args := m.Called(locationID, itemID)
	return args.Get(0).(float64)
}

func (m *MockMarketData) GetLocalDemand(locationID, itemID uuid.UUID) int {
	args := m.Called(locationID, itemID)
	return args.Int(0)
}

func TestProcessTransaction(t *testing.T) {
	repo := new(MockMerchantRepository)
	manager := NewMerchantManager(repo)

	merchantID := uuid.New()
	customerID := uuid.New()
	itemID := uuid.New()

	merchant := &Merchant{
		NPCID:         merchantID,
		Wealth:        1000,
		PriceModifier: 1.0,
		SalesHistory:  []SaleRecord{},
	}

	repo.On("GetMerchant", merchantID).Return(merchant, nil)
	repo.On("UpdateMerchant", mock.MatchedBy(func(m *Merchant) bool {
		return m.Wealth > 1000 && len(m.SalesHistory) == 1
	})).Return(nil)
	repo.On("RecordSale", mock.Anything).Return(nil)

	err := manager.ProcessTransaction(context.Background(), merchantID, customerID, itemID, 1)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestRestock(t *testing.T) {
	repo := new(MockMerchantRepository)
	manager := NewMerchantManager(repo)
	marketData := new(MockMarketData)

	merchantID := uuid.New()
	itemID := uuid.New()

	// History shows 10 items sold recently
	salesHistory := []SaleRecord{
		{
			Timestamp: time.Now().AddDate(0, 0, -1),
			ItemID:    itemID,
			Quantity:  10,
		},
	}

	merchant := &Merchant{
		NPCID:        merchantID,
		Wealth:       1000,
		SalesHistory: salesHistory,
	}

	repo.On("GetMerchant", merchantID).Return(merchant, nil)

	// Expect purchase: target 15 (1.5 * 10), current 0 -> buy 15
	repo.On("RecordPurchase", mock.MatchedBy(func(p *PurchaseRecord) bool {
		return p.ItemID == itemID && p.Quantity == 15
	})).Return(nil)

	repo.On("UpdateMerchant", mock.MatchedBy(func(m *Merchant) bool {
		return m.Wealth < 1000 // Spent money
	})).Return(nil)

	err := manager.Restock(context.Background(), merchantID, marketData)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestUpdatePrices(t *testing.T) {
	repo := new(MockMerchantRepository)
	manager := NewMerchantManager(repo)
	marketData := new(MockMarketData)

	merchantID := uuid.New()

	// High volume sales
	salesHistory := []SaleRecord{
		{
			Timestamp: time.Now(),
			Quantity:  60,
		},
	}

	merchant := &Merchant{
		NPCID:         merchantID,
		PriceModifier: 1.0,
		SalesHistory:  salesHistory,
	}

	repo.On("GetMerchant", merchantID).Return(merchant, nil)

	// Expect price increase
	repo.On("UpdateMerchant", mock.MatchedBy(func(m *Merchant) bool {
		return m.PriceModifier > 1.0
	})).Return(nil)

	err := manager.UpdatePrices(context.Background(), merchantID, marketData)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}
