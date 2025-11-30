package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/character"
)

type SessionHandler struct {
	authRepo auth.Repository
}

func NewSessionHandler(authRepo auth.Repository) *SessionHandler {
	return &SessionHandler{authRepo: authRepo}
}

// CreateCharacterRequest represents the request to create a new character
type CreateCharacterRequest struct {
	WorldID uuid.UUID `json:"world_id"`
	Name    string    `json:"name"`
	Species string    `json:"species"`
}

// CreateCharacterResponse represents the character creation response
type CreateCharacterResponse struct {
	Character      *auth.Character                `json:"character"`
	Attributes     *character.Attributes          `json:"attributes"`
	SecondaryAttrs *character.SecondaryAttributes `json:"secondary_attributes"`
}

func (h *SessionHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate inputs
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Character name is required")
		return
	}

	if req.WorldID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "World ID is required")
		return
	}

	// Get species template
	template := character.GetSpeciesTemplate(req.Species)
	if template.Name == "" {
		respondError(w, http.StatusBadRequest, "Invalid species")
		return
	}

	// Create character in database
	char := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     req.WorldID,
		Name:        req.Name,
		Position:    nil, // Will be set when character spawns
	}

	if err := h.authRepo.CreateCharacter(r.Context(), char); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create character: "+err.Error())
		return
	}

	// Calculate secondary attributes
	secAttrs := calculateSecondaryAttributes(&template.BaseAttrs)

	respondJSON(w, http.StatusCreated, CreateCharacterResponse{
		Character:      char,
		Attributes:     &template.BaseAttrs,
		SecondaryAttrs: secAttrs,
	})
}

// GetCharactersResponse represents the list of characters
type GetCharactersResponse struct {
	Characters []*auth.Character `json:"characters"`
}

func (h *SessionHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	characters, err := h.authRepo.GetUserCharacters(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve characters: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetCharactersResponse{
		Characters: characters,
	})
}

// JoinGameRequest represents the request to join a game world
type JoinGameRequest struct {
	CharacterID uuid.UUID `json:"character_id"`
}

// JoinGameResponse represents the join game response
type JoinGameResponse struct {
	Character *auth.Character `json:"character"`
	WorldID   uuid.UUID       `json:"world_id"`
	Message   string          `json:"message"`
}

func (h *SessionHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req JoinGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get character
	char, err := h.authRepo.GetCharacter(r.Context(), req.CharacterID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Character not found")
		return
	}

	// Verify character belongs to user
	if char.UserID != userID {
		respondError(w, http.StatusForbidden, "Not your character")
		return
	}

	// In a real implementation, we would:
	// 1. Load world state
	// 2. Set character spawn position
	// 3. Initialize game session
	// For now, just return success

	respondJSON(w, http.StatusOK, JoinGameResponse{
		Character: char,
		WorldID:   char.WorldID,
		Message:   "Successfully joined world. Connect via WebSocket to begin playing.",
	})
}

// Helper function to calculate secondary attributes
func calculateSecondaryAttributes(attrs *character.Attributes) *character.SecondaryAttributes {
	return &character.SecondaryAttributes{
		MaxHP:      (attrs.Vitality * 10) + (attrs.Endurance * 5),
		MaxStamina: (attrs.Endurance * 10) + (attrs.Agility * 5),
		MaxFocus:   (attrs.Intellect * 10) + (attrs.Willpower * 5),
		MaxMana:    (attrs.Willpower * 10) + (attrs.Intellect * 5),
		MaxNerve:   (attrs.Presence * 10) + (attrs.Willpower * 5),
	}
}
