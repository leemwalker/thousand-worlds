package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"tw-backend/internal/auth"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockPublisher struct{ mock.Mock }

func (m *MockPublisher) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

type MockTokenManager struct{ mock.Mock }

func (m *MockTokenManager) GenerateToken(userID, username string, roles []string) (string, error) {
	args := m.Called(userID, username, roles)
	return args.String(0), args.Error(1)
}

type MockPasswordHasher struct{ mock.Mock }

func (m *MockPasswordHasher) ComparePassword(password, encodedHash string) (bool, error) {
	args := m.Called(password, encodedHash)
	return args.Bool(0), args.Error(1)
}
func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

type MockSessionManager struct{ mock.Mock }

func (m *MockSessionManager) CreateSession(ctx context.Context, userID, username string) (*auth.Session, error) {
	args := m.Called(ctx, userID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Session), args.Error(1)
}

type MockRateLimiter struct{ mock.Mock }

func (m *MockRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Error(1)
}

func TestHandleLogin_Success(t *testing.T) {
	// Setup Mocks
	mockPub := new(MockPublisher)
	mockTM := new(MockTokenManager)
	mockPH := new(MockPasswordHasher)
	mockSM := new(MockSessionManager)
	mockRL := new(MockRateLimiter)

	handler := NewAuthHandler(mockPub, mockTM, mockPH, mockSM, mockRL)

	req := LoginRequest{Username: "admin", Password: "password123"}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{
		Data:  reqData,
		Reply: "reply-subject",
	}

	ctx := context.Background()

	// Expectations
	mockRL.On("Allow", ctx, "login:admin", 5, time.Minute).Return(true, nil)
	mockPH.On("HashPassword", "password123").Return("hashed", nil)
	mockPH.On("ComparePassword", "password123", "hashed").Return(true, nil)
	mockSM.On("CreateSession", ctx, "user-admin-id", "admin").Return(&auth.Session{ID: "sess-1"}, nil)
	mockTM.On("GenerateToken", "user-admin-id", "admin", []string{"admin"}).Return("valid-token", nil)

	// Expect response publish
	mockPub.On("Publish", "reply-subject", mock.MatchedBy(func(data []byte) bool {
		var resp LoginResponse
		json.Unmarshal(data, &resp)
		return resp.Token == "valid-token" && resp.Username == "admin" && resp.Error == ""
	})).Return(nil)

	// Execute
	err := handler.HandleLogin(ctx, msg)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestHandleLogin_RateLimit(t *testing.T) {
	mockPub := new(MockPublisher)
	mockRL := new(MockRateLimiter)
	// Other mocks not needed for early return

	handler := NewAuthHandler(mockPub, nil, nil, nil, mockRL) // nil for unused

	req := LoginRequest{Username: "admin"}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{Data: reqData, Reply: "reply"}

	mockRL.On("Allow", mock.Anything, "login:admin", 5, time.Minute).Return(false, nil)
	mockPub.On("Publish", "reply", mock.MatchedBy(func(data []byte) bool {
		var resp LoginResponse
		json.Unmarshal(data, &resp)
		return resp.Error != ""
	})).Return(nil)

	err := handler.HandleLogin(context.Background(), msg)
	assert.NoError(t, err)
}
