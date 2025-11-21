package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPinger struct {
	mock.Mock
}

func (m *MockPinger) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockNATS struct {
	mock.Mock
}

func (m *MockNATS) Status() nats.Status {
	args := m.Called()
	return args.Get(0).(nats.Status)
}

func TestHealthChecker_Check(t *testing.T) {
	t.Run("all healthy", func(t *testing.T) {
		db := new(MockPinger)
		redis := new(MockPinger)
		nc := new(MockNATS)

		db.On("Ping", mock.Anything).Return(nil)
		redis.On("Ping", mock.Anything).Return(nil)
		nc.On("Status").Return(nats.CONNECTED)

		hc := NewHealthChecker(db, redis, nc)
		ctx := context.Background()

		status := hc.Check(ctx)
		assert.Equal(t, "ok", status["status"])
		assert.Equal(t, "healthy", status["database"])
		assert.Equal(t, "healthy", status["redis"])
		assert.Equal(t, "healthy", status["nats"])
	})

	t.Run("database unhealthy", func(t *testing.T) {
		db := new(MockPinger)
		redis := new(MockPinger)
		nc := new(MockNATS)

		db.On("Ping", mock.Anything).Return(errors.New("connection refused"))
		redis.On("Ping", mock.Anything).Return(nil)
		nc.On("Status").Return(nats.CONNECTED)

		hc := NewHealthChecker(db, redis, nc)
		ctx := context.Background()

		status := hc.Check(ctx)
		assert.Equal(t, "degraded", status["status"])
		assert.Equal(t, "unhealthy", status["database"])
	})
}

func TestHealthChecker_Handler(t *testing.T) {
	db := new(MockPinger)
	redis := new(MockPinger)
	nc := new(MockNATS)

	db.On("Ping", mock.Anything).Return(nil)
	redis.On("Ping", mock.Anything).Return(nil)
	nc.On("Status").Return(nats.CONNECTED)

	hc := NewHealthChecker(db, redis, nc)

	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler := hc.Handler()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	if rr.Code != http.StatusOK {
		t.Logf("Response Body: %s", rr.Body.String())
		t.Logf("Status: %v", response)
	}

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok", response["status"])
}
