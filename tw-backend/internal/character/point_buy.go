package character

import (
	"fmt"
)

const (
	MaxPoints      = 100
	MaxIncrease    = 30
	Tier1Threshold = 10
	Tier2Threshold = 20
)

// ValidatePointBuy checks if the proposed attribute increases are valid
// base: The species baseline attributes
// variance: The random variance applied
// increases: Map of attribute name to points increased (raw value increase, not cost)
func ValidatePointBuy(base Attributes, variance Attributes, increases map[string]int) error {
	totalCost := 0

	for attr, increase := range increases {
		if increase < 0 {
			return fmt.Errorf("cannot decrease attribute %s below baseline", attr)
		}
		if increase == 0 {
			continue
		}
		if increase > MaxIncrease {
			return fmt.Errorf("attribute %s increase +%d exceeds max of +%d", attr, increase, MaxIncrease)
		}

		// Calculate cost
		cost := 0
		for i := 1; i <= increase; i++ {
			if i <= Tier1Threshold {
				cost += 1
			} else if i <= Tier2Threshold {
				cost += 2
			} else {
				cost += 3
			}
		}
		totalCost += cost
	}

	if totalCost > MaxPoints {
		return fmt.Errorf("total point cost %d exceeds budget of %d", totalCost, MaxPoints)
	}

	return nil
}

// ApplyPointBuy applies the valid increases to the base+variance attributes
func ApplyPointBuy(base Attributes, variance Attributes, increases map[string]int) Attributes {
	final := Attributes{
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

	for attr, increase := range increases {
		switch attr {
		case AttrMight:
			final.Might += increase
		case AttrAgility:
			final.Agility += increase
		case AttrEndurance:
			final.Endurance += increase
		case AttrReflexes:
			final.Reflexes += increase
		case AttrVitality:
			final.Vitality += increase
		case AttrIntellect:
			final.Intellect += increase
		case AttrCunning:
			final.Cunning += increase
		case AttrWillpower:
			final.Willpower += increase
		case AttrPresence:
			final.Presence += increase
		case AttrIntuition:
			final.Intuition += increase
		case AttrSight:
			final.Sight += increase
		case AttrHearing:
			final.Hearing += increase
		case AttrSmell:
			final.Smell += increase
		case AttrTaste:
			final.Taste += increase
		case AttrTouch:
			final.Touch += increase
		}
	}

	return final
}
