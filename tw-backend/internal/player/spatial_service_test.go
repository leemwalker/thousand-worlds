package player

import (
	"context"
	"strings"
	"testing"

	"tw-backend/internal/auth"
	"tw-backend/internal/repository"

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
func TestHandleMovementCommand_Bounded(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld, nil)

	charID := uuid.New()
	worldID := uuid.New()

	// Generic Bounded World (like Lobby)
	minX, minY := 0.0, 0.0
	maxX, maxY := 10.0, 10.0

	// Create map with colliders
	colliders := []map[string]interface{}{
		{
			"x":       5.0,
			"y":       5.0,
			"radius":  0.5,
			"message": "The massive statue blocks your path.",
		},
	}

	metadata := map[string]interface{}{
		"colliders": colliders,
	}

	boundedWorld := &repository.World{
		ID:        worldID,
		Name:      "BoundedWorld",
		BoundsMin: &repository.Vector3{X: minX, Y: minY, Z: 0},
		BoundsMax: &repository.Vector3{X: maxX, Y: maxY, Z: 0},
		Metadata:  metadata,
	}

	tests := []struct {
		name      string
		direction string
		startX    float64
		startY    float64
		expectedX float64
		expectedY float64
		hasError  bool
		errMatch  string
	}{
		// Valid moves
		{"North", "n", 5, 2, 5, 3, false, ""},
		{"South", "s", 5, 2, 5, 1, false, ""},
		{"East", "e", 5, 2, 6, 2, false, ""},
		{"West", "w", 5, 2, 4, 2, false, ""},

		// Walls (Generic Bounds 0-10)
		{"Wall North", "n", 5, 10, 5, 10, true, "wall"},
		{"Wall South", "s", 5, 0, 5, 0, true, "wall"},
		{"Wall East", "e", 10, 5, 10, 5, true, "wall"},
		{"Wall West", "w", 0, 5, 0, 5, true, "wall"},

		// Collider Check (Statue at 5,5 radius 0.5)
		{"Statue Collision North", "n", 5, 4, 5, 4, true, "statue"},
		{"Statue Collision South", "s", 5, 6, 5, 6, true, "statue"},
		{"Statue Collision East", "e", 4, 5, 4, 5, true, "statue"},
		{"Statue Collision West", "w", 6, 5, 6, 5, true, "statue"},
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
			mockWorld.On("GetWorld", ctx, worldID).Return(boundedWorld, nil).Once()

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
				if tt.errMatch != "" {
					assert.Contains(t, strings.ToLower(msg), tt.errMatch)
				}
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
	svc := NewSpatialService(mockAuth, mockWorld, nil)

	charID := uuid.New()
	worldID := uuid.New()

	circumference := 10000.0
	// 1 meter = 1 meter. No conversion needed for expectation.

	// Spherical World (No Bounds defined, defaults to Sphere logic)
	world := &repository.World{
		ID:            worldID,
		Name:          "TestWorld",
		Circumference: &circumference,
	}

	tests := []struct {
		name         string
		direction    string
		startX       float64 // Longitude (Meters)
		startY       float64 // Latitude (Meters)
		expectedX    float64 // Expected Longitude (Meters)
		expectedY    float64 // Expected Latitude (Meters)
		checkMessage string
	}{
		{
			name:         "Normal North",
			direction:    "n",
			startX:       1000,
			startY:       1000,
			expectedX:    1000,
			expectedY:    1001, // +1 meter
			checkMessage: "You move north",
		},
		{
			name:         "Normal East",
			direction:    "e",
			startX:       1000,
			startY:       1000,
			expectedX:    1001, // +1 meter
			expectedY:    1000,
			checkMessage: "You move east",
		},
		{
			name:      "Cross North Pole",
			direction: "n",
			startX:    1000,
			startY:    2499.5, // 0.5m from pole (2500)
			// Move 1m north. 2499.5 + 1 = 2500.5
			// Overshoot 0.5.
			// New Y = 2500 - 0.5 = 2499.5
			// New X = 1000 + 5000 (Half Circ) = 6000
			expectedX:    6000,
			expectedY:    2499.5,
			checkMessage: "cross the North Pole",
		},
		{
			name:      "Cross South Pole",
			direction: "s",
			startX:    5000,
			startY:    -2499.5, // 0.5m from pole (-2500)
			// Move 1m south. -2499.5 - 1 = -2500.5
			// Overshoot 0.5.
			// New Y = -2500 + 0.5 = -2499.5
			// New X = 5000 + 5000 = 10000 -> Wrap -> 0
			expectedX:    0,
			expectedY:    -2499.5,
			checkMessage: "cross the South Pole",
		},
		{
			name:      "Wrap Date Line East",
			direction: "e",
			startX:    9999.5, // 0.5m from edge (10000)
			startY:    0,
			// Move 1m East. 9999.5 + 1 = 10000.5
			// Wrap -> 0.5
			expectedX:    0.5,
			expectedY:    0,
			checkMessage: "circled back",
		},
		{
			name:      "Wrap Date Line West",
			direction: "w",
			startX:    0.5, // 0.5m from edge (0)
			startY:    0,
			// Move 1m West. 0.5 - 1 = -0.5
			// Wrap -> 9999.5
			expectedX:    9999.5,
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

			// Clear expectations
			mockAuth.ExpectedCalls = nil
			mockWorld.ExpectedCalls = nil

			mockAuth.On("GetCharacter", ctx, charID).Return(char, nil).Once()
			mockWorld.On("GetWorld", ctx, worldID).Return(world, nil).Once()

			var updatedChar *auth.Character
			mockAuth.On("UpdateCharacter", ctx, mock.MatchedBy(func(c *auth.Character) bool {
				updatedChar = c
				return true
			})).Return(nil).Once()

			msg, err := svc.HandleMovementCommand(ctx, charID, tt.direction)
			assert.NoError(t, err)

			// Floating point comparison
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
	svc := NewSpatialService(mockAuth, mockWorld, nil)

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
	svc := NewSpatialService(mockAuth, mockWorld, nil)

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
	svc := NewSpatialService(mockAuth, mockWorld, nil)

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

func TestGetPortalLocation(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld, nil)

	// Generic bounded world (0-10)
	minX, minY := 0.0, 0.0
	maxX, maxY := 10.0, 10.0
	world := &repository.World{
		ID:        uuid.New(),
		BoundsMin: &repository.Vector3{X: minX, Y: minY, Z: 0},
		BoundsMax: &repository.Vector3{X: maxX, Y: maxY, Z: 0},
	}

	// Test distribution across walls
	counts := map[string]int{
		"North": 0, "South": 0, "East": 0, "West": 0,
	}

	for i := 0; i < 100; i++ {
		targetID := uuid.New()
		x, y := svc.GetPortalLocation(world, targetID)

		// Check that it's on a wall
		onWall := (x == 0 && y >= 0 && y <= 10) || // West wall
			(x == 10 && y >= 0 && y <= 10) || // East wall
			(y == 0 && x >= 0 && x <= 10) || // South wall
			(y == 10 && x >= 0 && x <= 10) // North wall
		assert.True(t, onWall, "Point %f,%f should be on wall", x, y)

		// Check bounds
		assert.GreaterOrEqual(t, x, 0.0)
		assert.LessOrEqual(t, x, 10.0)
		assert.GreaterOrEqual(t, y, 0.0)
		assert.LessOrEqual(t, y, 10.0)

		// Count which wall it landed on
		if x == 0 {
			counts["West"]++
		} else if x == 10 {
			counts["East"]++
		} else if y == 0 {
			counts["South"]++
		} else if y == 10 {
			counts["North"]++
		}
	}

	// Rough check to ensure we use all walls (random chance of missing one in 100 is very low)
	assert.Greater(t, counts["North"], 0)
	assert.Greater(t, counts["South"], 0)
	assert.Greater(t, counts["East"], 0)
	assert.Greater(t, counts["West"], 0)
}

func TestCheckPortalProximity(t *testing.T) {
	mockAuth := new(MockAuthRepository)
	mockWorld := new(MockWorldRepository)
	svc := NewSpatialService(mockAuth, mockWorld, nil)

	// Portal at (0, 5) - West wall center
	portalX, portalY := 0.0, 5.0

	tests := []struct {
		name     string
		charX    float64
		charY    float64
		expected bool
	}{
		{"At Portal", 0.0, 5.0, true},
		{"1m away", 1.0, 5.0, true},
		{"4.9m away", 4.9, 5.0, true},
		{"5m away", 5.0, 5.0, true}, // Limit inclusive
		{"5.1m away", 5.1, 5.0, false},
		{"Center of Room (5,5)", 5.0, 5.0, true}, // Exact 5m
		{"Far corner (10,10)", 10.0, 10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := svc.CheckPortalProximity(tt.charX, tt.charY, portalX, portalY, false)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Additional tests for CheckPortalProximity
	// 5,5 to 10,10 distance is approx 7.07 > 5
	allowed, _ := svc.CheckPortalProximity(5, 5, 10, 10, false)
	assert.False(t, allowed)

	// 5,5 to 8,5 distance is 3 <= 5
	allowed, _ = svc.CheckPortalProximity(5, 5, 8, 5, false)
	assert.True(t, allowed)

	// Lobby Bypass check
	// 1000m away but isLobby=true
	allowed, _ = svc.CheckPortalProximity(0, 0, 1000, 1000, true)
	assert.True(t, allowed)
}
