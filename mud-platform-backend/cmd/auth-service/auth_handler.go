package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Publisher interface to decouple from NATS for testing
type Publisher interface {
	Publish(subject string, data []byte) error
}

type AuthHandler struct {
	publisher Publisher
}

func NewAuthHandler(pub Publisher) *AuthHandler {
	return &AuthHandler{
		publisher: pub,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (h *AuthHandler) HandleLogin(ctx context.Context, msg *nats.Msg) error {
	var req LoginRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("auth.HandleLogin: unmarshal: %w", err)
	}

	// TODO: Replace with real DB check
	// Simple validation for TDD
	if len(req.Password) < 8 {
		resp := LoginResponse{
			Error: "Invalid credentials",
		}
		return h.sendReply(msg.Reply, resp)
	}

	// Generate a fake token
	token := fmt.Sprintf("token-%s-%d", req.Username, time.Now().Unix())

	resp := LoginResponse{
		Token:    token,
		Username: req.Username,
	}

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
