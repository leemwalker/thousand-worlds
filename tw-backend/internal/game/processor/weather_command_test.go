package processor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/auth"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"
)

// MockWeatherRepo for testing
type MockWeatherRepo struct {
	SavedStates []*weather.WeatherState
}

func (m *MockWeatherRepo) SaveWeatherState(ctx context.Context, state *weather.WeatherState) error {
	m.SavedStates = append(m.SavedStates, state)
	return nil
}

func (m *MockWeatherRepo) GetWeatherState(ctx context.Context, cellID uuid.UUID, timestamp int64) (*weather.WeatherState, error) {
	return nil, nil
}

func (m *MockWeatherRepo) GetWeatherHistory(ctx context.Context, cellID uuid.UUID, days int) ([]*weather.WeatherState, error) {
	return nil, nil
}

func (m *MockWeatherRepo) GetAnnualPrecipitation(ctx context.Context, cellID uuid.UUID, year int) (float64, error) {
	return 0, nil
}

func (m *MockWeatherRepo) InitialiseValues(ctx context.Context, worldID uuid.UUID) error {
	return nil
}

func TestHandleWeather_GodMode(t *testing.T) {
	// 1. Setup
	mockAuthRepo := auth.NewMockRepository()
	mockWorldRepo := NewMockWorldRepository() // Assuming this exists in processor_test.go context

	// Create WeatherService with MockRepo
	mockWeatherRepo := &MockWeatherRepo{}
	weatherService := weather.NewService(mockWeatherRepo)

	// Create Processor
	// Note: lookService etc can be nil for this specific test as we won't trigger them
	// BUT NewGameProcessor might need them not to be nil if it uses them in constructor?
	// Based on code, it just assigns them.
	// But let's be safe and use basic mocks/nil where allowed.

	// We need a real LookService if we want to avoid panic if processor uses it?
	// Processor uses lookService in handleLook. handleWeather doesn't use it.

	proc := NewGameProcessor(mockAuthRepo, mockWorldRepo, nil, nil, nil, nil, weatherService)

	// Setup Client
	worldID := uuid.New()
	client := &mockClient{
		UserID:      uuid.New(),
		CharacterID: uuid.New(),
		WorldID:     worldID,
	}

	// Prime the WeatherService with some geography cells so ForceWorldWeather works
	// It checks s.geoCache[worldID]
	// We need to access the private cache or expose a method?
	// InitializeWorldWeather is public!

	cells := []*weather.GeographyCell{
		{
			CellID:      uuid.New(),
			Location:    geography.Point{X: 0, Y: 0},
			Elevation:   100,
			Temperature: 20,
		},
	}
	weatherService.InitializeWorldWeather(context.Background(), worldID, []*weather.WeatherState{}, cells)

	// 2. Execute Command: "weather storm"
	target := "storm"
	cmd := &websocket.CommandData{
		Action: "weather",
		Target: &target,
	}

	err := proc.ProcessCommand(context.Background(), client, cmd)
	require.NoError(t, err)

	// 3. Verify
	// Check Client received message
	assert.NotEmpty(t, client.messages)
	lastMsg := client.messages[len(client.messages)-1]
	assert.Contains(t, lastMsg.Text, "Weather changed to storm")

	// Check WeatherService Repo saved new state
	require.NotEmpty(t, mockWeatherRepo.SavedStates)
	lastState := mockWeatherRepo.SavedStates[len(mockWeatherRepo.SavedStates)-1]
	assert.Equal(t, weather.WeatherStorm, lastState.State)
	assert.Equal(t, 20.0, lastState.Precipitation) // Storm precip defined in ForceWorldWeather
}
