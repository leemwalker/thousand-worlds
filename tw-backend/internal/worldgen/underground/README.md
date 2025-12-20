# Underground System

**Package**: `internal/worldgen/underground/`  
**Coverage**: 85.8% (53 unit tests + 3 integration tests)

## Overview

Hybrid column-based underground system for mining, caves, fossils, and geological processes.

## Core Types

| File | Types | Purpose |
|------|-------|---------|
| `types.go` | WorldColumn, StrataLayer, VoidSpace, Deposit, MagmaInfo | Core data structures |
| `column_grid.go` | ColumnGrid | Thread-safe grid management |
| `caves.go` | Cave, CaveNode, CaveEdge | Cave network structures |

## Simulation Modules

| File | Functions | Description |
|------|-----------|-------------|
| `cave_formation.go` | SimulateCaveFormation, RegisterCaveInColumn | Limestone dissolution caves |
| `magma_simulation.go` | SimulateMagmaChambers, GetTectonicBoundaries | Magma and lava tubes |
| `deposits.go` | CreateOrganicDeposit, SimulateDepositEvolution | Fossil/oil formation |
| `mining.go` | Mine, CanMine, CreateBurrow, DigTunnel | Player mining operations |

## Usage

```go
import "tw-backend/internal/worldgen/underground"

// Create column grid
grid := underground.NewColumnGrid(width, height)

// Add strata to columns
col := grid.Get(x, y)
col.AddStratum("limestone", topZ, bottomZ, hardness, age, porosity)

// Simulate cave formation
caves := underground.SimulateCaveFormation(grid, rainfall, years, seed, config)

// Mining
tool := underground.StandardTools["iron_pick"]
result := underground.Mine(col, depth, tool, createTunnel)
```

## Integration with Geology

The underground system integrates with `WorldGeology` in `internal/ecosystem/geology.go`:

```go
// In SimulateGeology(), called automatically:
g.simulateCaveFormation(yearsElapsed)    // Every 100K+ years
g.simulateMagmaChambers(yearsElapsed)   // Every 10K+ years
g.simulateDepositEvolution(yearsElapsed) // Every 1K+ years
```

## World Composition

Underground strata vary by world type:

| Composition | Strata Sequence | Characteristics |
|-------------|-----------------|-----------------|
| `volcanic` | soil → basalt → gabbro → mantle | Low cave potential |
| `continental` | soil → sedimentary → limestone → granite | Balanced |
| `oceanic` | sediment → limestone → chalk → granite | High caves |
| `ancient` | soil → sandstone → schist → granite | Mineral rich |

## Tests

```bash
# Run all tests
go test ./internal/worldgen/underground/... -v

# Run with coverage
go test ./internal/worldgen/underground/... -cover

# Run integration tests only
go test ./internal/worldgen/underground/... -run Integration
```
