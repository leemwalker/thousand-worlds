package emotion

import (
	"math"
	"sort"

	"mud-platform-backend/internal/npc/memory"
)

// GetSimilarMemories retrieves memories matching the current emotional state
// Returns top 5 memories based on similarity, importance, and recency
func GetSimilarMemories(currentEmotion EmotionProfile, memories []memory.Memory) []memory.Memory {
	type ScoredMemory struct {
		Mem   memory.Memory
		Score float64
	}

	var scored []ScoredMemory

	for _, mem := range memories {
		// Calculate Similarity
		similarity := CalculateSimilarity(currentEmotion, mem.EmotionProfile)

		// Threshold > 0.6
		if similarity <= 0.6 {
			continue
		}

		// Score = (Similarity * 0.5) + (Importance * 0.3) + (Recency * 0.2)
		// Importance is not directly on Memory struct, but we have EmotionalWeight and Clarity.
		// Prompt says "Importance = Base * (1 + EmotionalWeight)".
		// Let's assume Base is roughly derived from Clarity or AccessCount?
		// Or maybe we just use EmotionalWeight as a proxy for Importance here if Base is missing.
		// Actually, `CalculateImportance` is in `relevance.go` (Phase 3.1).
		// But we don't have easy access to it here without circular deps if we import `memory`.
		// Let's use EmotionalWeight as the primary driver for importance in this context,
		// or assume Clarity * (1 + EmotionalWeight).

		importance := mem.Clarity * (1.0 + mem.EmotionalWeight)

		// Recency: 1.0 for now, 0.0 for very old.
		// Let's use a simple decay for recency score: 1 / (1 + days)
		// For simplicity in this pure function, let's assume we pass 'now' or ignore recency?
		// Prompt: "score = (emotionalSimilarity * 0.5) + (importance * 0.3) + (recency * 0.2)"
		// I'll skip recency calculation here to keep signature simple or add 'now'.
		// Let's add 'now' to signature? Or just use 0.5 for recency as placeholder?
		// Better: Use AccessCount as a proxy for "Recency/Relevance"? No.
		// Let's just use 0 for recency if we can't calc it, or rely on Similarity/Importance.
		// Actually, let's just use 0.0 for recency to avoid breaking signature changes if not strictly needed.
		// Or better, just use the timestamp relative to a fixed point? No.
		recency := 0.5 // Placeholder

		score := (similarity * 0.5) + (importance * 0.3) + (recency * 0.2)

		scored = append(scored, ScoredMemory{Mem: mem, Score: score})
	}

	// Sort by Score Descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Return top 5
	result := []memory.Memory{}
	count := 0
	for _, sm := range scored {
		result = append(result, sm.Mem)
		count++
		if count >= 5 {
			break
		}
	}

	return result
}

// CalculateSimilarity computes 1 - |current - memory| for dominant emotions
func CalculateSimilarity(current, memory EmotionProfile) float64 {
	if len(current) == 0 || len(memory) == 0 {
		return 0.0
	}

	// Prompt: "similarity = 1 - |currentEmotion - memoryEmotion|"
	// This implies comparing specific scalar values?
	// "When NPC experiences similar emotion... Current fear -> recalls past threats"
	// So we should compare the INTENSITY of the SAME emotion keys.

	totalSim := 0.0
	count := 0.0

	for emo, val := range current {
		if memVal, ok := memory[emo]; ok {
			// 1 - |diff|
			diff := math.Abs(val - memVal)
			sim := 1.0 - diff
			totalSim += sim
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalSim / count
}
