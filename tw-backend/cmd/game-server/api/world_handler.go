package api

import (
	"encoding/json"
	"net/http"

	"tw-backend/internal/repository"
)

type WorldHandler struct {
	repo repository.WorldRepository
}

func NewWorldHandler(repo repository.WorldRepository) *WorldHandler {
	return &WorldHandler{
		repo: repo,
	}
}

// ListWorlds returns a list of all worlds
func (h *WorldHandler) ListWorlds(w http.ResponseWriter, r *http.Request) {
	worlds, err := h.repo.ListWorlds(r.Context())
	if err != nil {
		http.Error(w, "Failed to list worlds", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worlds)
}
