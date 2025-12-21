# Population Package

Macro-level population dynamics and evolution simulation.

## Architecture

```
population/
├── dynamics.go       # PopulationSimulator - main engine
├── organism.go       # Organism lifecycle management
├── epochs.go         # Geological epoch effects
├── speciation.go     # Species divergence and creation
├── migration.go      # Population movement between biomes
├── genetics.go       # Genetic drift and bottlenecks
├── phylogeny.go      # Evolutionary trees
├── grid_integration.go # Geography integration
├── naming.go         # Species name generation
└── types.go          # Core types (SpeciesPopulation, Biome)
```

---

## Key Functions

| Function | Description |
|----------|-------------|
| `NewPopulationSimulator()` | Creates simulator |
| `SimulateStep()` | Advances one time step |
| `ApplyDisease()` | Density-dependent outbreaks |
| `ApplyNichePartitioning()` | Character displacement |
| `ApplySymbiosis()` | Mutualistic relationships |
| `CheckSpeciation()` | Trait divergence → new species |
| `UpdateOxygenLevel()` | Atmospheric changes |

---

## Biological Rules

| Rule | Implementation |
|------|----------------|
| Kleiber's Law | `CalculateMetabolicRate()` - mass^0.75 |
| r/K Selection | `CalculateReproductionModifier()` - size vs. reproduction |
| Lilliput Effect | Post-extinction small species advantage |

---

## Testing

```bash
go test -v ./internal/ecosystem/population/...
```
