package weather

import (
	"context"
	"testing"
	"time"

	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveWeatherState(ctx context.Context, state *WeatherState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockRepository) GetWeatherState(ctx context.Context, cellID uuid.UUID, timestamp int64) (*WeatherState, error) {
	args := m.Called(ctx, cellID, timestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WeatherState), args.Error(1)
}

func (m *MockRepository) GetWeatherHistory(ctx context.Context, cellID uuid.UUID, limit int) ([]*WeatherState, error) {
	args := m.Called(ctx, cellID, limit)
	return args.Get(0).([]*WeatherState), args.Error(1)
}

func (m *MockRepository) GetAnnualPrecipitation(ctx context.Context, cellID uuid.UUID, year int) (float64, error) {
	args := m.Called(ctx, cellID, year)
	return args.Get(0).(float64), args.Error(1)
}

func TestService_InitializeWorldWeather(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	ctx := context.Background()
	worldID := uuid.New()

	// Prepare data
	cellID1 := uuid.New()
	cellID2 := uuid.New()

	cells := []*GeographyCell{
		{CellID: cellID1, Location: geography.Point{X: 0, Y: 0}, Elevation: 100},
		{CellID: cellID2, Location: geography.Point{X: 1, Y: 0}, Elevation: 100},
	}

	states := []*WeatherState{
		{CellID: cellID1, State: WeatherClear, Temperature: 20},
		{CellID: cellID2, State: WeatherRain, Temperature: 15},
	}

	// Call Initialize
	service.InitializeWorldWeather(ctx, worldID, states, cells)

	// Verify caching
	cachedState1, err := service.GetCurrentWeather(ctx, worldID, cellID1)
	assert.NoError(t, err)
	assert.Equal(t, WeatherClear, cachedState1.State)

	cachedState2, err := service.GetCurrentWeather(ctx, worldID, cellID2)
	assert.NoError(t, err)
	assert.Equal(t, WeatherRain, cachedState2.State)

	// Verify geo caching
	// We can't access geoCache directly as it's private, but UpdateWorldWeather uses it.
	// So we'll test UpdateWorldWeather next.
}

func TestService_UpdateWorldWeather(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	ctx := context.Background()
	worldID := uuid.New()

	// Initial data
	cellID := uuid.New()
	cells := []*GeographyCell{
		{CellID: cellID, Location: geography.Point{X: 0, Y: 0}, Elevation: 100, Temperature: 20, IsOcean: false},
	}
	initialStates := []*WeatherState{
		{CellID: cellID, State: WeatherClear, Temperature: 20, Wind: Wind{Speed: 5}},
	}

	// Initialize
	service.InitializeWorldWeather(ctx, worldID, initialStates, cells)

	// Mock SaveWeatherState
	repo.On("SaveWeatherState", ctx, mock.Anything).Return(nil)

	// Update Weather
	// Force a change? Logic depends on UpdateWeather implementation which is deterministic based on inputs.
	// We'll just verify it runs and calls save.

	currentTime := time.Now()
	emotes, err := service.UpdateWorldWeather(ctx, worldID, currentTime, SeasonSummer)

	assert.NoError(t, err)
	repo.AssertExpectations(t)

	// Check if state updated in cache
	newState, _ := service.GetCurrentWeather(ctx, worldID, cellID)
	assert.NotNil(t, newState)
	assert.NotEqual(t, initialStates[0], newState, "State should be a new object")

	// Emotes might be empty if no significant change
	t.Logf("Emotes: %v", emotes)
}

func TestService_DetectWeatherChange(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	oldState := &WeatherState{State: WeatherClear, Wind: Wind{Speed: 5}}

	// Significant Change: Clear -> Storm
	newState := &WeatherState{State: WeatherStorm, Wind: Wind{Speed: 50}}
	emote := service.detectWeatherChange(oldState, newState)
	assert.Contains(t, emote, "storm", "Should detect storm")

	// Minor Change: Clear -> Clear, same wind
	newState2 := &WeatherState{State: WeatherClear, Wind: Wind{Speed: 6}}
	emote2 := service.detectWeatherChange(oldState, newState2)
	assert.Empty(t, emote2, "Should not emote for minor change")

	// Wind Change - increase
	newState3 := &WeatherState{State: WeatherClear, Wind: Wind{Speed: 20}}
	emote3 := service.detectWeatherChange(oldState, newState3)
	assert.Contains(t, emote3, "wind", "Should detect wind change")

	// Wind Change - decrease
	oldWindy := &WeatherState{State: WeatherClear, Wind: Wind{Speed: 25}}
	newCalm := &WeatherState{State: WeatherClear, Wind: Wind{Speed: 10}}
	emote4 := service.detectWeatherChange(oldWindy, newCalm)
	assert.Contains(t, emote4, "wind", "Should detect wind decrease")

	// Clear -> Cloudy
	cloudyState := &WeatherState{State: WeatherCloudy, Wind: Wind{Speed: 5}}
	emote5 := service.detectWeatherChange(oldState, cloudyState)
	assert.Contains(t, emote5, "Clouds", "Should detect cloudy")

	// Clear -> Rain
	rainState := &WeatherState{State: WeatherRain, Wind: Wind{Speed: 5}}
	emote6 := service.detectWeatherChange(oldState, rainState)
	assert.Contains(t, emote6, "Rain", "Should detect rain")

	// Rain -> Clear
	clearState := &WeatherState{State: WeatherClear, Wind: Wind{Speed: 5}}
	emote7 := service.detectWeatherChange(rainState, clearState)
	assert.Contains(t, emote7, "clear", "Should detect clear")

	// Clear -> Snow
	snowState := &WeatherState{State: WeatherSnow, Wind: Wind{Speed: 5}}
	emote8 := service.detectWeatherChange(oldState, snowState)
	assert.Contains(t, emote8, "Snow", "Should detect snow")
}

func TestService_ForceWorldWeather(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	ctx := context.Background()
	worldID := uuid.New()

	// Setup - no cells yet
	err := service.ForceWorldWeather(ctx, worldID, WeatherClear)
	assert.Error(t, err, "Should fail with no geography data")

	// Initialize with cells
	cellID := uuid.New()
	cells := []*GeographyCell{
		{CellID: cellID, Location: geography.Point{X: 0, Y: 0}, Elevation: 100, Temperature: 20},
	}
	service.InitializeWorldWeather(ctx, worldID, nil, cells)

	// Mock SaveWeatherState
	repo.On("SaveWeatherState", ctx, mock.Anything).Return(nil)

	// Test forcing each weather type
	weatherTypes := []WeatherType{
		WeatherClear,
		WeatherCloudy,
		WeatherRain,
		WeatherStorm,
		WeatherSnow,
	}

	for _, wt := range weatherTypes {
		t.Run(string(wt), func(t *testing.T) {
			err := service.ForceWorldWeather(ctx, worldID, wt)
			assert.NoError(t, err)

			// Verify cached state
			state, _ := service.GetCurrentWeather(ctx, worldID, cellID)
			assert.Equal(t, wt, state.State)
		})
	}
}

func TestService_UpdateWorldWeather_NoGeo(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	ctx := context.Background()
	worldID := uuid.New()

	// Call UpdateWorldWeather without initializing geo
	_, err := service.UpdateWorldWeather(ctx, worldID, time.Now(), SeasonSummer)
	assert.Error(t, err, "Should fail with no geography data")
}

func TestService_GetCurrentWeather_NotFound(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	ctx := context.Background()

	// Get weather for non-existent world
	state, err := service.GetCurrentWeather(ctx, uuid.New(), uuid.New())
	assert.NoError(t, err, "Should not error")
	assert.Nil(t, state, "Should return nil for not found")
}
