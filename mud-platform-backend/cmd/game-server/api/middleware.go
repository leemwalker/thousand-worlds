package api

import (
	"context"
	"net/http"
	"strings"

	"mud-platform-backend/internal/auth"

	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header or query parameter (for WebSocket)
			var token string
			authHeader := r.Header.Get("Authorization")

			if authHeader != "" {
				// Extract token from "Bearer <token>"
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					respondError(w, http.StatusUnauthorized, "Invalid authorization format")
					return
				}
				token = parts[1]
			} else {
				// Check query parameter (for WebSocket connections)
				token = r.URL.Query().Get("token")
				if token == "" {
					respondError(w, http.StatusUnauthorized, "Missing authorization")
					return
				}
			}

			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getUserIDFromContext retrieves user ID from context
func getUserIDFromContext(ctx context.Context) uuid.UUID {
	if userIDStr, ok := ctx.Value("userID").(string); ok {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return userID
		}
	}
	return uuid.Nil
}

// getCharacterIDFromContext retrieves character ID from context
func getCharacterIDFromContext(ctx context.Context) uuid.UUID {
	if characterIDStr, ok := ctx.Value("characterID").(string); ok {
		if characterID, err := uuid.Parse(characterIDStr); err == nil {
			return characterID
		}
	}
	return uuid.Nil
}
