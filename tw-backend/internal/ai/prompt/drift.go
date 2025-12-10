package prompt

import (
	"fmt"
	"tw-backend/internal/npc/relationship"
)

func buildDriftSection(drift *relationship.DriftMetrics, baseline relationship.BehavioralProfile, current relationship.BehavioralProfile) string {
	if drift == nil {
		return ""
	}

	// Baseline
	baselineStr := fmt.Sprintf(`   - Aggression: %.1f
   - Generosity: %.1f
   - Honesty: %.1f
   - Sociability: %.1f
   - Recklessness: %.1f
   - Loyalty: %.1f`,
		baseline.Aggression, baseline.Generosity, baseline.Honesty,
		baseline.Sociability, baseline.Recklessness, baseline.Loyalty)

	// Current
	currentStr := fmt.Sprintf(`   - Aggression: %.1f
   - Generosity: %.1f
   - Honesty: %.1f
   - Sociability: %.1f
   - Recklessness: %.1f
   - Loyalty: %.1f`,
		current.Aggression, current.Generosity, current.Honesty,
		current.Sociability, current.Recklessness, current.Loyalty)

	// Instruction based on level
	instruction := ""
	switch drift.DriftLevel {
	case "Severe":
		instruction = "You are alarmed by this drastic personality change. This is deeply unsettling to you."
	case "Moderate":
		instruction = "You are genuinely concerned about this behavior change. Express this clearly."
	case "Subtle":
		instruction = "You've noticed a slight difference. Mention it casually if appropriate."
	default:
		instruction = "You notice some changes."
	}

	// Fill Template
	// We manually construct it here to match the template structure defined in templates.go logic
	// Actually, we can just return the struct data or formatted string.
	// Let's return the formatted string ready for injection.

	return fmt.Sprintf(`PERSONALITY DRIFT DETECTED:
- Original Baseline:
%s

- Current Behavior (last 20 actions):
%s

- Drift Level: %s

%s`, baselineStr, currentStr, drift.DriftLevel, instruction)
}
