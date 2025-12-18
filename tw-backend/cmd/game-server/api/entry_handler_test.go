package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tw-backend/internal/game/entry"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEntryProvider struct {
	mock.Mock
}

func (m *MockEntryProvider) GetEntryOptions(ctx context.Context, worldID uuid.UUID) (*entry.EntryOptions, error) {
	args := m.Called(ctx, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entry.EntryOptions), args.Error(1)
}

func TestGetEntryOptions(t *testing.T) {
	mockService := new(MockEntryProvider)

	handler := NewEntryHandler(mockService)

	worldID := uuid.New()
	options := &entry.EntryOptions{
		CanEnterAsWatcher: true,
		AvailableNPCs:     []entry.NPCPreview{},
		CanCreateCustom:   true,
	}

	mockService.On("GetEntryOptions", mock.Anything, worldID).Return(options, nil)

	req, _ := http.NewRequest("GET", "/entry?world_id="+worldID.String(), nil)
	rr := httptest.NewRecorder()
	handler.GetEntryOptions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp entry.EntryOptions
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.True(t, resp.CanEnterAsWatcher)
	assert.True(t, resp.CanCreateCustom)
}

func TestGetEntryOptions_MissingWorldID(t *testing.T) {
	mockService := new(MockEntryProvider)
	handler := NewEntryHandler(mockService)

	req, _ := http.NewRequest("GET", "/entry", nil)
	rr := httptest.NewRecorder()
	handler.GetEntryOptions(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
