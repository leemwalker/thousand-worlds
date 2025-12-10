package effects

import "time"

// NewManager creates a new effect manager
func NewManager() *EffectManager {
	return &EffectManager{
		Modifiers: make([]*StatModifier, 0),
	}
}

// TickResult aggregates damage from DoTs
type TickResult struct {
	PoisonDamage int
	BleedDamage  int
}

// Tick processes all active effects
func (m *EffectManager) Tick(now time.Time) TickResult {
	res := TickResult{}

	// Poison
	if m.Poison != nil {
		if m.Poison.IsExpired(now) {
			m.Poison = nil
		} else {
			res.PoisonDamage = m.Poison.Tick(now)
		}
	}

	// Stun
	if m.Stun != nil {
		if now.After(m.Stun.EndsAt) {
			m.Stun = nil
		}
	}

	// Slow
	if m.Slow != nil {
		if now.Sub(m.Slow.AppliedAt) >= m.Slow.Duration {
			m.Slow = nil
		}
	}

	// Bleed
	if m.Bleed != nil {
		if m.Bleed.IsExpired(now) {
			m.Bleed = nil
		} else {
			res.BleedDamage = m.Bleed.Tick(now)
		}
	}

	// Modifiers cleanup
	activeModifiers := make([]*StatModifier, 0, len(m.Modifiers))
	for _, mod := range m.Modifiers {
		if now.Sub(mod.AppliedAt) < mod.Duration {
			activeModifiers = append(activeModifiers, mod)
		}
	}
	m.Modifiers = activeModifiers

	return res
}
