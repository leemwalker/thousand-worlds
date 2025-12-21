# Minerals Package

Geologically-realistic mineral deposit generation.

## Architecture

```
minerals/
├── formation.go      # Deposit generation by geological process
├── clusters.go       # Mineral clustering algorithms
├── concentration.go  # Ore grade calculation
├── depletion.go      # Resource extraction and depletion
├── discovery.go      # Deposit discovery mechanics
├── veins.go          # Vein structure generation
├── repository.go     # Persistence layer
└── types.go          # Core types (MineralDeposit, TectonicContext)
```

---

## Formation Types

| Process | Minerals | Requirements |
|---------|----------|--------------|
| Hydrothermal | Gold, Copper, Sulfides | Ocean ridges |
| Placer | Gold, Gems | River erosion |
| BIF | Iron | Oxygen spike (GOE) |
| Kimberlite | Diamonds | Ancient cratons >2.5B years |
| Evaporite | Salt, Gypsum | Arid climate |
| Coal | Coal (peat→anthracite) | Organic burial + time |

---

## Key Functions

| Function | Description |
|----------|-------------|
| `GenerateMineralVein()` | Creates deposit by tectonic context |
| `GenerateBIFDeposits()` | Banded Iron from oxygen events |
| `GeneratePlacerDeposits()` | Alluvial deposits at river bends |
| `GenerateKimberlitePipe()` | Diamond pipes in ancient cratons |
| `GenerateCoalDeposits()` | Coal seams by burial depth/age |
| `ExtractResource()` | Mining with quantity reduction |
| `SampleConcentration()` | Ore grade by distance from center |

---

## Usage

```go
// Generate vein deposit
deposit := minerals.GenerateMineralVein(tectonicCtx, minerals.MineralGold, epicenter)

// Extract resources
extracted := minerals.ExtractResource(deposit, 100)

// Sample concentration
grade := minerals.SampleConcentration(deposit, x, y)
```

---

## Testing

```bash
go test -v ./internal/worldgen/minerals/...
go test -cover ./internal/worldgen/minerals/...
```
