package auth

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// MockRepository is an in-memory repository for testing/development
type MockRepository struct {
	users      map[uuid.UUID]*User
	characters map[uuid.UUID]*Character
	userEmails map[string]uuid.UUID // email -> userID
	mu         sync.RWMutex
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:      make(map[uuid.UUID]*User),
		characters: make(map[uuid.UUID]*Character),
		userEmails: make(map[string]uuid.UUID),
	}
}

// CreateUser creates a new user
func (r *MockRepository) CreateUser(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	if _, exists := r.userEmails[user.Email]; exists {
		return errors.New("user already exists")
	}

	r.users[user.UserID] = user
	r.userEmails[user.Email] = user.UserID
	return nil
}

// GetUserByID retrieves a user by ID
func (r *MockRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, exists := r.userEmails[email]
	if !exists {
		return nil, nil // Not found is not an error for this method
	}

	return r.users[userID], nil
}

// UpdateUser updates an existing user
func (r *MockRepository) UpdateUser(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.UserID]; !exists {
		return errors.New("user not found")
	}

	r.users[user.UserID] = user
	return nil
}

// CreateCharacter creates a new character
func (r *MockRepository) CreateCharacter(ctx context.Context, char *Character) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.characters[char.CharacterID] = char
	return nil
}

// GetCharacter retrieves a character by ID
func (r *MockRepository) GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	char, exists := r.characters[characterID]
	if !exists {
		return nil, errors.New("character not found")
	}
	return char, nil
}

// GetUserCharacters retrieves all characters for a user
func (r *MockRepository) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*Character, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var characters []*Character
	for _, char := range r.characters {
		if char.UserID == userID {
			characters = append(characters, char)
		}
	}
	return characters, nil
}

// GetCharacterByUserAndWorld retrieves a character by user and world
func (r *MockRepository) GetCharacterByUserAndWorld(ctx context.Context, userID, worldID uuid.UUID) (*Character, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, char := range r.characters {
		if char.UserID == userID && char.WorldID == worldID {
			return char, nil
		}
	}
	return nil, nil
}

// UpdateCharacter updates an existing character
func (r *MockRepository) UpdateCharacter(ctx context.Context, char *Character) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.characters[char.CharacterID]; !exists {
		return errors.New("character not found")
	}

	r.characters[char.CharacterID] = char
	return nil
}
