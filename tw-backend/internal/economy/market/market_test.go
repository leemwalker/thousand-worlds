package market

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetMarketData(locationID, itemID uuid.UUID) (*MarketData, error) {
	args := m.Called(locationID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketData), args.Error(1)
}

func (m *MockRepository) UpdateMarketData(data *MarketData) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockRepository) RecordPriceHistory(point *PriceHistory) error {
	args := m.Called(point)
	return args.Error(0)
}

func TestUpdateMarketData(t *testing.T) {
	repo := new(MockRepository)
	manager := NewMarketManager(repo)

	locationID := uuid.New()
	itemID := uuid.New()

	// Mock merchants
	merchants := []MerchantSnapshot{
		{
			LocationID: locationID,
			Inventory: map[uuid.UUID]struct {
				Price    int
				Quantity int
			}{
				itemID: {Price: 100, Quantity: 10},
			},
		},
		{
			LocationID: locationID,
			Inventory: map[uuid.UUID]struct {
				Price    int
				Quantity int
			}{
				itemID: {Price: 120, Quantity: 5},
			},
		},
	}

	// Expectation: Get existing data (returns nil/error to simulate new)
	repo.On("GetMarketData", locationID, itemID).Return(nil, assert.AnError)

	// Expectation: Update data
	// Avg price = (100 + 120) / 2 = 110
	// Total supply = 10 + 5 = 15
	repo.On("UpdateMarketData", mock.MatchedBy(func(d *MarketData) bool {
		return d.AveragePrice == 110.0 && d.LocalSupply == 15
	})).Return(nil)

	// Expectation: Record history
	repo.On("RecordPriceHistory", mock.MatchedBy(func(h *PriceHistory) bool {
		return h.Price == 110.0
	})).Return(nil)

	err := manager.UpdateMarketData(context.Background(), locationID, itemID, merchants)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestRecordDemand(t *testing.T) {
	repo := new(MockRepository)
	manager := NewMarketManager(repo)

	locationID := uuid.New()
	itemID := uuid.New()

	existingData := &MarketData{
		LocationID:  locationID,
		ItemID:      itemID,
		LocalDemand: 10,
	}

	repo.On("GetMarketData", locationID, itemID).Return(existingData, nil)

	repo.On("UpdateMarketData", mock.MatchedBy(func(d *MarketData) bool {
		return d.LocalDemand == 15 // 10 + 5
	})).Return(nil)

	err := manager.RecordDemand(context.Background(), locationID, itemID, 5)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}
