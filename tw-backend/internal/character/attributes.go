package character

// CalculateSecondaryAttributes derives secondary stats from primary attributes
func CalculateSecondaryAttributes(attrs Attributes) SecondaryAttributes {
	return SecondaryAttributes{
		MaxHP:      attrs.Vitality * 10,
		MaxStamina: (attrs.Endurance * 7) + (attrs.Might * 3),
		MaxFocus:   (attrs.Intellect * 6) + (attrs.Willpower * 4),
		MaxMana:    (attrs.Intuition * 6) + (attrs.Willpower * 4),
		MaxNerve:   (attrs.Willpower * 5) + (attrs.Presence * 3) + (attrs.Reflexes * 2),
	}
}
