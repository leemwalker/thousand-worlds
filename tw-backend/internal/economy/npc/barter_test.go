package npc

import (
	"context"
	"testing"
	"time"

	"mud-platform-backend/internal/economy/market"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateOffer_Accept(t *testing.T) {
	marketData := new(MockMarketData)
	manager := NewBarterManager(marketData)

	npcID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()

	// Offer: 10 items worth 10 each (100 total)
	// Request: 5 items worth 15 each (75 total)
	// Should accept

	offer := &market.BarterOffer{
		OfferID:   uuid.New(),
		OfferedBy: uuid.New(),
		OfferedTo: npcID,
		OfferedItems: []market.ItemStack{
			{ItemID: itemID1, Quantity: 10},
		},
		RequestedItems: []market.ItemStack{
			{ItemID: itemID2, Quantity: 5},
		},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	marketData.On("GetAveragePrice", npcID, itemID1).Return(10.0)
	marketData.On("GetAveragePrice", npcID, itemID2).Return(15.0)

	status, counter, err := manager.EvaluateOffer(context.Background(), offer, npcID)

	assert.NoError(t, err)
	assert.Equal(t, market.BarterAccepted, status)
	assert.Nil(t, counter)
}

func TestEvaluateOffer_Reject(t *testing.T) {
	marketData := new(MockMarketData)
	manager := NewBarterManager(marketData)

	npcID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()

	// Offer: 10 items worth 5 each (50 total)
	// Request: 5 items worth 20 each (100 total)
	// Should reject (50 < 80% of 100)

	offer := &market.BarterOffer{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		OfferedItems: []market.ItemStack{
			{ItemID: itemID1, Quantity: 10},
		},
		RequestedItems: []market.ItemStack{
			{ItemID: itemID2, Quantity: 5},
		},
	}

	marketData.On("GetAveragePrice", npcID, itemID1).Return(5.0)
	marketData.On("GetAveragePrice", npcID, itemID2).Return(20.0)

	status, _, err := manager.EvaluateOffer(context.Background(), offer, npcID)

	assert.NoError(t, err)
	assert.Equal(t, market.BarterRejected, status)
}

func TestEvaluateOffer_Counter(t *testing.T) {
	marketData := new(MockMarketData)
	manager := NewBarterManager(marketData)

	npcID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()

	// Offer: 10 items worth 9 each (90 total)
	// Request: 5 items worth 20 each (100 total)
	// Should counter (90 >= 80% of 100)

	offer := &market.BarterOffer{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		OfferedItems: []market.ItemStack{
			{ItemID: itemID1, Quantity: 10},
		},
		RequestedItems: []market.ItemStack{
			{ItemID: itemID2, Quantity: 5},
		},
	}

	marketData.On("GetAveragePrice", npcID, itemID1).Return(9.0)
	marketData.On("GetAveragePrice", npcID, itemID2).Return(20.0)

	status, counter, err := manager.EvaluateOffer(context.Background(), offer, npcID)

	assert.NoError(t, err)
	assert.Equal(t, market.BarterCountered, status)
	assert.NotNil(t, counter)
	// Counter should ask for more of itemID1
	assert.Greater(t, counter.OfferedItems[0].Quantity, 10)
}
