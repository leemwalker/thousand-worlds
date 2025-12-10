package look

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/internal/auth"
	"tw-backend/internal/game/services/entity"
)

func TestDescribeEntity_Self(t *testing.T) {
	s := &LookService{} // Dependencies not needed for self
	char := &auth.Character{
		Name:        "Hero",
		Description: "A brave hero.",
		Occupation:  "Warrior",
	}

	// "self"
	desc, err := s.DescribeEntity(context.Background(), char, "self")
	require.NoError(t, err)
	assert.Contains(t, desc, "You are Hero, the Warrior.")
	assert.Contains(t, desc, "A brave hero.")

	// "me"
	desc, err = s.DescribeEntity(context.Background(), char, "me")
	require.NoError(t, err)
	assert.Contains(t, desc, "You are Hero")

	// "Hero" (case insensitive)
	desc, err = s.DescribeEntity(context.Background(), char, "hero")
	require.NoError(t, err)
	assert.Contains(t, desc, "You are Hero")
}

func TestDescribeEntity_Other(t *testing.T) {
	entityService := entity.NewService()
	s := &LookService{
		entityService: entityService,
	}

	worldID := uuid.New()
	char := &auth.Character{
		WorldID:   worldID,
		PositionX: 10,
		PositionY: 10,
	}

	// Add entity
	ent := &entity.Entity{
		ID:          uuid.New(),
		Name:        "Sword",
		Description: "A sharp iron sword.",
		WorldID:     worldID,
		X:           12, // nearby
		Y:           10,
	}
	entityService.AddEntity(context.Background(), ent)

	// Look for it
	desc, err := s.DescribeEntity(context.Background(), char, "Sword")
	require.NoError(t, err)
	assert.Equal(t, "A sharp iron sword.", desc)

	// Look for missing
	_, err = s.DescribeEntity(context.Background(), char, "Shield")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "don't see any 'Shield' here")
}
