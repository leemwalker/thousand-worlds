package underground

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStandardTools(t *testing.T) {
	assert.Equal(t, 1.0, StandardTools["hands"].MaxHardness)
	assert.Equal(t, 4.0, StandardTools["stone_pick"].MaxHardness)
	assert.Equal(t, 6.0, StandardTools["iron_pick"].MaxHardness)
	assert.Equal(t, 10.0, StandardTools["diamond_pick"].MaxHardness)
}

func TestCanMine_Success(t *testing.T) {
	stratum := &StrataLayer{
		Material: "soil",
		Hardness: 2,
	}
	tool := StandardTools["stone_pick"]

	canMine, reason := CanMine(tool, stratum, 50)

	assert.True(t, canMine)
	assert.Empty(t, reason)
}

func TestCanMine_TooHard(t *testing.T) {
	stratum := &StrataLayer{
		Material: "granite",
		Hardness: 8,
	}
	tool := StandardTools["stone_pick"] // Max hardness 4

	canMine, reason := CanMine(tool, stratum, 50)

	assert.False(t, canMine)
	assert.Contains(t, reason, "tool too weak")
}

func TestCanMine_TooDeep(t *testing.T) {
	stratum := &StrataLayer{
		Material: "soil",
		Hardness: 2,
	}
	tool := StandardTools["wooden_pick"] // Depth limit 50m

	canMine, reason := CanMine(tool, stratum, 100)

	assert.False(t, canMine)
	assert.Contains(t, reason, "cannot reach this depth")
}

func TestMine_Success(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}
	col.AddStratum("soil", 100, 0, 2, 100, 0.4)

	tool := StandardTools["stone_pick"]
	result := Mine(col, 90, tool, false)

	assert.True(t, result.Success)
	assert.Greater(t, result.TimeRequired, 0.0)
	assert.Nil(t, result.VoidCreated)
}

func TestMine_WithTunnel(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}
	col.AddStratum("soil", 100, 0, 2, 100, 0.4)

	tool := StandardTools["iron_pick"]
	result := Mine(col, 90, tool, true)

	assert.True(t, result.Success)
	assert.NotNil(t, result.VoidCreated)
	assert.Equal(t, "mine", result.VoidCreated.VoidType)
	assert.Equal(t, 1, len(col.Voids))
}

func TestMine_FindsResource(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}
	col.AddStratum("rock", 100, 0, 4, 100, 0.1)
	col.Resources = []Deposit{
		{Type: "iron", DepthZ: 90, Quantity: 10, Discovered: false},
	}

	tool := StandardTools["iron_pick"]
	result := Mine(col, 90, tool, false)

	assert.True(t, result.Success)
	assert.NotNil(t, result.ResourceFound)
	assert.True(t, col.Resources[0].Discovered)
}

func TestMine_VoidAlreadyExists(t *testing.T) {
	col := &WorldColumn{X: 0, Y: 0, Surface: 100}
	col.AddStratum("soil", 100, 0, 2, 100, 0.4)
	col.Voids = []VoidSpace{
		{MinZ: 85, MaxZ: 95, VoidType: "cave"},
	}

	tool := StandardTools["iron_pick"]
	result := Mine(col, 90, tool, false)

	assert.False(t, result.Success)
	assert.Contains(t, result.Reason, "already a void")
}

func TestExtractResource(t *testing.T) {
	deposit := &Deposit{Type: "iron", Quantity: 100}

	extracted, ok := ExtractResource(deposit, 30)

	assert.True(t, ok)
	assert.Equal(t, 30.0, extracted)
	assert.Equal(t, 70.0, deposit.Quantity)
}

func TestExtractResource_ExceedsQuantity(t *testing.T) {
	deposit := &Deposit{Type: "iron", Quantity: 20}

	extracted, ok := ExtractResource(deposit, 50)

	assert.True(t, ok)
	assert.Equal(t, 20.0, extracted)
	assert.Equal(t, 0.0, deposit.Quantity)
}

func TestCreateBurrow_Success(t *testing.T) {
	col := &WorldColumn{X: 10, Y: 20, Surface: 100}
	col.AddStratum("soil", 100, 0, 2, 100, 0.5) // Soft enough for burrowing

	ownerID := uuid.New()
	burrow, err := CreateBurrow(col, ownerID, 100, 10, 3)

	assert.NoError(t, err)
	assert.NotNil(t, burrow)
	assert.Equal(t, ownerID, burrow.OwnerID)
	assert.Equal(t, 3, len(burrow.Chambers))
	assert.Equal(t, 3, len(burrow.Tunnels))
	assert.Equal(t, "nest", burrow.Chambers[2].Purpose)
	assert.GreaterOrEqual(t, len(col.Voids), 3) // At least 3 chamber voids
}

func TestCreateBurrow_GroundTooHard(t *testing.T) {
	col := &WorldColumn{X: 10, Y: 20, Surface: 100}
	col.AddStratum("granite", 100, 0, 8, 100, 0.05) // Too hard

	ownerID := uuid.New()
	burrow, err := CreateBurrow(col, ownerID, 100, 10, 3)

	assert.Error(t, err)
	assert.Nil(t, burrow)
	assert.Contains(t, err.Error(), "too hard")
}

func TestDigTunnel(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	// Set up soft soil in all columns
	for _, col := range grid.AllColumns() {
		col.Surface = 100
		col.AddStratum("soil", 100, 0, 2, 100, 0.4)
	}

	tool := StandardTools["iron_pick"]
	results, err := DigTunnel(grid, 2, 2, 90, 5, 2, 90, tool)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(results)) // 4 steps from (2,2) to (5,2)

	// Check that voids were created along the path
	for x := 2; x <= 5; x++ {
		col := grid.Get(x, 2)
		assert.NotEmpty(t, col.Voids, "Column at x=%d should have void", x)
	}
}

func TestDigTunnel_FailsOnHardRock(t *testing.T) {
	grid := NewColumnGrid(10, 10)

	// Mix of soft and hard
	for _, col := range grid.AllColumns() {
		col.Surface = 100
		if col.X == 4 {
			col.AddStratum("granite", 100, 0, 8, 100, 0.05) // Hard rock
		} else {
			col.AddStratum("soil", 100, 0, 2, 100, 0.4)
		}
	}

	tool := StandardTools["stone_pick"] // Can't mine hardness 8
	results, err := DigTunnel(grid, 2, 2, 90, 5, 2, 90, tool)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mining failed")
	// Some results before failure
	assert.GreaterOrEqual(t, len(results), 2)
}
