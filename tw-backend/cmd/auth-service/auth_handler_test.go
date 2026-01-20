package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"tw-backend/internal/auth"

	"github.com/google/uuid"
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

type MockAuthService struct{ mock.Mock }

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, *auth.User, error) {
	args := m.Called(ctx, email, password)
	token := args.String(0)
	user := args.Get(1)
	if user == nil {
		return token, nil, args.Error(2)
	}
	return token, user.(*auth.User), args.Error(2)
}

func (m *MockAuthService) Register(ctx context.Context, email, username, password string) (*auth.User, error) {
	args := m.Called(ctx, email, username, password)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*auth.User), args.Error(1)
}

type MockRateLimiter struct{ mock.Mock }

func (m *MockRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Error(1)
}

func TestHandleLogin_Success(t *testing.T) {
	// Setup Mocks
	mockPub := new(MockPublisher)
	mockService := new(MockAuthService)
	mockRL := new(MockRateLimiter)

	handler := NewAuthHandler(mockPub, mockService, mockRL)

	req := LoginRequest{Username: "admin", Password: "password123"}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{
		Data:  reqData,
		Reply: "reply-subject",
	}

	ctx := context.Background()

	// Expectations
	mockRL.On("Allow", ctx, "login:admin", 10, time.Minute).Return(true, nil)

	expectedUser := &auth.User{
		Username: "admin",
		UserID:   uuid.New(),
	}
	mockService.On("Login", ctx, "admin", "password123").Return("valid-token", expectedUser, nil)

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
	mockService := new(MockAuthService)

	handler := NewAuthHandler(mockPub, mockService, mockRL)

	req := LoginRequest{Username: "admin"}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{Data: reqData, Reply: "reply"} // Reply subject is "reply"

	mockRL.On("Allow", mock.Anything, "login:admin", 10, time.Minute).Return(false, nil)
	mockPub.On("Publish", "reply", mock.MatchedBy(func(data []byte) bool {
		var resp LoginResponse
		json.Unmarshal(data, &resp)
		return resp.Error == "Too many login attempts"
	})).Return(nil)

	err := handler.HandleLogin(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleLogin_ServiceError(t *testing.T) {
	mockPub := new(MockPublisher)
	mockService := new(MockAuthService)
	mockRL := new(MockRateLimiter)

	handler := NewAuthHandler(mockPub, mockService, mockRL)

	req := LoginRequest{Username: "admin", Password: "wrong"}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{Data: reqData, Reply: "reply"}

	mockRL.On("Allow", mock.Anything, "login:admin", 10, time.Minute).Return(true, nil)
	mockService.On("Login", mock.Anything, "admin", "wrong").Return("", nil, errors.New("invalid credentials"))

	mockPub.On("Publish", "reply", mock.MatchedBy(func(data []byte) bool {
		var resp LoginResponse
		json.Unmarshal(data, &resp)
		return resp.Error == "Invalid credentials"
	})).Return(nil)

	err := handler.HandleLogin(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleRegister_Success(t *testing.T) {
	mockPub := new(MockPublisher)
	mockService := new(MockAuthService)
	mockRL := new(MockRateLimiter)

	handler := NewAuthHandler(mockPub, mockService, mockRL)

	req := RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{Data: reqData, Reply: "reply"}

	mockRL.On("Allow", mock.Anything, "register_global", 100, time.Minute).Return(true, nil)

	createdUser := &auth.User{
		Username: "newuser",
		UserID:   uuid.New(),
		Email:    "new@example.com",
	}
	mockService.On("Register", mock.Anything, "new@example.com", "newuser", "password123").Return(createdUser, nil)

	mockPub.On("Publish", "reply", mock.MatchedBy(func(data []byte) bool {
		var resp RegisterResponse
		json.Unmarshal(data, &resp)
		return resp.Username == "newuser" && resp.ID == createdUser.UserID.String() && resp.Error == ""
	})).Return(nil)

	err := handler.HandleRegister(context.Background(), msg)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestHandleRegister_MissingFields(t *testing.T) {
	handler := NewAuthHandler(new(MockPublisher), nil, nil)
	mockPub := handler.publisher.(*MockPublisher)

	// Missing email
	req := RegisterRequest{Username: "user", Password: "pass"}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{Data: reqData, Reply: "reply"}

	mockPub.On("Publish", "reply", mock.MatchedBy(func(data []byte) bool {
		var resp RegisterResponse
		json.Unmarshal(data, &resp)
		return resp.Error == "Missing fields"
	})).Return(nil)

	err := handler.HandleRegister(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleRegister_ServiceError(t *testing.T) {
	mockPub := new(MockPublisher)
	mockService := new(MockAuthService)
	mockRL := new(MockRateLimiter)

	handler := NewAuthHandler(mockPub, mockService, mockRL)

	req := RegisterRequest{
		Username: "user",
		Email:    "exists@example.com",
		Password: "pass",
	}
	reqData, _ := json.Marshal(req)
	msg := &nats.Msg{Data: reqData, Reply: "reply"}

	mockRL.On("Allow", mock.Anything, "register_global", 100, time.Minute).Return(true, nil)
	mockService.On("Register", mock.Anything, "exists@example.com", "user", "pass").Return(nil, errors.New("user exists"))

	mockPub.On("Publish", "reply", mock.MatchedBy(func(data []byte) bool {
		var resp RegisterResponse
		json.Unmarshal(data, &resp)
		return resp.Error == "user exists"
	})).Return(nil)

	err := handler.HandleRegister(context.Background(), msg)
	assert.NoError(t, err)
}
