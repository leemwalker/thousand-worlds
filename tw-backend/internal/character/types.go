package character

import (
	"time"

	"github.com/google/uuid"
)

// Attribute Categories
const (
	CategoryPhysical = "Physical"
	CategoryMental   = "Mental"
	CategorySensory  = "Sensory"
)

// Physical Attributes
const (
	AttrMight     = "Might"
	AttrAgility   = "Agility"
	AttrEndurance = "Endurance"
	AttrReflexes  = "Reflexes"
	AttrVitality  = "Vitality"
)

// Mental Attributes
const (
	AttrIntellect = "Intellect"
	AttrCunning   = "Cunning"
	AttrWillpower = "Willpower"
	AttrPresence  = "Presence"
	AttrIntuition = "Intuition"
)

// Sensory Attributes
const (
	AttrSight   = "Sight"
	AttrHearing = "Hearing"
	AttrSmell   = "Smell"
	AttrTaste   = "Taste"
	AttrTouch   = "Touch"
)

// Secondary Attributes
const (
	AttrHP      = "HP"
	AttrStamina = "Stamina"
	AttrFocus   = "Focus"
	AttrMana    = "Mana"
	AttrNerve   = "Nerve"
)

// Species Constants
const (
	SpeciesHuman = "Human"
	SpeciesDwarf = "Dwarf"
	SpeciesElf   = "Elf"
)

// Attributes holds all primary and sensory attributes
type Attributes struct {
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

// SecondaryAttributes holds calculated derived stats
type SecondaryAttributes struct {
	MaxHP      int `json:"max_hp"`
	MaxStamina int `json:"max_stamina"`
	MaxFocus   int `json:"max_focus"`
	MaxMana    int `json:"max_mana"`
	MaxNerve   int `json:"max_nerve"`
}

// Character represents a player character or NPC
type Character struct {
	ID        uuid.UUID           `json:"id"`
	PlayerID  uuid.UUID           `json:"player_id"` // Null/Zero for NPCs
	Name      string              `json:"name"`
	Species   string              `json:"species"`
	BaseAttrs Attributes          `json:"attributes"`
	SecAttrs  SecondaryAttributes `json:"secondary_attributes"`
	PositionX float64             `json:"position_x"`
	PositionY float64             `json:"position_y"`
	PositionZ float64             `json:"position_z"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// SpeciesTemplate defines the baseline attributes for a species
type SpeciesTemplate struct {
	Name      string
	BaseAttrs Attributes
}

// GetSpeciesTemplate returns the template for a given species name
func GetSpeciesTemplate(species string) SpeciesTemplate {
	switch species {
	case SpeciesDwarf:
		return SpeciesTemplate{
			Name: SpeciesDwarf,
			BaseAttrs: Attributes{
				// Physical: Strong and tough, but slower
				Might:     60,
				Agility:   40,
				Endurance: 65,
				Reflexes:  45,
				Vitality:  60,
				// Mental: Stubborn and practical
				Intellect: 50,
				Cunning:   45,
				Willpower: 60,
				Presence:  55,
				Intuition: 50,
				// Sensory: Good underground (hearing/touch), poor sight range
				Sight:   45,
				Hearing: 65,
				Smell:   50,
				Taste:   55,
				Touch:   60,
			},
		}
	case SpeciesElf:
		return SpeciesTemplate{
			Name: SpeciesElf,
			BaseAttrs: Attributes{
				// Physical: Agile and quick, but frail
				Might:     40,
				Agility:   65,
				Endurance: 45,
				Reflexes:  60,
				Vitality:  45,
				// Mental: Wise and perceptive
				Intellect: 60,
				Cunning:   55,
				Willpower: 50,
				Presence:  60,
				Intuition: 65,
				// Sensory: Excellent sight and hearing
				Sight:   70,
				Hearing: 65,
				Smell:   55,
				Taste:   50,
				Touch:   50,
			},
		}
	case SpeciesHuman:
		return SpeciesTemplate{
			Name: SpeciesHuman,
			BaseAttrs: Attributes{
				// Balanced across the board
				Might:     50,
				Agility:   50,
				Endurance: 50,
				Reflexes:  50,
				Vitality:  50,
				Intellect: 50,
				Cunning:   50,
				Willpower: 50,
				Presence:  50,
				Intuition: 50,
				Sight:     50,
				Hearing:   50,
				Smell:     50,
				Taste:     50,
				Touch:     50,
			},
		}
	default:
		return SpeciesTemplate{}
	}
}
