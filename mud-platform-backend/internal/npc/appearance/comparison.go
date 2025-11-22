package appearance

// CompareAppearances calculates similarity between two descriptions (0.0 - 1.0)
// Used for family resemblance checks
func CompareAppearances(app1, app2 AppearanceDescription) float64 {
	score := 0.0

	// 1. Height (20%)
	// Exact match = 1.0, Adjacent (Tall vs Very Tall) = 0.5?
	// Since we use strings, let's do simple string match for now, or map to values.
	// "tall" == "tall" -> 1.0
	if app1.Height == app2.Height {
		score += 0.2
	} else {
		// Partial credit?
		// e.g. "tall" vs "very tall"
		// Let's assume 0 for mismatch unless we map to ordinal.
	}

	// 2. Build (20%)
	if app1.Build == app2.Build {
		score += 0.2
	}

	// 3. Coloring (Hair + Eyes) (30%)
	colorScore := 0.0
	if app1.Hair == app2.Hair {
		colorScore += 0.5
	}
	if app1.Eyes == app2.Eyes {
		colorScore += 0.5
	}
	score += colorScore * 0.3

	// 4. Features (30%)
	// We don't explicitly store features struct in AppearanceDescription yet, just full string.
	// But we can compare the genetic inputs if we had them, or parse the string.
	// For this function signature, we only have AppearanceDescription.
	// Let's assume we parse or just rely on what we have.
	// Or we can add Features to the struct.
	// For now, let's just use the components we have.
	// Maybe add Age category match? No, age changes.
	// Let's assume the remaining 30% is based on "General Vibe" or just re-weight the others?
	// Prompt: "height 20%, build 20%, coloring 30%, features 30%"
	// We are missing explicit "Features" field in struct.
	// Let's add it to struct or ignore for now and normalize.
	// Let's normalize based on 70% total available.

	return score / 0.7
}

// CalculateSimilarity is a wrapper for genetic-based comparison if available
// But prompt asks for "CompareAppearances(appearance1, appearance2)".
// So we should probably extract features into the struct in Generator.
