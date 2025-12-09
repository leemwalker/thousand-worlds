package lobby

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/world/interview"
)

// MockInterviewRepository for testing
type MockInterviewRepository struct {
	mock.Mock
}

func (m *MockInterviewRepository) GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*interview.WorldConfiguration, error) {
	args := m.Called(worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interview.WorldConfiguration), args.Error(1)
}

func (m *MockInterviewRepository) GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*interview.WorldConfiguration, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interview.WorldConfiguration), args.Error(1)
}

// MockWorldRepository for testing
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
	return args.Get(0).([]repository.World), args.Error(1)
}

func (m *MockWorldRepository) GetWorldsByOwner(ctx context.Context, ownerID uuid.UUID) ([]repository.World, error) {
	args := m.Called(ctx, ownerID)
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

func setupLookServiceTest() (*LookService, *auth.MockRepository, *MockWorldRepository, *MockInterviewRepository) {
	authRepo := auth.NewMockRepository()
	worldRepo := &MockWorldRepository{}
	interviewRepo := &MockInterviewRepository{}
	service := NewLookService(authRepo, worldRepo, interviewRepo, nil)
	return service, authRepo, worldRepo, interviewRepo
}

func TestDescribePlayer_NewPlayer(t *testing.T) {
	service, authRepo, _, _ := setupLookServiceTest()
	ctx := context.Background()
	username := "NewUser"
	userID := uuid.New()

	// Mock user found
	authRepo.CreateUser(ctx, &auth.User{
		UserID:   userID,
		Username: username,
		Email:    "test@example.com",
	})

	// Mock no characters
	// Note: MockRepository.GetUserCharacters returns empty list by default if none added

	desc, err := service.DescribePlayer(ctx, username)
	require.NoError(t, err)
	assert.Contains(t, desc, "shapeless gray spirit")
	assert.Contains(t, desc, username)
}

func TestDescribePlayer_ReturningPlayer(t *testing.T) {
	service, authRepo, worldRepo, interviewRepo := setupLookServiceTest()
	ctx := context.Background()
	username := "ReturningUser"
	userID := uuid.New()
	worldID := uuid.New()

	// Mock user
	authRepo.CreateUser(ctx, &auth.User{
		UserID:   userID,
		Username: username,
		Email:    "test@example.com",
	})

	// Mock character
	lastPlayed := time.Now()
	char := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     worldID,
		Name:        "Hero",
		LastPlayed:  &lastPlayed,
	}
	authRepo.CreateCharacter(ctx, char)

	// Mock world
	world := &repository.World{ID: worldID, Name: "FantasyWorld"}
	worldRepo.On("GetWorld", ctx, worldID).Return(world, nil)

	// Mock config
	config := &interview.WorldConfiguration{
		Theme:     "Forest",
		WorldName: "FantasyWorld",
	}
	interviewRepo.On("GetConfigurationByWorldID", worldID).Return(config, nil)

	desc, err := service.DescribePlayer(ctx, username)
	require.NoError(t, err)
	assert.Contains(t, desc, "Hero")
	assert.Contains(t, desc, "FantasyWorld")
	assert.Contains(t, desc, "earthy scent")
}

func TestDescribePortal_Found(t *testing.T) {
	service, _, worldRepo, interviewRepo := setupLookServiceTest()
	ctx := context.Background()
	worldName := "FireRealm"
	worldID := uuid.New()

	// Mock list worlds
	worlds := []repository.World{
		{ID: worldID, Name: worldName},
	}
	worldRepo.On("ListWorlds", ctx).Return(worlds, nil)

	// Mock config
	config := &interview.WorldConfiguration{
		Theme:     "Desert",
		WorldName: worldName,
	}
	interviewRepo.On("GetConfigurationByWorldID", worldID).Return(config, nil)

	desc, err := service.DescribePortal(ctx, worldName)
	require.NoError(t, err)
	assert.Contains(t, desc, "sun-bleached stone")
	assert.Contains(t, desc, worldName)
}

func TestDescribePortal_NotFound(t *testing.T) {
	service, _, worldRepo, _ := setupLookServiceTest()
	ctx := context.Background()

	// Mock empty world list
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	desc, err := service.DescribePortal(ctx, "UnknownWorld")
	require.Error(t, err)
	assert.Equal(t, "", desc)
	assert.Contains(t, err.Error(), "portal not found")
}

func TestDescribeStatue_Neutral(t *testing.T) {
	service, _, _, interviewRepo := setupLookServiceTest()
	ctx := context.Background()
	userID := uuid.New()

	// Mock config not found (user hasn't created a world)
	interviewRepo.On("GetConfigurationByUserID", userID).Return(nil, errors.New("not found"))

	desc, err := service.DescribeStatue(ctx, userID)
	require.NoError(t, err)
	assert.Contains(t, desc, "weathered stone statue")
	assert.Contains(t, desc, "waiting patiently")
	assert.Contains(t, desc, "Send it a tell")
	assert.NotContains(t, desc, "create world")
}

func TestDescribeStatue_Themed(t *testing.T) {
	service, _, _, interviewRepo := setupLookServiceTest()
	ctx := context.Background()
	userID := uuid.New()
	worldName := "MyOceanWorld"

	// Mock config created by user
	config := &interview.WorldConfiguration{
		CreatedBy: userID,
		WorldName: worldName,
		Theme:     "Ocean",
	}
	interviewRepo.On("GetConfigurationByUserID", userID).Return(config, nil)

	desc, err := service.DescribeStatue(ctx, userID)
	require.NoError(t, err)
	assert.Contains(t, desc, "coral and shell")
	assert.Contains(t, desc, worldName)
}

func TestDescribeStatue_AllThemes(t *testing.T) {
	themes := []struct {
		theme    string
		expected string
	}{
		{"Forest", "intertwined vines"},
		{"Desert", "sun-bleached sandstone"},
		{"Ocean", "coral and shell"},
		{"Mountain", "granite and ice"},
		{"Tech", "chrome and circuitry"},
		{"Magic", "crystalline statue"},
		{"Unknown", "essence of"},
	}

	for _, tt := range themes {
		t.Run(tt.theme, func(t *testing.T) {
			service, _, _, interviewRepo := setupLookServiceTest()
			ctx := context.Background()
			userID := uuid.New()
			worldName := "ThemedWorld"

			interviewRepo.On("GetConfigurationByUserID", userID).Return(&interview.WorldConfiguration{
				CreatedBy: userID,
				WorldName: worldName,
				Theme:     tt.theme,
			}, nil)

			desc, err := service.DescribeStatue(ctx, userID)
			require.NoError(t, err)
			assert.Contains(t, desc, tt.expected)
		})
	}
}

func TestDescribePortal_AllThemes(t *testing.T) {
	themes := []struct {
		theme    string
		expected string
	}{
		{"Forest", "living vines"},
		{"Desert", "sun-bleached stone"},
		{"Ocean", "coral and driftwood"},
		{"Mountain", "ancient granite"},
		{"Tech", "sleek alloy"},
		{"Magic", "otherworldly light"},
		{"Unknown", "reflecting its nature"},
	}

	for _, tt := range themes {
		t.Run(tt.theme, func(t *testing.T) {
			service, _, worldRepo, interviewRepo := setupLookServiceTest()
			ctx := context.Background()
			worldID := uuid.New()
			worldName := "ThemedWorld"

			worldRepo.On("ListWorlds", ctx).Return([]repository.World{{ID: worldID, Name: worldName}}, nil)
			interviewRepo.On("GetConfigurationByWorldID", worldID).Return(&interview.WorldConfiguration{
				WorldName: worldName,
				Theme:     tt.theme,
			}, nil)

			desc, err := service.DescribePortal(ctx, worldName)
			require.NoError(t, err)
			assert.Contains(t, desc, tt.expected)
		})
	}
}

func TestDescribePortal_Basic(t *testing.T) {
	service, _, worldRepo, interviewRepo := setupLookServiceTest()
	ctx := context.Background()
	worldID := uuid.New()
	worldName := "BasicWorld"

	worldRepo.On("ListWorlds", ctx).Return([]repository.World{{ID: worldID, Name: worldName}}, nil)
	// Return error to trigger basic description
	interviewRepo.On("GetConfigurationByWorldID", worldID).Return(nil, errors.New("not found"))

	desc, err := service.DescribePortal(ctx, worldName)
	require.NoError(t, err)
	assert.Contains(t, desc, "shimmers before you")
	assert.Contains(t, desc, "rippling like water")
}

// TestGetLobbyDescription tests the dynamic lobby description
func TestGetLobbyDescription(t *testing.T) {
	service, authRepo, worldRepo, _ := setupLookServiceTest()
	ctx := context.Background()
	userID := uuid.New()

	// Mock user (authRepo is a real in-memory implementation, not a testify mock)
	user := &auth.User{
		UserID:   userID,
		Username: "Observer",
		Email:    "observer@example.com",
	}
	err := authRepo.CreateUser(ctx, user)
	require.NoError(t, err)

	// Mock worlds (worldRepo IS a testify mock)
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{
		{ID: uuid.New(), Name: "Wonderland"},
	}, nil)

	// Mock players - None

	// Create character
	charID := uuid.New()
	char := &auth.Character{
		CharacterID: charID,
		UserID:      userID,
		WorldID:     LobbyWorldID,
		Name:        "TestChar",
		PositionX:   500, // Central Hub
	}
	err = authRepo.CreateCharacter(ctx, char)
	require.NoError(t, err)

	desc, err := service.GetLobbyDescription(ctx, userID, charID, nil)
	require.NoError(t, err)
	assert.Contains(t, desc, "low gray fog")
	assert.Contains(t, desc, "Wonderland")
}
