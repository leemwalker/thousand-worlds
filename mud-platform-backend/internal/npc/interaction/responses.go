package interaction

import (
	"math/rand"
	"mud-platform-backend/internal/npc/personality"
	"strings"
)

// GenerateResponse creates a response based on personality and relationship
func GenerateResponse(p *personality.Personality, affection int, topic Topic) (string, string) {
	// Determine Response Type
	// High Agreeableness + High Affection -> Agreeable
	// Low Agreeableness + Low Affection -> Disagreeable

	agreeableness := p.Agreeableness.Value
	score := (agreeableness + float64(affection)) / 2.0 // -50 to 100 range roughly

	var responseType string
	if score > 60 {
		responseType = "agreeable"
	} else if score < 30 {
		responseType = "disagreeable"
	} else {
		responseType = "neutral"
	}

	// Select Template
	template := ResponseTemplates[responseType]

	// Select Adjective
	var adj string
	if responseType == "agreeable" {
		adj = Adjectives["positive"][rand.Intn(len(Adjectives["positive"]))]
	} else if responseType == "disagreeable" {
		adj = Adjectives["negative"][rand.Intn(len(Adjectives["negative"]))]
	} else {
		adj = "interesting"
	}

	text := strings.Replace(template, "{adjective}", adj, 1)
	return text, responseType
}
