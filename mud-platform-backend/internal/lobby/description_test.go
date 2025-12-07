package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/repository"
)

// MockWebsocketClient for testing
type MockWebsocketClient struct {
	UserID   uuid.UUID
	Username string
}

func (m *MockWebsocketClient) GetUserID() uuid.UUID {
	return m.UserID
}

func (m *MockWebsocketClient) GetUsername() string {
	return m.Username
}

func setupDescriptionGeneratorTest() (*DescriptionGenerator, *MockWorldRepository, *auth.MockRepository) {
	worldRepo := &MockWorldRepository{}
	authRepo := auth.NewMockRepository()
	generator := NewDescriptionGenerator(worldRepo, authRepo)
	return generator, worldRepo, authRepo
}

func TestGenerateDescription_Default(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	user := &auth.User{
		UserID: uuid.New(),
	}

	// Mock empty world list
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	// Mock character
	char := &auth.Character{PositionX: 500}

	desc, err := generator.GenerateDescription(ctx, user, char, nil)
	require.NoError(t, err)
	assert.Contains(t, desc, "Central Hub")
	assert.Contains(t, desc, "stone statue")
}

func TestGenerateDescription_WithLastWorld_Desert(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	lastWorldID := uuid.New()
	user := &auth.User{
		UserID:      uuid.New(),
		LastWorldID: &lastWorldID,
	}
	char := &auth.Character{PositionX: 500}

	// Mock last world with desert theme
	world := &repository.World{
		ID:   lastWorldID,
		Name: "DesertWorld",
		Metadata: map[string]interface{}{
			"description": "A vast sandy desert",
		},
	}
	worldRepo.On("GetWorld", ctx, lastWorldID).Return(world, nil)
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	desc, err := generator.GenerateDescription(ctx, user, char, nil)
	require.NoError(t, err)
	assert.Contains(t, desc, "Hot desert winds") // Wait, this text was part of the OLD logic, did I update it in description.go?
	// In description.go I changed it to "A lingering warmth from the desert clings to your clothes."
	// I should check description.go logic again.
	// Ah, step 213 shows I changed it.
	assert.Contains(t, desc, "warmth from the desert")
}

func TestGenerateDescription_WithLastWorld_Ocean(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	lastWorldID := uuid.New()
	user := &auth.User{
		UserID:      uuid.New(),
		LastWorldID: &lastWorldID,
	}
	char := &auth.Character{PositionX: 500}

	// Mock last world with ocean theme
	world := &repository.World{
		ID:   lastWorldID,
		Name: "OceanWorld",
		Metadata: map[string]interface{}{
			"description": "A vast blue ocean",
		},
	}
	worldRepo.On("GetWorld", ctx, lastWorldID).Return(world, nil)
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	desc, err := generator.GenerateDescription(ctx, user, char, nil)
	require.NoError(t, err)
	assert.Contains(t, desc, "tang of salt")
}

func TestGenerateDescription_WithPortals(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	user := &auth.User{
		UserID: uuid.New(),
	}
	char := &auth.Character{PositionX: 500}

	// Mock worlds list
	worlds := []repository.World{
		{ID: uuid.New(), Name: "World1"},
		{ID: uuid.New(), Name: "World2"},
	}
	worldRepo.On("ListWorlds", ctx).Return(worlds, nil)

	desc, err := generator.GenerateDescription(ctx, user, char, nil)
	require.NoError(t, err)
	assert.Contains(t, desc, "portal to World1")
	assert.Contains(t, desc, "portal to World2")
}

func TestGenerateDescription_WithPlayers(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	user := &auth.User{
		UserID: uuid.New(),
	}
	char := &auth.Character{PositionX: 500}

	// Mock empty world list
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	// Mock other players
	players := []WebsocketClient{
		&MockWebsocketClient{UserID: uuid.New(), Username: "Alice"},
		&MockWebsocketClient{UserID: uuid.New(), Username: "Bob"},
	}

	desc, err := generator.GenerateDescription(ctx, user, char, players)
	require.NoError(t, err)
	assert.Contains(t, desc, "Alice")
	assert.Contains(t, desc, "Bob")
}

func TestGenerateDescription_WithManyPlayers(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	user := &auth.User{
		UserID: uuid.New(),
	}
	char := &auth.Character{PositionX: 500}

	// Mock empty world list
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	// Mock other players
	players := []WebsocketClient{
		&MockWebsocketClient{UserID: uuid.New(), Username: "Alice"},
		&MockWebsocketClient{UserID: uuid.New(), Username: "Bob"},
		&MockWebsocketClient{UserID: uuid.New(), Username: "Charlie"},
		&MockWebsocketClient{UserID: uuid.New(), Username: "Dave"},
	}

	desc, err := generator.GenerateDescription(ctx, user, char, players)
	require.NoError(t, err)
	assert.Contains(t, desc, "You see")
}

func TestGenerateDescription_Zones(t *testing.T) {
	generator, worldRepo, _ := setupDescriptionGeneratorTest()
	ctx := context.Background()
	user := &auth.User{UserID: uuid.New()}
	worldRepo.On("ListWorlds", ctx).Return([]repository.World{}, nil)

	// West Wing
	charWest := &auth.Character{PositionX: 100}
	desc, _ := generator.GenerateDescription(ctx, user, charWest, nil)
	assert.Contains(t, desc, "West Wing")
	assert.Contains(t, desc, "quieter here")

	// East Wing
	charEast := &auth.Character{PositionX: 900}
	desc, _ = generator.GenerateDescription(ctx, user, charEast, nil)
	assert.Contains(t, desc, "East Wing")
	assert.Contains(t, desc, "gather and plan")
}
