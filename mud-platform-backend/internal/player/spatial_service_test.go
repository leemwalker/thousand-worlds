package player

import (
	"context"
	"testing"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthRepository matches auth.Repository interface
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockAuthRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*auth.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *MockAuthRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *MockAuthRepository) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *MockAuthRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockAuthRepository) CreateCharacter(ctx context.Context, char *auth.Character) error {
	args := m.Called(ctx, char)
	return args.Error(0)
}
func (m *MockAuthRepository) GetCharacter(ctx context.Context, characterID uuid.UUID) (*auth.Character, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Character), args.Error(1)
}
func (m *MockAuthRepository) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*auth.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.Character), args.Error(1)
}
func (m *MockAuthRepository) GetCharacterByUserAndWorld(ctx context.Context, userID, worldID uuid.UUID) (*auth.Character, error) {
	args := m.Called(ctx, userID, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Character), args.Error(1)
}
func (m *MockAuthRepository) UpdateCharacter(ctx context.Context, char *auth.Character) error {
	args := m.Called(ctx, char)
	return args.Error(0)
}

// MockWorldRepository matches repository.WorldRepository interface
type MockWorldRepository struct {
	mock.Mock
}

func (m *MockWorldRepository) CreateWorld(ctx context.Context, world *repository.World) error {
	args := m.Called(ctx, world)
	return args.Error(0)
}
func (m *MockWorldRepository) GetWorld(ctx context.Context, worldID uuid.UUID) (*repository.World, error) {
	args := m.Called(ctx, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.World), args.Error(1)
}
func (m *MockWorldRepository) ListWorlds(ctx context.Context) ([]repository.World, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.World), args.Error(1)
}
func (m *MockWorldRepository) GetWorldsByOwner(ctx context.Context, ownerID uuid.UUID) ([]repository.World, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.World), args.Error(1)
}
func (m *MockWorldRepository) UpdateWorld(ctx context.Context, world *repository.World) error {
	args := m.Called(ctx, world)
	return args.Error(0)
}
func (m *MockWorldRepository) DeleteWorld(ctx context.Context, worldID uuid.UUID) error {
	args := m.Called(ctx, worldID)
	return args.Error(0)
}

// Tests
func TestHandleMovementCommand_Lobby(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld)

	charID := uuid.New()
	worldID := uuid.New()

	// Lobby World
	lobbyWorld := &repository.World{
		ID:   worldID,
		Name: "Lobby",
	}

	tests := []struct {
		name      string
		direction string
		startX    float64
		startY    float64
		expectedX float64
		expectedY float64
		hasError  bool
	}{
		{"North", "n", 5, 500, 5, 501, false},
		{"South", "s", 5, 500, 5, 499, false},
		{"East", "e", 5, 500, 6, 500, false},
		{"West", "w", 5, 500, 4, 500, false},
		{"Wall North", "n", 5, 1000, 5, 1000, true},
		{"Wall South", "s", 5, 0, 5, 0, true},
		{"Wall East", "e", 10, 500, 10, 500, true},
		{"Wall West", "w", 0, 500, 0, 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			char := &auth.Character{
				CharacterID: charID,
				WorldID:     worldID,
				PositionX:   tt.startX,
				PositionY:   tt.startY,
			}

			// Mock Expectations
			mockAuth.On("GetCharacter", ctx, charID).Return(char, nil).Once()
			mockWorld.On("GetWorld", ctx, worldID).Return(lobbyWorld, nil).Once()

			if !tt.hasError {
				mockAuth.On("UpdateCharacter", ctx, mock.MatchedBy(func(c *auth.Character) bool {
					return c.PositionX == tt.expectedX && c.PositionY == tt.expectedY
				})).Return(nil).Once()
			}

			msg, err := svc.HandleMovementCommand(ctx, charID, tt.direction)

			if tt.hasError {
				// We return user friendly errors as strings in message sometimes, but here we implemented error return for walls validation logic
				// In SpatialService I returned error for walls.
				// Wait, in calculateLobbyPosition I returned error for walls.
				// But HandleMovementCommand caught it? No, it returns error string.
				// "return err.Error(), nil"
				// So validation errors are NOT Go errors, they are successful call with error message strings.
				assert.NoError(t, err) // Technical error is nil
				assert.Contains(t, msg, "cannot go further")
			} else {
				assert.NoError(t, err)
				assert.Contains(t, msg, "You move")
				assert.Equal(t, tt.expectedX, char.PositionX)
				assert.Equal(t, tt.expectedY, char.PositionY)
			}
		})
	}
}

func TestHandleMovementCommand_Spherical(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld)

	charID := uuid.New()
	worldID := uuid.New()

	circumference := 10000.0
	// 1 degree = 10000 / 360 = 27.777... meters
	degPerMeter := 360.0 / circumference

	// Spherical World
	world := &repository.World{
		ID:            worldID,
		Name:          "TestWorld",
		Circumference: &circumference,
	}

	tests := []struct {
		name         string
		direction    string
		startX       float64 // Longitude
		startY       float64 // Latitude
		expectedX    float64 // Expected Longitude
		expectedY    float64 // Expected Latitude
		checkMessage string
	}{
		{
			name:         "Normal North",
			direction:    "n",
			startX:       0,
			startY:       0,
			expectedX:    0,
			expectedY:    degPerMeter, // 0 + 1m worth of degrees
			checkMessage: "You move north",
		},
		{
			name:         "Normal East",
			direction:    "e",
			startX:       0,
			startY:       0,
			expectedX:    degPerMeter, // 0 + 1m worth of degrees
			expectedY:    0,
			checkMessage: "You move east",
		},
		{
			name:      "Cross North Pole",
			direction: "n",
			startX:    0,
			startY:    89.99, // Very close to pole
			// Move 1m north (approx 0.036 degrees). 89.99 + 0.036 = 90.026
			// Crosses pole.
			// New Lat = 180 - 90.026 = 89.974
			// New Lon = 0 + 180 = 180
			expectedX:    180,
			expectedY:    89.974, // Approximate
			checkMessage: "cross the pole",
		},
		{
			name:      "Cross South Pole",
			direction: "s",
			startX:    45,
			startY:    -89.99,
			// Move 1m south (-0.036 deg). -89.99 - 0.036 = -90.026
			// New Lat = -180 - (-90.026) = -89.974
			// New Lon = 45 + 180 = 225 -> Normalize -> -135
			expectedX:    -135,
			expectedY:    -89.974,
			checkMessage: "cross the pole",
		},
		{
			name:      "Wrap Date Line East",
			direction: "e",
			startX:    179.99,
			startY:    0,
			// Move 1m east (+0.036 deg). 179.99 + 0.036 = 180.026
			// Normalize -> -179.974
			expectedX:    -179.974,
			expectedY:    0,
			checkMessage: "circled back",
		},
		{
			name:      "Wrap Date Line West",
			direction: "w",
			startX:    -179.99,
			startY:    0,
			// Move 1m west (-0.036 deg). -179.99 - 0.036 = -180.026
			// Normalize -> 179.974
			expectedX:    179.974,
			expectedY:    0,
			checkMessage: "circled back",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			char := &auth.Character{
				CharacterID: charID,
				WorldID:     worldID,
				PositionX:   tt.startX,
				PositionY:   tt.startY,
			}

			// Reset mock calls for each iteration to avoid aggregation issues
			// Actually we create new args but mocks are reused.
			// Best to use .Maybe() or clean call management.
			// Or just set expectations for this specific call.
			mockAuth.ExpectedCalls = nil
			mockWorld.ExpectedCalls = nil

			mockAuth.On("GetCharacter", ctx, charID).Return(char, nil).Once()
			mockWorld.On("GetWorld", ctx, worldID).Return(world, nil).Once()
			// Capture the updated character
			var updatedChar *auth.Character
			mockAuth.On("UpdateCharacter", ctx, mock.MatchedBy(func(c *auth.Character) bool {
				updatedChar = c
				return true
			})).Return(nil).Once()

			msg, err := svc.HandleMovementCommand(ctx, charID, tt.direction)
			assert.NoError(t, err)

			// Floating point comparison
			// Note: Precision might need tuning based on degPerMeter
			// degPerMeter is ~0.036

			// For pole crossing, calculation is: 180 - (start + delta)
			// expectedY in struct is rough approximation logic wise
			// Let's rely on checking if it flipped hemisphere or wrapped significantly if that's easier,
			// but InDelta is better.

			// Re-calculate expected for verification precise matching or use loose delta
			assert.InDelta(t, tt.expectedY, updatedChar.PositionY, 0.1, "Latitude mismatch")
			assert.InDelta(t, tt.expectedX, updatedChar.PositionX, 0.1, "Longitude mismatch")

			if tt.checkMessage != "" {
				assert.Contains(t, msg, tt.checkMessage)
			}
		})
	}
}

// TestCalculateDistance tests the CalculateDistance method
func TestCalculateDistance(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld)

	// Earth radius in meters (approximately)
	radius := 6371000.0

	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
		delta    float64
	}{
		{
			name: "Same point",
			lat1: 0, lon1: 0,
			lat2: 0, lon2: 0,
			expected: 0,
			delta:    0.1,
		},
		{
			name: "Equator: 0 to 90E (quarter circumference)",
			lat1: 0, lon1: 0,
			lat2: 0, lon2: 90,
			expected: 10007543, // ~10,000 km
			delta:    10000,    // 10km tolerance
		},
		{
			name: "Pole to pole (half circumference)",
			lat1: 90, lon1: 0,
			lat2: -90, lon2: 0,
			expected: 20015087, // ~20,000 km
			delta:    10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2, radius)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

// TestHandleMovementCommand_CharacterNotFound tests error path when character doesn't exist
func TestHandleMovementCommand_CharacterNotFound(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld)

	charID := uuid.New()
	ctx := context.Background()

	mockAuth.On("GetCharacter", ctx, charID).Return(nil, assert.AnError).Once()

	_, err := svc.HandleMovementCommand(ctx, charID, "n")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "character")
}

// TestHandleMovementCommand_WorldNotFound tests error path when world doesn't exist
func TestHandleMovementCommand_WorldNotFound(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld)

	charID := uuid.New()
	worldID := uuid.New()
	ctx := context.Background()

	char := &auth.Character{
		CharacterID: charID,
		WorldID:     worldID,
		PositionX:   5,
		PositionY:   500,
	}

	mockAuth.On("GetCharacter", ctx, charID).Return(char, nil).Once()
	mockWorld.On("GetWorld", ctx, worldID).Return(nil, assert.AnError).Once()

	_, err := svc.HandleMovementCommand(ctx, charID, "n")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "world")
}
