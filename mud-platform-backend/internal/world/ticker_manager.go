package world

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"mud-platform-backend/internal/eventstore"
)

// NATSPublisher interface for publishing messages to NATS
type NATSPublisher interface {
	Publish(subject string, data []byte) error
}

// TickerManager manages world tickers
type TickerManager struct {
	mu            sync.RWMutex
	tickers       map[uuid.UUID]*ticker
	registry      *Registry
	eventStore    eventstore.EventStore
	natsPublisher NATSPublisher
}

type ticker struct {
	worldID        uuid.UUID
	worldName      string
	stopCh         chan struct{}
	tickInterval   time.Duration
	dilationFactor float64
	version        int64     // Event version counter
	pausedAt       time.Time // When ticker was paused
	lastTickAt     time.Time // Last successful tick
}

const DefaultTickInterval = 100 * time.Millisecond

// NewTickerManager creates a new ticker manager
func NewTickerManager(registry *Registry, eventStore eventstore.EventStore, natsPublisher NATSPublisher) *TickerManager {
	return &TickerManager{
		tickers:       make(map[uuid.UUID]*ticker),
		registry:      registry,
		eventStore:    eventStore,
		natsPublisher: natsPublisher,
	}
}

// SpawnTicker creates and starts a new ticker for a world
func (tm *TickerManager) SpawnTicker(worldID uuid.UUID, worldName string, dilationFactor float64) error {
	tm.mu.Lock()

	// Check if ticker already exists
	if _, exists := tm.tickers[worldID]; exists {
		tm.mu.Unlock()
		return fmt.Errorf("ticker for world %s already running", worldID)
	}

	// Create ticker
	t := &ticker{
		worldID:        worldID,
		worldName:      worldName,
		stopCh:         make(chan struct{}),
		tickInterval:   DefaultTickInterval,
		dilationFactor: dilationFactor,
		version:        0,
	}

	tm.tickers[worldID] = t
	tm.mu.Unlock()

	// Register world in registry
	world := &WorldState{
		ID:             worldID,
		Name:           worldName,
		Status:         StatusRunning,
		TickCount:      0,
		GameTime:       0,
		DilationFactor: dilationFactor,
		CreatedAt:      time.Now(),
		LastTickAt:     time.Time{},
	}

	if err := tm.registry.RegisterWorld(world); err != nil {
		// If world already exists, update it instead
		tm.registry.UpdateWorld(worldID, func(w *WorldState) {
			w.Status = StatusRunning
			w.DilationFactor = dilationFactor
		})
	}

	// Emit WorldCreated event
	t.version++
	payload := WorldCreatedPayload{
		WorldID:        worldID.String(),
		Name:           worldName,
		DilationFactor: dilationFactor,
		CreatedAt:      time.Now(),
	}

	if tm.eventStore != nil {
		payloadJSON, _ := json.Marshal(payload)
		event := eventstore.Event{
			ID:            uuid.New().String(),
			EventType:     eventstore.EventType("WorldCreated"),
			AggregateID:   worldID.String(),
			AggregateType: eventstore.AggregateType("World"),
			Version:       t.version,
			Timestamp:     time.Now(),
			Payload:       payloadJSON,
		}
		if err := tm.eventStore.AppendEvent(context.Background(), event); err != nil {
			log.Error().Err(err).Str("world_id", worldID.String()).Msg("Failed to emit WorldCreated event")
		}
	}

	// Check if this is a resume (world exists in registry and was paused)
	existingWorld, err := tm.registry.GetWorld(worldID)
	isResume := err == nil && !existingWorld.PausedAt.IsZero()

	// Start ticker goroutine
	if isResume {
		// Perform catch-up, then run normal ticker
		go tm.performCatchupThenRun(t, existingWorld)
	} else {
		// Normal startup
		go tm.runTicker(t)
	}

	return nil
}

// StopTicker stops a running ticker
func (tm *TickerManager) StopTicker(worldID uuid.UUID) error {
	tm.mu.Lock()

	t, exists := tm.tickers[worldID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("ticker for world %s not running", worldID)
	}

	// Remove from map before unlocking
	delete(tm.tickers, worldID)
	tm.mu.Unlock()

	// Signal stop
	close(t.stopCh)

	// Update registry and record pause time
	var tickCount int64
	pauseTime := time.Now()
	tm.registry.UpdateWorld(worldID, func(w *WorldState) {
		w.Status = StatusPaused
		w.PausedAt = pauseTime
		tickCount = w.TickCount
	})

	// Emit WorldPaused event
	t.version++
	payload := WorldPausedPayload{
		WorldID:   worldID.String(),
		PausedAt:  time.Now(),
		TickCount: tickCount,
	}

	if tm.eventStore != nil {
		payloadJSON, _ := json.Marshal(payload)
		event := eventstore.Event{
			ID:            uuid.New().String(),
			EventType:     eventstore.EventType("WorldPaused"),
			AggregateID:   worldID.String(),
			AggregateType: eventstore.AggregateType("World"),
			Version:       t.version,
			Timestamp:     time.Now(),
			Payload:       payloadJSON,
		}
		if err := tm.eventStore.AppendEvent(context.Background(), event); err != nil {
			log.Error().Err(err).Str("world_id", worldID.String()).Msg("Failed to emit WorldPaused event")
		}
	}

	return nil
}

// GetTickerStatus returns the current status of a ticker
func (tm *TickerManager) GetTickerStatus(worldID uuid.UUID) (running bool, tickCount int64, gameTime time.Duration) {
	tm.mu.RLock()
	_, running = tm.tickers[worldID]
	tm.mu.RUnlock()

	// Get world state from registry
	world, err := tm.registry.GetWorld(worldID)
	if err != nil {
		return false, 0, 0
	}

	return running, world.TickCount, world.GameTime
}

// StopAll stops all running tickers (for graceful shutdown)
func (tm *TickerManager) StopAll() {
	tm.mu.Lock()
	worldIDs := make([]uuid.UUID, 0, len(tm.tickers))
	for worldID := range tm.tickers {
		worldIDs = append(worldIDs, worldID)
	}
	tm.mu.Unlock()

	for _, worldID := range worldIDs {
		if err := tm.StopTicker(worldID); err != nil {
			log.Error().Err(err).Str("world_id", worldID.String()).Msg("Failed to stop ticker during shutdown")
		}
	}
}

// runTicker is the main ticker loop (runs in a goroutine)
func (tm *TickerManager) runTicker(t *ticker) {
	ticker := time.NewTicker(t.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			tm.tick(t)
		}
	}
}

// tick performs a single tick update
func (tm *TickerManager) tick(t *ticker) {
	now := time.Now()

	// Calculate game time delta based on dilation factor
	gameTimeDelta := time.Duration(float64(t.tickInterval) * t.dilationFactor)

	var newTickCount int64
	var newGameTime time.Duration

	// Update world state in registry
	err := tm.registry.UpdateWorld(t.worldID, func(w *WorldState) {
		w.TickCount++
		w.GameTime += gameTimeDelta
		w.LastTickAt = now
		newTickCount = w.TickCount
		newGameTime = w.GameTime
	})

	if err != nil {
		log.Error().Err(err).Str("world_id", t.worldID.String()).Msg("Failed to update world state on tick")
		return
	}

	// Broadcast to NATS
	if tm.natsPublisher != nil {
		// Calculate time of day and season
		sunPos := CalculateSunPosition(newGameTime, DefaultDayLength)
		timeOfDay := GetTimeOfDay(sunPos)
		season, seasonProgress := CalculateSeason(newGameTime, DefaultSeasonLength)

		broadcast := TickBroadcast{
			WorldID:        t.worldID.String(),
			TickNumber:     newTickCount,
			GameTimeMs:     int64(newGameTime / time.Millisecond),
			RealTimeMs:     t.tickInterval.Milliseconds(),
			DilationFactor: t.dilationFactor,
			TimeOfDay:      string(timeOfDay),
			SunPosition:    sunPos,
			CurrentSeason:  string(season),
			SeasonProgress: seasonProgress,
		}
		data, _ := json.Marshal(broadcast)
		subject := fmt.Sprintf("world.tick.%s", t.worldID)
		if err := tm.natsPublisher.Publish(subject, data); err != nil {
			log.Error().Err(err).Str("world_id", t.worldID.String()).Msg("Failed to publish tick to NATS")
		}
	}

	// Emit WorldTicked event
	t.version++
	payload := WorldTickedPayload{
		WorldID:            t.worldID.String(),
		TickCount:          newTickCount,
		GameTimeNs:         int64(newGameTime),
		RealTickDurationMs: t.tickInterval.Milliseconds(),
	}

	if tm.eventStore != nil {
		payloadJSON, _ := json.Marshal(payload)
		event := eventstore.Event{
			ID:            uuid.New().String(),
			EventType:     eventstore.EventType("WorldTicked"),
			AggregateID:   t.worldID.String(),
			AggregateType: eventstore.AggregateType("World"),
			Version:       t.version,
			Timestamp:     time.Now(),
			Payload:       payloadJSON,
		}
		if err := tm.eventStore.AppendEvent(context.Background(), event); err != nil {
			log.Error().Err(err).Str("world_id", t.worldID.String()).Msg("Failed to emit WorldTicked event")
		}
	}
}
