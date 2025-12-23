package weather

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"tw-backend/internal/spatial"

	"github.com/google/uuid"
)

// Service handles weather updates and retrieval
type Service struct {
	repo       Repository
	stateCache map[uuid.UUID]map[uuid.UUID]*WeatherState // worldID -> cellID -> State
	geoCache   map[uuid.UUID][]*GeographyCell            // worldID -> cells
	cacheMutex sync.RWMutex
	topology   spatial.Topology // Optional: nil = flat mode
}

// NewService creates a new weather service
func NewService(repo Repository) *Service {
	return &Service{
		repo:       repo,
		stateCache: make(map[uuid.UUID]map[uuid.UUID]*WeatherState),
		geoCache:   make(map[uuid.UUID][]*GeographyCell),
	}
}

// WithTopology sets the spherical topology for the service.
// When set, the service uses spherical coordinates for weather calculations.
func (s *Service) WithTopology(t spatial.Topology) {
	s.topology = t
}

// Topology returns the current topology (nil if flat mode)
func (s *Service) Topology() spatial.Topology {
	return s.topology
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

// ForceWorldWeather forces a specific weather type for an entire world
func (s *Service) ForceWorldWeather(ctx context.Context, worldID uuid.UUID, weatherType WeatherType) error {
	s.cacheMutex.RLock()
	cells, ok := s.geoCache[worldID]
	s.cacheMutex.RUnlock()

	if !ok || len(cells) == 0 {
		return fmt.Errorf("no geography data found for world %s", worldID)
	}

	// Lock cache
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	worldCache, ok := s.stateCache[worldID]
	if !ok {
		worldCache = make(map[uuid.UUID]*WeatherState)
		s.stateCache[worldID] = worldCache
	}

	currentTime := time.Now()

	// Update all cells
	for _, cell := range cells {
		// Create forced state
		// Ideally we simulate "real" values for this weather type
		// For now, construct a basic state matching the type

		newState := &WeatherState{
			CellID:      cell.CellID,
			Timestamp:   currentTime,
			State:       weatherType,
			Temperature: cell.Temperature, // Keep base temp? Or adjust?
		}

		// Adjust params based on forced weather to be consistent
		switch weatherType {
		case WeatherClear:
			newState.Precipitation = 0
			newState.Humidity = 0.3
			newState.Wind = Wind{Speed: 5, Direction: 0}
			newState.Visibility = 10000
		case WeatherCloudy:
			newState.Precipitation = 0
			newState.Humidity = 0.6
			newState.Wind = Wind{Speed: 10, Direction: 0}
			newState.Visibility = 5000
		case WeatherRain:
			newState.Precipitation = 5.0
			newState.Humidity = 0.9
			newState.Wind = Wind{Speed: 15, Direction: 0}
			newState.Visibility = 2000
		case WeatherStorm:
			newState.Precipitation = 20.0
			newState.Humidity = 1.0
			newState.Wind = Wind{Speed: 50, Direction: 0}
			newState.Visibility = 500
		case WeatherSnow:
			newState.Precipitation = 2.0 // Snow water equivalent
			newState.Humidity = 0.5
			newState.Temperature = -5.0 // Force cold for snow
			newState.Wind = Wind{Speed: 10, Direction: 0}
			newState.Visibility = 1000
		}

		// Save to DB
		if err := s.repo.SaveWeatherState(ctx, newState); err != nil {
			return fmt.Errorf("failed to save forced weather state: %w", err)
		}

		// Update cache
		worldCache[cell.CellID] = newState
	}

	return nil
}
