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
	"mud-platform-backend/internal/spatial"
	"mud-platform-backend/internal/worldgen/weather"
)

// NATSPublisher interface for publishing messages to NATS
type NATSPublisher interface {
	Publish(subject string, data []byte) error
}

// AreaBroadcaster interface for broadcasting messages to spatial areas
type AreaBroadcaster interface {
	BroadcastToArea(center spatial.Position, radius float64, msgType string, data interface{})
}

// TickerManager manages world tickers
type TickerManager struct {
	mu             sync.RWMutex
	tickers        map[uuid.UUID]*ticker
	registry       *Registry
	eventStore     eventstore.EventStore
	natsPublisher  NATSPublisher
	weatherService *weather.Service
	broadcaster    AreaBroadcaster
}

type ticker struct {
	worldID             uuid.UUID
	worldName           string
	stopCh              chan struct{}
	tickInterval        time.Duration
	dilationFactor      float64
	version             int64         // Event version counter
	pausedAt            time.Time     // When ticker was paused
	lastTickAt          time.Time     // Last successful tick
	lastWeatherGameTime time.Duration // Game time of last weather update
}

const DefaultTickInterval = 100 * time.Millisecond

// NewTickerManager creates a new ticker manager
func NewTickerManager(registry *Registry, eventStore eventstore.EventStore, natsPublisher NATSPublisher, weatherService *weather.Service, broadcaster AreaBroadcaster) *TickerManager {
	return &TickerManager{
		tickers:        make(map[uuid.UUID]*ticker),
		registry:       registry,
		eventStore:     eventStore,
		natsPublisher:  natsPublisher,
		weatherService: weatherService,
		broadcaster:    broadcaster,
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

	// Update Weather (every 30 game minutes)
	if tm.weatherService != nil && (newGameTime-t.lastWeatherGameTime >= 30*time.Minute || t.lastWeatherGameTime == 0) {
		currentSeason, _ := CalculateSeason(newGameTime, DefaultSeasonLength)
		// We need to map time.Duration to time.Time for day/night calc in weather?
		// Weather uses time.Time effectively for seasonality but passed separately.
		// UpdateWorldWeather uses time for diurnal temp variation. We should synthesize a time.
		// World CreatedAt + GameTime?
		// But UpdateWeather logic uses time.Hour() for diurnal.
		// So we construct a fake time or used stored CreatedAt + GameTime.
		// We don't have CreatedAt easily here on ticker struct, need to fetch from registry or add to ticker.
		// It was added to WorldState. Let's assume we can get it or just use reference time.
		// Ticker does not have CreatedAt. But we can ignore exact date if we just want time of day.
		// 0 game time = 00:00?
		// Let's assume GameTime starts at some point.

		// Simplest: time.Unix(0, 0).Add(newGameTime)
		calcTime := time.Unix(0, 0).Add(newGameTime)

		emotes, err := tm.weatherService.UpdateWorldWeather(context.Background(), t.worldID, calcTime, weather.Season(currentSeason))
		if err != nil {
			log.Error().Err(err).Str("world_id", t.worldID.String()).Msg("Failed to update weather")
		} else {
			t.lastWeatherGameTime = newGameTime

			// Broadcast emotes
			if tm.broadcaster != nil && len(emotes) > 0 {
				// We need cell locations to broadcast to area.
				// Emotes map is cellID -> text.
				// We need cellID -> location.
				// WeatherService has this internally but doesn't expose it easily in return.
				// Iterate all cells? Or WeatherService should return map[cellID]struct{Text, Location}?
				// Re-fetching per cell is expensive if we don't have location.
				// But we can get GeographyCell?
				// WeatherService could just broadcast itself if we passed broadcaster to it?
				// Or return `[]WeatherEvent` where event has Location.

				// For now, let's assume we can't broadcast efficiently without location.
				// I'll update WeatherService to return map[uuid.UUID]WeatherEvent
				// type WeatherEvent struct { Text string; Location geography.Point }
				// Checking WeatherService return... currently map[uuid.UUID]string.

				// Quick fix: Retrieve location from WeatherService? No method.
				// Iterate stateCache? No access.
				// I should update WeatherService.UpdateWorldWeather to return more info.

				// Since I'm in TickerManager, I'll defer this and update WeatherService return type in next step.
				// For this file update, I'll comment the broadcast part or use placeholder.
			}
		}
	}
}
