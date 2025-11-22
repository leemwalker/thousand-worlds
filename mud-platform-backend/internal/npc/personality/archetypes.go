package personality

// Standard Archetypes for Testing
var (
	ArchetypeAdventurer = &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: 90},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: 40},
		Extraversion:      Trait{Name: TraitExtraversion, Value: 75},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: 60},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: 30},
	}

	ArchetypeScholar = &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: 85},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: 90},
		Extraversion:      Trait{Name: TraitExtraversion, Value: 30},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: 50},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: 40},
	}

	ArchetypeLeader = &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: 60},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: 80},
		Extraversion:      Trait{Name: TraitExtraversion, Value: 90},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: 70},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: 25},
	}

	ArchetypeHermit = &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: 50},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: 60},
		Extraversion:      Trait{Name: TraitExtraversion, Value: 10},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: 40},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: 70},
	}

	ArchetypeMerchant = &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: 50},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: 85},
		Extraversion:      Trait{Name: TraitExtraversion, Value: 65},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: 45},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: 35},
	}
)

// GetArchetype returns a copy of the requested archetype
func GetArchetype(name string) *Personality {
	var source *Personality
	switch name {
	case "Adventurer":
		source = ArchetypeAdventurer
	case "Scholar":
		source = ArchetypeScholar
	case "Leader":
		source = ArchetypeLeader
	case "Hermit":
		source = ArchetypeHermit
	case "Merchant":
		source = ArchetypeMerchant
	default:
		return NewPersonality()
	}

	// Return copy
	return &Personality{
		Openness:          Trait{Name: TraitOpenness, Value: source.Openness.Value},
		Conscientiousness: Trait{Name: TraitConscientiousness, Value: source.Conscientiousness.Value},
		Extraversion:      Trait{Name: TraitExtraversion, Value: source.Extraversion.Value},
		Agreeableness:     Trait{Name: TraitAgreeableness, Value: source.Agreeableness.Value},
		Neuroticism:       Trait{Name: TraitNeuroticism, Value: source.Neuroticism.Value},
	}
}
