package npc

import (
	"tw-backend/internal/economy/resources"
)

// Occupation defines an NPC's economic role and priorities
// Redefined here to match the updated types
type Occupation struct {
	Name               string
	PrimaryResources   []resources.ResourceCategory
	SecondaryResources []resources.ResourceCategory
	PreferredSkills    []string
	GatheringRadius    float64 // How far they travel to gather
}

// Standard occupations
var (
	OccupationFarmer = Occupation{
		Name: "farmer",
		PrimaryResources: []resources.ResourceCategory{
			resources.CategoryGrain,
			resources.CategoryVegetable,
			resources.CategoryFruit,
		},
		SecondaryResources: []resources.ResourceCategory{
			resources.CategoryFiber,
			resources.CategoryHerb,
		},
		PreferredSkills: []string{"farming", "herbalism"},
		GatheringRadius: 200.0,
	}

	OccupationMiner = Occupation{
		Name: "miner",
		PrimaryResources: []resources.ResourceCategory{
			resources.CategoryOre,
			resources.CategoryCoal,
			resources.CategoryGem,
		},
		SecondaryResources: []resources.ResourceCategory{
			resources.CategoryStone,
		},
		PreferredSkills: []string{"mining"},
		GatheringRadius: 1000.0,
	}

	OccupationWoodcutter = Occupation{
		Name: "woodcutter",
		PrimaryResources: []resources.ResourceCategory{
			resources.CategoryWood,
		},
		SecondaryResources: []resources.ResourceCategory{
			resources.CategoryResin,
			resources.CategoryHerb,
		},
		PreferredSkills: []string{"logging"},
		GatheringRadius: 500.0,
	}

	OccupationHunter = Occupation{
		Name: "hunter",
		PrimaryResources: []resources.ResourceCategory{
			resources.CategoryMeat,
			resources.CategoryHide,
			resources.CategoryBone,
		},
		SecondaryResources: []resources.ResourceCategory{
			resources.CategoryFeather,
		},
		PreferredSkills: []string{"hunting", "tracking"},
		GatheringRadius: 2000.0,
	}

	OccupationHerbalist = Occupation{
		Name: "herbalist",
		PrimaryResources: []resources.ResourceCategory{
			resources.CategoryHerb,
			resources.CategoryFlower,
		},
		SecondaryResources: []resources.ResourceCategory{
			resources.CategoryMushroom,
			resources.CategoryBerry,
		},
		PreferredSkills: []string{"herbalism", "alchemy"},
		GatheringRadius: 800.0,
	}
)

// GetOccupation returns the occupation definition by name
func GetOccupation(name string) (Occupation, bool) {
	switch name {
	case "farmer":
		return OccupationFarmer, true
	case "miner":
		return OccupationMiner, true
	case "woodcutter":
		return OccupationWoodcutter, true
	case "hunter":
		return OccupationHunter, true
	case "herbalist":
		return OccupationHerbalist, true
	default:
		return Occupation{}, false
	}
}
