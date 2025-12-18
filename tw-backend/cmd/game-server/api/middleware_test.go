package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tw-backend/internal/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService helpers
// Since AuthMiddleware takes *auth.Service struct, we can't mock it easily without an interface.
// HOWEVER, looking at auth.Service, it uses a Config and Repository.
// We can instantiate a real auth.Service with a MockRepository and custom Config.

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) CreateUser(ctx context.Context, user *auth.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *MockAuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *MockAuthRepo) UpdateUser(ctx context.Context, user *auth.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockAuthRepo) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthRepo) CreateCharacter(ctx context.Context, char *auth.Character) error {
	return m.Called(ctx, char).Error(0)
}

func (m *MockAuthRepo) GetCharacter(ctx context.Context, characterID uuid.UUID) (*auth.Character, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Character), args.Error(1)
}

func (m *MockAuthRepo) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*auth.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.Character), args.Error(1)
}

func (m *MockAuthRepo) GetCharacterByUserAndWorld(ctx context.Context, userID, worldID uuid.UUID) (*auth.Character, error) {
	args := m.Called(ctx, userID, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Character), args.Error(1)
}

func (m *MockAuthRepo) UpdateCharacter(ctx context.Context, char *auth.Character) error {
	return m.Called(ctx, char).Error(0)
}

func TestAuthMiddleware(t *testing.T) {
	// Setup real service with mock repo
	mockRepo := new(MockAuthRepo)
	secretKey := []byte("secret")
	config := &auth.Config{
		SecretKey:       secretKey,
		TokenExpiration: time.Hour,
	}
	authService := auth.NewService(config, mockRepo)

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := getUserIDFromContext(r.Context())
		if userID == uuid.Nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authService)
	handler := middleware(nextHandler)

	// Generate valid token
	userID := uuid.New()
	token, _ := authService.GenerateToken(userID, uuid.Nil)

	// Case 1: Valid Header
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case 2: Valid Cookie
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case 3: Missing Token
	req, _ = http.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Case 4: Invalid Token
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
