package market

import (
	"time"

	"github.com/google/uuid"
)

// MarketData tracks supply and demand for an item in a location
type MarketData struct {
	LocationID   uuid.UUID
	ItemID       uuid.UUID
	LocalSupply  int // Total quantity available from all merchants
	LocalDemand  int // Purchase attempts in past 7 days
	AveragePrice float64
	PriceHistory []PricePoint
	LastUpdated  time.Time

	// Indicators
	ShortageLevel float64 // 0.0 to 1.0
	InflationRate float64 // % change over 30 days
}

// PricePoint represents a historical price record
type PricePoint struct {
	Date  time.Time
	Price float64
}

// BarterOffer represents a proposed non-currency trade
type BarterOffer struct {
	OfferID             uuid.UUID
	OfferedBy           uuid.UUID
	OfferedTo           uuid.UUID
	OfferedItems        []ItemStack
	RequestedItems      []ItemStack
	TotalOfferedValue   int
	TotalRequestedValue int
	Status              BarterStatus
	CounterOffer        *BarterOffer
	CreatedAt           time.Time
	ExpiresAt           time.Time
}

// ItemStack represents a quantity of items
type ItemStack struct {
	ItemID   uuid.UUID
	Quantity int
	Quality  int // 0-4
}

// BarterStatus defines the state of a barter offer
type BarterStatus string

const (
	BarterPending   BarterStatus = "pending"
	BarterAccepted  BarterStatus = "accepted"
	BarterRejected  BarterStatus = "rejected"
	BarterCountered BarterStatus = "countered"
	BarterExpired   BarterStatus = "expired"
)
