package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	InitLogger()

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correlation ID is present
		cid := GetCorrelationID(r.Context())
		assert.NotEmpty(t, cid)

		// Verify logger is in context
		logger := FromContext(r.Context())
		assert.NotNil(t, logger)

		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddleware_ExistingCorrelationID(t *testing.T) {
	InitLogger()

	existingID := "existing-id-123"

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := GetCorrelationID(r.Context())
		assert.Equal(t, existingID, cid)
	}))

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Correlation-ID", existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
}
