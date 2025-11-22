package player

import (
	"fmt"
	"sync"
)

// StaminaManager handles stamina tracking for a character
type StaminaManager struct {
	mu             sync.RWMutex
	currentStamina int
	maxStamina     int
}

// NewStaminaManager creates a new manager initialized to max stamina
func NewStaminaManager(maxStamina int) *StaminaManager {
	return &StaminaManager{
		currentStamina: maxStamina,
		maxStamina:     maxStamina,
	}
}

// Current returns the current stamina
func (sm *StaminaManager) Current() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentStamina
}

// Max returns the max stamina
func (sm *StaminaManager) Max() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.maxStamina
}

// Consume deducts stamina if available
func (sm *StaminaManager) Consume(amount int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentStamina < amount {
		return fmt.Errorf("insufficient stamina")
	}
	sm.currentStamina -= amount
	return nil
}

// SetCurrent sets the current stamina directly (e.g. from event replay)
func (sm *StaminaManager) SetCurrent(amount int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentStamina = amount
}
