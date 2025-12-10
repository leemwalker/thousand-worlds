package trade

import (
	"context"
	"errors"
	"time"

	"tw-backend/internal/economy/npc"

	"github.com/google/uuid"
)

var (
	ErrNoProfitableRoutes = errors.New("no profitable routes found")
)

// RoutePlanner finds profitable trade routes
type RoutePlanner struct {
	marketData MarketDataProvider
	mapService MapService
}

// MarketDataProvider interface for dependency injection
type MarketDataProvider interface {
	GetAveragePrice(locationID, itemID uuid.UUID) float64
	GetLocalDemand(locationID, itemID uuid.UUID) int
	GetLocalSupply(locationID, itemID uuid.UUID) int
}

// MapService interface for distance calculations
type MapService interface {
	GetDistance(from, to uuid.UUID) float64
	GetTravelTime(distance float64, speed float64) time.Duration
}

func NewRoutePlanner(marketData MarketDataProvider, mapService MapService) *RoutePlanner {
	return &RoutePlanner{
		marketData: marketData,
		mapService: mapService,
	}
}

// PlanRoute finds the best trade route for a merchant
func (p *RoutePlanner) PlanRoute(ctx context.Context, merchant *npc.Merchant, knownLocations []uuid.UUID, knownItems []uuid.UUID) (*npc.TradeRoute, error) {
	var bestRoute *npc.TradeRoute
	maxProfit := 0

	// Simplified algorithm:
	// 1. For each known destination (different from current)
	// 2. For each known item
	// 3. Check price diff: Buy at current, Sell at dest
	// 4. Calculate travel cost
	// 5. Estimate profit

	for _, destID := range knownLocations {
		if destID == merchant.Location {
			continue
		}

		distance := p.mapService.GetDistance(merchant.Location, destID)
		travelTime := p.mapService.GetTravelTime(distance, 1.0) // Assume speed 1.0
		travelCost := int(distance * 0.5)                       // Cost per unit distance

		var buyItems []npc.TradeItem
		totalProfit := 0
		currentWealth := merchant.Wealth
		remainingCapacity := 100 // Placeholder capacity

		for _, itemID := range knownItems {
			buyPrice := p.marketData.GetAveragePrice(merchant.Location, itemID)
			sellPrice := p.marketData.GetAveragePrice(destID, itemID)

			if buyPrice <= 0 || sellPrice <= 0 {
				continue
			}

			margin := int(sellPrice - buyPrice)
			if margin <= 0 {
				continue
			}

			// Check supply at origin and demand at dest
			supply := p.marketData.GetLocalSupply(merchant.Location, itemID)
			demand := p.marketData.GetLocalDemand(destID, itemID)

			if supply <= 0 {
				continue
			}

			// Determine quantity
			maxBuy := int(float64(currentWealth) / buyPrice)
			qty := min(supply, demand, maxBuy, remainingCapacity)

			if qty > 0 {
				profit := margin * qty
				totalProfit += profit
				currentWealth -= int(buyPrice) * qty
				remainingCapacity -= qty

				buyItems = append(buyItems, npc.TradeItem{
					ItemID:        itemID,
					Quantity:      qty,
					BuyPriceEach:  int(buyPrice),
					SellPriceEach: int(sellPrice),
					Margin:        margin,
				})
			}

			if remainingCapacity <= 0 {
				break
			}
		}

		netProfit := totalProfit - travelCost

		if netProfit > maxProfit && len(buyItems) > 0 {
			maxProfit = netProfit
			bestRoute = &npc.TradeRoute{
				RouteID:         uuid.New(),
				MerchantID:      merchant.NPCID,
				Origin:          merchant.Location,
				Destination:     destID,
				Distance:        distance,
				TravelTime:      travelTime,
				CargoCapacity:   100,
				EstimatedProfit: totalProfit,
				TravelCost:      travelCost,
				NetProfit:       netProfit,
				BuyItems:        buyItems,
				Status:          npc.RoutePlanning,
			}
		}
	}

	if bestRoute == nil {
		return nil, ErrNoProfitableRoutes
	}

	return bestRoute, nil
}

func min(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, v := range nums {
		if v < m {
			m = v
		}
	}
	return m
}
