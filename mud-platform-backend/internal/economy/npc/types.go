package npc

import (
	"time"

	"github.com/google/uuid"
)

// Merchant represents an NPC that trades goods
type Merchant struct {
	NPCID          uuid.UUID
	ShopName       string
	Specialization string    // blacksmith, general_goods, etc.
	Location       uuid.UUID // Location ID (settlement/zone)
	Wealth         int
	PriceModifier  float64
	Reputation     float64

	// Business hours
	OpeningHour int
	ClosingHour int
	DaysOpen    []time.Weekday

	// Relationships
	Suppliers   []uuid.UUID
	Competitors []uuid.UUID

	// Data tracking
	SalesHistory    []SaleRecord
	PurchaseHistory []PurchaseRecord
}

// SaleRecord tracks items sold by the merchant
type SaleRecord struct {
	Timestamp  time.Time
	ItemID     uuid.UUID
	Quantity   int
	PricePer   int
	CustomerID uuid.UUID
	Profit     int
}

// PurchaseRecord tracks items bought by the merchant
type PurchaseRecord struct {
	Timestamp  time.Time
	ItemID     uuid.UUID
	Quantity   int
	PricePer   int
	SupplierID uuid.UUID
}

// TradeRoute represents a planned journey to trade goods
type TradeRoute struct {
	RouteID       uuid.UUID
	MerchantID    uuid.UUID
	Origin        uuid.UUID // Location ID
	Destination   uuid.UUID // Location ID
	Distance      float64
	TravelTime    time.Duration
	CargoCapacity int

	// Economics
	EstimatedProfit int
	TravelCost      int
	NetProfit       int

	// Items
	BuyItems  []TradeItem
	SellItems []TradeItem

	// Status
	Status      RouteStatus
	StartedAt   time.Time
	CompletedAt time.Time
}

// TradeItem represents an item in a trade route
type TradeItem struct {
	ItemID        uuid.UUID
	Quantity      int
	BuyPriceEach  int
	SellPriceEach int
	Margin        int
}

// RouteStatus defines the current state of a trade route
type RouteStatus string

const (
	RoutePlanning  RouteStatus = "planning"
	RouteTraveling RouteStatus = "traveling"
	RouteTrading   RouteStatus = "trading"
	RouteReturning RouteStatus = "returning"
	RouteCompleted RouteStatus = "completed"
	RouteFailed    RouteStatus = "failed"
)
