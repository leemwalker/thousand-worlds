// Package ecosystem provides turning point presentation for player interaction.
package ecosystem

import (
	"fmt"
	"strings"
)

// TurningPointPresentation contains formatted text for player display
type TurningPointPresentation struct {
	Header       string   // Title and description
	Options      []string // Formatted options (0-3)
	Instructions string   // How to respond
}

// FormatTurningPoint creates a player-friendly presentation of a turning point
func FormatTurningPoint(tp *TurningPoint) *TurningPointPresentation {
	if tp == nil {
		return nil
	}

	var sb strings.Builder

	// Header with emoji based on trigger type
	emoji := getTriggerEmoji(tp.Trigger)
	sb.WriteString(fmt.Sprintf("%s TURNING POINT: %s\n", emoji, tp.Title))
	sb.WriteString(fmt.Sprintf("Year %d\n", tp.Year))
	sb.WriteString("\n")
	sb.WriteString(tp.Description)
	sb.WriteString("\n\n")

	// World state summary
	sb.WriteString(fmt.Sprintf("Current State: %d species (%d extinct)\n",
		tp.TotalSpecies, tp.ExtinctSpecies))
	sb.WriteString("\n")

	// Format options - always include option 0 as "observe"
	options := make([]string, 0, 4)
	options = append(options, "0. Observe (let nature take its course)")

	// Add up to 3 interventions from available list
	interventionCount := 0
	for _, intervention := range tp.Interventions {
		if intervention.Type == InterventionNone {
			continue // Already added as option 0
		}
		if interventionCount >= 3 {
			break
		}
		interventionCount++

		optionText := formatIntervention(interventionCount, &intervention)
		options = append(options, optionText)
	}

	return &TurningPointPresentation{
		Header:       sb.String(),
		Options:      options,
		Instructions: "Enter 0-3 to choose, or 4 for different options.",
	}
}

// formatIntervention creates a readable option string for an intervention
func formatIntervention(num int, intervention *Intervention) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%d. %s", num, intervention.Name))

	if intervention.Description != "" {
		sb.WriteString(fmt.Sprintf(" - %s", intervention.Description))
	}

	// Add cost if significant
	if intervention.Cost > 0 {
		sb.WriteString(fmt.Sprintf(" (Cost: %d)", intervention.Cost))
	}

	// Add risk level indicator
	if intervention.RiskLevel >= 0.7 {
		sb.WriteString(" ⚠️ High Risk")
	} else if intervention.RiskLevel >= 0.4 {
		sb.WriteString(" ⚡ Moderate Risk")
	}

	return sb.String()
}

// getTriggerEmoji returns an appropriate emoji for the trigger type
func getTriggerEmoji(trigger TurningPointTrigger) string {
	switch trigger {
	case TriggerExtinction:
		return "💀"
	case TriggerSapience:
		return "🧠"
	case TriggerInterval:
		return "🔮"
	case TriggerClimateShift:
		return "🌡️"
	case TriggerTectonicEvent:
		return "🌋"
	case TriggerPandemic:
		return "🦠"
	default:
		return "⭐"
	}
}

// FormatTurningPointMessage creates a complete message string for display
func FormatTurningPointMessage(tp *TurningPoint) string {
	presentation := FormatTurningPoint(tp)
	if presentation == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(presentation.Header)
	sb.WriteString("Choose your intervention:\n")
	for _, opt := range presentation.Options {
		sb.WriteString(opt)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString(presentation.Instructions)

	return sb.String()
}

// ParseTurningPointResponse parses player input and returns the selected intervention
// Returns nil for option 0 (observe) or if invalid input
func ParseTurningPointResponse(tp *TurningPoint, input string) *Intervention {
	if tp == nil {
		return nil
	}

	input = strings.TrimSpace(input)

	switch input {
	case "0":
		// Observe - return nil to indicate no intervention
		return nil
	case "1", "2", "3":
		// Get corresponding intervention
		optNum := int(input[0] - '0')
		interventionIndex := 0
		for i, intervention := range tp.Interventions {
			if intervention.Type == InterventionNone {
				continue
			}
			interventionIndex++
			if interventionIndex == optNum {
				return &tp.Interventions[i]
			}
		}
		return nil
	case "4":
		// Request different options - handled by caller
		return nil
	default:
		return nil
	}
}

// NeedsRegeneration returns true if input is "4" (request different options)
func NeedsRegeneration(input string) bool {
	return strings.TrimSpace(input) == "4"
}
