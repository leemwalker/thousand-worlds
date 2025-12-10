package auth_test

import (
	"context"
	"testing"

	"tw-backend/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionManager_Lifecycle(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	sm := auth.NewSessionManager(client)
	ctx := context.Background()

	t.Run("creates, retrieves, and invalidates session", func(t *testing.T) {
		userID := "user-session-1"
		username := "session-user"

		// Create
		session, err := sm.CreateSession(ctx, userID, username)
		require.NoError(t, err)
		require.NotEmpty(t, session.ID)
		assert.Equal(t, userID, session.UserID)

		// Get
		retrieved, err := sm.GetSession(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, userID, retrieved.UserID)

		// Invalidate
		err = sm.InvalidateSession(ctx, session.ID)
		require.NoError(t, err)

		// Get should fail
		_, err = sm.GetSession(ctx, session.ID)
		assert.Error(t, err)
	})
}
