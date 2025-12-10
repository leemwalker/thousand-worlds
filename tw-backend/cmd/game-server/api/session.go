package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"

	"tw-backend/internal/auth"
	"tw-backend/internal/character"
	"tw-backend/internal/errors"
	"tw-backend/internal/game/constants"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/validation"
)

type SessionHandler struct {
	authRepo    auth.Repository
	lookService *look.LookService
}

func NewSessionHandler(authRepo auth.Repository, lookService *look.LookService) *SessionHandler {
	return &SessionHandler{
		authRepo:    authRepo,
		lookService: lookService,
	}
}

// CreateCharacterRequest represents the request to create a new character
type CreateCharacterRequest struct {
	WorldID     uuid.UUID `json:"world_id"`
	Name        string    `json:"name"`
	Species     string    `json:"species"`
	Role        string    `json:"role,omitempty"`
	Appearance  string    `json:"appearance,omitempty"`
	Description string    `json:"description,omitempty"`
	Occupation  string    `json:"occupation,omitempty"`
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
		errors.RespondWithError(w, errors.ErrUnauthorized)
		return
	}

	var req CreateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
			"Failed to parse request body", err))
		return
	}

	// Validate inputs using validation layer
	validator := validation.New()
	validationErrs := &validation.ValidationErrors{}

	validationErrs.Add(validator.ValidateRequired(req.Name, "name"))
	validationErrs.Add(validator.ValidateStringLength(req.Name, "name", 1, 50))
	validationErrs.Add(validator.ValidateUUID(req.WorldID, "world_id"))

	if req.Role != "" {
		validationErrs.Add(validator.ValidateOneOf(req.Role, "role", []string{"player", "watcher", "admin"}))
	}

	if req.Description != "" {
		validationErrs.Add(validator.ValidateStringLength(req.Description, "description", 0, 500))
	}

	if req.Occupation != "" {
		validationErrs.Add(validator.ValidateStringLength(req.Occupation, "occupation", 0, 100))
	}

	if validationErrs.HasErrors() {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
			validationErrs.Error(), nil))
		return
	}

	// Check if character already exists for this world
	existingChar, err := h.authRepo.GetCharacterByUserAndWorld(r.Context(), userID, req.WorldID)
	if err == nil && existingChar != nil {
		// Character exists
		if req.Role == "watcher" && existingChar.Role == "watcher" {
			// Idempotent success for watchers - return existing character
			respondJSON(w, http.StatusOK, CreateCharacterResponse{
				Character:      existingChar,
				Attributes:     nil, // Watchers have no attributes
				SecondaryAttrs: nil,
			})
			return
		}
		// Failure for other cases (e.g. trying to create a 2nd player char)
		// Or trying to create a player when you are a watcher (must delete watcher first? or switch?)
		// For now, return Conflict
		errors.RespondWithError(w, errors.Wrap(errors.ErrConflict,
			"You already have a character in this world", nil))
		return
	} else if err != nil && err != auth.ErrCharacterNotFound {
		// DB Error
		errors.RespondWithError(w, errors.Wrap(errors.ErrInternalServer,
			"Failed to check existing characters", err))
		return
	}

	// Get species template (skip for watcher role)
	var template *character.SpeciesTemplate
	if req.Role != "watcher" {
		t := character.GetSpeciesTemplate(req.Species)
		if t.Name == "" {
			errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
				"Invalid species", nil))
			return
		}
		template = &t
	}

	// Create character in database
	char := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     req.WorldID,
		Name:        req.Name,
		Role:        req.Role,
		Appearance:  req.Appearance,
		Description: req.Description,
		Occupation:  req.Occupation,
		Position:    nil, // Will be set when character spawns
	}

	if char.Role == "" {
		char.Role = "player"
	}

	if err := h.authRepo.CreateCharacter(r.Context(), char); err != nil {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInternalServer,
			"Failed to create character", err))
		return
	}

	// Calculate secondary attributes (skip for watcher)
	if req.Role == "watcher" || template == nil {
		respondJSON(w, http.StatusCreated, CreateCharacterResponse{
			Character:      char,
			Attributes:     nil,
			SecondaryAttrs: nil,
		})
		return
	}

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

	// 1. Load world state
	// 2. Set character spawn position
	// 3. Initialize game session

	// Requirement: On login, player must spawn in Lobby.
	if !constants.IsLobby(char.WorldID) {
		currentWorldID := char.WorldID
		char.LastWorldVisited = &currentWorldID
		char.WorldID = constants.LobbyWorldID
		char.PositionX = 500.0
		char.PositionY = 500.0
		char.PositionZ = 0.0

		if err := h.authRepo.UpdateCharacter(r.Context(), char); err != nil {
			log.Printf("Failed to update character spawn to lobby: %v", err)
			// Continue anyway? Or fail? Best to fail if we can't ensure consistency
			respondError(w, http.StatusInternalServerError, "Failed to join game")
			return
		}
	}

	// Update user's LastWorldID (legacy/user preference tracking)
	user, err := h.authRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		// Should not happen as we have a valid token
		log.Printf("Failed to get user for JoinGame: %v", err)
	}

	welcomeMsg := "Welcome back to Thousand Worlds"
	if user != nil && user.LastWorldID == nil {
		welcomeMsg = "Welcome to Thousand Worlds, we're excited you're here! If you need help at any point, type 'help' and 'help new' to get some pointers. Enjoy!"
	}

	if err == nil {
		user.LastWorldID = &char.WorldID
		// We ignore the error here as it's non-critical for gameplay
		_ = h.authRepo.UpdateUser(r.Context(), user)
	}

	// Get initial view
	dc := look.DescribeContext{
		WorldID:     char.WorldID,
		Character:   char,
		Orientation: "", // No orientation on spawn unless we set it
		DetailLevel: 1,
	}
	viewDesc, err := h.lookService.Describe(r.Context(), dc)
	if err != nil {
		log.Printf("Failed to get initial view: %v", err)
		viewDesc = "You have entered the world."
	}

	respondJSON(w, http.StatusOK, JoinGameResponse{
		Character: char,
		WorldID:   char.WorldID,
		Message:   fmt.Sprintf("%s\n\n%s", welcomeMsg, viewDesc),
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
