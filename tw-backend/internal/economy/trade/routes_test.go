package trade

import (
	"context"
	"testing"
	"time"

	"mud-platform-backend/internal/economy/npc"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockMarketDataProvider struct {
	mock.Mock
}

func (m *MockMarketDataProvider) GetAveragePrice(locationID, itemID uuid.UUID) float64 {
	args := m.Called(locationID, itemID)
	return args.Get(0).(float64)
}

func (m *MockMarketDataProvider) GetLocalDemand(locationID, itemID uuid.UUID) int {
	args := m.Called(locationID, itemID)
	return args.Int(0)
}

func (m *MockMarketDataProvider) GetLocalSupply(locationID, itemID uuid.UUID) int {
	args := m.Called(locationID, itemID)
	return args.Int(0)
}

type MockMapService struct {
	mock.Mock
}

func (m *MockMapService) GetDistance(from, to uuid.UUID) float64 {
	args := m.Called(from, to)
	return args.Get(0).(float64)
}

func (m *MockMapService) GetTravelTime(distance float64, speed float64) time.Duration {
	args := m.Called(distance, speed)
	return args.Get(0).(time.Duration)
}

func TestPlanRoute(t *testing.T) {
	marketData := new(MockMarketDataProvider)
	mapService := new(MockMapService)
	planner := NewRoutePlanner(marketData, mapService)

	merchantID := uuid.New()
	originID := uuid.New()
	destID := uuid.New()
	itemID := uuid.New()

	merchant := &npc.Merchant{
		NPCID:    merchantID,
		Location: originID,
		Wealth:   1000,
	}

	knownLocations := []uuid.UUID{destID}
	knownItems := []uuid.UUID{itemID}

	// Expectations
	mapService.On("GetDistance", originID, destID).Return(100.0)
	mapService.On("GetTravelTime", 100.0, 1.0).Return(1 * time.Hour)

	// Buy at origin: 10
	marketData.On("GetAveragePrice", originID, itemID).Return(10.0)
	marketData.On("GetLocalSupply", originID, itemID).Return(50)

	// Sell at dest: 20
	marketData.On("GetAveragePrice", destID, itemID).Return(20.0)
	marketData.On("GetLocalDemand", destID, itemID).Return(20)

	route, err := planner.PlanRoute(context.Background(), merchant, knownLocations, knownItems)

	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, originID, route.Origin)
	assert.Equal(t, destID, route.Destination)
	assert.Equal(t, 1, len(route.BuyItems))

	// Check quantity: min(supply=50, demand=20, afford=1000/10=100, cap=100) -> 20
	assert.Equal(t, 20, route.BuyItems[0].Quantity)

	// Profit: (20 - 10) * 20 = 200
	// Travel cost: 100 * 0.5 = 50
	// Net: 150
	assert.Equal(t, 150, route.NetProfit)
}

func TestPlanRoute_NoProfitableRoutes(t *testing.T) {
	marketData := new(MockMarketDataProvider)
	mapService := new(MockMapService)
	planner := NewRoutePlanner(marketData, mapService)

	merchant := &npc.Merchant{
		Location: uuid.New(),
		Wealth:   1000,
	}

	destID := uuid.New()
	itemID := uuid.New()

	mapService.On("GetDistance", mock.Anything, mock.Anything).Return(100.0)
	mapService.On("GetTravelTime", mock.Anything, mock.Anything).Return(1 * time.Hour)

	// Buy high, sell low
	marketData.On("GetAveragePrice", merchant.Location, itemID).Return(20.0)
	marketData.On("GetAveragePrice", destID, itemID).Return(10.0)

	route, err := planner.PlanRoute(context.Background(), merchant, []uuid.UUID{destID}, []uuid.UUID{itemID})

	assert.Error(t, err)
	assert.Equal(t, ErrNoProfitableRoutes, err)
	assert.Nil(t, route)
}
