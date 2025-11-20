package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNATS is a mock for NATS connection/subscription if needed, 
// but for handler logic we might just test the handler function directly 
// if we decouple it well. 
// For now, let's assume we want to test the "HandleLogin" method of an AuthHandler struct.

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func TestAuthHandler_HandleLogin(t *testing.T) {
	t.Run("Successful Login", func(t *testing.T) {
		// Setup
		mockPub := new(MockPublisher)
		handler := NewAuthHandler(mockPub)

		payload := LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		data, _ := json.Marshal(payload)
		msg := &nats.Msg{
			Subject: "auth.login",
			Data:    data,
			Reply:   "auth.login.reply",
		}

		// Expect a reply with a token
		mockPub.On("Publish", msg.Reply, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			// Verify the response body
			respData := args.Get(1).([]byte)
			var resp LoginResponse
			err := json.Unmarshal(respData, &resp)
			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Token)
			assert.Equal(t, "testuser", resp.Username)
		})

		err := handler.HandleLogin(context.Background(), msg)
		assert.NoError(t, err)
		mockPub.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		// Setup
		mockPub := new(MockPublisher)
		handler := NewAuthHandler(mockPub)

		// For simplicity in this TDD step, let's assume any password < 8 chars is invalid
		// or we mock a DB. Let's just enforce a simple rule for now.
		payload := LoginRequest{
			Username: "testuser",
			Password: "short",
		}
		data, _ := json.Marshal(payload)
		msg := &nats.Msg{
			Subject: "auth.login",
			Data:    data,
			Reply:   "auth.login.reply",
		}

		// Expect an error response
		mockPub.On("Publish", msg.Reply, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			respData := args.Get(1).([]byte)
			var resp LoginResponse
			err := json.Unmarshal(respData, &resp)
			assert.NoError(t, err)
			assert.Empty(t, resp.Token)
			assert.NotEmpty(t, resp.Error)
		})

		err := handler.HandleLogin(context.Background(), msg)
		assert.NoError(t, err) // The handler itself shouldn't error, it should send an error response
		mockPub.AssertExpectations(t)
	})
}
