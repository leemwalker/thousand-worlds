// Package ecosystem provides turning point presentation for player interaction.
package ecosystem

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// TurningPointPresentation contains formatted text for player display
type TurningPointPresentation struct {
	Header       string   // Title and description
	Options      []string // Formatted options (0-3)
	Instructions string   // How to respond
}

// SpeciesOption represents a species available for targeting
type SpeciesOption struct {
	Index      int
	SpeciesID  uuid.UUID
	Name       string
	Population int64
	Status     string // "endangered", "at_risk", "stable"
}

// SpeciesSelectionPresentation contains formatted species selection
type SpeciesSelectionPresentation struct {
	Header       string
	Species      []SpeciesOption
	Instructions string
}

// FormatTurningPoint creates a player-friendly presentation of a turning point
func FormatTurningPoint(tp *TurningPoint, divineEnergy int) *TurningPointPresentation {
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

	// Divine Energy display
	sb.WriteString(fmt.Sprintf("⚡ Divine Energy: %d\n", divineEnergy))

	// World state summary
	sb.WriteString(fmt.Sprintf("🌍 Species: %d living, %d extinct\n",
		tp.TotalSpecies-tp.ExtinctSpecies, tp.ExtinctSpecies))
	sb.WriteString("\n")

	// Format options - always include option 0 as "observe"
	options := make([]string, 0, 4)
	options = append(options, "0. Observe (let nature take its course) - FREE")

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

		optionText := formatIntervention(interventionCount, &intervention, divineEnergy)
		options = append(options, optionText)
	}

	return &TurningPointPresentation{
		Header:       sb.String(),
		Options:      options,
		Instructions: "Enter 0-3 to choose, or 4 for different options.",
	}
}

// formatIntervention creates a readable option string for an intervention
func formatIntervention(num int, intervention *Intervention, divineEnergy int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%d. %s", num, intervention.Name))

	if intervention.Description != "" {
		sb.WriteString(fmt.Sprintf(" - %s", intervention.Description))
	}

	// Add cost with affordability indicator
	if intervention.Cost > 0 {
		if divineEnergy >= intervention.Cost {
			sb.WriteString(fmt.Sprintf(" (Cost: %d⚡)", intervention.Cost))
		} else {
			sb.WriteString(fmt.Sprintf(" (Cost: %d⚡ ❌ INSUFFICIENT)", intervention.Cost))
		}
	}

	// Add targeting info
	if intervention.TargetType == "species" {
		sb.WriteString(" [Targets: Species]")
	} else if intervention.TargetType == "biome" {
		sb.WriteString(" [Targets: Biome]")
	}

	// Add risk level indicator
	if intervention.RiskLevel >= 0.7 {
		sb.WriteString(" ⚠️ High Risk")
	} else if intervention.RiskLevel >= 0.4 {
		sb.WriteString(" ⚡ Moderate Risk")
	}

	return sb.String()
}

// FormatSpeciesSelection creates a species selection prompt for targeted interventions
func FormatSpeciesSelection(intervention *Intervention, species []SpeciesOption) *SpeciesSelectionPresentation {
	if intervention == nil || len(species) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You chose: %s\n\n", intervention.Name))
	sb.WriteString("Select a species to target:\n")

	return &SpeciesSelectionPresentation{
		Header:       sb.String(),
		Species:      species,
		Instructions: "Enter species number (or 0 to go back):",
	}
}

// FormatSpeciesSelectionMessage creates the complete species selection message
func FormatSpeciesSelectionMessage(intervention *Intervention, species []SpeciesOption) string {
	presentation := FormatSpeciesSelection(intervention, species)
	if presentation == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(presentation.Header)

	for _, sp := range species {
		statusEmoji := getStatusEmoji(sp.Status)
		sb.WriteString(fmt.Sprintf("%d. %s (Pop: %d) %s %s\n",
			sp.Index, sp.Name, sp.Population, statusEmoji, sp.Status))
	}

	sb.WriteString("\n")
	sb.WriteString(presentation.Instructions)

	return sb.String()
}

// getStatusEmoji returns emoji for species status
func getStatusEmoji(status string) string {
	switch status {
	case "endangered":
		return "🔴"
	case "at_risk":
		return "🟡"
	case "stable":
		return "🟢"
	default:
		return "⚪"
	}
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
func FormatTurningPointMessage(tp *TurningPoint, divineEnergy int) string {
	presentation := FormatTurningPoint(tp, divineEnergy)
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

// ParseSpeciesSelection parses species selection input
// Returns the selected species index (1-based) or 0 if going back
func ParseSpeciesSelection(input string, maxIndex int) int {
	input = strings.TrimSpace(input)

	if input == "0" {
		return 0 // Go back
	}

	if len(input) == 1 && input[0] >= '1' && input[0] <= '9' {
		index := int(input[0] - '0')
		if index <= maxIndex {
			return index
		}
	}

	return -1 // Invalid
}

// NeedsRegeneration returns true if input is "4" (request different options)
func NeedsRegeneration(input string) bool {
	return strings.TrimSpace(input) == "4"
}

// NeedsSpeciesSelection returns true if the intervention requires species targeting
func NeedsSpeciesSelection(intervention *Intervention) bool {
	if intervention == nil {
		return false
	}
	return intervention.TargetType == "species"
}
