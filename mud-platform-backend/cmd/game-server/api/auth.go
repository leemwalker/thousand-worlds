package api

import (
	"encoding/json"
	"net/http"

	"mud-platform-backend/internal/auth"
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Create user
	user, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrUserExists {
			respondError(w, http.StatusConflict, "User already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Authenticate
	token, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, "Login failed")
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
