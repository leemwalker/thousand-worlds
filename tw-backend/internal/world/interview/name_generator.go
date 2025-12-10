package interview

import (
	"fmt"
	"strings"
)

// NameGenerator generates world names based on interview context
type NameGenerator struct {
	client LLMClient
}

// NewNameGenerator creates a new name generator
func NewNameGenerator(client LLMClient) *NameGenerator {
	return &NameGenerator{client: client}
}

// GenerateWorldNames generates unique world name suggestions based on interview answers
func (g *NameGenerator) GenerateWorldNames(session *InterviewSession, count int) ([]string, error) {
	// Build context from interview answers
	theme := session.State.Answers["Core Concept"]
	species := session.State.Answers["Sentient Species"]
	environment := session.State.Answers["Environment"]
	techMagic := session.State.Answers["Magic & Tech"]
	conflict := session.State.Answers["Conflict"]

	prompt := fmt.Sprintf(`Based on the following world description, generate EXACTLY %d unique, creative world names. Output only the names, one per line, with NO numbering, bullets, or other formatting.

World Description:
- Core Concept: %s
- Sentient Species: %s
- Environment: %s
- Technology & Magic: %s
- Central Conflict: %s

The names should:
- Be memorable and evocative
- Reflect the world's theme and character
- Be between 1-4 words
- Not use generic patterns like "World of X" or "Realm of Y"
- Sound like actual place names

Generate %d names:`, count, theme, species, environment, techMagic, conflict, count)

	response, err := g.client.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate names: %w", err)
	}

	// Parse names from response
	lines := strings.Split(response, "\n")
	var uniqueNames []string
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove any numbering, bullets, or markdown formatting
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Remove numbering like "1." or "1)"
		if len(line) > 2 && line[0] >= '0' && line[0] <= '9' {
			if line[1] == '.' || line[1] == ')' {
				line = strings.TrimSpace(line[2:])
			}
		}

		if line != "" {
			lowerName := strings.ToLower(line)
			if !seen[lowerName] {
				seen[lowerName] = true
				uniqueNames = append(uniqueNames, line)
			}
		}

		// Stop once we have enough unique names
		if len(uniqueNames) >= count {
			break
		}
	}

	// If we didn't get enough names, return what we have
	if len(uniqueNames) == 0 {
		return nil, fmt.Errorf("failed to extract names from LLM response")
	}

	return uniqueNames, nil
}
