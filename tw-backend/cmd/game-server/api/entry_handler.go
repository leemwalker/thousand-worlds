package api

import (
	"context"
	"encoding/json"
	"net/http"

	"tw-backend/internal/game/entry"

	"github.com/google/uuid"
)

type EntryProvider interface {
	GetEntryOptions(ctx context.Context, worldID uuid.UUID) (*entry.EntryOptions, error)
}

type EntryHandler struct {
	service EntryProvider
}

func NewEntryHandler(service EntryProvider) *EntryHandler {
	return &EntryHandler{
		service: service,
	}
}

// GetEntryOptions returns available entry modes for a world
func (h *EntryHandler) GetEntryOptions(w http.ResponseWriter, r *http.Request) {
	worldIDStr := r.URL.Query().Get("world_id")
	if worldIDStr == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		return
	}

	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "invalid world_id", http.StatusBadRequest)
		return
	}

	options, err := h.service.GetEntryOptions(r.Context(), worldID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}
