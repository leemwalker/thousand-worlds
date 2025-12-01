package auth_test

import (
	"strings"
	"testing"

	"mud-platform-backend/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_HashAndCompare(t *testing.T) {
	ph := auth.NewPasswordHasher()

	t.Run("hashes and verifies correct password", func(t *testing.T) {
		password := "secure-password-123"
		hash, err := ph.HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		// Verify structure (e.g., $argon2id$v=19$m=65536,t=3,p=4$...)
		assert.True(t, strings.HasPrefix(hash, "$argon2id$"))

		match, err := ph.ComparePassword(password, hash)
		require.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("rejects incorrect password", func(t *testing.T) {
		password := "secure-password-123"
		hash, err := ph.HashPassword(password)
		require.NoError(t, err)

		match, err := ph.ComparePassword("wrong-password", hash)
		require.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("rejects invalid hash formats", func(t *testing.T) {
		invalidHashes := []string{
			"invalid",
			"$argon2id$v=19$m=65536,t=3,p=4$salt",            // Missing hash
			"$bcrypt$v=19$m=65536,t=3,p=4$salt$hash",         // Wrong variant
			"$argon2id$v=99$m=65536,t=3,p=4$salt$hash",       // Wrong version
			"$argon2id$v=19$m=bad,t=3,p=4$salt$hash",         // Bad params
			"$argon2id$v=19$m=65536,t=3,p=4$bad-base64$hash", // Bad salt base64
			"$argon2id$v=19$m=65536,t=3,p=4$salt$bad-base64", // Bad hash base64
		}

		for _, h := range invalidHashes {
			match, err := ph.ComparePassword("password", h)
			assert.Error(t, err, "expected error for hash: %s", h)
			assert.False(t, match)
		}
	})
}
