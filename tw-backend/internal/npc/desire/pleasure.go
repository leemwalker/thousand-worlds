package desire

// PleasureContext holds data for pleasure needs
type PleasureContext struct {
	NearUnexplored bool
	BoredHours     float64
	IsCreating     bool
}

// UpdatePleasureNeeds updates curiosity, hedonism, and creativity
func UpdatePleasureNeeds(profile *DesireProfile, timeDeltaHours float64, traits PersonalityTraits, ctx PleasureContext) {
	// Curiosity
	// High Openness: base value 40-60?
	// Increases near unexplored areas
	if ctx.NearUnexplored {
		rate := 5.0 * traits.Openness
		profile.Needs[NeedCuriosity].Value += rate * timeDeltaHours
	}
	// Decay if satisfied? Or just constant drive?
	// Let's assume it decays slowly if not stimulated?
	// Or increases slowly if bored?
	// Prompt: "Increases near unexplored areas..."
	clamp(profile.Needs[NeedCuriosity])

	// Hedonism
	// High Extraversion + Low Conscientiousness = High Hedonism
	// Increases when bored (no stimulation for 2+ hours)
	if ctx.BoredHours >= 2.0 {
		// Rate based on personality
		// High E (1.0) + Low C (0.0) -> Max rate
		hedonismFactor := traits.Extraversion + (1.0 - traits.Conscientiousness)
		rate := 2.0 * hedonismFactor
		profile.Needs[NeedHedonism].Value += rate * timeDeltaHours
	}
	clamp(profile.Needs[NeedHedonism])

	// Creativity
	// High in NPCs with artist/craftsman occupations (Context?)
	// Let's assume Openness drives it too.
	if ctx.IsCreating {
		profile.Needs[NeedCreativity].Value -= 10.0 * timeDeltaHours
	} else {
		rate := 0.5 * traits.Openness
		profile.Needs[NeedCreativity].Value += rate * timeDeltaHours
	}
	clamp(profile.Needs[NeedCreativity])
}
