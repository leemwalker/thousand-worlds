package api

import (
	"encoding/json"
	"net/http"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/errors"
	"mud-platform-backend/internal/validation"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService    *auth.Service
	sessionManager *auth.SessionManager
	rateLimiter    *auth.RateLimiter
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service, sessionManager *auth.SessionManager, rateLimiter *auth.RateLimiter) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		sessionManager: sessionManager,
		rateLimiter:    rateLimiter,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput, 
			"Failed to parse request body", err))
		return
	}

	// Validate input using validation layer
	validator := validation.New()
	validationErrs := &validation.ValidationErrors{}
	
	validationErrs.Add(validator.ValidateEmail(req.Email))
	validationErrs.Add(validator.ValidatePassword(req.Password))
	
	if validationErrs.HasErrors() {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
			validationErrs.Error(), nil))
		return
	}

	// Create user
	user, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrUserExists {
			errors.RespondWithError(w, errors.Wrap(errors.ErrConflict,
				"User already exists", err))
			return
		}
		errors.RespondWithError(w, errors.Wrap(errors.ErrInternalServer,
			"Failed to create user", err))
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
			"Failed to parse request body", err))
		return
	}

	// Validate input
	validator := validation.New()
	validationErrs := &validation.ValidationErrors{}
	
	validationErrs.Add(validator.ValidateRequired(req.Email, "email"))
	validationErrs.Add(validator.ValidateRequired(req.Password, "password"))
	
	if validationErrs.HasErrors() {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
			validationErrs.Error(), nil))
		return
	}

	// Authenticate
	token, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			errors.RespondWithError(w, errors.Wrap(errors.ErrUnauthorized,
				"Invalid credentials", err))
			return
		}
		errors.RespondWithError(w, errors.Wrap(errors.ErrInternalServer,
			"Login failed", err))
		return
	}

	respondJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// GetMe returns the current authenticated user
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
	})
}

// Logout invalidates the current session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session ID from header or query
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
	}

	if sessionID != "" && h.sessionManager != nil {
		if err := h.sessionManager.InvalidateSession(r.Context(), sessionID); err != nil {
			// Log error but don't fail - session may already be expired
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
