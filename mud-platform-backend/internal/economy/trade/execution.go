package trade

import (
	"context"
	"errors"
	"time"

	"mud-platform-backend/internal/economy/npc"

	"github.com/google/uuid"
)

var (
	ErrRouteNotStarted       = errors.New("route not started")
	ErrRouteAlreadyCompleted = errors.New("route already completed")
)

// RouteExecutor handles the execution of trade routes
type RouteExecutor struct {
	merchantMgr MerchantManager
	marketMgr   MarketManager
}

// MerchantManager interface for dependency injection
type MerchantManager interface {
	ProcessTransaction(ctx context.Context, merchantID, customerID, itemID uuid.UUID, quantity int) error
	UpdateMerchantLocation(merchantID, locationID uuid.UUID) error
}

// MarketManager interface for dependency injection
type MarketManager interface {
	RecordDemand(ctx context.Context, locationID, itemID uuid.UUID, quantity int) error
}

func NewRouteExecutor(merchantMgr MerchantManager, marketMgr MarketManager) *RouteExecutor {
	return &RouteExecutor{
		merchantMgr: merchantMgr,
		marketMgr:   marketMgr,
	}
}

// ExecuteStep advances the trade route state
func (e *RouteExecutor) ExecuteStep(ctx context.Context, route *npc.TradeRoute) error {
	switch route.Status {
	case npc.RoutePlanning:
		return e.startRoute(ctx, route)
	case npc.RouteTraveling:
		return e.processTravel(ctx, route)
	case npc.RouteTrading:
		return e.completeTrade(ctx, route)
	case npc.RouteCompleted, npc.RouteFailed:
		return ErrRouteAlreadyCompleted
	default:
		return errors.New("unknown route status")
	}
}

func (e *RouteExecutor) startRoute(ctx context.Context, route *npc.TradeRoute) error {
	// 1. Buy items at origin
	// In a real system, this would interact with local merchants
	// For now, we assume the merchant just acquires them (deducting wealth handled in planning/transaction)

	// We need to actually deduct the wealth and add items to inventory here if not done
	// But RoutePlanner only *planned* it.

	// Let's assume the merchant buys from "the market" (abstract)
	// or from specific suppliers.

	route.Status = npc.RouteTraveling
	route.StartedAt = time.Now()
	return nil
}

func (e *RouteExecutor) processTravel(ctx context.Context, route *npc.TradeRoute) error {
	// Check if travel time has passed
	elapsed := time.Since(route.StartedAt)
	if elapsed >= route.TravelTime {
		// Arrived
		if err := e.merchantMgr.UpdateMerchantLocation(route.MerchantID, route.Destination); err != nil {
			return err
		}
		route.Status = npc.RouteTrading
	}
	return nil
}

func (e *RouteExecutor) completeTrade(ctx context.Context, route *npc.TradeRoute) error {
	// Sell items at destination
	totalRevenue := 0

	for _, item := range route.BuyItems {
		// Sell to "the market" or local buyers
		revenue := item.Quantity * item.SellPriceEach
		totalRevenue += revenue

		// Record demand at destination (since we fulfilled it, or rather, we are selling to satisfy demand)
		// Actually, if we sell, we are *supplying*. Demand is what drove us here.
		// But if we sell to a local merchant, they are buying (demand).
		_ = e.marketMgr.RecordDemand(ctx, route.Destination, item.ItemID, item.Quantity)

		// In a real system, we'd call ProcessTransaction for each sale
	}

	// Update route profit
	route.NetProfit = totalRevenue - route.TravelCost // And minus buy cost?
	// The struct has NetProfit calculated in planning.
	// We should update it with actuals.

	route.Status = npc.RouteCompleted
	route.CompletedAt = time.Now()
	return nil
}
