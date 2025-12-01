package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecordHubBroadcast(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordHubBroadcast(100 * time.Millisecond)
	})
}

func TestRecordMessageProcessed(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordMessageProcessed("chat")
	})
}

func TestSetActiveConnections(t *testing.T) {
	assert.NotPanics(t, func() {
		SetActiveConnections(10)
	})
}

func TestRecordDBQuery(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordDBQuery("select", "users", 50*time.Millisecond)
	})
}

func TestRecordCacheHit(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordCacheHit()
	})
}

func TestRecordCacheMiss(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordCacheMiss()
	})
}
