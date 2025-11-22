package appearance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareAppearances(t *testing.T) {
	app1 := AppearanceDescription{
		Height: "tall",
		Build:  "muscular",
		Hair:   "black",
		Eyes:   "brown",
	}

	app2 := AppearanceDescription{
		Height: "tall",     // Match
		Build:  "muscular", // Match
		Hair:   "black",    // Match
		Eyes:   "blue",     // Mismatch
	}

	// Score:
	// Height: 0.2
	// Build: 0.2
	// Color: (0.5 + 0.0) * 0.3 = 0.15
	// Total: 0.55 / 0.7 = 0.78

	score := CompareAppearances(app1, app2)
	assert.InDelta(t, 0.785, score, 0.01)

	// Identical
	score = CompareAppearances(app1, app1)
	assert.InDelta(t, 1.0, score, 0.01)
}
