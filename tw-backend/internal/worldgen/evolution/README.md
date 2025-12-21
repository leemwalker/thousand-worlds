# Evolution Package

Time-accelerated species evolution during world creation.

## Architecture

```
evolution/
├── fitness.go         # Fitness calculation (climate, food, predation)
├── mutation.go        # Genetic mutation and speciation
├── food_chain.go      # Trophic levels and energy transfer
├── predation.go       # Predator-prey dynamics
├── competition.go     # Interspecific competition
├── diversity.go       # Biodiversity metrics
├── species.go         # Species type definition
├── ecosystem.go       # Ecosystem balance
└── types.go           # Core types (Species, Environment)
```

---

## Key Concepts

### Fitness Calculation

Species fitness is calculated as a product of multiple factors:

| Factor | Function |
|--------|----------|
| Climate | `CalculateClimateFitness()` - temperature tolerance |
| Food | `CalculateFoodFitness()` - food availability |
| Predation | `CalculatePredationFitness()` - defensive traits |
| Competition | `CalculateCompetitionFitness()` - niche overlap |

```go
totalFitness := CalculateTotalFitness(species, env, food, pred, competitors)
```

### Mutation

| Constant | Value | Description |
|----------|-------|-------------|
| Base Rate | 1-5% | Normal mutation rate |
| Bottleneck Rate | 10-15% | Elevated during genetic bottleneck |
| Speciation Threshold | 1.5x | 50% change triggers new species |

### Food Chain

| Trophic Level | Energy Transfer |
|---------------|-----------------|
| Producer → Herbivore | 10% |
| Herbivore → Carnivore | 10% |

---

## Mass Extinction Events

| Event | Severity | Effect |
|-------|----------|--------|
| Asteroid | 0.9 | ~75% species loss, large animals hit hardest |
| Volcanic | 0.6 | ~40% species loss |
| Ice Age | 0.4 | ~25% species loss |

---

## Usage

```go
// Calculate fitness
fitness := evolution.CalculateTotalFitness(species, env, food, pred, nil)

// Apply mutations
newSpecies := evolution.ApplyMutationsToPopulation(species)

// Calculate biomass
biomass := evolution.CalculateBiomassProduction(0.8, 1500, 20)
```

---

## Testing

```bash
go test -v ./internal/worldgen/evolution/...
go test -cover ./internal/worldgen/evolution/...
```
