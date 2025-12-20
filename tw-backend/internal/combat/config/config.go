// Package config provides externalized combat configuration
// allowing game balancing without recompilation.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// CombatConfig holds all combat-related coefficients and thresholds.
// Values can be loaded from JSON for runtime configuration.
type CombatConfig struct {
	mu sync.RWMutex

	// Damage calculation divisors
	SkillDivisor          float64 `json:"skill_divisor"`
	MightDivisor          float64 `json:"might_divisor"`
	AgilityDivisor        float64 `json:"agility_divisor"`
	MixedAttributeDivisor float64 `json:"mixed_attribute_divisor"`
	RollDivisor           float64 `json:"roll_divisor"`

	// Critical hit settings
	CriticalFailureThreshold int     `json:"critical_failure_threshold"`
	CriticalHitBaseThreshold int     `json:"critical_hit_base_threshold"`
	CunningBonusDivisor      int     `json:"cunning_bonus_divisor"`
	HeavyAttackBonus         int     `json:"heavy_attack_bonus"`
	CriticalMultiplier       float64 `json:"critical_multiplier"`
	CriticalIgnoreArmor      float64 `json:"critical_ignore_armor"`
}

// Default returns a CombatConfig with values matching the original hardcoded constants.
func Default() *CombatConfig {
	return &CombatConfig{
		// Damage calculation divisors (from calculator.go)
		SkillDivisor:          200.0,
		MightDivisor:          200.0,
		AgilityDivisor:        200.0,
		MixedAttributeDivisor: 400.0,
		RollDivisor:           100.0,

		// Critical hit settings (from critical.go)
		CriticalFailureThreshold: 5,
		CriticalHitBaseThreshold: 95,
		CunningBonusDivisor:      50,
		HeavyAttackBonus:         5,
		CriticalMultiplier:       2.0,
		CriticalIgnoreArmor:      0.5,
	}
}

// LoadFromFile loads combat configuration from a JSON file.
func LoadFromFile(path string) (*CombatConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Start with defaults so missing fields use default values
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return cfg, nil
}

// Reload reloads the configuration from the specified file path.
// Thread-safe for use with SIGHUP handlers.
func (c *CombatConfig) Reload(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse into temporary struct
	temp := Default()
	if err := json.Unmarshal(data, temp); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Atomically update all fields
	c.mu.Lock()
	defer c.mu.Unlock()

	c.SkillDivisor = temp.SkillDivisor
	c.MightDivisor = temp.MightDivisor
	c.AgilityDivisor = temp.AgilityDivisor
	c.MixedAttributeDivisor = temp.MixedAttributeDivisor
	c.RollDivisor = temp.RollDivisor
	c.CriticalFailureThreshold = temp.CriticalFailureThreshold
	c.CriticalHitBaseThreshold = temp.CriticalHitBaseThreshold
	c.CunningBonusDivisor = temp.CunningBonusDivisor
	c.HeavyAttackBonus = temp.HeavyAttackBonus
	c.CriticalMultiplier = temp.CriticalMultiplier
	c.CriticalIgnoreArmor = temp.CriticalIgnoreArmor

	return nil
}

// Thread-safe getters for hot-reload support

// GetSkillDivisor returns the skill divisor (thread-safe).
func (c *CombatConfig) GetSkillDivisor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SkillDivisor
}

// GetMightDivisor returns the might divisor (thread-safe).
func (c *CombatConfig) GetMightDivisor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MightDivisor
}

// GetAgilityDivisor returns the agility divisor (thread-safe).
func (c *CombatConfig) GetAgilityDivisor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AgilityDivisor
}

// GetMixedAttributeDivisor returns the mixed attribute divisor (thread-safe).
func (c *CombatConfig) GetMixedAttributeDivisor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MixedAttributeDivisor
}

// GetRollDivisor returns the roll divisor (thread-safe).
func (c *CombatConfig) GetRollDivisor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RollDivisor
}

// GetCriticalFailureThreshold returns the critical failure threshold (thread-safe).
func (c *CombatConfig) GetCriticalFailureThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CriticalFailureThreshold
}

// GetCriticalHitBaseThreshold returns the critical hit base threshold (thread-safe).
func (c *CombatConfig) GetCriticalHitBaseThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CriticalHitBaseThreshold
}

// GetCunningBonusDivisor returns the cunning bonus divisor (thread-safe).
func (c *CombatConfig) GetCunningBonusDivisor() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CunningBonusDivisor
}

// GetHeavyAttackBonus returns the heavy attack bonus (thread-safe).
func (c *CombatConfig) GetHeavyAttackBonus() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HeavyAttackBonus
}

// GetCriticalMultiplier returns the critical multiplier (thread-safe).
func (c *CombatConfig) GetCriticalMultiplier() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CriticalMultiplier
}

// GetCriticalIgnoreArmor returns the critical ignore armor percentage (thread-safe).
func (c *CombatConfig) GetCriticalIgnoreArmor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CriticalIgnoreArmor
}
