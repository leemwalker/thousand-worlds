package gateway

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriber is a mock implementation of Subscriber
type MockSubscriber struct {
	mock.Mock
	callback nats.MsgHandler
}

func (m *MockSubscriber) Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	m.callback = cb
	args := m.Called(subj, cb)
	return nil, args.Error(1)
}

func TestStartListener_Success(t *testing.T) {
	// Drain queue first
loop:
	for {
		select {
		case <-RequestQueue:
		default:
			break loop
		}
	}

	mockSub := new(MockSubscriber)
	mockSub.On("Subscribe", "ai.request.>", mock.Anything).Return(nil, nil)

	err := StartListener(mockSub)
	assert.NoError(t, err)

	// Simulate message
	req := AIRequest{ID: "123", Prompt: "test"}
	data, _ := json.Marshal(req)
	msg := &nats.Msg{
		Subject: "ai.request.123",
		Data:    data,
	}

	mockSub.callback(msg)

	// Verify queue
	select {
	case received := <-RequestQueue:
		assert.Equal(t, "123", received.ID)
		assert.Equal(t, "ai.response.123", received.ResponseSubject)
	case <-time.After(1 * time.Second):
		t.Fatal("Request not queued")
	}
}

func TestStartListener_InvalidJSON(t *testing.T) {
	mockSub := new(MockSubscriber)
	mockSub.On("Subscribe", "ai.request.>", mock.Anything).Return(nil, nil)

	StartListener(mockSub)

	// Simulate invalid message
	msg := &nats.Msg{
		Subject: "ai.request.123",
		Data:    []byte("invalid json"),
	}

	mockSub.callback(msg)

	// Verify queue is empty
	select {
	case <-RequestQueue:
		t.Fatal("Invalid request queued")
	default:
		// OK
	}
}

func TestStartListener_QueueFull(t *testing.T) {
	// Fill queue
	for i := 0; i < 100; i++ {
		RequestQueue <- AIRequest{ID: "filler"}
	}

	mockSub := new(MockSubscriber)
	mockSub.On("Subscribe", "ai.request.>", mock.Anything).Return(nil, nil)

	StartListener(mockSub)

	// Simulate message
	req := AIRequest{ID: "overflow"}
	data, _ := json.Marshal(req)
	msg := &nats.Msg{
		Subject: "ai.request.overflow",
		Data:    data,
	}

	mockSub.callback(msg)

	// Verify queue still full (should not block)
	assert.Len(t, RequestQueue, 100)

	// Drain to clean up
loop:
	for {
		select {
		case <-RequestQueue:
		default:
			break loop
		}
	}
}
