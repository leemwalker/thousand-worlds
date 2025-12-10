package memory

import (
	"sort"
	"time"
)

// RelevanceContext defines the current situation for relevance scoring
type RelevanceContext struct {
	Tags            []string
	CurrentLocation *Location
	CurrentTime     time.Time
}

// CalculateImportance computes the base importance of a memory
// Formula: baseImportance × (1 + recencyBonus + accessBonus)
func CalculateImportance(memory Memory) float64 {
	// Base importance: emotionalWeight × clarity
	// If clarity or emotion is 0, we ensure a minimum importance to avoid 0
	base := memory.EmotionalWeight * memory.Clarity
	if base < 0.01 {
		base = 0.01
	}

	// Recency bonus: 1.0 - (daysSinceCreation / 365) capped at 0.0
	daysSince := time.Since(memory.Timestamp).Hours() / 24.0
	recencyBonus := 1.0 - (daysSince / 365.0)
	if recencyBonus < 0 {
		recencyBonus = 0
	}

	// Access frequency bonus: min(accessCount / 10, 1.0)
	accessBonus := float64(memory.AccessCount) / 10.0
	if accessBonus > 1.0 {
		accessBonus = 1.0
	}

	return base * (1.0 + recencyBonus + accessBonus)
}

// CalculateRelevance computes the relevance score for a memory in a given context
// Formula: (recency × 0.3) + (emotionalWeight × 0.4) + (accessCount × 0.1) + (contextMatch × 0.2)
func CalculateRelevance(memory Memory, context RelevanceContext) float64 {
	// Normalize recency (0-1): 1.0 = now, 0.0 = 1 year ago
	daysSince := context.CurrentTime.Sub(memory.Timestamp).Hours() / 24.0
	recency := 1.0 - (daysSince / 365.0)
	if recency < 0 {
		recency = 0
	}

	// Normalize access count (0-1): 10+ accesses = 1.0
	access := float64(memory.AccessCount) / 10.0
	if access > 1.0 {
		access = 1.0
	}

	// Context match (0-1): Percentage of context tags found in memory tags
	matchScore := 0.0
	if len(context.Tags) > 0 {
		matches := 0
		for _, ctxTag := range context.Tags {
			for _, memTag := range memory.Tags {
				if ctxTag == memTag {
					matches++
					break
				}
			}
		}
		matchScore = float64(matches) / float64(len(context.Tags))
	}

	// Weighted sum
	score := (recency * 0.3) + (memory.EmotionalWeight * 0.4) + (access * 0.1) + (matchScore * 0.2)
	return score
}

// GetRelevantMemories sorts memories by relevance score
func GetRelevantMemories(memories []Memory, context RelevanceContext) []Memory {
	type ScoredMemory struct {
		Memory Memory
		Score  float64
	}

	scored := make([]ScoredMemory, len(memories))
	for i, m := range memories {
		scored[i] = ScoredMemory{
			Memory: m,
			Score:  CalculateRelevance(m, context),
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	result := make([]Memory, len(memories))
	for i, s := range scored {
		result[i] = s.Memory
	}
	return result
}
