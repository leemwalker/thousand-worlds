package resources

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResourceNode_GetCategory(t *testing.T) {
	tests := []struct {
		name     string
		resType  ResourceType
		resName  string
		expected ResourceCategory
	}{
		// Minerals
		{"Iron Ore", ResourceMineral, "Iron Ore", CategoryOre},
		{"Coal", ResourceMineral, "Coal Deposit", CategoryCoal},
		{"Ruby", ResourceMineral, "Ruby Vein", CategoryGem},
		{"Stone", ResourceMineral, "Limestone", CategoryStone},

		// Vegetation
		{"Oak Wood", ResourceVegetation, "Oak Wood", CategoryWood},
		{"Wheat", ResourceVegetation, "Wild Wheat", CategoryGrain},
		{"Cotton", ResourceVegetation, "Cotton Plant", CategoryFiber},
		{"Basil", ResourceVegetation, "Basil Herb", CategoryHerb},
		{"Blueberry", ResourceVegetation, "Blueberry Bush", CategoryBerry},
		{"Mushroom", ResourceVegetation, "Red Mushroom", CategoryMushroom},
		{"Rose", ResourceVegetation, "Wild Rose Flower", CategoryFlower},
		{"Apple", ResourceVegetation, "Apple", CategoryFruit},
		{"Pine Resin", ResourceVegetation, "Pine Resin", CategoryResin},
		{"Unknown Veg", ResourceVegetation, "Strange Plant", CategoryVegetable},

		// Animal
		{"Beef", ResourceAnimal, "Raw Meat", CategoryMeat},
		{"Leather", ResourceAnimal, "Wolf Hide", CategoryHide},
		{"Bone", ResourceAnimal, "Bone Shard", CategoryBone},
		{"Feather", ResourceAnimal, "Eagle Feather", CategoryFeather},
		{"Unknown Animal", ResourceAnimal, "Strange Part", CategoryMeat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ResourceNode{
				Name:      tt.resName,
				Type:      tt.resType,
				CreatedAt: time.Now(),
			}
			assert.Equal(t, tt.expected, node.GetCategory())
		})
	}
}
