package personality

// Trait Names (Big Five OCEAN)
const (
	TraitOpenness          = "Openness"
	TraitConscientiousness = "Conscientiousness"
	TraitExtraversion      = "Extraversion"
	TraitAgreeableness     = "Agreeableness"
	TraitNeuroticism       = "Neuroticism"
)

// Mood Types
const (
	MoodCheerful   = "cheerful"
	MoodMelancholy = "melancholy"
	MoodAnxious    = "anxious"
	MoodAngry      = "angry"
	MoodCalm       = "calm"
	MoodExcited    = "excited"
)

// Trait represents a single personality dimension
type Trait struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"` // 0-100
}

// Personality represents the Big Five profile
type Personality struct {
	Openness          Trait `json:"openness"`
	Conscientiousness Trait `json:"conscientiousness"`
	Extraversion      Trait `json:"extraversion"`
	Agreeableness     Trait `json:"agreeableness"`
	Neuroticism       Trait `json:"neuroticism"`
}

// Mood represents a temporary emotional state
type Mood struct {
	Type      string             `json:"type"`
	Duration  float64            `json:"duration"` // Remaining duration in hours
	Modifiers map[string]float64 `json:"modifiers"`
}

// NewPersonality creates a default neutral personality
func NewPersonality() *Personality {
	return &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: 50},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: 50},
		Extraversion:      Trait{Name: TraitExtraversion, Value: 50},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: 50},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: 50},
	}
}

// NewMood creates a new mood instance
func NewMood(moodType string, duration float64) *Mood {
	m := &Mood{
		Type:      moodType,
		Duration:  duration,
		Modifiers: make(map[string]float64),
	}

	// Initialize default modifiers based on type
	switch moodType {
	case MoodCheerful:
		m.Modifiers[TraitExtraversion] = 5.0
	case MoodMelancholy:
		m.Modifiers[TraitExtraversion] = -5.0 // Indirectly handled via social penalty usually
	case MoodAnxious:
		m.Modifiers[TraitNeuroticism] = 15.0
	case MoodAngry:
		m.Modifiers[TraitAgreeableness] = -15.0
	case MoodExcited:
		m.Modifiers[TraitOpenness] = 10.0
	case MoodCalm:
		// No modifiers
	}

	return m
}
