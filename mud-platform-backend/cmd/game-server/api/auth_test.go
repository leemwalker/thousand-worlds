package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mud-platform-backend/internal/auth"
)

// MockRepository is a mock implementation of auth.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Implement other interface methods with stubs
func (m *MockRepository) CreateCharacter(ctx context.Context, char *auth.Character) error {
	return nil
}
func (m *MockRepository) GetCharacter(ctx context.Context, id string) (*auth.Character, error) {
	return nil, nil
}
func (m *MockRepository) GetUserCharacters(ctx context.Context, id string) ([]*auth.Character, error) {
	return nil, nil
}
func (m *MockRepository) GetCharacterByUserAndWorld(ctx context.Context, uid, wid string) (*auth.Character, error) {
	return nil, nil
}
func (m *MockRepository) UpdateCharacter(ctx context.Context, char *auth.Character) error {
	return nil
}

// Note: We need to adapt the mock to match the actual interface types (uuid.UUID)
// But since we are defining the mock here, we can just use the real types if we import them.
// However, the interface in internal/auth/types.go uses uuid.UUID.
// So let's use the real MockRepository from internal/auth if possible, or define a proper one here.
// The internal/auth/mock_repository.go exists, let's use that one if it's exported.
// It is exported as NewMockRepository.

func TestAuthHandler_Register(t *testing.T) {
	// Setup
	repo := auth.NewMockRepository()
	config := &auth.Config{
		SecretKey:       []byte("test-secret"),
		TokenExpiration: time.Hour,
	}
	service := auth.NewService(config, repo)
	handler := NewAuthHandler(service, nil, nil)

	t.Run("Successful Registration", func(t *testing.T) {
		payload := RegisterRequest{
			Email:    "newuser@example.com",
			Password: "Password123", // Now needs uppercase, lowercase, and digit
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp auth.User
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, payload.Email, resp.Email)
		assert.Empty(t, resp.PasswordHash) // Should not return hash
	})

	t.Run("Invalid Input", func(t *testing.T) {
		payload := RegisterRequest{
			Email:    "",
			Password: "short",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	repo := auth.NewMockRepository()
	config := &auth.Config{
		SecretKey:       []byte("test-secret"),
		TokenExpiration: time.Hour,
	}
	service := auth.NewService(config, repo)
	handler := NewAuthHandler(service, nil, nil)

	// Create a user first
	ctx := context.Background()
	_, err := service.Register(ctx, "existing@example.com", "Password123")
	assert.NoError(t, err)

	t.Run("Successful Login", func(t *testing.T) {
		payload := LoginRequest{
			Email:    "existing@example.com",
			Password: "Password123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp LoginResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, payload.Email, resp.User.Email)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		payload := LoginRequest{
			Email:    "existing@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
