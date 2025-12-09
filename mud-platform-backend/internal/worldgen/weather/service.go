package weather

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service handles weather updates and retrieval
type Service struct {
	repo       Repository
	stateCache map[uuid.UUID]map[uuid.UUID]*WeatherState // worldID -> cellID -> State
	geoCache   map[uuid.UUID][]*GeographyCell            // worldID -> cells
	cacheMutex sync.RWMutex
}

// NewService creates a new weather service
func NewService(repo Repository) *Service {
	return &Service{
		repo:       repo,
		stateCache: make(map[uuid.UUID]map[uuid.UUID]*WeatherState),
		geoCache:   make(map[uuid.UUID][]*GeographyCell),
	}
}

// UpdateWorldWeather updates weather for all cells in a world
func (s *Service) UpdateWorldWeather(ctx context.Context, worldID uuid.UUID, currentTime time.Time, season Season) (map[uuid.UUID]string, error) {
	s.cacheMutex.RLock()
	cells, ok := s.geoCache[worldID]
	s.cacheMutex.RUnlock()

	if !ok || len(cells) == 0 {
		return nil, fmt.Errorf("no geography data found for world %s", worldID)
	}

	// Calculate new states
	newStates := UpdateWeather(cells, currentTime, season)

	// Persist states and detect changes
	emotes := make(map[uuid.UUID]string)

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	worldCache, ok := s.stateCache[worldID]
	if !ok {
		worldCache = make(map[uuid.UUID]*WeatherState)
		s.stateCache[worldID] = worldCache
	}

	for _, newState := range newStates {
		// Save to DB (async? for now sync to be safe)
		if err := s.repo.SaveWeatherState(ctx, newState); err != nil {
			return nil, fmt.Errorf("failed to save weather state: %w", err)
		}

		// Check for changes
		oldState, exists := worldCache[newState.CellID]
		if exists {
			if emote := s.detectWeatherChange(oldState, newState); emote != "" {
				emotes[newState.CellID] = emote
			}
		}

		// Update cache
		worldCache[newState.CellID] = newState
	}

	return emotes, nil
}

// GetCurrentWeather retrieves the latest weather state for a cell
func (s *Service) GetCurrentWeather(ctx context.Context, worldID, cellID uuid.UUID) (*WeatherState, error) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if worldCache, ok := s.stateCache[worldID]; ok {
		if state, ok := worldCache[cellID]; ok {
			return state, nil
		}
	}

	// Fallback to DB if not in cache (could implement if needed, but cache should be primed by UpdateWorld)
	// For now, return nil if not found
	return nil, nil // Or specific error
}

// InitializeWorldWeather loads initial weather states and geography into the cache
func (s *Service) InitializeWorldWeather(ctx context.Context, worldID uuid.UUID, states []*WeatherState, cells []*GeographyCell) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Initialize state cache
	worldCache, ok := s.stateCache[worldID]
	if !ok {
		worldCache = make(map[uuid.UUID]*WeatherState)
		s.stateCache[worldID] = worldCache
	}

	for _, state := range states {
		worldCache[state.CellID] = state
	}

	// Initialize geography cache
	s.geoCache[worldID] = cells
}

// detectWeatherChange returns an emote string if the weather has changed significantly
func (s *Service) detectWeatherChange(old, new *WeatherState) string {
	if old.State != new.State {
		switch new.State {
		case WeatherClear:
			return "The clouds part, revealing a clear sky."
		case WeatherCloudy:
			return "Clouds gather overhead, obscuring the sun."
		case WeatherRain:
			return "Rain begins to fall from the grey sky."
		case WeatherStorm:
			return "The wind howls as a storm breaks overhead!"
		case WeatherSnow:
			return "Snowflakes begin to drift down gently."
		}
	}

	// Wind changes
	if math.Abs(new.Wind.Speed-old.Wind.Speed) > 10 {
		if new.Wind.Speed > old.Wind.Speed {
			return "The wind picks up intensity."
		} else {
			return "The wind dies down."
		}
	}

	return ""
}
