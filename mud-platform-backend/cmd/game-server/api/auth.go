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
	Username string `json:"username"`
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
	validationErrs.Add(validator.ValidateRequired(req.Username, "username"))
	validationErrs.Add(validator.ValidatePassword(req.Password))

	if validationErrs.HasErrors() {
		errors.RespondWithError(w, errors.Wrap(errors.ErrInvalidInput,
			validationErrs.Error(), nil))
		return
	}

	// Create user
	user, err := h.authService.Register(r.Context(), req.Email, req.Username, req.Password)
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

	// Set authentication token as HttpOnly cookie for security
	// This prevents XSS attacks from stealing the token
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,                    // Prevents JavaScript access
		Secure:   isSecureContext(r),      // HTTPS only in production
		SameSite: http.SameSiteStrictMode, // CSRF protection
		Path:     "/",
		MaxAge:   86400, // 24 hours (should match JWT expiration)
	})

	// Return user info only (no token in response)
	respondJSON(w, http.StatusOK, LoginResponse{
		Token: "", // Empty token field for backward compatibility
		User:  user,
	})
}

// GetMe returns the current authenticated user
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	// Get user from database to return full details
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": user.UserID,
		"email":   user.Email,
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

	// Clear the auth_token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // Delete cookie immediately
	})

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

// isSecureContext determines if the request is over HTTPS
func isSecureContext(r *http.Request) bool {
	// Check if request is HTTPS
	if r.TLS != nil {
		return true
	}
	// Check X-Forwarded-Proto header (for reverse proxies)
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}
