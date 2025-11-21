package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Session represents a user session.
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	LoginTime  time.Time `json:"login_time"`
	LastAccess time.Time `json:"last_access"`
}

// SessionManager handles session storage in Redis.
type SessionManager struct {
	client *redis.Client
	ttl    time.Duration
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(client *redis.Client) *SessionManager {
	return &SessionManager{
		client: client,
		ttl:    24 * time.Hour,
	}
}

// CreateSession creates a new session for a user.
func (sm *SessionManager) CreateSession(ctx context.Context, userID, username string) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now().UTC()

	session := &Session{
		ID:         sessionID,
		UserID:     userID,
		Username:   username,
		LoginTime:  now,
		LastAccess: now,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	key := "session:" + sessionID
	if err := sm.client.Set(ctx, key, data, sm.ttl).Err(); err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by ID and extends its TTL.
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := "session:" + sessionID
	data, err := sm.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	// Update LastAccess and extend TTL
	session.LastAccess = time.Now().UTC()
	updatedData, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	if err := sm.client.Set(ctx, key, updatedData, sm.ttl).Err(); err != nil {
		return nil, err
	}

	return &session, nil
}

// InvalidateSession removes a session.
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	key := "session:" + sessionID
	return sm.client.Del(ctx, key).Err()
}
