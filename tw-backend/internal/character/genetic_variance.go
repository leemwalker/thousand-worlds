package character

import (
	"math/rand"
	"time"
)

// VarianceData holds the random adjustments applied to attributes
type VarianceData struct {
	// Physical
	Might     int `json:"might"`
	Agility   int `json:"agility"`
	Endurance int `json:"endurance"`
	Reflexes  int `json:"reflexes"`
	Vitality  int `json:"vitality"`

	// Mental
	Intellect int `json:"intellect"`
	Cunning   int `json:"cunning"`
	Willpower int `json:"willpower"`
	Presence  int `json:"presence"`
	Intuition int `json:"intuition"`

	// Sensory
	Sight   int `json:"sight"`
	Hearing int `json:"hearing"`
	Smell   int `json:"smell"`
	Taste   int `json:"taste"`
	Touch   int `json:"touch"`
}

// ApplyVariance applies a random variance (+/- 5) to each attribute
// Returns the new attributes and the variance data used
func ApplyVariance(base Attributes, seed int64) (Attributes, VarianceData) {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))

	variance := VarianceData{
		Might:     rng.Intn(11) - 5, // -5 to +5
		Agility:   rng.Intn(11) - 5,
		Endurance: rng.Intn(11) - 5,
		Reflexes:  rng.Intn(11) - 5,
		Vitality:  rng.Intn(11) - 5,
		Intellect: rng.Intn(11) - 5,
		Cunning:   rng.Intn(11) - 5,
		Willpower: rng.Intn(11) - 5,
		Presence:  rng.Intn(11) - 5,
		Intuition: rng.Intn(11) - 5,
		Sight:     rng.Intn(11) - 5,
		Hearing:   rng.Intn(11) - 5,
		Smell:     rng.Intn(11) - 5,
		Taste:     rng.Intn(11) - 5,
		Touch:     rng.Intn(11) - 5,
	}

	newAttrs := Attributes{
		Might:     base.Might + variance.Might,
		Agility:   base.Agility + variance.Agility,
		Endurance: base.Endurance + variance.Endurance,
		Reflexes:  base.Reflexes + variance.Reflexes,
		Vitality:  base.Vitality + variance.Vitality,
		Intellect: base.Intellect + variance.Intellect,
		Cunning:   base.Cunning + variance.Cunning,
		Willpower: base.Willpower + variance.Willpower,
		Presence:  base.Presence + variance.Presence,
		Intuition: base.Intuition + variance.Intuition,
		Sight:     base.Sight + variance.Sight,
		Hearing:   base.Hearing + variance.Hearing,
		Smell:     base.Smell + variance.Smell,
		Taste:     base.Taste + variance.Taste,
		Touch:     base.Touch + variance.Touch,
	}

	return newAttrs, variance
}
