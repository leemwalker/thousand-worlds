package api

import (
	"context"
	"log"
	"net/http"
	"strings"

	"mud-platform-backend/internal/auth"

	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[AUTH] Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

			// Get token from cookie (most secure), Authorization header, or query parameter (WebSocket)
			var token string

			// Priority 1: Check HttpOnly cookie
			if cookie, err := r.Cookie("auth_token"); err == nil && cookie.Value != "" {
				token = cookie.Value
				log.Printf("[AUTH] Found token in cookie")
			} else {
				// Priority 2: Check Authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "" {
					log.Printf("[AUTH] Found Authorization header")
					// Extract token from "Bearer <token>"
					parts := strings.Split(authHeader, " ")
					if len(parts) != 2 || parts[0] != "Bearer" {
						log.Printf("[AUTH] Invalid authorization format")
						respondError(w, http.StatusUnauthorized, "Invalid authorization format")
						return
					}
					token = parts[1]
				} else {
					// Priority 3: Check query parameter (for WebSocket connections)
					token = r.URL.Query().Get("token")
					if token == "" {
						log.Printf("[AUTH] No token found in cookie, header, or query")
						respondError(w, http.StatusUnauthorized, "Missing authorization")
						return
					}
					log.Printf("[AUTH] Found token in query parameter: %s...", token[:min(20, len(token))])
				}
			}

			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				log.Printf("[AUTH] Token validation failed: %v", err)
				respondError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			log.Printf("[AUTH] Token validated for user %s", claims.UserID)

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
