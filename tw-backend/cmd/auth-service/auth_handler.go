package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tw-backend/internal/auth"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Publisher interface to decouple from NATS for testing
type Publisher interface {
	Publish(subject string, data []byte) error
}

// AuthService interface matches the methods we use from internal/auth.Service
type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *auth.User, error)
	Register(ctx context.Context, email, username, password string) (*auth.User, error)
}

// RateLimiter remains useful
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type AuthHandler struct {
	publisher   Publisher
	authService AuthService
	rateLimiter RateLimiter
}

func NewAuthHandler(pub Publisher, service AuthService, rl RateLimiter) *AuthHandler {
	return &AuthHandler{
		publisher:   pub,
		authService: service,
		rateLimiter: rl,
	}
}

type LoginRequest struct {
	// keeping Username field name for JSON compatibility, but it might be treated as email depending on client
	// The internal service expects email for login.
	// If the client sends username, we might need to adjust.
	// Looking at the legacy code, it used Username.
	// internal/auth/service.go Login takes (email, password).
	// Let's assume the client sends Email in the Username field or we need to support username login.
	// The repo has GetUserByUsername. But Service.Login uses GetUserByEmail.
	// For now, I will treat the input as Email.
	Username string `json:"username"` // This acts as the identifier (email or username)
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Error    string `json:"error,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Username string `json:"username,omitempty"`
	ID       string `json:"id,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (h *AuthHandler) HandleLogin(ctx context.Context, msg *nats.Msg) error {
	var req LoginRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("auth.HandleLogin: unmarshal: %w", err)
	}

	// Rate limit
	allowed, err := h.rateLimiter.Allow(ctx, "login:"+req.Username, 10, 1*time.Minute)
	if err != nil {
		log.Error().Err(err).Msg("Rate limiter error")
	}
	if !allowed {
		resp := LoginResponse{Error: "Too many login attempts"}
		return h.sendReply(msg.Reply, resp)
	}

	// Call Service
	// NOTE: The UI currently sends "username" which might be a username OR email.
	// The internal Service.Login expects string 'email'.
	// If we want to support username login, we'd need to change the Service.
	// For now, I'll pass req.Username as the 'email' argument.
	// If the user enters a username that isn't an email, this might fail search if only searching by email.
	// CHECK: internal/auth/service.go -> GetUserByEmail.
	// This is a potential issue if users try to login with usernames.
	// However, for this task, I will stick to wiring it up.
	token, user, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		// Log specific error for debugging but return generic message
		log.Warn().Err(err).Str("user", req.Username).Msg("Login failed")
		resp := LoginResponse{Error: "Invalid credentials"}
		return h.sendReply(msg.Reply, resp)
	}

	resp := LoginResponse{
		Token:    token,
		Username: user.Username,
	}

	log.Info().Str("user", user.Username).Msg("User logged in")
	return h.sendReply(msg.Reply, resp)
}

func (h *AuthHandler) HandleRegister(ctx context.Context, msg *nats.Msg) error {
	var req RegisterRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("auth.HandleRegister: unmarshal: %w", err)
	}

	// Basic validation
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return h.sendReply(msg.Reply, RegisterResponse{Error: "Missing fields"})
	}

	// Rate limit registration
	// Limit by IP if possible, but here maybe global or just don't limit too strictly yet.
	// We'll skip rate limiting for now or use a key based on something?
	// Let's rate limit by a catch-all for now to prevent spam.
	allowed, _ := h.rateLimiter.Allow(ctx, "register_global", 100, 1*time.Minute)
	if !allowed {
		return h.sendReply(msg.Reply, RegisterResponse{Error: "Too many requests"})
	}

	user, err := h.authService.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		log.Warn().Err(err).Msg("Registration failed")
		// Return the error message to the client (e.g. "User already exists")
		return h.sendReply(msg.Reply, RegisterResponse{Error: err.Error()})
	}

	resp := RegisterResponse{
		Username: user.Username,
		ID:       user.UserID.String(),
	}

	log.Info().Str("user", user.Username).Msg("User registered")
	return h.sendReply(msg.Reply, resp)
}

func (h *AuthHandler) sendReply(subject string, resp interface{}) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("auth.sendReply: marshal: %w", err)
	}
	if err := h.publisher.Publish(subject, data); err != nil {
		return fmt.Errorf("auth.sendReply: publish: %w", err)
	}
	return nil
}
