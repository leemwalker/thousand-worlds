package underground

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrganicDeposit(t *testing.T) {
	entityID := uuid.New()
	deposit := CreateOrganicDeposit(entityID, "fish", 10, 20, 100, 1000, false)

	assert.Equal(t, string(DepositRemains), deposit.Type)
	assert.Equal(t, 100.0, deposit.DepthZ)
	assert.Equal(t, 1.0, deposit.Quantity)
	assert.NotNil(t, deposit.Source)
	assert.Equal(t, entityID, deposit.Source.OriginalEntityID)
	assert.Equal(t, "fish", deposit.Source.Species)
	assert.Equal(t, int64(1000), deposit.Source.DeathYear)
}

func TestCreateOrganicDeposit_Plant(t *testing.T) {
	entityID := uuid.New()
	deposit := CreateOrganicDeposit(entityID, "tree", 10, 20, 100, 1000, true)

	assert.Equal(t, string(DepositRemains)+"_plant", deposit.Type)
}

func TestDefaultDepositConfig(t *testing.T) {
	config := DefaultDepositConfig()

	assert.Equal(t, 10.0, config.BurialDepthForMineralization)
	assert.Equal(t, int64(1_000), config.MineralizationAge)
	assert.Equal(t, int64(100_000), config.FossilizationAge)
	assert.Equal(t, int64(5_000_000), config.OilFormationAge)
	assert.Equal(t, 3000.0, config.OilFormationDepth)
}

func TestTransformDeposit_ToMineralizing(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}

	deposit := Deposit{
		Type:   string(DepositRemains),
		DepthZ: 80, // 20m below surface
		Source: &OrganicSource{
			DeathYear:  0,
			BurialYear: 100, // Buried at year 100
		},
	}

	config := DefaultDepositConfig()
	currentYear := int64(2000) // 2000 years old, 1900 years buried

	transformDeposit(&deposit, col, currentYear, config)

	assert.Equal(t, string(DepositMineralizing), deposit.Type)
}

func TestTransformDeposit_ToFossil(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}

	deposit := Deposit{
		Type:   string(DepositMineralizing),
		DepthZ: 50, // 50m below surface
		Source: &OrganicSource{
			DeathYear:  0,
			BurialYear: 100,
		},
	}

	config := DefaultDepositConfig()
	currentYear := int64(200_000) // 200K years old

	transformDeposit(&deposit, col, currentYear, config)

	assert.Equal(t, string(DepositFossil), deposit.Type)
}

func TestTransformDeposit_PlantToCoal(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}

	deposit := Deposit{
		Type:   string(DepositMineralizing) + "_plant",
		DepthZ: 50,
		Source: &OrganicSource{
			DeathYear:  0,
			BurialYear: 100,
		},
	}

	config := DefaultDepositConfig()
	currentYear := int64(200_000)

	transformDeposit(&deposit, col, currentYear, config)

	assert.Equal(t, string(DepositCoal), deposit.Type)
}

func TestTransformDeposit_FossilToOil(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}

	deposit := Deposit{
		Type:     string(DepositFossil),
		DepthZ:   -3500, // 3.5km below surface (at 100m, so depth = 3600m)
		Quantity: 1.0,
		Source: &OrganicSource{
			Species:    "fish", // Oil producer
			DeathYear:  0,
			BurialYear: 100,
		},
	}

	config := DefaultDepositConfig()
	currentYear := int64(10_000_000) // 10M years old

	transformDeposit(&deposit, col, currentYear, config)

	assert.Equal(t, string(DepositOil), deposit.Type)
	assert.Greater(t, deposit.Quantity, 1.0, "Oil quantity should be multiplied")
}

func TestTransformDeposit_NonOilSpecies(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}

	deposit := Deposit{
		Type:   string(DepositFossil),
		DepthZ: -3500,
		Source: &OrganicSource{
			Species:    "rabbit", // NOT an oil producer
			DeathYear:  0,
			BurialYear: 100,
		},
	}

	config := DefaultDepositConfig()
	currentYear := int64(10_000_000)

	transformDeposit(&deposit, col, currentYear, config)

	// Should remain fossil, not become oil
	assert.Equal(t, string(DepositFossil), deposit.Type)
}

func TestIsOrganicRich(t *testing.T) {
	tests := []struct {
		species  string
		expected bool
	}{
		{"fish", true},
		{"whale", true},
		{"plankton", true},
		{"algae", true},
		{"dinosaur", true},
		{"rabbit", false},
		{"human", false},
		{"wolf", false},
	}

	for _, tt := range tests {
		result := isOrganicRich(tt.species)
		assert.Equal(t, tt.expected, result, "species: %s", tt.species)
	}
}

func TestWorldColumn_GetDepositByType(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0}

	col.Resources = []Deposit{
		{Type: string(DepositFossil)},
		{Type: string(DepositOil)},
		{Type: string(DepositFossil)},
		{Type: string(DepositCoal)},
	}

	fossils := col.GetDepositByType(DepositFossil)
	assert.Equal(t, 2, len(fossils))

	oil := col.GetDepositByType(DepositOil)
	assert.Equal(t, 1, len(oil))

	coal := col.GetDepositByType(DepositCoal)
	assert.Equal(t, 1, len(coal))
}

func TestSimulateDepositEvolution(t *testing.T) {
	grid := NewColumnGrid(5, 5)

	col := grid.Get(2, 2)
	col.Surface = 100

	// Add a fresh deposit
	deposit := CreateOrganicDeposit(uuid.New(), "fish", 2, 2, 100, 0, false)
	col.AddOrganicDeposit(deposit)

	config := DefaultDepositConfig()
	config.SedimentRatePerYear = 1.0 // Fast sedimentation for testing

	// Simulate many years
	SimulateDepositEvolution(grid, 10_000, config, nil, 42)

	// Check deposit was buried
	assert.Less(t, col.Resources[0].DepthZ, 100.0, "Deposit should be buried deeper")
}
