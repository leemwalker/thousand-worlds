package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Register(t *testing.T) {
	repo := NewMockRepository()
	s := NewService(&Config{SecretKey: []byte("secret"), TokenExpiration: time.Hour}, repo)
	ctx := context.Background()

	// Success
	user, err := s.Register(ctx, "test@example.com", "user1", "pass")
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)

	// Duplicate
	_, err = s.Register(ctx, "test@example.com", "user2", "pass")
	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
}

func TestService_Login(t *testing.T) {
	repo := NewMockRepository()
	s := NewService(&Config{SecretKey: []byte("secret"), TokenExpiration: time.Hour}, repo)
	ctx := context.Background()

	_, err := s.Register(ctx, "test@example.com", "user1", "pass")
	require.NoError(t, err)

	// Success
	token, user, err := s.Login(ctx, "test@example.com", "pass")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "user1", user.Username)

	// Invalid Password
	_, _, err = s.Login(ctx, "test@example.com", "wrong")
	assert.Equal(t, ErrInvalidCredentials, err)

	// User Not Found
	_, _, err = s.Login(ctx, "missing@example.com", "pass")
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestService_TokenValidation(t *testing.T) {
	repo := NewMockRepository()
	s := NewService(&Config{SecretKey: []byte("secret"), TokenExpiration: time.Hour}, repo)

	userID := uuid.New()
	token, err := s.GenerateToken(userID, uuid.Nil)
	require.NoError(t, err)

	claims, err := s.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)

	// Invalid Token
	_, err = s.ValidateToken("invalid.token.here")
	assert.Error(t, err)
}

func TestService_GetUserByID(t *testing.T) {
	repo := NewMockRepository()
	s := NewService(&Config{}, repo)
	ctx := context.Background()

	user, _ := s.Register(ctx, "test@example.com", "user1", "pass")

	fetched, err := s.GetUserByID(ctx, user.UserID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, fetched.Email)

	_, err = s.GetUserByID(ctx, uuid.New())
	assert.Error(t, err)
}
