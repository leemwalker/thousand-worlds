package skills

import (
	"math/rand"
	"time"
)

// CheckResult represents the outcome of a skill check
type CheckResult struct {
	Success  bool
	Critical bool // True if critical success or critical failure
	Roll     int  // The base d100 roll
	Total    int  // Roll + Skill + Bonus
	Margin   int  // Total - Difficulty
}

// PerformCheck executes a skill check
// Roll: d100 + skillLevel + (attributeVal / 5)
// Critical Success: Natural 96-100
// Critical Failure: Natural 1-5
func PerformCheck(skillLevel int, attributeVal int, difficulty int) CheckResult {
	rand.Seed(time.Now().UnixNano()) // Ensure randomness
	roll := rand.Intn(100) + 1       // 1-100

	// Criticals based on natural roll
	if roll >= 96 {
		return CheckResult{
			Success:  true,
			Critical: true,
			Roll:     roll,
			Total:    roll + skillLevel + (attributeVal / 5),
			Margin:   (roll + skillLevel + (attributeVal / 5)) - difficulty,
		}
	}
	if roll <= 5 {
		return CheckResult{
			Success:  false,
			Critical: true,
			Roll:     roll,
			Total:    roll + skillLevel + (attributeVal / 5),
			Margin:   (roll + skillLevel + (attributeVal / 5)) - difficulty,
		}
	}

	total := roll + skillLevel + (attributeVal / 5)
	success := total >= difficulty

	return CheckResult{
		Success:  success,
		Critical: false,
		Roll:     roll,
		Total:    total,
		Margin:   total - difficulty,
	}
}

// PerformCheckWithSynergy executes a skill check including synergy bonuses
func PerformCheckWithSynergy(skillLevel int, attributeVal int, synergyBonus int, difficulty int) CheckResult {
	// We can reuse logic but we need to inject the synergy into the total.
	// However, criticals are based on natural roll, so synergy doesn't affect crit chance, only success.

	rand.Seed(time.Now().UnixNano())
	roll := rand.Intn(100) + 1

	if roll >= 96 {
		return CheckResult{
			Success:  true,
			Critical: true,
			Roll:     roll,
			Total:    roll + skillLevel + (attributeVal / 5) + synergyBonus,
			Margin:   (roll + skillLevel + (attributeVal / 5) + synergyBonus) - difficulty,
		}
	}
	if roll <= 5 {
		return CheckResult{
			Success:  false,
			Critical: true,
			Roll:     roll,
			Total:    roll + skillLevel + (attributeVal / 5) + synergyBonus,
			Margin:   (roll + skillLevel + (attributeVal / 5) + synergyBonus) - difficulty,
		}
	}

	total := roll + skillLevel + (attributeVal / 5) + synergyBonus
	success := total >= difficulty

	return CheckResult{
		Success:  success,
		Critical: false,
		Roll:     roll,
		Total:    total,
		Margin:   total - difficulty,
	}
}
