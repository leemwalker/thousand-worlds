package emotion

// Primary Emotions (Ekman model)
const (
	Joy      = "joy"
	Anger    = "anger"
	Fear     = "fear"
	Sadness  = "sadness"
	Surprise = "surprise"
	Disgust  = "disgust"
)

// Complex Emotions
const (
	Anticipation = "anticipation"
	Contempt     = "contempt"
	Anxiety      = "anxiety"
)

// EmotionProfile maps emotion names to intensity (0.0-1.0)
type EmotionProfile map[string]float64

// EventContext provides details for analysis
type EventContext struct {
	EventType      string
	DamageTaken    float64
	MaxHP          float64
	GiftValue      float64
	WealthLevel    float64
	IsFirstMeeting bool
	IsBetrayal     bool
	IsDeath        bool
	IsThreat       bool
}

// PersonalityTraits modifies emotional response
type PersonalityTraits struct {
	Neuroticism float64 // Increases Fear/Sadness
	Aggression  float64 // Increases Anger
	Optimism    float64 // Increases Joy, Decreases Sadness
}
