# Ecosystem Simulation Mechanics - Complete Reference

**Version 2.0** | Last Updated: December 2025

The Thousand Worlds simulation engine drives the geological, ecological, and evolutionary history of generated worlds. It simulates millions to billions of years of history to generate rich, believable backstories, fossil records, and species distributions.

---

## Table of Contents
1. [Simulation Architecture](#1-simulation-architecture)
2. [Geological & Environmental Simulation](#2-geological--environmental-simulation)
3. [Species & Traits](#3-species--traits)
4. [Population Dynamics](#4-population-dynamics)
5. [Evolution Mechanisms](#5-evolution-mechanisms)
6. [Ecological Interactions](#6-ecological-interactions)
7. [Migration & Biome Transitions](#7-migration--biome-transitions)
8. [Performance & Optimization](#8-performance--optimization)
9. [Testing & Validation](#9-testing--validation)

---

## 1. Simulation Architecture

The simulation runs in adaptive time-steps, processing sequentially:

1. **Geological Events**: Tectonics, climate shifts, disasters
2. **Biome Updates**: Transitions, fragmentation, carrying capacity
3. **Population Dynamics**: Births, deaths, metabolism, predation
4. **Ecological Interactions**: Symbiosis, competition, disease
5. **Evolution**: Mutation, selection, speciation, extinction

**Core Logic**: `internal/ecosystem/population/dynamics.go`

### Time-step Resolution

```go
// Adaptive timestep based on simulation epoch
func getTimestep(year int64) int64 {
    if year < 100_000_000 {
        return 1_000 // 1,000-year steps (rapid early evolution)
    } else if year < 500_000_000 {
        return 10_000 // 10,000-year steps (stable ecosystems)
    } else {
        return 100_000 // 100,000-year steps (mature biosphere)
    }
}
```

**Total Iterations**: ~150,000 for 1 billion years (vs 1,000,000 with fixed 1-year steps)

---

## 2. Geological & Environmental Simulation

### 2.1 Tectonic Activity

#### Continental Drift
Plates move at 2-10 cm/year, causing:
- **Fragmentation**: Continents split apart
- **Collision**: Continents merge (supercontinent formation)
- **Ocean Current Disruption**: Changes global climate patterns

**Fragmentation Parameter**: `0.0` (supercontinent) to `1.0` (maximum fragmentation)

**Effects**:
- **High Fragmentation (>0.7)**:
  - Speciation rate +200%
  - Genetic drift rate +150%
  - Large animal stress (island dwarfism): Size penalty -20%
  - Coastal area +300% → marine biodiversity increases
  
- **Low Fragmentation (<0.3)** (Supercontinent):
  - Competition intensity +100%
  - Generalist species favored: Fitness +15%
  - Interior desert expansion: Arid biome area +50%
  - Speciation rate -60%

```go
// Fragmentation affects speciation rate
speciationRate := baseRate * (1.0 + fragmentation * 2.0)

// Large animals penalized on small landmasses
if fragmentation > 0.7 && traits.Size > 3.0 {
    fitness -= 0.2
}
```

#### Continental Configuration Effects

**🔧 IMPLEMENTATION NEEDED** (Priority 8):

```go
// internal/ecosystem/geography/continents.go

type ContinentalConfiguration struct {
    ContinentCount    int     // Number of separate landmasses
    LargestContinent  float64 // km²
    CoastalPercentage float64 // 0.0-1.0
    PangaeaIndex      float64 // 0.0 (fragmented) to 1.0 (supercontinent)
}

func (config *ContinentalConfiguration) calculateClimaticEffects() {
    // Supercontinent → interior deserts
    // Fragmented → more coastal climates
}
```

**Required Tests**:
- `TestSupercontinentInteriorDeserts`: Verify >70% becomes arid
- `TestFragmentationSpeciationBoost`: Verify 2x speciation rate
- `TestCoastalAreaCalculation`: Verify coastline increases with fragmentation

---

### 2.2 Geological Events

Events disrupt equilibrium, causing extinctions and opening niches.

| Event Type | Frequency | Severity Range | Duration | Primary Effects |
|------------|-----------|----------------|----------|-----------------|
| **Volcanic Winter** | Every 50-100M years | 0.3-0.7 | 10-100 years | Sunlight ↓70%, Flora mortality ↑80% |
| **Ice Age** | Every 10-50M years | 0.4-0.8 | 10,000-100,000 years | Temperature ↓15°C, biomes shift poleward |
| **Asteroid Impact** | Every 100-500M years | 0.7-0.95 | 1-10 years | Mass extinction 70-90%, "nuclear winter" |
| **Ocean Anoxia** | Every 20-80M years | 0.4-0.7 | 100,000-1M years | Marine O₂ ↓90%, ocean life collapse |
| **Flood Basalt** | Every 50-200M years | 0.5-0.8 | 1-2M years | Acid rain, warming, SO₂ poisoning |

#### Event Severity Formula

```go
func applyExtinctionEvent(eventType ExtinctionEventType, severity float64) int {
    baseMortality := getBaseMortality(eventType) // e.g., 0.5 for ice age
    
    // Exponential scaling for catastrophic events
    mortalityRate := baseMortality * math.Pow(severity, 1.5)
    
    deaths := 0
    for _, pop := range populations {
        // Size-based vulnerability
        sizeMultiplier := 1.0
        if eventType == EventAsteroidImpact && pop.Species.Traits.Size > 2.0 {
            sizeMultiplier = 2.0 // Large animals 2x more vulnerable
        }
        
        popDeaths := int(float64(pop.Count) * mortalityRate * sizeMultiplier)
        pop.Count -= popDeaths
        deaths += popDeaths
    }
    
    return deaths
}
```

**🔧 IMPLEMENTATION STATUS**: ✅ Basic events implemented, needs severity refinement

---

### 2.3 Oxygen Cycle

Atmospheric O₂ is dynamic, ranging from 15% (Permian) to 35% (Carboniferous).

#### Sources & Sinks

```go
type AtmosphericGases struct {
    O2Level  float64 // 0.15 to 0.35 (15% to 35%)
    CO2Level float64 // 180 to 7000+ ppm
}

func (atm *AtmosphericGases) UpdateOxygen(floraBiomass, faunaBiomass, volcanicActivity float64) {
    // Production: Photosynthesis
    production := floraBiomass * 0.0001 // 0.01% per biomass unit
    
    // Consumption: Respiration
    consumption := faunaBiomass * 0.00005 // 0.005% per biomass unit
    
    // Volcanic release (consumes O2 through oxidation)
    volcanicSink := volcanicActivity * 0.00002
    
    atm.O2Level += production - consumption - volcanicSink
    
    // Clamp to realistic range
    atm.O2Level = math.Max(0.10, math.Min(0.40, atm.O2Level))
}
```

#### Effects on Life

```go
func calculateO2Effects(o2Level float64, size float64) (maxSizeMultiplier, fitnessModifier float64) {
    // Low O2 (<15%): Limits maximum size
    if o2Level < 0.15 {
        deficit := (0.15 - o2Level) / 0.15
        maxSizeMultiplier = 1.0 - (deficit * 0.5) // Max 50% size reduction
        
        // Large animals suffer more
        if size > 3.0 {
            fitnessModifier = -0.2 * deficit
        }
    }
    
    // High O2 (>25%): Allows giant sizes
    if o2Level > 0.25 {
        excess := (o2Level - 0.25) / 0.10 // Normalized to [0, 1] for 25-35%
        maxSizeMultiplier = 1.0 + (excess * 2.0) // Up to 3x size
        
        // Fire frequency increases (ecological stress)
        fireFrequency := excess * 0.5 // Up to 50% increase
    }
    
    return maxSizeMultiplier, fitnessModifier
}
```

**Historical O₂ Levels**:
- **4.0-2.5 Ga** (billion years ago): <1% (anoxic atmosphere)
- **2.5-0.54 Ga**: 1-10% (Great Oxygenation Event)
- **540 Ma** (Cambrian): ~13%
- **320 Ma** (Carboniferous): ~35% (giant insects, 2m dragonflies)
- **250 Ma** (Permian): ~15% (mass extinction)
- **Present**: 21%

**🔧 IMPLEMENTATION NEEDED** (Priority 9):

```go
// internal/ecosystem/atmosphere/oxygen_cycle_test.go

func TestO2ProductionByFlora(t *testing.T) {
    // Verify photosynthesis increases O2
}

func TestO2ConsumptionByFauna(t *testing.T) {
    // Verify respiration decreases O2
}

func TestCarboniferousGiantInsects(t *testing.T) {
    // At 35% O2, verify arthropods can reach Size 5.0+
}

func TestPermianSizeLimitation(t *testing.T) {
    // At 15% O2, verify large animals face fitness penalty
}
```

---

### 2.4 Solar Evolution (Billion-Year Timescales)

**🔧 IMPLEMENTATION NEEDED** (Priority 11 - Long-term feature):

The Sun brightens ~10% per billion years due to hydrogen fusion.

```go
// internal/ecosystem/astronomy/solar_evolution.go

func calculateSolarLuminosity(ageInYears int64) float64 {
    // Age of simulation (0 = present, negative = past, positive = future)
    ageInBillionYears := float64(ageInYears) / 1e9
    
    // Sun was 70% luminosity 4 billion years ago
    // Sun will be 110% luminosity 1 billion years from now
    return 1.0 + (ageInBillionYears * 0.1)
}

func applySolarLuminosityEffects(luminosity float64) {
    // Temperature scales with √Luminosity (Stefan-Boltzmann)
    temperatureMultiplier := math.Sqrt(luminosity)
    
    // Effects on biomes
    if luminosity > 1.05 { // +500M years
        // Tropical expansion
        tropicalArea *= 1.3
        // Desert growth
        desertArea *= 1.5
    }
    
    if luminosity > 1.08 { // +800M years
        // Ocean evaporation increases
        oceanArea *= 0.9
        // Biosphere stress
        globalFitness *= 0.8
    }
    
    if luminosity > 1.10 { // +1B years
        // Approaching runaway greenhouse
        // Earth becoming Venus-like
        habitableArea *= 0.3
    }
}
```

**Required Tests**:
- `TestSolarBrighteningRate`: Verify 10% per billion years
- `TestTemperatureScaling`: Verify T ∝ √L
- `TestHabitableZoneShift`: Verify biosphere stress at +800M years

---

## 3. Species & Traits

### 3.1 Trait System

Species are defined by quantitative traits (most scaled 0.0-1.0):

#### Physical Traits
```go
type PhysicalTraits struct {
    Size     float64 // 0.1 to 10.0 (0.1 = mouse, 10.0 = elephant)
    Speed    SpeedType // Slow, Medium, Fast, VeryFast
    Strength float64 // 0.0 to 1.0 (affects predation success)
    Covering CoveringType // None, Fur, Scales, Feathers, Shell, Blubber
}
```

#### Survival Traits
```go
type SurvivalTraits struct {
    ColdResistance    float64 // 0.0 to 1.0
    HeatResistance    float64 // 0.0 to 1.0
    NightVision       float64 // 0.0 to 1.0
    Camouflage        float64 // 0.0 to 1.0 (prey defense)
    DiseaseResistance float64 // 0.0 to 1.0
}
```

#### Behavioral Traits
```go
type BehavioralTraits struct {
    Aggression   float64 // 0.0 to 1.0 (affects predation)
    Social       float64 // 0.0 (solitary) to 1.0 (pack/herd)
    Intelligence float64 // 0.0 to 1.0 (tool use, problem-solving)
}
```

#### Reproductive Traits
```go
type ReproductiveTraits struct {
    Fertility     float64 // 0.0 to 1.0 (base reproduction rate)
    MaturityAge   float64 // Years to sexual maturity
    LitterSize    float64 // 1.0 to 10.0+
    Lifespan      float64 // Years (max age)
}
```

#### Dietary Traits
```go
type DietaryTraits struct {
    Diet            DietType // Herbivore, Carnivore, Omnivore
    CarnivoreTendency float64 // 0.0 (pure herbivore) to 1.0 (pure carnivore)
    HasVenom        bool
    PoisonResistance float64 // 0.0 to 1.0
}

type DietType int
const (
    DietHerbivore DietType = iota // Primary consumer (eats plants)
    DietCarnivore                  // Secondary consumer (eats herbivores)
    DietOmnivore                   // Generalist (eats both)
)
```

---

### 3.2 Genetic Code Representation

**🔧 IMPLEMENTATION NEEDED** (Priority 5 - Part of Speciation):

Species carry a simplified genetic code as a 50-dimensional vector.

```go
// internal/ecosystem/population/genetics.go

type GeneticCode []float64 // Length 50, each gene in [0.0, 1.0]

type Species struct {
    ID             string
    Traits         *Traits
    GeneticCode    GeneticCode // 50D vector
    AncestorID     string      // For phylogeny tracking
    OriginYear     int64       // When species first appeared
    ExtinctionYear int64       // 0 if extant (still alive)
}

// generateGeneticCode creates initial random genome
func generateGeneticCode() GeneticCode {
    code := make(GeneticCode, 50)
    for i := range code {
        code[i] = rand.Float64()
    }
    return code
}

// Genotype-to-Phenotype Mapping
func updateTraitsFromGenetics(species *Species) {
    gc := species.GeneticCode
    
    // Genes 0-9: Size (average value → size 0.1 to 10.0)
    avgSizeGenes := average(gc[0:10])
    species.Traits.Size = 0.1 + (avgSizeGenes * 9.9)
    
    // Genes 10-14: Diet (argmax → diet type)
    species.Traits.Diet = DietType(argmax(gc[10:15]))
    
    // Genes 15-19: Covering (argmax → covering type)
    species.Traits.Covering = CoveringType(argmax(gc[15:21]))
    
    // Genes 21-24: Speed (argmax → speed type)
    species.Traits.Speed = SpeedType(argmax(gc[21:25]))
    
    // Genes 25-29: Cold resistance
    species.Traits.ColdResistance = average(gc[25:30])
    
    // Genes 30-34: Heat resistance
    species.Traits.HeatResistance = average(gc[30:35])
    
    // Genes 35-39: Camouflage
    species.Traits.Camouflage = average(gc[35:40])
    
    // Genes 40-44: Disease resistance
    species.Traits.DiseaseResistance = average(gc[40:45])
    
    // Genes 45-49: Intelligence
    species.Traits.Intelligence = average(gc[45:50])
}

// Phenotype-to-Genotype Reverse Mapping (for trait-based mutations)
func updateGeneticsFromTraits(species *Species) {
    // Convert traits back to genetic values
    // Used when traits mutate directly
}
```

**Required Tests**:
```go
func TestGeneticCodeGeneration(t *testing.T) {
    // Verify 50D vector, all values [0, 1]
}

func TestGenotypeToPhentotype(t *testing.T) {
    // Verify size genes → size trait mapping
}

func TestPhenotypeToGenotype(t *testing.T) {
    // Verify reverse mapping consistency
}
```

---

### 3.3 Taxonomy & Naming

Species names are procedurally generated based on traits and lineage:

```go
func generateSpeciesName(traits *Traits, biome geography.BiomeType) string {
    var parts []string
    
    // Size descriptor
    if traits.Size < 0.5 {
        parts = append(parts, "Tiny")
    } else if traits.Size < 1.0 {
        parts = append(parts, "Small")
    } else if traits.Size > 5.0 {
        parts = append(parts, "Giant")
    } else if traits.Size > 3.0 {
        parts = append(parts, "Large")
    }
    
    // Covering descriptor
    switch traits.Covering {
    case CoveringFur:
        parts = append(parts, "Woolly")
    case CoveringFeathers:
        parts = append(parts, "Feathered")
    case CoveringScales:
        parts = append(parts, "Scaled")
    case CoveringShell:
        parts = append(parts, "Armored")
    }
    
    // Speed descriptor
    if traits.Speed >= SpeedVeryFast {
        parts = append(parts, "Swift")
    }
    
    // Diet descriptor
    switch traits.Diet {
    case DietHerbivore:
        parts = append(parts, "Grazer")
    case DietCarnivore:
        if traits.Size > 3.0 {
            parts = append(parts, "Hunter")
        } else {
            parts = append(parts, "Stalker")
        }
    case DietOmnivore:
        parts = append(parts, "Forager")
    }
    
    // Biome descriptor (optional)
    biomeNames := map[geography.BiomeType]string{
        geography.BiomeTundra:     "Arctic",
        geography.BiomeDesert:     "Desert",
        geography.BiomeAlpine:     "Mountain",
        geography.BiomeRainforest: "Jungle",
    }
    if biomeName, exists := biomeNames[biome]; exists {
        parts = append(parts, biomeName)
    }
    
    return strings.Join(parts, " ")
}
```

**Examples**:
- `"Tiny Swift Stalker"` (small fast carnivore)
- `"Giant Woolly Grazer"` (large furred herbivore)
- `"Armored Desert Forager"` (shelled omnivore in desert)

**🔧 IMPLEMENTATION STATUS**: ✅ Basic naming exists, needs expansion

---

## 4. Population Dynamics

### 4.1 Size-Dependent Metabolism (Kleiber's Law)

**🔧 IMPLEMENTATION NEEDED** (Priority 2 - CRITICAL):

Metabolic rate doesn't scale linearly with mass—it follows a 3/4 power law.

```go
// internal/ecosystem/population/metabolism.go

// calculateMetabolicRate returns relative metabolic cost
// Kleiber's Law: Metabolic Rate ∝ Mass^0.75
func calculateMetabolicRate(size float64) float64 {
    return math.Pow(size, 0.75)
}

// Examples:
// Size 1.0 (baseline): metabolic rate = 1.0
// Size 8.0 (8x larger): metabolic rate = 4.76 (not 8.0!)
// Size 0.125 (8x smaller): metabolic rate = 0.21
```

#### Reproduction Rate Scaling

```go
// calculateReproductionRate returns relative reproduction speed
// Smaller animals reproduce faster (r vs K selection)
func calculateReproductionRate(size float64) float64 {
    // Inverse square root relationship
    return 1.0 / math.Sqrt(size)
}

// Examples:
// Size 0.125 (mouse): repro rate = 2.83 (fast breeding)
// Size 1.0 (baseline): repro rate = 1.0
// Size 8.0 (elephant): repro rate = 0.35 (slow breeding)
```

#### Carrying Capacity

```go
// calculatePopulationCapacity returns max sustainable population
func calculatePopulationCapacity(resourceUnits float64, size float64) float64 {
    metabolicCost := calculateMetabolicRate(size)
    return resourceUnits / metabolicCost
}

// Example with 10,000 resource units:
// Mouse (size 0.125): capacity = 10,000 / 0.21 = 47,619
// Baseline (size 1.0): capacity = 10,000 / 1.0 = 10,000
// Elephant (size 8.0): capacity = 10,000 / 4.76 = 2,101
```

#### Starvation Risk

```go
// calculateStarvationRate returns mortality during scarcity
func calculateStarvationRate(size float64, resourceAvailability float64) float64 {
    metabolicNeed := calculateMetabolicRate(size)
    
    // No starvation if resources meet need
    if resourceAvailability >= metabolicNeed {
        return 0.0
    }
    
    // Starvation risk proportional to unmet need
    deficit := (metabolicNeed - resourceAvailability) / metabolicNeed
    baseStarvationRate := 0.1 // 10% base mortality
    
    return baseStarvationRate * deficit
}

// Example with 50% resource availability:
// Mouse (need 0.21): deficit = 0%, no starvation
// Baseline (need 1.0): deficit = 50%, starvation rate = 5%
// Elephant (need 4.76): deficit = 89.5%, starvation rate = 8.95%
```

#### Integration with Population Updates

```go
func (p *Population) UpdatePopulation(biome geography.BiomeType, resourceAvailability float64) {
    traits := p.Species.Traits
    
    // Apply starvation
    starvationRate := calculateStarvationRate(traits.Size, resourceAvailability)
    deaths := int(float64(p.Count) * starvationRate)
    p.Count -= deaths
    
    if p.Count <= 0 {
        p.Count = 0
        return
    }
    
    // Apply reproduction
    baseGrowthRate := 0.02 // 2% per timestep
    reproModifier := calculateReproductionRate(traits.Size)
    births := int(float64(p.Count) * baseGrowthRate * reproModifier)
    
    // Apply carrying capacity limit
    capacity := p.calculateCarryingCapacity(biome, resourceAvailability)
    if p.Count + births > int(capacity) {
        births = int(capacity) - p.Count
        if births < 0 {
            births = 0
        }
    }
    
    p.Count += births
}
```

**Required Tests** (>80% coverage):
```go
func TestKleibersLaw(t *testing.T) {
    // Verify 8x size = 4.76x metabolic rate, not 8x
}

func TestReproductionInverseScaling(t *testing.T) {
    // Verify small animals reproduce faster
}

func TestCarryingCapacityBySize(t *testing.T) {
    // Verify small animals have higher population limits
}

func TestLargeAnimalStarvationVulnerability(t *testing.T) {
    // Verify large animals starve first during scarcity
}

func TestRvsKSelection(t *testing.T) {
    // Simulate 1000 years, verify small pops grow faster
}
```

---

### 4.2 Trophic Level Interactions

**🔧 IMPLEMENTATION NEEDED** (Priority 7):

Energy flows through trophic levels with 10% efficiency (ecological pyramid).

```go
// internal/ecosystem/population/trophic.go

type TrophicLevel int
const (
    Producer TrophicLevel = iota // Plants (Flora)
    PrimaryConsumer                // Herbivores
    SecondaryConsumer              // Carnivores
    ApexPredator                   // Top carnivores
)

const TrophicEfficiency = 0.1 // 10% energy transfer

func calculateTrophicCapacity(biome *BiomeState) map[TrophicLevel]float64 {
    capacity := make(map[TrophicLevel]float64)
    
    // Base production from sunlight (varies by biome)
    primaryProduction := biome.SunlightLevel * biome.WaterAvailability * 10000
    capacity[Producer] = primaryProduction
    
    // Each level gets 10% of previous
    capacity[PrimaryConsumer] = capacity[Producer] * TrophicEfficiency
    capacity[SecondaryConsumer] = capacity[PrimaryConsumer] * TrophicEfficiency
    capacity[ApexPredator] = capacity[SecondaryConsumer] * TrophicEfficiency
    
    return capacity
}

func (p *Population) getTrophicLevel() TrophicLevel {
    switch p.Species.Traits.Diet {
    case DietHerbivore:
        return PrimaryConsumer
    case DietCarnivore:
        // Check if it preys on carnivores (apex) or herbivores (secondary)
        if p.PreysOnCarnivores {
            return ApexPredator
        }
        return SecondaryConsumer
    case DietOmnivore:
        // Omnivores are intermediate
        return PrimaryConsumer
    default:
        return PrimaryConsumer
    }
}

// Lotka-Volterra Predator-Prey Dynamics
func applyPredatorPreyDynamics(prey, predator *Population, timestep float64) {
    // Parameters
    α := 0.1  // Prey birth rate
    β := 0.02 // Predation rate
    γ := 0.01 // Predator death rate
    δ := 0.01 // Predator birth per prey eaten
    
    preyCount := float64(prey.Count)
    predCount := float64(predator.Count)
    
    // Differential equations
    dPrey := (α*preyCount - β*preyCount*predCount) * timestep
    dPredator := (δ*preyCount*predCount - γ*predCount) * timestep
    
    prey.Count += int(dPrey)
    predator.Count += int(dPredator)
    
    // Prevent negative populations
    if prey.Count < 0 {
        prey.Count = 0
    }
    if predator.Count < 0 {
        predator.Count = 0
    }
}
```

**Trophic Cascade Effects**:
```go
// If top predators go extinct, herbivores explode, flora collapses
func checkTrophicCascade(ecosystem *Ecosystem) {
    if ecosystem.GetApexPredatorCount() == 0 {
        // No population control on herbivores
        for _, herbivore := range ecosystem.GetHerbivores() {
            herbivore.Count *= 2 // Population explosion
        }
        
        // Overgrazing reduces flora
        for _, flora := range ecosystem.GetFlora() {
            flora.Count = int(float64(flora.Count) * 0.5)
        }
    }
}
```

**Required Tests**:
```go
func TestTrophicEfficiency(t *testing.T) {
    // Verify 10% energy transfer between levels
}

func TestLotkaVolterraCycles(t *testing.T) {
    // Verify predator-prey oscillations
}

func TestApexExtinctionCascade(t *testing.T) {
    // Remove apex predators, verify herbivore explosion
}

func TestBiomassePyramid(t *testing.T) {
    // Verify producers >> primary >> secondary >> apex
}
```

---

### 4.3 Covering Effects on Survival

**🔧 IMPLEMENTATION NEEDED** (Priority 1 - HIGHEST PRIORITY):

Covering type dramatically affects fitness in different biomes.

```go
// internal/ecosystem/population/covering.go

func calculateCoveringFitness(traits *Traits, biome geography.BiomeType) float64 {
    coveringFitness := 0.0
    
    // Size modifier for thermal regulation (square-cube law)
    // Larger animals have lower surface-to-volume ratio
    thermalEfficiency := 1.0 / math.Sqrt(traits.Size)
    
    switch traits.Covering {
    case CoveringFur:
        // INSULATION: Excellent in cold climates
        if biome == geography.BiomeTundra || biome == geography.BiomeTaiga {
            coveringFitness += 0.2 // +20% fitness
        }
        if biome == geography.BiomeAlpine {
            coveringFitness += 0.15 // +15% fitness
        }
        
        // OVERHEATING: Terrible in hot climates
        if biome == geography.BiomeDesert {
            // Base penalty, worse for large animals
            heatPenalty := -0.15 * (1.0 + (traits.Size - 1.0) * 0.5)
            coveringFitness += heatPenalty
            // Size 1.0: -15% fitness
            // Size 3.0: -30% fitness (square-cube law)
        }
        if biome == geography.BiomeRainforest {
            coveringFitness -= 0.1 // Humidity makes fur problematic
        }
        
    case CoveringScales:
        // WATER RETENTION: Excellent in arid environments
        if biome == geography.BiomeDesert {
            coveringFitness += 0.15 // +15% fitness
        }
        
        // HYDRODYNAMICS: Good in aquatic environments
        if biome == geography.BiomeOcean || biome == geography.BiomeCoastal {
            coveringFitness += 0.1
            }
        
        // POOR INSULATION: Bad in extreme cold
        if biome == geography.BiomeTundra {
            coveringFitness -= 0.1
        }
        
    case CoveringFeathers:
        // BEST INSULATION-TO-WEIGHT RATIO
        if biome == geography.BiomeAlpine || biome == geography.BiomeTaiga {
            coveringFitness += 0.18 // +18% fitness
        }
        
        // WATER RESISTANCE: Good in humid/coastal areas
        if biome == geography.BiomeRainforest || biome == geography.BiomeCoastal {
            coveringFitness += 0.05
        }
        
    case CoveringShell:
        // MOBILITY PENALTY: Universal disadvantage
        coveringFitness -= 0.05 // -5% fitness everywhere
        
        // DESICCATION RESISTANCE: Good in deserts
        if biome == geography.BiomeDesert {
            coveringFitness += 0.1 // Net +5% in deserts
        }
        
        // Physical protection handled in predation calculation
        
    case CoveringBlubber:
        // EXTREME INSULATION + BUOYANCY
        if biome == geography.BiomeOcean || biome == geography.BiomeTundra {
            coveringFitness += 0.25 // +25% fitness (best for cold water)
        }
        
        // FATAL OVERHEATING: Terrible in hot climates
        if biome == geography.BiomeDesert {
            coveringFitness -= 0.3 // -30% fitness (nearly unsurvivable)
        }
        if biome == geography.BiomeRainforest {
            coveringFitness -= 0.2 // -20% fitness
        }
    }
    
    return coveringFitness
}

// calculatePredationRate handles predation with covering effects
func calculatePredationRate(traits *Traits, basePredationRate float64) float64 {
    modifier := 1.0
    
    switch traits.Covering {
    case CoveringShell:
        modifier = 0.6 // Shell reduces predation by 40%
    case CoveringScales:
        modifier = 0.9 // Scales reduce predation by 10%
    case CoveringFeathers:
        modifier = 0.95 // Feathers help escape (visual confusion) by 5%
    }
    
    return basePredationRate * modifier
}
```

**Integration with Fitness Calculation**:
```go
func (p *Population) CalculateBiomeFitness(biome geography.BiomeType, worldState *WorldState) float64 {
    fitness := 0.5 // Base fitness
    traits := p.Species.Traits
    
    // ... existing diet, size, speed calculations ...
    
    // ADD: Covering effects
    fitness += calculateCoveringFitness(traits, biome)
    
    // Clamp to [0.0, 1.0]
    return math.Max(0.0, math.Min(1.0, fitness))
}
```

**Required Tests** (>80% coverage):
```go
func TestFurInTundra(t *testing.T) {
    // Verify +20% fitness for furred species in tundra
}

func TestFurInDesert(t *testing.T) {
    // Verify negative fitness, worse for large animals
}

func TestScalesInDesert(t *testing.T) {
    // Verify +15% fitness for scaled species
}

func TestShellMobilityPenalty(t *testing.T) {
    // Verify -5% fitness everywhere
}

func TestShellPredationReduction(t *testing.T) {
    // Verify 40% predation reduction
}

func TestBlubberExtremes(t *testing.T) {
    // Verify +25% in ocean, -30% in desert
}

func TestSizeCoveringInteraction(t *testing.T) {
    // Verify large furred animals worse in heat
}

func TestCoveringEvolution(t *testing.T) {
    // Run 10M year sim, verify tundra → fur, desert → scales
}
```

---

## 5. Evolution Mechanisms

### 5.1 Mutation

```go
// internal/ecosystem/population/mutation.go

const BaseMutationRate = 0.001 // 0.1% per generation

func applyMutation(pop *Population, mutationRate float64) {
    for i := range pop.Species.GeneticCode {
        if rand.Float64() < mutationRate {
            // Random walk mutation
            delta := (rand.Float64() - 0.5) * 0.1 // ±5% change
            pop.Species.GeneticCode[i] += delta
            
            // Keep bounded [0, 1]
            pop.Species.GeneticCode[i] = math.Max(0.0, math.Min(1.0, pop.Species.GeneticCode[i]))
        }
    }
    
    // Update phenotype from genotype
    updateTraitsFromGenetics(pop.Species)
}

// Increased mutation after bottlenecks
func calculateAdaptiveMutationRate(populationSize int, postExtinction bool) float64 {
    baseMutation := BaseMutationRate
    
    // Small populations: increased drift allows fixation of mutations
    if populationSize < 500 {
        baseMutation *= 2.0
    }
    
    // Post-extinction: accelerated evolution
    if postExtinction {
        baseMutation *= 1.5
    }
    
    return baseMutation
}
```

**🔧 IMPLEMENTATION STATUS**: ✅ Basic mutations exist, needs bottleneck effects

---

### 5.2 Natural Selection

Selection pressure varies by biome and ecological context.

```go
// internal/ecosystem/population/selection.go

func applyBiomeSelection(pop *Population, biome geography.BiomeType) {
    traits := pop.Species.Traits
    selectionPressure := 0.0
    
    switch biome {
    case geography.BiomeTundra, geography.BiomeAlpine:
        // Favors: Cold resistance, large size (Bergmann's Rule), fur/blubber
        if traits.ColdResistance < 0.5 {
            selectionPressure -= 0.3 // Strong negative selection
        }
        if traits.Size < 1.0 {
            selectionPressure -= 0.2 // Small animals lose heat faster
        }
        if traits.Covering == CoveringFur || traits.Covering == CoveringBlubber {
            selectionPressure += 0.2 // Positive selection
        }
        
    case geography.BiomeDesert:
        // Favors: Heat resistance, small size, water conservation
        if traits.HeatResistance < 0.5 {
            selectionPressure -= 0.3
        }
        if traits.Size > 2.0 {
            selectionPressure -= 0.2 // Large animals overheat
        }
        if traits.Covering == CoveringScales {
            selectionPressure += 0.15 // Water retention
        }
        
    case geography.BiomeRainforest:
        // Favors: Camouflage, climbing, disease resistance
        if traits.Camouflage < 0.5 {
            selectionPressure -= 0.15 // High predation pressure
        }
        if traits.DiseaseResistance < 0.5 {
            selectionPressure -= 0.25 // High pathogen load
        }
    }
    
    // Apply selection as mortality
    if selectionPressure < 0 {
        mortalityRate := -selectionPressure
        deaths := int(float64(pop.Count) * mortalityRate)
        pop.Count -= deaths
    }
}
```

#### Predator-Prey Coevolution (Arms Race)

```go
func applyCoevolutionaryPressure(prey, predator *Population) {
    // Prey evolve: Speed, camouflage
    if prey.Species.Traits.Speed < SpeedFast {
        prey.Species.Traits.Speed++ // Increase speed
        prey.Species.Traits.Camouflage += 0.1
    }
    
    // Predators evolve: Speed, strength, intelligence
    if predator.Species.Traits.Speed <= prey.Species.Traits.Speed {
        predator.Species.Traits.Speed++ // Keep pace
    }
    predator.Species.Traits.Strength += 0.05
    predator.Species.Traits.Intelligence += 0.05 // Better hunting strategies
}
```

**🔧 IMPLEMENTATION STATUS**: ✅ Basic selection exists, needs coevolution

---

### 5.3 Genetic Drift (Neutral Evolution)

**🔧 IMPLEMENTATION NEEDED** (Priority 10):

Random allele frequency changes, especially in small populations.

```go
// internal/ecosystem/population/drift.go

// applyGeneticDrift simulates random allele frequency changes
func applyGeneticDrift(pop *Population) {
    driftRate := calculateDriftRate(pop.Count)
    
    for i := range pop.Species.GeneticCode {
        // Random walk independent of fitness
        if rand.Float64() < driftRate {
            delta := (rand.Float64() - 0.5) * 0.05 // Smaller than mutation
            pop.Species.GeneticCode[i] += delta
            
            // Keep bounded
            pop.Species.GeneticCode[i] = math.Max(0.0, math.Min(1.0, pop.Species.GeneticCode[i]))
        }
    }
    
    updateTraitsFromGenetics(pop.Species)
}

// calculateDriftRate based on population size
// Formula: drift rate ≈ 1 / (2 * N)
func calculateDriftRate(populationSize int) float64 {
    if populationSize == 0 {
        return 0.0
    }
    
    // Effective population size (often smaller than census size)
    Ne := float64(populationSize) * 0.7
    
    driftRate := 1.0 / (2.0 * Ne)
    
    // Cap at reasonable bounds
    return math.Max(0.0001, math.Min(0.05, driftRate))
}

// Founder effect: New populations start with reduced diversity
func applyFounderEffect(newPop *Population, founderCount int) {
    if founderCount < 100 {
        // Severe bottleneck: random subset of parent genes fixed
        for i := range newPop.Species.GeneticCode {
            if rand.Float64() < 0.5 {
                // Fix allele at random extreme
                if rand.Float64() < 0.5 {
                    newPop.Species.GeneticCode[i] = 0.0
                } else {
                    newPop.Species.GeneticCode[i] = 1.0
                }
            }
        }
    }
}
```

**Interaction with Selection**:
- **Selection dominant when**: `N * s >> 1` (large population, strong selection)
- **Drift dominant when**: `N * s << 1` (small population, weak selection)
- **Neutral evolution**: Most mutations have `s ≈ 0`, so drift always relevant

**Required Tests**:
```go
func TestDriftRateInversePopulation(t *testing.T) {
    // Verify drift rate = 1/(2N)
}

func TestSmallPopulationHighDrift(t *testing.T) {
    // Population of 50 should have high drift rate
}

func TestFounderEffect(t *testing.T) {
    // New population from 10 founders should have reduced diversity
}

func TestDriftVsSelection(t *testing.T) {
    // Small pop: drift can override weak selection
    // Large pop: selection dominates
}
```

---

### 5.4 Speciation (CRITICAL - Most Important Feature)

**🔧 IMPLEMENTATION NEEDED** (Priority 5 - HIGHEST PRIORITY after Covering):

New species branch off when populations diverge genetically and reproductively isolate.

#### Thresholds & Criteria

```go
// internal/ecosystem/population/speciation.go

const (
    SpeciationDistanceThreshold = 0.3     // 30% genetic divergence
    MinGenerationsForSpeciation = 10000   // 10,000 generations minimum
    AdaptiveRadiationMultiplier = 3.0     // Post-extinction speciation boost
)

// calculateGeneticDistance computes Euclidean distance between species
func calculateGeneticDistance(sp1, sp2 *Species) float64 {
    if len(sp1.GeneticCode) != len(sp2.GeneticCode) {
        return 1.0 // Maximum distance if incompatible
    }
    
    sumSqDiff := 0.0
    for i := range sp1.GeneticCode {
        diff := sp1.GeneticCode[i] - sp2.GeneticCode[i]
        sumSqDiff += diff * diff
    }
    
    // Euclidean distance, normalized by dimensionality
    distance := math.Sqrt(sumSqDiff) / math.Sqrt(float64(len(sp1.GeneticCode)))
    
    return distance
}

// canInterbreed determines if two species can produce viable offspring
func canInterbreed(sp1, sp2 *Species) bool {
    distance := calculateGeneticDistance(sp1, sp2)
    return distance < SpeciationDistanceThreshold
}

// shouldSpeciate determines if population becomes new species
func shouldSpeciate(geneticDistance float64, generationsSeparated int64) bool {
    // Need BOTH genetic distance AND time separation
    if generationsSeparated < MinGenerationsForSpeciation {
        return false
    }
    
    if geneticDistance < SpeciationDistanceThreshold {
        return false
    }
    
    // Probabilistic based on how far past thresholds
    distanceExcess := (geneticDistance - SpeciationDistanceThreshold) / SpeciationDistanceThreshold
    timeExcess := float64(generationsSeparated - MinGenerationsForSpeciation) / float64(MinGenerationsForSpeciation)
    
    speciationProb := math.Min((distanceExcess + timeExcess) / 4.0, 0.9)
    
    return rand.Float64() < speciationProb
}
```

#### Allopatric Speciation (Geographic Isolation)

```go
// CheckForSpeciation scans all populations for speciation events
func (ps *PopulationSimulator) CheckForSpeciation() int {
    speciationEvents := 0
    
    // Group populations by ancestral species
    speciesGroups := ps.groupByAncestor()
    
    for ancestorID, populations := range speciesGroups {
        if len(populations) < 2 {
            continue // Need at least 2 populations to diverge
        }
        
        // Check each pair for speciation
        for i := 0; i < len(populations); i++ {
            for j := i + 1; j < len(populations); j++ {
                pop1 := populations[i]
                pop2 := populations[j]
                
                // Calculate isolation duration
                isolation := ps.CurrentYear - pop1.LastContactYear
                
                // Calculate genetic distance
                distance := calculateGeneticDistance(pop1.Species, pop2.Species)
                
                // Check speciation criteria
                if shouldSpeciate(distance, isolation) {
                    // Create new species from pop2
                    newSpecies := ps.speciatePopulation(pop2, ancestorID)
                    speciationEvents++
                    
                    ps.logSpeciationEvent(pop1.Species, newSpecies, "allopatric")
                }
            }
        }
    }
    
    return speciationEvents
}

// speciatePopulation creates new species from population
func (ps *PopulationSimulator) speciatePopulation(pop *Population, ancestorID string) *Species {
    newSpecies := &Species{
        ID:          generateSpeciesID(),
        Traits:      copyTraits(pop.Species.Traits),
        GeneticCode: copyGeneticCode(pop.Species.GeneticCode),
        AncestorID:  ancestorID,
        OriginYear:  ps.CurrentYear,
        ExtinctionYear: 0, // Extant
    }
    
    // Update population to use new species
    pop.Species = newSpecies
    pop.LastContactYear = ps.CurrentYear
    
    // Track in phylogeny if enabled
    if ps.PhylogenyEnabled {
        ps.addToPhylogeny(ancestorID, newSpecies)
    }
    
    return newSpecies
}

// groupByAncestor organizes populations by ancestral lineage
func (ps *PopulationSimulator) groupByAncestor() map[string][]*Population {
    groups := make(map[string][]*Population)
    
    for _, pop := range ps.Populations {
        ancestorID := pop.Species.AncestorID
        if ancestorID == "" {
            ancestorID = pop.Species.ID // Is its own ancestor
        }
        groups[ancestorID] = append(groups[ancestorID], pop)
    }
    
    return groups
}
```

#### Sympatric Speciation (Same Location)

```go
// checkSympatricSpeciation looks for disruptive selection
func (ps *PopulationSimulator) checkSympatricSpeciation() int {
    speciationEvents := 0
    
    // For each biome with high resource diversity
    for biomeKey, biome := range ps.World.Biomes {
        if biome.ResourceDiversity < 0.7 {
            continue // Need high niche diversity
        }
        
        // Get populations in this biome
        pops := ps.GetPopulationsInBiome(biomeKey)
        
        for _, pop := range pops {
            // Large, generalist populations can split into specialists
            if pop.Count < 5000 {
                continue // Need large population for split
            }
            
            if pop.Species.Traits.Diet != DietOmnivore {
                continue // Need generalist to specialize
            }
            
            // Chance of sympatric speciation
            if rand.Float64() < 0.001 { // 0.1% per check
                ps.splitPopulationSympatric(pop, biomeKey)
                speciationEvents++
            }
        }
    }
    
    return speciationEvents
}

// splitPopulationSympatric divides population into two species
func (ps *PopulationSimulator) splitPopulationSympatric(pop *Population, biomeKey string) {
    // Keep half as original
    originalCount := pop.Count / 2
    pop.Count = originalCount
    
    // Create specialist 1: Herbivore
    herbivore := &Species{
        ID:          generateSpeciesID(),
        Traits:      copyTraits(pop.Species.Traits),
        GeneticCode: copyGeneticCode(pop.Species.GeneticCode),
        AncestorID:  pop.Species.ID,
        OriginYear:  ps.CurrentYear,
    }
    herbivore.Traits.Diet = DietHerbivore
    updateGeneticsFromTraits(herbivore)
    
    // Create specialist 2: Carnivore
    carnivore := &Species{
        ID:          generateSpeciesID(),
        Traits:      copyTraits(pop.Species.Traits),
        GeneticCode: copyGeneticCode(pop.Species.GeneticCode),
        AncestorID:  pop.Species.ID,
        OriginYear:  ps.CurrentYear,
    }
    carnivore.Traits.Diet = DietCarnivore
    updateGeneticsFromTraits(carnivore)
    
    // Add new populations
    ps.AddPopulation(herbivore, biomeKey, originalCount/2)
    ps.AddPopulation(carnivore, biomeKey, originalCount/2)
    
    ps.logSympatricSpeciation(pop.Species, herbivore, carnivore)
}

// ApplyDisruptiveSelection increases speciation pressure
func (ps *PopulationSimulator) ApplyDisruptiveSelection(biomeKey string) {
    pops := ps.GetPopulationsInBiome(biomeKey)
    
    for _, pop := range pops {
        // Disruptive selection favors extremes, disfavors intermediates
        // Temporarily increase mutation rate
        applyMutation(pop, BaseMutationRate * 2.0)
    }
}
```

#### Adaptive Radiation (Post-Extinction)

```go
// ApplyAdaptiveRadiation increases speciation after mass extinction
func (ps *PopulationSimulator) ApplyAdaptiveRadiation(duration int64) {
    ps.InRadiationPeriod = true
    ps.RadiationEndYear = ps.CurrentYear + duration
    
    // Speciation rate multiplied during this period
    ps.SpeciationRateMultiplier = AdaptiveRadiationMultiplier
}

// CheckForSpeciation (modified to include adaptive radiation)
func (ps *PopulationSimulator) CheckForSpeciation() int {
    baseEvents := ps.checkAllopatricSpeciation()
    baseEvents += ps.checkSympatricSpeciation()
    
    // Apply radiation multiplier if in recovery period
    if ps.InRadiationPeriod {
        if ps.CurrentYear > ps.RadiationEndYear {
            ps.InRadiationPeriod = false
            ps.SpeciationRateMultiplier = 1.0
        }
        
        // Additional speciation events during radiation
        additionalEvents := int(float64(baseEvents) * (ps.SpeciationRateMultiplier - 1.0))
        baseEvents += additionalEvents
    }
    
    return baseEvents
}
```

#### Phylogenetic Tree Construction

```go
// BuildPhylogeneticTree constructs full evolutionary tree
func (ps *PopulationSimulator) BuildPhylogeneticTree() *PhylogeneticTree {
    if !ps.PhylogenyEnabled {
        return nil
    }
    
    // Find root (earliest ancestor)
    var root *Species
    earliestYear := int64(math.MaxInt64)
    
    for _, species := range ps.AllSpecies {
        if species.OriginYear < earliestYear {
            earliestYear = species.OriginYear
            root = species
        }
    }
    
    if root == nil {
        return nil
    }
    
    // Build tree recursively
    rootNode := &PhylogeneticNode{
        Species:      root,
        Children:     []*PhylogeneticNode{},
        BranchLength: 0,
    }
    
    ps.buildTreeRecursive(rootNode)
    
    return &PhylogeneticTree{Root: rootNode}
}

func (ps *PopulationSimulator) buildTreeRecursive(node *PhylogeneticNode) {
    // Find all direct descendants
    for _, species := range ps.AllSpecies {
        if species.AncestorID == node.Species.ID {
            distance := calculateGeneticDistance(node.Species, species)
            
            childNode := &PhylogeneticNode{
                Species:      species,
                Children:     []*PhylogeneticNode{},
                BranchLength: distance,
            }
            
            node.Children = append(node.Children, childNode)
            
            // Recurse
            ps.buildTreeRecursive(childNode)
        }
    }
}

type PhylogeneticNode struct {
    Species      *Species
    Children     []*PhylogeneticNode
    BranchLength float64 // Genetic distance to parent
}

type PhylogeneticTree struct {
    Root *PhylogeneticNode
}

// GetLeaves returns all terminal nodes (extant species)
func (tree *PhylogeneticTree) GetLeaves() []*PhylogeneticNode {
    leaves := []*PhylogeneticNode{}
    tree.collectLeaves(tree.Root, &leaves)
    return leaves
}

func (tree *PhylogeneticTree) collectLeaves(node *PhylogeneticNode, leaves *[]*PhylogeneticNode) {
    if len(node.Children) == 0 {
        *leaves = append(*leaves, node)
        return
    }
    
    for _, child := range node.Children {
        tree.collectLeaves(child, leaves)
    }
}
```

**Required Tests** (>80% coverage):
```go
func TestGeneticDistanceCalculation(t *testing.T) {
    // Verify Euclidean distance calculation
}

func TestReproductiveIsolation(t *testing.T) {
    // Distance >0.3 prevents interbreeding
}

func TestAllopatricSpeciation(t *testing.T) {
    // 100k years isolation → speciation
}

func TestSympatricSpeciation(t *testing.T) {
    // High diversity + large generalist → specialists
}

func TestAdaptiveRadiation(t *testing.T) {
    // Post-extinction speciation rate 3x normal
}

func TestPhylogeneticTreeConstruction(t *testing.T) {
    // Build tree, verify ancestor-descendant relationships
}

func TestSpeciationThresholds(t *testing.T) {
    // Test various distance/time combinations
}
```

---

### 5.5 Mass Extinction & Recovery

#### Extinction Events
Already covered in §2.2. Effects:
- **70-90% mortality** for severe events (asteroids)
- **Lilliput Effect**: Large species (Size >3.0) face 2x mortality
- **Selective Pressure**: Cold-adapted species survive ice ages better

#### Recovery Phase

```go
const RecoveryPeriodYears = 20000 // 20,000 year "healing" period

func (ps *PopulationSimulator) ApplyExtinctionEvent(eventType ExtinctionEventType, severity float64) int {
    deaths := ps.applyExtinctionMortality(eventType, severity)
    
    // Start recovery period
    ps.ApplyAdaptiveRadiation(RecoveryPeriodYears)
    
    // Increase mutation rate temporarily
    ps.PostExtinctionMutationBoost = true
    ps.MutationBoostEndYear = ps.CurrentYear + RecoveryPeriodYears
    
    return deaths
}

// During recovery: elevated speciation, increased mutation
func (ps *PopulationSimulator) isInRecoveryPeriod() bool {
    return ps.InRadiationPeriod || ps.PostExtinctionMutationBoost
}
```

**🔧 IMPLEMENTATION STATUS**: ✅ Basic extinctions exist, needs recovery mechanics

---

## 6. Ecological Interactions

### 6.1 Symbiosis & Mutualism

**🔧 IMPLEMENTATION NEEDED** (Priority 6):

Symbiotic relationships create interdependencies and co-extinction risks.

```go
// internal/ecosystem/population/symbiosis.go

type SymbiosisType int
const (
    Mutualism SymbiosisType = iota // Both benefit
    Commensalism                     // One benefits, other neutral
    Parasitism                       // One benefits, other harmed
)

type SymbioticLink struct {
    PartnerA      *Species
    PartnerB      *Species
    Type          SymbiosisType
    Strength      float64 // 0.0 to 1.0 (how dependent)
    BenefitA      float64 // Fitness benefit to A
    BenefitB      float64 // Fitness benefit to B
}

// createMutualisticLink establishes pollination or seed dispersal
func createMutualisticLink(flora *Species, fauna *Species) *SymbioticLink {
    return &SymbioticLink{
        PartnerA: flora,
        PartnerB: fauna,
        Type:     Mutualism,
        Strength: 0.6, // Moderately dependent
        BenefitA: 0.3, // Flora: +30% reproduction (pollination)
        BenefitB: 0.2, // Fauna: +20% food availability
    }
}

// applySymbioticEffects modifies population growth
func applySymbioticEffects(pop *Population, links []*SymbioticLink) {
    totalBenefit := 0.0
    
    for _, link := range links {
        // Check if partner is still alive
        var benefit float64
        var partnerAlive bool
        
        if link.PartnerA.ID == pop.Species.ID {
            benefit = link.BenefitA
            partnerAlive = isSpeciesExtant(link.PartnerB)
        } else {
            benefit = link.BenefitB
            partnerAlive = isSpeciesExtant(link.PartnerA)
        }
        
        if partnerAlive {
            totalBenefit += benefit
        } else {
            // Co-extinction risk
            coExtinctionPenalty := -link.Strength * 0.2 // Lose 20% fitness
            totalBenefit += coExtinctionPenalty
        }
    }
    
    // Apply benefits to growth rate
    pop.GrowthRateModifier += totalBenefit
}

// checkCoExtinction handles partner death
func checkCoExtinction(extinctSpecies *Species, links []*SymbioticLink) []*Species {
    atRisk := []*Species{}
    
    for _, link := range links {
        var partner *Species
        if link.PartnerA.ID == extinctSpecies.ID {
            partner = link.PartnerB
        } else {
            partner = link.PartnerA
        }
        
        // Strong mutualistic links create extinction risk
        if link.Type == Mutualism && link.Strength > 0.7 {
            coExtinctionProb := link.Strength * 0.5 // Up to 50% chance
            if rand.Float64() < coExtinctionProb {
                atRisk = append(atRisk, partner)
            }
        }
    }
    
    return atRisk
}

// evolveSymbiosis creates new relationships over time
func (ps *PopulationSimulator) evolveSymbiosis() {
    // Flora-Fauna partnerships (pollination, seed dispersal)
    for _, flora := range ps.GetFlora() {
        for _, fauna := range ps.GetSmallFauna() { // Small = pollinators
            // Check proximity and compatibility
            if ps.areInSameBiome(flora, fauna) && rand.Float64() < 0.001 {
                link := createMutualisticLink(flora.Species, fauna.Species)
                ps.SymbioticLinks = append(ps.SymbioticLinks, link)
            }
        }
    }
}
```

**Symbiosis Types**:

| Type | Example | Effect on A | Effect on B |
|------|---------|-------------|-------------|
| **Mutualism** | Pollination (flower + bee) | +30% reproduction | +20% food |
| **Commensalism** | Scavenging (vulture + lion) | 0% | +10% food |
| **Parasitism** | Tapeworm + host | +15% growth | -10% fitness |

**Required Tests**:
```go
func TestMutualismBenefits(t *testing.T) {
    // Both partners grow faster
}

func TestCoExtinctionRisk(t *testing.T) {
    // Partner death causes fitness drop
}

func TestStrongMutualismCoExtinction(t *testing.T) {
    // Strength >0.7 → high co-extinction probability
}

func TestSymbiosisEvolution(t *testing.T) {
    // Over time, mutualistic links emerge
}
```

---
### 6.2 Niche Partitioning & Competition

**🔧 IMPLEMENTATION NEEDED** (Priority 6):

```go
// internal/ecosystem/population/competition.go

// hasNicheOverlap checks if two species compete for resources
func hasNicheOverlap(sp1, sp2 *Species) bool {
    // Same diet = overlap
    if sp1.Traits.Diet == sp2.Traits.Diet {
        // Similar size = more overlap (within 50%)
        sizeRatio := sp1.Traits.Size / sp2.Traits.Size
        if sizeRatio > 0.5 && sizeRatio < 2.0 {
            return true
        }
    }
    
    return false
}

// calculateCompetitionIntensity measures resource overlap
func calculateCompetitionIntensity(sp1, sp2 *Species) float64 {
    intensity := 0.0
    
    // Diet similarity
    if sp1.Traits.Diet == sp2.Traits.Diet {
        intensity += 0.5
    }
    
    // Size similarity (Gaussian kernel)
    sizeDiff := math.Abs(sp1.Traits.Size - sp2.Traits.Size)
    sizeSimilarity := math.Exp(-sizeDiff * sizeDiff / 2.0)
    intensity += sizeSimilarity * 0.3
    
    // Activity time overlap (if implemented)
    // intensity += activityOverlap * 0.2
    
    return intensity
}

// applyCompetition reduces growth rates of competing species
func applyCompetition(pop1, pop2 *Population) {
    if !hasNicheOverlap(pop1.Species, pop2.Species) {
        return
    }
    
    intensity := calculateCompetitionIntensity(pop1.Species, pop2.Species)
    
    // Competitive exclusion principle: more fit species wins
    fitness1 := pop1.CalculateBiomeFitness(pop1.CurrentBiome, nil)
    fitness2 := pop2.CalculateBiomeFitness(pop2.CurrentBiome, nil)
    
    if fitness1 > fitness2 {
        // Species 2 suffers
        pop2.GrowthRateModifier -= intensity * 0.2
    } else {
        // Species 1 suffers
        pop1.GrowthRateModifier -= intensity * 0.2
    }
}

// characterDisplacement forces divergence to reduce competition
func applyCharacterDisplacement(pop1, pop2 *Population) {
    if calculateCompetitionIntensity(pop1.Species, pop2.Species) < 0.5 {
        return // Not enough competition pressure
    }
    
    // Diverge in traits to reduce overlap
    // Example: One becomes nocturnal, other diurnal
    if rand.Float64() < 0.01 { // 1% chance per check
        // Diverge size
        if pop1.Species.Traits.Size < pop2.Species.Traits.Size {
            pop1.Species.Traits.Size *= 0.9 // Get smaller
            pop2.Species.Traits.Size *= 1.1 // Get larger
        } else {
            pop1.Species.Traits.Size *= 1.1
            pop2.Species.Traits.Size *= 0.9
        }
        
        // Update genetics to match
        updateGeneticsFromTraits(pop1.Species)
        updateGeneticsFromTraits(pop2.Species)
    }
}
```

**Required Tests**:
```go
func TestNicheOverlapDetection(t *testing.T) {
    // Same diet + similar size = overlap
}

func TestCompetitiveExclusion(t *testing.T) {
    // Less fit species declines when competing
}

func TestCharacterDisplacement(t *testing.T) {
    // Competing species diverge over time
}
```

---

### 6.3 Disease Dynamics

**🔧 IMPLEMENTATION NEEDED** (Priority 6):

```go
// internal/ecosystem/population/disease.go

type Disease struct {
    ID               string
    HostSpecies      *Species
    InfectionRate    float64 // 0.0 to 1.0
    MortalityRate    float64 // 0.0 to 1.0
    TransmissionType TransmissionType
}

type TransmissionType int
const (
    DensityDependent TransmissionType = iota // Higher density = more spread
    FrequencyDependent                        // Constant infection rate
)

// checkForOutbreak determines if disease emerges
func checkForOutbreak(pop *Population, carryingCapacity float64) *Disease {
    // Trigger: Population > 1.5x carrying capacity
    overcrowding := float64(pop.Count) / carryingCapacity
    
    if overcrowding < 1.5 {
        return nil
    }
    
    // Outbreak probability increases with overcrowding
    outbreakProb := (overcrowding - 1.5) * 0.2 // Up to 20% per check
    
    if rand.Float64() < outbreakProb {
        return &Disease{
            ID:            generateDiseaseID(),
            HostSpecies:   pop.Species,
            InfectionRate: 0.05 + (overcrowding-1.5)*0.1, // 5-25%
            MortalityRate: 0.5 * (1.0 - pop.Species.Traits.DiseaseResistance),
            TransmissionType: DensityDependent,
        }
    }
    
    return nil
}

// applyDiseaseOutbreak reduces population
func applyDiseaseOutbreak(pop *Population, disease *Disease) int {
    // Calculate infections
    infectionRate := disease.InfectionRate
    
    // Social species spread disease faster
    if pop.Species.Traits.Social > 0.7 {
        infectionRate *= 2.0 // Pack/herd species 2x infection
    } else if pop.Species.Traits.Social < 0.3 {
        infectionRate *= 0.5 // Solitary species 0.5x infection
    }
    
    infected := int(float64(pop.Count) * infectionRate)
    
    // Calculate deaths
    deaths := int(float64(infected) * disease.MortalityRate)
    
    pop.Count -= deaths
    
    // Survivors evolve resistance
    pop.Species.Traits.DiseaseResistance += 0.1
    if pop.Species.Traits.DiseaseResistance > 1.0 {
        pop.Species.Traits.DiseaseResistance = 1.0
    }
    
    return deaths
}

// Disease prevents single-species dominance (negative feedback)
func (ps *PopulationSimulator) regulateWithDisease() {
    for _, pop := range ps.Populations {
        biome := ps.World.Biomes[pop.CurrentBiomeKey]
        capacity := calculateCarryingCapacity(biome, pop.Species.Traits.Size)
        
        disease := checkForOutbreak(pop, capacity)
        if disease != nil {
            deaths := applyDiseaseOutbreak(pop, disease)
            ps.logDiseaseOutbreak(pop.Species, deaths)
        }
    }
}
```

**Disease Effects Summary**:
- **Trigger**: Density > 1.5x carrying capacity
- **Base Infection**: 5% per year, scales with overcrowding
- **Mortality**: 50% * (1 - DiseaseResistance)
- **Social Penalty**: Pack/herd species 2x infection rate
- **Evolution**: Survivors gain +0.1 resistance per outbreak

**Required Tests**:
```go
func TestDiseaseOutbreakTrigger(t *testing.T) {
    // Verify outbreak at 1.5x capacity
}

func TestSocialSpeciesSpreadFaster(t *testing.T) {
    // Social species 2x infection rate
}

func TestResistanceEvolution(t *testing.T) {
    // Survivors gain resistance
}

func TestDiseaseRegulation(t *testing.T) {
    // Prevents single species monopoly
}
```

---

### 6.4 Seasonal Cycles

```go
// internal/ecosystem/population/seasons.go

type Season int
const (
    Spring Season = iota // Breeding season
    Summer               // Growth season
    Fall                 // Migration season
    Winter               // Scarcity season
)

func getSeasonalModifiers(season Season, biome geography.BiomeType, latitude float64) (foodMod, survivalMod float64) {
    switch season {
    case Spring:
        foodMod = 1.2      // Abundant food
        survivalMod = 1.0  // Normal survival
        
    case Summer:
        foodMod = 1.5      // Peak food availability
        survivalMod = 1.0
        
    case Fall:
        foodMod = 1.0      // Normal food
        survivalMod = 1.0
        
    case Winter:
        // High latitudes face harsh winters
        winterSeverity := latitude / 90.0 // 0.0 at equator, 1.0 at pole
        
        if biome == geography.BiomeTundra || biome == geography.BiomeTaiga {
            foodMod = 0.2 // Severe scarcity
            survivalMod = 0.7 - (winterSeverity * 0.3) // 70% to 40% survival
        } else if biome == geography.BiomeTemperate {
            foodMod = 0.6
            survivalMod = 0.9
        } else {
            foodMod = 1.0 // Tropical = no winter
            survivalMod = 1.0
        }
    }
    
    return foodMod, survivalMod
}
```

**🔧 IMPLEMENTATION STATUS**: ⭕ Optional feature (low priority)

---

## 7. Migration & Biome Transitions

### 7.1 Fitness-Gradient Migration

**🔧 IMPLEMENTATION NEEDED** (Priority 4):

Populations migrate toward higher-fitness biomes.

```go
// internal/ecosystem/population/migration.go

// calculateMigrationGradient returns fitness difference (target - source)
func calculateMigrationGradient(sourceFitness, targetFitness float64) float64 {
    return targetFitness - sourceFitness
}

// calculateMigrationRate returns proportion of population that migrates
func calculateMigrationRate(sourceFitness, targetFitness float64, populationSize int) float64 {
    // Minimum viable population for migration
    const minMigrationPop = 100
    
    if populationSize < minMigrationPop {
        return 0.0
    }
    
    gradient := calculateMigrationGradient(sourceFitness, targetFitness)
    
    // Only migrate toward better conditions
    if gradient <= 0 {
        return 0.0
    }
    
    // Base migration rate: 5% per cycle
    baseMigrationRate := 0.05
    
    // Scale by gradient (max 3x for gradient of 0.6+)
    scalingFactor := math.Min(gradient / 0.2, 3.0)
    
    return baseMigrationRate * scalingFactor
}

// Examples:
// Gradient 0.6 (0.2 → 0.8): rate = 15%
// Gradient 0.1 (0.5 → 0.6): rate = 2.5%
// Gradient -0.3 (0.8 → 0.5): rate = 0% (no migration to worse biome)
```

#### Adaptive Migration Intervals

```go
// calculateMigrationInterval returns years between migration checks
func calculateMigrationInterval(year int64, ecosystemStability float64) int64 {
    baseInterval := int64(1000)
    
    // After 100M years, use longer intervals
    if year > 100_000_000 {
        baseInterval = 10_000
    }
    
    // After 500M years, even longer
    if year > 500_000_000 {
        baseInterval = 20_000
    }
    
    // During instability (post-extinction), check more frequently
    if ecosystemStability < 0.5 {
        return baseInterval / 2
    }
    
    // During high stability, check less frequently
    if ecosystemStability > 0.8 {
        return baseInterval
    }
    
    // Normal stability
    return baseInterval * 3 / 4
}

// calculateEcosystemStability measures recent extinction rate + population variance
func (ps *PopulationSimulator) calculateEcosystemStability() float64 {
    if len(ps.Populations) == 0 {
        return 0.0
    }
    
    // Check recent extinction rate
    recentExtinctions := ps.getRecentExtinctions(10_000) // Last 10k years
    extinctionPressure := 1.0 - math.Min(float64(recentExtinctions)/10.0, 1.0)
    
    // Check population variance (stable pops = stable ecosystem)
    variance := ps.calculatePopulationVariance()
    varianceStability := 1.0 - math.Min(variance/0.5, 1.0)
    
    // Average the factors
    stability := (extinctionPressure + varianceStability) / 2.0
    
    return math.Max(0.0, math.Min(1.0, stability))
}
```

#### Migration Cycle Execution

```go
// ApplyMigrationCycle processes one round of migration
func (ps *PopulationSimulator) ApplyMigrationCycle() (migrants int, extinctions int) {
    totalMigrants := 0
    totalExtinctions := 0
    
    // Process each population
    for _, pop := range ps.Populations {
        if pop.Count < 100 {
            continue // Too small to migrate
        }
        
        // Find adjacent biomes
        adjacentBiomes := ps.getAdjacentBiomes(pop.CurrentBiomeKey)
        
        if len(adjacentBiomes) == 0 {
            continue
        }
        
        // Calculate fitness in current biome
        currentBiome := ps.getBiomeType(pop.CurrentBiomeKey)
        sourceFitness := pop.CalculateBiomeFitness(currentBiome, ps.World)
        
        // Find best target biome
        var bestTarget string
        var bestFitness float64 = sourceFitness
        
        for _, adjKey := range adjacentBiomes {
            adjBiome := ps.getBiomeType(adjKey)
            targetFitness := pop.CalculateBiomeFitness(adjBiome, ps.World)
            
            if targetFitness > bestFitness {
                bestFitness = targetFitness
                bestTarget = adjKey
            }
        }
        
        // No better biome found
        if bestTarget == "" {
            continue
        }
        
        // Calculate migration
        migrationRate := calculateMigrationRate(sourceFitness, bestFitness, pop.Count)
        migrantCount := int(float64(pop.Count) * migrationRate)
        
        if migrantCount == 0 {
            continue
        }
        
        // Process migration
        ext := ps.ProcessMigration(pop.Species, bestTarget, migrantCount)
        
        // Update source population
        pop.Count -= migrantCount
        
        totalMigrants += migrantCount
        totalExtinctions += ext
    }
    
    return totalMigrants, totalExtinctions
}
```

#### Competitive Exclusion from Migration

```go
// ProcessMigration handles migrants arriving in target biome
func (ps *PopulationSimulator) ProcessMigration(
    species *Species,
    targetBiomeKey string,
    migrantCount int,
) int {
    extinctions := 0
    
    // Check if population already exists in target
    existingPop := ps.GetPopulation(species.ID, targetBiomeKey)
    
    if existingPop != nil {
        // Add to existing population
        existingPop.Count += migrantCount
        return 0
    }
    
    // Create new population
    newPop := &Population{
        Species:         species,
        Count:          migrantCount,
        CurrentBiomeKey: targetBiomeKey,
        BirthYear:      ps.CurrentYear,
    }
    
    ps.Populations = append(ps.Populations, newPop)
    
    // Check for competitive exclusion with existing species
    extinctions += ps.checkCompetitiveExclusion(targetBiomeKey, species)
    
    return extinctions
}

// checkCompetitiveExclusion determines if new arrivals cause extinctions
func (ps *PopulationSimulator) checkCompetitiveExclusion(
    biomeKey string,
    newSpecies *Species,
) int {
    extinctions := 0
    targetBiome := ps.getBiomeType(biomeKey)
    
    newFitness := calculateBiomeFitness(newSpecies.Traits, targetBiome)
    
    // Check all populations in same biome with similar niches
    for _, pop := range ps.Populations {
        if pop.CurrentBiomeKey != biomeKey {
            continue
        }
        
        if pop.Species.ID == newSpecies.ID {
            continue
        }
        
        // Check for niche overlap
        if !hasNicheOverlap(pop.Species, newSpecies) {
            continue
        }
        
        // Calculate relative fitness
        existingFitness := pop.CalculateBiomeFitness(targetBiome, ps.World)
        
        // If new species is significantly fitter (>0.2 advantage)
        if newFitness > existingFitness+0.2 {
            // Existing species faces extinction pressure
            extinctionProb := 0.1 * (newFitness - existingFitness)
            
            if rand.Float64() < extinctionProb {
                pop.Count = 0 // Mark for extinction
                extinctions++
            }
        }
    }
    
    return extinctions
}
```

**Required Tests** (>80% coverage):
```go
func TestFitnessGradientCalculation(t *testing.T) {
    // Verify gradient = target - source
}

func TestMigrationRateScaling(t *testing.T) {
    // High gradient → high rate, negative gradient → 0
}

func TestMinimumMigrationPopulation(t *testing.T) {
    // <100 individuals cannot migrate
}

func TestAdaptiveMigrationInterval(t *testing.T) {
    // Early sim: 1k years, late sim: 20k years
}

func TestMigrationTowardBetterFitness(t *testing.T) {
    // Pop in desert migrates to adjacent tundra if furred
}

func TestCompetitiveExclusionFromMigration(t *testing.T) {
    // Strong invader causes native extinction
}

func TestEcosystemStabilityCalculation(t *testing.T) {
    // Post-extinction → low stability
}
```

---

### 7.2 Dynamic Biome Shifts with Transition Speed

**🔧 IMPLEMENTATION NEEDED** (Priority 3):

Different geological events cause biome changes at different rates.

```go
// internal/ecosystem/geography/biome_transitions.go

type BiomeCell struct {
    Type               geography.BiomeType
    Latitude           float64 // 0° to 90°
    Elevation          float64 // meters
    IsCoastal          bool
    ResourceDiversity  float64 // 0.0 to 1.0
    TransitionProgress float64 // 0.0 to 1.0
    TargetType         geography.BiomeType
}

type WorldState struct {
    Biomes          map[string]*BiomeCell
    BiomeAdjacency  map[string][]string // Which biomes are adjacent
    Year            int64
}

// determineTransitionParameters returns event type and speed
func determineTransitionParameters(eventType ExtinctionEventType, severity float64) (string, float64) {
    var event string
    var transitionSpeed float64
    
    switch eventType {
    case EventIceAge:
        event = "ice_age"
        // Glacial advance: 1,000-10,000 years
        transitionSpeed = 0.001 * severity
        
    case EventAsteroidImpact:
        if severity > 0.8 {
            event = "ice_age" // Nuclear winter
            transitionSpeed = 0.1 // Years to decades
        } else if severity > 0.5 {
            event = "impact_winter"
            transitionSpeed = 0.05 // Decades
        } else {
            event = "regional_devastation"
            transitionSpeed = 0.2 // Immediate
        }
        
    case EventFloodBasalt:
        event = "warming"
        // Eruptions last 1-2 million years
        transitionSpeed = 0.0001
        
    case EventVolcanicWinter:
        if severity > 0.7 {
            event = "ice_age"
            transitionSpeed = 0.01 // Centuries
        } else {
            event = "cooling"
            transitionSpeed = 0.05
        }
        
    case EventContinentalDrift:
        event = "tectonic_reorganization"
        // 10-100 million years
        transitionSpeed = 0.00001
        
    case EventOceanAnoxia:
        event = "anoxic_event"
        transitionSpeed = 0.001 // Millennial scale
    }
    
    return event, transitionSpeed
}

// calculateTransitionProbability determines if a biome cell should transition
func calculateTransitionProbability(cell *BiomeCell, event string, transitionSpeed float64) float64 {
    baseProb := transitionSpeed
    
    // Refugia resistance factors
    resistanceFactor := 1.0
    
    // Coastal areas moderated by ocean thermal mass
    if cell.IsCoastal {
        resistanceFactor *= 0.6 // 40% resistance
    }
    
    // Elevation effects
    if event == "ice_age" {
        // High elevations cool faster
        if cell.Elevation > 2000 {
            resistanceFactor *= 1.5 // 50% faster transition
        }
        // Low elevations resist more
        if cell.Elevation < 500 {
            resistanceFactor *= 0.7
        }
    }
    
    // Latitude effects
    if event == "ice_age" {
        // Poles transition faster, equator resists
        latitudeFactor := cell.Latitude / 90.0 // 0.0 at equator, 1.0 at pole
        resistanceFactor *= (0.3 + 0.7*latitudeFactor)
    } else if event == "warming" {
        // Warming affects poles more (polar amplification)
        latitudeFactor := cell.Latitude / 90.0
        resistanceFactor *= (0.8 + 0.4*latitudeFactor)
    }
    
    // Ocean anoxia only affects water biomes
    if event == "anoxic_event" {
        if cell.Type == geography.BiomeOcean {
            return 0.8 // Highly affected
        } else if cell.Type == geography.BiomeCoastal {
            return 0.3 // Moderately affected
        } else {
            return 0.0 // Not affected
        }
    }
    
    return baseProb * resistanceFactor
}

// getTransitionTarget determines what biome type to transition to
func getTransitionTarget(current geography.BiomeType, event string, latitude float64) geography.BiomeType {
    switch event {
    case "ice_age":
        if latitude > 60 {
            return geography.BiomeTundra
        } else if latitude > 45 {
            return geography.BiomeTaiga
        } else if current == geography.BiomeDesert {
            return geography.BiomeDesert // Deserts can expand during ice ages
        } else {
            return geography.BiomeTemperate
        }
        
    case "warming":
        if latitude > 75 {
            return geography.BiomeTundra // Still cold at extreme poles
        } else if latitude > 55 {
            return geography.BiomeTaiga
        } else if latitude > 30 {
            return geography.BiomeTemperate
        } else if current == geography.BiomeRainforest {
            return geography.BiomeRainforest // Rainforests expand
        } else {
            return geography.BiomeDesert // More desertification
        }
        
    case "anoxic_event":
        if current == geography.BiomeOcean {
            // Dead zones expand (same type but degraded)
            return geography.BiomeOcean
        }
        return current
        
    case "tectonic_reorganization":
        // Complex - depends on new continental positions
        // Simplified: change based on latitude
        return determineBiomeByLatitude(latitude)
    }
    
    return current
}

// applyGradualTransition updates biomes over time
func applyGradualTransition(world *WorldState, event string, transitionSpeed float64, years int) {
    for key, cell := range world.Biomes {
        // Calculate if this cell should transition
        prob := calculateTransitionProbability(cell, event, transitionSpeed)
        
        // Accumulate transition progress
        cell.TransitionProgress += prob * float64(years)
        
        // If progress >= 1.0, complete the transition
        if cell.TransitionProgress >= 1.0 {
            target := getTransitionTarget(cell.Type, event, cell.Latitude)
            
            if target != cell.Type {
                cell.Type = target
                cell.TransitionProgress = 0.0
            }
        }
        
        world.Biomes[key] = cell
    }
}

// Integration with simulation
func (ps *PopulationSimulator) ApplyBiomeTransitions(
    eventType ExtinctionEventType,
    severity float64,
    year int64,
) {
    event, speed := determineTransitionParameters(eventType, severity)
    
    // Apply transition over appropriate timescale
    yearsToSimulate := 1000 // Chunk size
    
    applyGradualTransition(ps.World, event, speed, yearsToSimulate)
    
    // Log significant changes
    ps.logBiomeChanges()
}
```

**Transition Speed Summary**:

| Event | Speed | Timescale | Example |
|-------|-------|-----------|---------|
| Ice Age | 0.001 * severity | 1,000-10,000 years | Glacial advance |
| Asteroid Impact | 0.1 | Years-decades | Nuclear winter |
| Flood Basalt | 0.0001 | 1-2 million years | Siberian Traps |
| Continental Drift | 0.00001 | 10-100 million years | Pangaea breakup |
| Volcanic Winter | 0.01 | Centuries | Tambora eruption |

**Refugia Effects**:
- **Coastal areas**: 40% resistance (ocean thermal buffering)
- **High elevation during ice ages**: 50% faster cooling
- **Equatorial regions**: Resist ice age transitions

**Required Tests** (>80% coverage):
```go
func TestTransitionSpeedByEvent(t *testing.T) {
    // Verify ice age slower than asteroid
}

func TestCoastalRefugia(t *testing.T) {
    // Coastal areas resist transitions
}

func TestLatitudinalIceAgeProgression(t *testing.T) {
    // Ice ages advance from poles
}

func TestOceanAnoxiaLandImmunity(t *testing.T) {
    // Inland biomes unaffected by ocean anoxia
}

func TestGradualTransitionProgress(t *testing.T) {
    // Verify accumulation until threshold
}
```

---

## 8. Performance & Optimization

### 8.1 Adaptive Timesteps

For billion-year simulations, use variable resolution:

```go
// internal/ecosystem/simulation/timestep.go

type SimulationEpoch struct {
    StartYear        int64
    EndYear          int64
    Timestep         int64
    RecordFrequency  int64 // How often to save state
    EventCheckFreq   int64 // How often to check for extinctions
}

var epochs = []SimulationEpoch{
    {
        StartYear:       0,
        EndYear:         100_000_000,
        Timestep:        1_000,
        RecordFrequency: 10_000,
        EventCheckFreq:  10_000_000,
    },
    {
        StartYear:       100_000_000,
        EndYear:         500_000_000,
        Timestep:        10_000,
        RecordFrequency: 100_000,
        EventCheckFreq:  50_000_000,
    },
    {
        StartYear:       500_000_000,
        EndYear:         1_000_000_000,
        Timestep:        100_000,
        RecordFrequency: 1_000_000,
        EventCheckFreq:  100_000_000,
    },
}

func getEpoch(year int64) *SimulationEpoch {
    for _, epoch := range epochs {
        if year >= epoch.StartYear && year < epoch.EndYear {
            return &epoch
        }
    }
    return &epochs[len(epochs)-1] // Default to last epoch
}

// Main simulation loop with adaptive timesteps
func (sim *Simulator) Run(totalYears int64) {
    currentYear := int64(0)
    
    for currentYear < totalYears {
        epoch := getEpoch(currentYear)
        
        // Process one timestep
        sim.ProcessTimestep(currentYear, epoch.Timestep)
        
        // Conditional checks based on epoch
        if currentYear % epoch.EventCheckFreq == 0 {
            sim.CheckForExtinctionEvents()
        }
        
        if currentYear % epoch.RecordFrequency == 0 {
            sim.RecordState()
        }
        
        currentYear += epoch.Timestep
    }
}
```

**Performance Gain**: ~150,000 iterations for 1 billion years vs 1,000,000 with fixed 1-year steps

---

### 8.2 Memory Management

```go
// Extinct species: Move to fossil record after 10M years
func (ps *PopulationSimulator) pruneExtinctSpecies() {
    cutoffYear := ps.CurrentYear - 10_000_000
    
    for i := len(ps.AllSpecies) - 1; i >= 0; i-- {
        species := ps.AllSpecies[i]
        
        if species.ExtinctionYear > 0 && species.ExtinctionYear < cutoffYear {
            // Move to fossil record
            ps.FossilRecord = append(ps.FossilRecord, species)
            
            // Remove from active species
            ps.AllSpecies = append(ps.AllSpecies[:i], ps.AllSpecies[i+1:]...)
        }
    }
}

// Population culling: Remove pops <10 individuals
func (ps *PopulationSimulator) cullSmallPopulations() {
    for i := len(ps.Populations) - 1; i >= 0; i-- {
        if ps.Populations[i].Count < 10 {
            ps.Populations = append(ps.Populations[:i], ps.Populations[i+1:]...)
        }
    }
}

// Call periodically
if year % 1_000_000 == 0 {
    ps.pruneExtinctSpecies()
    ps.cullSmallPopulations()
}
```

---

### 8.3
Benchmark Targets

```go
// internal/ecosystem/simulation/benchmark_test.go

func BenchmarkOneMillion Years(b *testing.B) {
    sim := NewSimulator()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        sim.Run(1_000_000)
    }
    // Target: <1 second
}

func BenchmarkOneHundredMillionYears(b *testing.B) {
    sim := NewSimulator()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        sim.Run(100_000_000)
    }
    // Target: <2 minutes
}

func BenchmarkOneBillionYears(b *testing.B) {
    sim := NewSimulator()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        sim.Run(1_000_000_000)
    }
    // Target: <30 minutes
}
```

**Target Performance**:
- 1 million years: <1 second
- 100 million years: <2 minutes
- 1 billion years: <30 minutes

---

## 9. Testing & Validation

### 9.1 TDD Requirements

**ALL new code MUST have**:
- ✅ Tests written BEFORE implementation
- ✅ Minimum 80% code coverage
- ✅ Integration tests for multi-system interactions
- ✅ Benchmark tests for performance-critical paths

### 9.2 Test Categories

#### Unit Tests
Test individual functions in isolation.

```go
// Example: Testing Kleiber's Law
func TestMetabolicRateScaling(t *testing.T) {
    tests := []struct {
        size           float64
        expectedRate   float64
    }{
        {1.0, 1.0},
        {8.0, 4.76},
        {0.125, 0.21},
    }
    
    for _, tt := range tests {
        rate := calculateMetabolicRate(tt.size)
        if math.Abs(rate-tt.expectedRate) > 0.1 {
            t.Errorf("Size %v: expected %v, got %v", tt.size, tt.expectedRate, rate)
        }
    }
}
```

#### Integration Tests
Test multiple systems working together.

```go
func TestFurredSpeciesEvolutionInTundra(t *testing.T) {
    sim := NewPopulationSimulator()
    
    // Start with naked species in tundra
    nakedSpecies := &Species{
        Traits: &Traits{Covering: CoveringNone},
    }
    sim.AddPopulation(nakedSpecies, "tundra_1", 1000)
    
    // Run 10M years
    for year := 0; year < 10_000_000; year += 10000 {
        sim.ProcessTimestep(year, 10000)
    }
    
    // Check that fur evolved
    tundraSpecies := sim.GetSpeciesInBiome("tundra_1")
    furCount := 0
    for _, sp := range tundraSpecies {
        if sp.Traits.Covering == CoveringFur {
            furCount++
        }
    }
    
    if furCount == 0 {
        t.Errorf("Expected fur to evolve in tundra")
    }
}
```

#### Scientific Validation Tests
Verify simulation matches known evolutionary patterns.

```go
func TestCarboniferousGiantInsects(t *testing.T) {
    sim := NewSimulatorWithO2(0.35) // 35% oxygen
    
    // Add arthropods
    arthropod := &Species{
        Traits: &Traits{Size: 1.0},
        Taxonomy: Arthropod,
    }
    sim.AddPopulation(arthropod, "forest_1", 1000)
    
    // Run 50M years
    sim.Run(50_000_000)
    
    // Check for giant sizes
    arthropods := sim.GetArthropods()
    hasGiant := false
    for _, arth := range arthropods {
        if arth.Traits.Size > 5.0 {
            hasGiant = true
        }
    }
    
    if !hasGiant {
        t.Errorf("Expected giant arthropods in high O2 environment")
    }
}

func TestMassExtinctionRecovery(t *testing.T) {
    sim := NewPopulationSimulator()
    
    // Pre-extinction diversity
    initialSpecies := 100
    for i := 0; i < initialSpecies; i++ {
        sim.AddRandomSpecies()
    }
    
    preExtinction := sim.GetSpeciesCount()
    
    // Asteroid impact (85% mortality)
    sim.ApplyExtinctionEvent(EventAsteroidImpact, 0.85)
    
    postExtinction := sim.GetSpeciesCount()
    
    // Should lose ~60-80% of species (not just individuals)
    if float64(postExtinction) > float64(preExtinction)*0.4 {
        t.Errorf("Extinction not severe enough")
    }
    
    // Track recovery
    speciationRate := sim.GetSpeciationRate()
    
    // Run 20k year recovery
    sim.Run(20_000)
    
    recoveryRate := sim.GetSpeciationRate()
    
    // Should have 3x normal speciation during recovery
    if recoveryRate < speciationRate*2.5 {
        t.Errorf("Expected elevated speciation during recovery")
    }
}
```

### 9.3 Coverage Verification

```bash
# Run all tests with coverage
go test ./internal/ecosystem/... -coverprofile=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Fail if coverage <80%
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 80% threshold"
    exit 1
fi
```

---

## Implementation Roadmap Summary

### Phase 1: Core Survival Mechanics (Weeks 1-3)
- **Priority 1**: ✅ Covering → Survivability (HIGHEST)
  - Tests: >80% coverage
  - Expected: 10M years → tundra dominated by fur, desert by scales
- **Priority 2**: ✅ Size-Dependent Metabolism (Kleiber's Law)
  - Tests: >80% coverage
  - Expected: Large animals starve first, small animals reproduce faster

### Phase 2: Population Dynamics (Weeks 4-6)
- **Priority 3**: ✅ Dynamic Biome Shifts with Speed
  - Tests: >80% coverage
  - Expected: Ice ages advance slowly, asteroids rapid
- **Priority 4**: ✅ Migration with Fitness Gradients
  - Tests: >80% coverage
  - Expected: Populations migrate toward better biomes

### Phase 3: Speciation & Genetics (Weeks 7-10)
- **Priority 5**: ✅ Speciation Events (CRITICAL)
  - Tests: >80% coverage
  - Expected: 100k years isolation → new species
- **Priority 6**: ✅ Symbiosis & Disease
  - Tests: >80% coverage
  - Expected: Mutualism, co-extinction, disease regulation
- **Priority 7**: ✅ Trophic Levels & Predation
  - Tests: >80% coverage
  - Expected: Lotka-Volterra cycles, cascades

### Phase 4: Advanced Features (Weeks 11-14)
- **Priority 8**: ⭕ Continental Configuration
- **Priority 9**: ⭕ Oxygen Cycle Effects
- **Priority 10**: ⭕ Genetic Drift
- **Priority 11**: ⭕ Solar Evolution (billion-year scale)

### Phase 5: Optimization (Weeks 15-16)
- Adaptive timesteps
- Memory management
- Benchmark to targets
- Final integration testing

---

## Critical Missing Features Checklist

Use this checklist to track implementation:

### HIGHEST PRIORITY (Must implement first):
- [ ] **Priority 1**: Covering → Survivability (with size interaction)
- [ ] **Priority 2**: Size-Dependent Metabolism (Kleiber's Law)
- [ ] **Priority 5**: Speciation Mechanics (genetic distance, thresholds)

### HIGH PRIORITY (Core gameplay):
- [ ] **Priority 3**: Dynamic Biome Transitions (with speeds)
- [ ] **Priority 4**: Migration (fitness gradients)
- [ ] **Priority 7**: Trophic Levels (Lotka-Volterra)

### MEDIUM PRIORITY (Depth):
- [ ] **Priority 6**: Symbiosis & Disease
- [ ] **Priority 10**: Genetic Drift

### LOW PRIORITY (Long-term):
- [ ] **Priority 8**: Continental Configuration
- [ ] **Priority 9**: Oxygen Cycle
- [ ] **Priority 11**: Solar Evolution

### PERFORMANCE:
- [ ] Adaptive timesteps
- [ ] Memory management
- [ ] Benchmark tests
- [ ] Coverage >80% all modules

---

## Quick Reference: Key Formulae

### Metabolism (Kleiber's Law)
```
Metabolic Rate = Size^0.75
Reproduction Rate = 1 / √Size
Carrying Capacity = Resources / Metabolic Rate
```

### Fitness Modifiers
```
Fur in Tundra: +20%
Fur in Desert: -15% (worse for large animals)
Scales in Desert: +15%
Shell everywhere: -5% (but -40% predation)
Blubber in Ocean: +25%
Blubber in Desert: -30%
```

### Speciation Thresholds
```
Genetic Distance: >0.3 (30% divergence)
Time Separation: >10,000 generations
Probability = min((distanceExcess + timeExcess) / 4, 0.9)
```

### Biome Transition Speeds
```
Ice Age: 0.001 * severity (1k-10k years)
Asteroid: 0.1 (years-decades)
Flood Basalt: 0.0001 (1-2M years)
Continental Drift: 0.00001 (10-100M years)
```

### Migration
```
Rate = 5% * min(gradient / 0.2, 3.0)
Min Population: 100 individuals
Only toward positive gradients
```

---

## Questions for Clarification

Before implementing, please clarify:

1. **Genetic Code Dimensionality**: Is 50D sufficient, or should we use 100D+ for more trait complexity?
1. **Answer**: Let's use 100D+ - Is it possible to have 100 defined genetic traits and have 100 blank that can be added to as mutations arise?

2. **Timestep Granularity**: Are the proposed adaptive timesteps (1k → 10k → 100k years) acceptable for gameplay, or too coarse?
2. **Answer**: Currently the game freezes if we try to simulate more than a million years, I would like to see enhancements to the performance that allow us to simulate a billion years at our current granularity before we move to adaptive timesteps.

3. **Speciation Rate**: Should sympatric speciation be rarer (currently 0.1% per check)?
3. **Answer**: Sympatric speciation should be rarer (0.001% per check) unless speciation has just occured in which case it is 0.01% for the next 100,000 years. Allopatric speciation should be much more common.

4. **Disease Mechanics**: Should diseases persist across generations, or are outbreaks transient?
4. **Answer**: Diseases should persist, we want them to be realistic and have them be dormant in some population and spread into others so technically both states are true. Diseases persist across generation once they become endimec but when they first show up they are transient if the mutation doesn't allow for them to persist without killing off their hosts before spreading.

5. **Performance Targets**: Are 30 minutes for 1 billion years acceptable, or should we target faster?
5. **Answer**: 30 minutes for 1 billion years is acceptable as long as it is happening asynchronously and a player can fly about the world, watching it change beneath them.

6. **Flora Simulation**: Should plants have their own trait system, or remain abstract "resources"?
6. **Answer**: Plants should have their own trait system, we want them to be just as rich as the animals. We also want to keep the separation between flora and fauna fuzzy to account for lifeforms like fungi and other organisms that don't fall comfortable into the flora and fauna categories.

7. **Player Interaction**: At what timescale can players intervene (introduce species, trigger events)?
7. **Answer**: That's a great question, should we introduce turning points every 100k years where players are given three options for the direction the world goes?
