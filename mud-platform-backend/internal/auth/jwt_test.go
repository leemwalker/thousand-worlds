package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager_GenerateAndValidateToken(t *testing.T) {
	signingKey := []byte("secret-signing-key-must-be-long-enough")
	encryptionKey := []byte("01234567890123456789012345678901") // 32 bytes
	tm, err := NewTokenManager(signingKey, encryptionKey)
	require.NoError(t, err)

	t.Run("generates and validates valid token", func(t *testing.T) {
		userID := "user-123"
		username := "testuser"
		roles := []string{"admin", "player"}

		token, err := tm.GenerateToken(userID, username, roles)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := tm.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, roles, claims.Roles)

		// Check expiration
		assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 1*time.Minute)
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		userID := "user-bad"
		username := "baduser"
		roles := []string{"player"}

		token, err := tm.GenerateToken(userID, username, roles)
		require.NoError(t, err)

		// Create another manager with different signing key
		otherTM, _ := NewTokenManager([]byte("wrong-signing-key-00000000000000"), encryptionKey)
		_, err = otherTM.ValidateToken(token)
		assert.Error(t, err)
	})
}

func TestNewTokenManager_Validation(t *testing.T) {
	t.Run("rejects invalid encryption key length", func(t *testing.T) {
		_, err := NewTokenManager([]byte("sign"), []byte("short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be 32 bytes")
	})
}

func TestTokenManager_EncryptionErrors(t *testing.T) {
	signingKey := []byte("secret-signing-key-must-be-long-enough")
	encryptionKey := []byte("01234567890123456789012345678901")
	tm, err := NewTokenManager(signingKey, encryptionKey)
	require.NoError(t, err)

	t.Run("decrypts fails on short ciphertext", func(t *testing.T) {
		_, err := tm.decrypt([]byte("short"))
		assert.Error(t, err)
		assert.Equal(t, "malformed ciphertext", err.Error())
	})
}
