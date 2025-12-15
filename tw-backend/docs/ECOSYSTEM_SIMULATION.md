# Ecosystem Simulation Mechanics

The Thousand Worlds simulation engine drives the geological, ecological, and evolutionary history of generated worlds. It simulates millions of years of history to generate rich, believable backstories, fossil records, and species distributions.

## 1. Simulation Architecture

The simulation runs in time-steps of **1 year** (SimulateYear), processing sequentially:
1.  **Geological Events**: Tectonics, climate shifts, disasters.
2.  **Biome Updates**: Fragmentation, carrying capacity adjustments.
3.  **Population Dynamics**: Births, deaths, predation (Lotka-Volterra).
4.  **Ecological Interactions**: Symbiosis, competition, disease.
5.  **Evolution**: Mutation, selection, speciation, extinction.

The core logic resides in `internal/ecosystem/population/dynamics.go`.

---

## 2. Geological & Weather Simulation

The world is not static. Geological processes drive evolution by changing the environment.

### Tectonic Activity
-   **Continental Drift**: Plates move, causing fragmentation or supercontinent formation.
-   **Fragmentation**: Biomes can split (high fragmentation) or merge (low fragmentation).
    -   High fragmentation increases **genetic drift** and **speciation** but stresses large animals.
    -   Low fragmentation (supercontinents) encourages **competition** and **generalist** species.

### Geological Events (`events.go`)
Major events disrupt the status quo, causing extinctions and opening niches:
-   **Volcanic Winter**: Blocks sunlight, killing flora and starving herbivores.
-   **Ice Age**: Lowers global temperatures; favors cold-adapted, large (Bergmann's Rule) species.
-   **Asteroid Impact**: Catastrophic mass extinction (70-90% mortality).
-   **Ocean Anoxia**: Marine die-off due to low oxygen.
-   **Flood Basalt**: Long-term poisoning and warming.

Events have **Severity** (0.0-1.0) and **Duration**.

### Oxygen Cycle (`OxygenLevel`)
Atmospheric Oxygen (O2) is dynamic:
-   **Sources**: Produced by Flora (photosynthesis).
-   **Sinks**: Consumed by Fauna (respiration), volcanic release.
-   **Effects**:
    -   **High O2 (>25%)**: Allows giant sizes (e.g., carboniferous insects).
    -   **Low O2 (<15%)**: Stresses large organisms, limits max size.

---

## 3. Flora & Fauna Simulation

Life is modeled as **SpeciesPopulations** rather than individuals, tracking aggregate stats (Count, Genes).

### Traits (`EvolvableTraits`)
Species are defined by quantitative traits (0.0 - 1.0 or Scaled):
-   **Physical**: Size, Speed, Strength, Covering (Fur, Scales, Feathers).
-   **Survival**: Cold/Heat Resistance, Night Vision, Camouflage, Disease Resistance.
-   **Behavior**: Aggression, Social (Pack/Herd), Intelligence.
-   **Reproduction**: Fertility, Maturity Age, Litter Size, Lifespan.
-   **Diet**: Carnivore Tendency, Venom, Poison Resistance.

### Taxonomy & Naming
Species names are procedurally generated based on traits and lineage:
-   *Example*: "Woolly Swift Grazer", "Giant Armored Hunter".
-   Names reflect: **Size** (Small/Giant), **Covering** (Feathered/Scaled), **Diet** (Grazer/Hunter), **Biome** (Alpine/Desert).

### Needs-Based Engine
Populations grow/shrink based on:
-   **Food Availability**: Calculated per trophic level (Producer -> Primary -> Secondary).
-   **Biotic Potential**: Breeding rate adjusted by size (r/K selection).
-   **Environmental Resistance**: Predation, Disease, Climate mismatch.

---

## 4. Evolution & Biomes

Evolution is driven by **Mutation**, **Selection**, and **Drift**.

### Natural Selection (`applyBiomeSelection`)
Species must adapt to their biome:
-   **Tundra/Alpine**: Favors Cold Resistance, Large Size (heat retention), Fur/Fat.
-   **Desert**: Favors Heat Resistance, Small Size, Water Conservation traits.
-   **Rainforest**: Favors Camouflage, Climbing, Disease Resistance.
-   **Predation**: Prey evolve Speed/Camouflage; Predators evolve Speed/Strength/Intelligence.

### Speciation (`CheckSpeciation`)
New species branch off when:
-   Populations are isolated (Fragmentation).
-   New niches open (Adaptive Radiation after extinction).
-   Traits diverge significantly from ancestors.

### Mass Extinction & Recovery
-   **Recovery Phase**: A 20,000-year "healing" period after major events.
-   **Lilliput Effect**: Large species die off; survivors are small ("disaster taxa").
-   **Explosion**: Rapid speciation follows as life recolonizes empty biomes.

---

## 5. Ecological Mechanics

Advanced interactions add depth to the simulation.

### Symbiosis & Mutualism
-   **Mechanism**: Links Flora (producers) with small Fauna (pollinators/dispersers).
-   **Benefit**: Both partners gain population growth bonuses.
-   **Risk**: If one partner goes extinct, the other suffers ("co-extinction").

### Niche Partitioning
-   **Competition**: Species with similar traits (Diet, Size, Active Time) compete.
-   **Character Displacement**: Competition forces species to evolve apart (e.g., one becomes nocturnal, the other diurnal) to avoid conflict.

### Disease Dynamics
-   **Epidemics**: Density-dependent outbreaks occur in overcrowded populations.
-   **Resistance**: Survivors evolve higher `DiseaseResistance`.
-   **Regulation**: Prevents single species from permanently dominating a biome (negative feedback).

### Seasonal Cycles
-   **Seasons**: Spring (Breeding), Summer (Growth), Fall (Migration), Winter (Scarcity).
-   **Impact**: Food and survival rates vary by season and biome latitude.
