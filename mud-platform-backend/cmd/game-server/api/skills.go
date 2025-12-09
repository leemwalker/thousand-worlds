package api

import (
	"encoding/json"
	"net/http"

	"mud-platform-backend/internal/skills"

	"github.com/google/uuid"
)

type SkillsHandler struct {
	service *skills.Service
}

func NewSkillsHandler(service *skills.Service) *SkillsHandler {
	return &SkillsHandler{service: service}
}

func (h *SkillsHandler) HandleGetSkills(w http.ResponseWriter, r *http.Request) {
	// Parse character_id from query params
	characterIDStr := r.URL.Query().Get("character_id")
	if characterIDStr == "" {
		http.Error(w, "character_id is required", http.StatusBadRequest)
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		http.Error(w, "invalid character_id", http.StatusBadRequest)
		return
	}

	// We could also enforce that the logged-in user owns this character using request context user info.
	// For now, assuming public or frontend context validity.
	// Ideally:
	// userID := r.Context().Value(auth.ContextKeyUserID).(uuid.UUID)
	// Check if user owns character...

	sheet, err := h.service.GetSkillSheet(r.Context(), characterID)
	if err != nil {
		http.Error(w, "failed to get skills", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sheet)
}
