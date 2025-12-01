package auth

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
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
// Implements batch updates to reduce Redis write frequency
type SessionManager struct {
	client *redis.Client
	ttl    time.Duration

	// In-memory cache for LastAccess times
	// Flushed to Redis periodically to reduce write ops
	lastAccessCache map[string]time.Time
	cacheMu         sync.RWMutex

	// Background flush control
	flushInterval time.Duration
	stopFlush     chan struct{}
	flushDone     chan struct{}
}

// NewSessionManager creates a new SessionManager.
// Starts background goroutine for periodic session flush
func NewSessionManager(client *redis.Client) *SessionManager {
	sm := &SessionManager{
		client:          client,
		ttl:             24 * time.Hour,
		lastAccessCache: make(map[string]time.Time),
		flushInterval:   5 * time.Minute,
		stopFlush:       make(chan struct{}),
		flushDone:       make(chan struct{}),
	}

	// Start background flush worker
	go sm.flushWorker()

	return sm
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
// LastAccess is tracked in-memory and flushed periodically
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

	// Update LastAccess in memory only (batched to Redis later)
	now := time.Now().UTC()
	sm.cacheMu.Lock()
	sm.lastAccessCache[sessionID] = now
	sm.cacheMu.Unlock()

	session.LastAccess = now
	return &session, nil
}

// InvalidateSession removes a session.
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	// Remove from in-memory cache
	sm.cacheMu.Lock()
	delete(sm.lastAccessCache, sessionID)
	sm.cacheMu.Unlock()

	key := "session:" + sessionID
	return sm.client.Del(ctx, key).Err()
}

// Close stops the background flush worker and performs final flush
func (sm *SessionManager) Close(ctx context.Context) error {
	close(sm.stopFlush)
	<-sm.flushDone // Wait for flush worker to finish

	// Final flush
	return sm.flushSessions(ctx)
}

// flushWorker runs in background, periodically flushing session updates to Redis
func (sm *SessionManager) flushWorker() {
	defer close(sm.flushDone)

	ticker := time.NewTicker(sm.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := sm.flushSessions(ctx); err != nil {
				// Log error but continue (non-critical)
				// In production, use proper logger
			}
			cancel()
		case <-sm.stopFlush:
			return
		}
	}
}

// flushSessions writes all pending LastAccess updates to Redis
// Complexity: O(N) where N = sessions with pending updates (not all sessions)
func (sm *SessionManager) flushSessions(ctx context.Context) error {
	sm.cacheMu.Lock()
	if len(sm.lastAccessCache) == 0 {
		sm.cacheMu.Unlock()
		return nil
	}

	// Copy cache and clear
	pendingUpdates := make(map[string]time.Time, len(sm.lastAccessCache))
	for k, v := range sm.lastAccessCache {
		pendingUpdates[k] = v
	}
	sm.lastAccessCache = make(map[string]time.Time)
	sm.cacheMu.Unlock()

	// Update Redis (outside lock)
	for sessionID, lastAccess := range pendingUpdates {
		key := "session:" + sessionID

		// Get current session data
		data, err := sm.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				// Session expired, skip
				continue
			}
			return err
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue // Skip malformed sessions
		}

		// Update LastAccess
		session.LastAccess = lastAccess
		updatedData, err := json.Marshal(session)
		if err != nil {
			continue
		}

		// Write back with extended TTL
		if err := sm.client.Set(ctx, key, updatedData, sm.ttl).Err(); err != nil {
			return err
		}
	}

	return nil
}
