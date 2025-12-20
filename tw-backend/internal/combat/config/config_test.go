package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	// Verify all defaults match current hardcoded values
	assert.Equal(t, 200.0, cfg.SkillDivisor)
	assert.Equal(t, 200.0, cfg.MightDivisor)
	assert.Equal(t, 200.0, cfg.AgilityDivisor)
	assert.Equal(t, 400.0, cfg.MixedAttributeDivisor)
	assert.Equal(t, 100.0, cfg.RollDivisor)

	// Critical settings
	assert.Equal(t, 5, cfg.CriticalFailureThreshold)
	assert.Equal(t, 95, cfg.CriticalHitBaseThreshold)
	assert.Equal(t, 50, cfg.CunningBonusDivisor)
	assert.Equal(t, 5, cfg.HeavyAttackBonus)
	assert.Equal(t, 2.0, cfg.CriticalMultiplier)
	assert.Equal(t, 0.5, cfg.CriticalIgnoreArmor)
}

func TestLoadFromFile(t *testing.T) {
	// Create temp config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "combat_config.json")

	configJSON := `{
		"skill_divisor": 250.0,
		"might_divisor": 180.0,
		"agility_divisor": 220.0,
		"mixed_attribute_divisor": 350.0,
		"roll_divisor": 100.0,
		"critical_failure_threshold": 3,
		"critical_hit_base_threshold": 90,
		"cunning_bonus_divisor": 40,
		"heavy_attack_bonus": 7,
		"critical_multiplier": 2.5,
		"critical_ignore_armor": 0.6
	}`
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	// Load and verify
	cfg, err := LoadFromFile(configPath)
	require.NoError(t, err)

	assert.Equal(t, 250.0, cfg.SkillDivisor)
	assert.Equal(t, 180.0, cfg.MightDivisor)
	assert.Equal(t, 220.0, cfg.AgilityDivisor)
	assert.Equal(t, 350.0, cfg.MixedAttributeDivisor)
	assert.Equal(t, 3, cfg.CriticalFailureThreshold)
	assert.Equal(t, 90, cfg.CriticalHitBaseThreshold)
	assert.Equal(t, 40, cfg.CunningBonusDivisor)
	assert.Equal(t, 7, cfg.HeavyAttackBonus)
	assert.Equal(t, 2.5, cfg.CriticalMultiplier)
	assert.Equal(t, 0.6, cfg.CriticalIgnoreArmor)
}

func TestLoadFromFile_FileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.json")
	assert.Error(t, err)
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")
	err := os.WriteFile(configPath, []byte("not valid json"), 0644)
	require.NoError(t, err)

	_, err = LoadFromFile(configPath)
	assert.Error(t, err)
}

func TestReload(t *testing.T) {
	// Create temp config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "combat_config.json")

	// Write initial config
	initialJSON := `{"skill_divisor": 200.0}`
	err := os.WriteFile(configPath, []byte(initialJSON), 0644)
	require.NoError(t, err)

	// Load initial
	cfg, err := LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, 200.0, cfg.SkillDivisor)

	// Modify file
	updatedJSON := `{"skill_divisor": 300.0}`
	err = os.WriteFile(configPath, []byte(updatedJSON), 0644)
	require.NoError(t, err)

	// Reload
	err = cfg.Reload(configPath)
	require.NoError(t, err)
	assert.Equal(t, 300.0, cfg.SkillDivisor)
}
