# Ecosystem Simulation V2 - Implementation Plan

**Created**: December 15, 2025
**Status**: Draft - Pending Review
**Reference**: [ECOSYSTEM_SIMULATION_V2.md](file:///Users/walker/git/thousand-worlds/tw-backend/docs/ECOSYSTEM_SIMULATION_V2.md)

---

## Overview

This plan outlines the implementation work needed to bring the Thousand Worlds ecosystem simulation from its current state to the V2 design. The work is organized into phases with dependencies clearly marked.

---

## Phase 0: Foundation & Logging (Week 1)

### 0.1 World Simulation Logger
**Files**: `internal/ecosystem/logger.go` (NEW)

- [ ] Create dedicated simulation logger with dual output (file + DB)
- [ ] Implement verbosity levels: TRACE, DEBUG, INFO, WARN, ERROR
- [ ] Log file path: `/logs/world_simulation.log`
- [ ] Database table: `simulation_events` for queryable history
- [ ] Add profiling hooks to identify performance bottlenecks

```go
type SimulationLogger struct {
    FileLogger   *log.Logger
    DBLogger     *DBEventLogger
    Verbosity    LogLevel
    WorldID      uuid.UUID
}

type SimulationEvent struct {
    ID        uuid.UUID
    WorldID   uuid.UUID
    Year      int64
    EventType string  // "speciation", "extinction", "disease", "turning_point"
    Severity  float64
    Details   json.RawMessage
    Timestamp time.Time
}
```

### 0.2 Checkpoint System
**Files**: `internal/ecosystem/checkpoint.go` (NEW)

- [ ] Full world state snapshot every 1,000,000 years
- [ ] Delta storage between snapshots
- [ ] Checkpoint restoration for rewind feature
- [ ] Database table: `world_checkpoints`

---

## Phase 1: Unified Organism Model (Weeks 2-4)

### 1.1 New Organism Type
**Files**: `internal/ecosystem/population/organism.go` (NEW)

Replace `EvolvableTraits` with unified `Organism`:

```go
type Organism struct {
    ID              uuid.UUID
    Name            string
    GeneticCode     GeneticCode     // 200D vector (100 defined + 100 blank)
    Traits          OrganismTraits  // Derived from genetic code
    AncestorID      *uuid.UUID
    OriginYear      int64
    ExtinctionYear  int64           // 0 if extant
}

type OrganismTraits struct {
    // Continuous values 0.0-1.0
    Autotrophy      float64  // 0=heterotroph, 1=autotroph
    Complexity      float64  // 0=prokaryote, 1=complex multicellular
    Motility        float64  // 0=sessile, 1=highly mobile
    
    // Physical (existing)
    Size            float64
    Speed           float64
    Strength        float64
    
    // ... (all existing traits)
    
    // New traits
    ToolUse         float64  // 0=none, 1=advanced
    Communication   float64  // 0=none, 1=complex language
    MagicAffinity   float64  // 0=mundane, 1=highly magical (if enabled)
}
```

### 1.2 Genetic Code System
**Files**: `internal/ecosystem/population/genetics.go` (NEW)

```go
// IMPORTANT: Use float32 to save 50% memory bandwidth
type GeneticCode struct {
    DefinedGenes  [100]float32  // Map to phenotypic traits via Expression Matrix
    BlankGenes    [100]float32  // Unlockable exotic traits
    ActiveBlanks  []int         // Indices of unlocked blank genes
}

// Genotype → Phenotype via Expression Matrix (P = G × E)
// Enables punctuated equilibrium: silent mutations until threshold crossed
type ExpressionMatrix struct {
    Weights [100][25]float32  // 100 genes → 25 phenotypic traits
}

type GeneMapping struct {
    TraitName    string
    GeneIndices  []int        // Polygenic: multiple genes affect one trait
    Weights      []float32    // How much each gene contributes
    Category     GeneCategory // For weighted speciation distance
}

type GeneCategory int
const (
    GeneBodyPlan    GeneCategory = iota  // Genes 0-5: 10x weight
    GeneMorphology                        // Genes 6-20: 5x weight
    GeneBehavior                          // Genes 21-50: 2x weight
    GeneMinor                             // Genes 51-100: 1x weight
)

// Universal exotic traits (available to all worlds)
var UniversalExoticTraits = []ExoticTrait{
    {Name: "Bioluminescence", GeneIndex: 0},
    {Name: "Echolocation", GeneIndex: 1},
    {Name: "Regeneration", GeneIndex: 2},
    {Name: "Camouflage_Active", GeneIndex: 3},
    {Name: "Venom_Enhanced", GeneIndex: 4},
    // ... 50+ universal traits
}

// Fantastical traits (require magic-enabled world)
var FantasticalTraits = []ExoticTrait{
    {Name: "Magic_Affinity", GeneIndex: 50, RequiresMagic: true},
    {Name: "Teleportation", GeneIndex: 51, RequiresMagic: true},
    {Name: "Shield_Generation", GeneIndex: 52, RequiresMagic: true},
    {Name: "Telekinesis", GeneIndex: 53, RequiresMagic: true},
    {Name: "Shadowy", GeneIndex: 54, RequiresMagic: true},
    // ... more fantastical traits
}
```

### 1.3 World Settings
**Files**: `internal/worldgen/settings.go` (MODIFY)

```go
type WorldSettings struct {
    // Existing fields...
    
    // New fields
    MagicSetting        MagicType  // Mundane, Magical, OpenToMagic
    AllowMagicalCreatures bool     // For mundane worlds
    LifeformBasis       LifeBasis  // Carbon, Silicon, Energy, Crystalline
    ExoticTraitSeed     int64      // For world-specific trait generation
}

type MagicType string
const (
    MagicMundane    MagicType = "mundane"
    MagicEnabled    MagicType = "magical"
    MagicOpenTo     MagicType = "open_to_magic"
)

type LifeBasis string
const (
    LifeCarbon      LifeBasis = "carbon"
    LifeSilicon     LifeBasis = "silicon"
    LifeEnergy      LifeBasis = "energy"
    LifeCrystalline LifeBasis = "crystalline"
)
```

---

## Phase 2: Geographic Isolation (Weeks 5-6)

### 2.1 Region System
**Files**: `internal/ecosystem/geography/regions.go` (NEW)

```go
// Hex Grid Tectonic System (NOT a simple float!)
type TectonicSystem struct {
    Plates       []TectonicPlate
    CellToPlate  map[HexCoord]uuid.UUID
    SubductionZones []SubductionZone
    RiftZones    []RiftZone
}

type TectonicPlate struct {
    ID           uuid.UUID
    Name         string
    Type         PlateType  // Continental, Oceanic
    Velocity     Vector2D
    LandmassPct  float32
    Cells        []HexCoord
}

type HexCoord struct {
    Q, R int  // Axial coordinates
}

type HexCell struct {
    Coord       HexCoord
    PlateID     uuid.UUID
    BiomeID     uuid.UUID
    RegionID    uuid.UUID
    Elevation   float32
    Temperature float32
    Moisture    float32
}

// Fragmentation is a DERIVED statistic
func (ts *TectonicSystem) CalculateFragmentation() float32 {
    continentalPlates := 0
    for _, plate := range ts.Plates {
        if plate.Type == Continental && plate.LandmassPct > 0.2 {
            continentalPlates++
        }
    }
    return float32(math.Min(float64(continentalPlates-1)/6.0, 1.0))
}

type Region struct {
    ID              uuid.UUID
    BiomeID         uuid.UUID
    Cells           []HexCoord  // Cells in this region
    TerrainType     TerrainType
    Connections     []RegionConnection
    LastUpdated     int64
    IsolationYears  int64       // For gigantism/dwarfism calculation
}

type RegionConnection struct {
    TargetRegionID  uuid.UUID
    Passable        bool
    Difficulty      float32  // 0=easy, 1=impassable
    ObstacleType    string   // "mountain", "river", "ocean", "none"
}

type TerrainType string
const (
    TerrainPlains    TerrainType = "plains"     // 2x migration
    TerrainForest    TerrainType = "forest"     // 0.7x migration
    TerrainMountain  TerrainType = "mountain"   // 0.3x migration
    TerrainSwamp     TerrainType = "swamp"      // 0.5x migration
    TerrainDesert    TerrainType = "desert"     // 0.6x migration
    TerrainOcean     TerrainType = "ocean"      // 0 for land species
)
```

### 2.2 Isolation Tracking
**Files**: `internal/ecosystem/population/isolation.go` (NEW)

```go
type PopulationIsolation struct {
    PopulationID    uuid.UUID
    RegionID        uuid.UUID
    LastContactYear map[uuid.UUID]int64  // Last contact with other populations
}

func (ps *PopulationSimulator) UpdateRegions() {
    // Every 10,000 years, re-evaluate geography
    // Update region boundaries based on terrain changes
    // Update connections based on new obstacles
}

func CalculateMigrationRange(traits OrganismTraits, terrain TerrainType, social float64) float64 {
    baseRange := traits.Size*0.5 + traits.Speed*0.3 + traits.Intelligence*0.2
    terrainMod := getTerrainModifier(terrain)
    herdBonus := 1.0
    if social > 0.7 {
        herdBonus = 1.5
    }
    return baseRange * terrainMod * herdBonus
}
```

---

## Phase 3: Advanced Speciation (Weeks 7-9)

### 3.1 Genetic Distance Calculation
**Files**: `internal/ecosystem/population/speciation.go` (NEW/MAJOR UPDATE)

```go
const (
    SpeciationDistanceThreshold = 0.3   // 30% genetic divergence
    MinGenerationsForSpeciation = 10000
    SympatricBaseRate           = 0.00001  // 0.001%
    SympatricPostSpeciationRate = 0.0001   // 0.01% for 100k years
    MinSympatricPopulation      = 5000     // Minimum pop for sympatric
)

// Weighted genetic distance - body plan genes count 10x more
func CalculateGeneticDistance(g1, g2 GeneticCode) float32 {
    var weightedSumSq float32 = 0.0
    for i := 0; i < 100; i++ {
        weight := getGeneWeight(i)  // 10x for 0-5, 5x for 6-20, 2x for 21-50, 1x for 51-100
        diff := g1.DefinedGenes[i] - g2.DefinedGenes[i]
        weightedSumSq += weight * diff * diff
    }
    totalWeight := 10*6 + 5*15 + 2*30 + 1*49  // = 60 + 75 + 60 + 49 = 244
    return float32(math.Sqrt(float64(weightedSumSq / float32(totalWeight))))
}

// Inbreeding depression for small populations
func CalculateInbreedingPenalty(populationSize int64) float64 {
    if populationSize >= 50 {
        return 1.0  // No penalty
    }
    if populationSize < 2 {
        return 0.1  // Near extinction
    }
    return 0.1 + 0.9*(float64(populationSize-2)/48.0)
}

func (ps *PopulationSimulator) CheckAllopatricSpeciation() int {
    // For each species with populations in different regions
    // If isolation > MinGenerationsForSpeciation AND genetic distance > threshold
    // Create new species
}

func (ps *PopulationSimulator) CheckSympatricSpeciation() int {
    // For large generalist populations in high-diversity biomes
    // Low probability split into specialists
    // Minimum population size: 5000
}
```

### 3.2 Phylogenetic Tree
**Files**: `internal/ecosystem/population/phylogeny.go` (NEW)

```go
type PhylogeneticTree struct {
    Root     *PhylogeneticNode
    WorldID  uuid.UUID
}

type PhylogeneticNode struct {
    Species      *Organism
    Children     []*PhylogeneticNode
    BranchLength float64  // Genetic distance from parent
    BranchTime   int64    // Years since split
}

func (ps *PopulationSimulator) BuildPhylogeneticTree() *PhylogeneticTree
func (tree *PhylogeneticTree) GetLeaves() []*PhylogeneticNode
func (tree *PhylogeneticTree) GetExtinct() []*PhylogeneticNode
func (tree *PhylogeneticTree) GetMRCA(sp1, sp2 uuid.UUID) *PhylogeneticNode
```

---

## Phase 4: Pathogen System (Weeks 10-11)

### 4.1 Pathogen Types
**Files**: `internal/ecosystem/pathogen/types.go` (NEW)

```go
type PathogenType string
const (
    PathogenVirus    PathogenType = "virus"
    PathogenBacteria PathogenType = "bacteria"
    PathogenFungus   PathogenType = "fungus"
    PathogenPrion    PathogenType = "prion"
)

type Pathogen struct {
    ID              uuid.UUID
    Name            string
    Type            PathogenType
    Traits          PathogenTraits
    HostSpeciesIDs  []uuid.UUID
    OriginYear      int64
    Status          PathogenStatus  // Transient, Endemic, Dormant
}

type PathogenTraits struct {
    Virulence       float64  // 0-1, lethality
    Transmissibility float64 // 0-1, spread rate
    IncubationDays  float64  // Days before symptoms
    HostSpecificity float64  // 0=broad, 1=narrow
    MutationRate    float64  // How fast it evolves
}
```

### 4.2 Disease Simulation
**Files**: `internal/ecosystem/pathogen/simulation.go` (NEW)

```go
func (ps *PopulationSimulator) SimulatePathogens() {
    // Check for new outbreaks (density-dependent)
    // Spread existing pathogens
    // Apply mortality
    // Evolve endemic pathogens (reduce virulence over time)
    // Check for zoonotic transfer
}

func CanInfect(pathogen *Pathogen, species *Organism, phylogeny *PhylogeneticTree) bool {
    // Check if species is in host list
    // OR closely related species (shared ancestor within 1M years)
}

func (p *Pathogen) Evolve(yearsEndemic int64) {
    // Endemic pathogens become less virulent over time
    p.Traits.Virulence *= 0.9999  // Slow decrease
}
```

---

## Phase 5: Extinction Cascades (Week 12)

### 5.1 Cascade Effects
**Files**: `internal/ecosystem/population/cascades.go` (NEW)

```go
func (ps *PopulationSimulator) ApplyExtinctionCascade(extinctSpecies *Organism) []ExtinctionEvent {
    cascadeEvents := []ExtinctionEvent{}
    
    // 1. Symbiotic partners (co-extinction)
    // 2. Predators (food source loss)
    // 3. Prey (predator release → population explosion)
    // 4. Keystone species check
    
    return cascadeEvents
}
```

---

## Phase 6: Sapience & Turning Points (Weeks 13-14)

### 6.1 Sapience Detection
**Files**: `internal/ecosystem/sapience/detection.go` (NEW)

```go
type SapienceThreshold struct {
    MinIntelligence   float64  // 0.7
    MinSocial         float64  // 0.6
    RequiresToolUse   bool     // true
    RequiresComm      bool     // true
}

func (ps *PopulationSimulator) CheckProtoSapience() []*Organism {
    candidates := []*Organism{}
    for _, pop := range ps.Populations {
        if pop.Traits.Intelligence > 0.7 &&
           pop.Traits.Social > 0.6 &&
           pop.Traits.ToolUse > 0.3 &&
           pop.Traits.Communication > 0.3 {
            candidates = append(candidates, pop.Species)
        }
    }
    return candidates
}
```

### 6.2 Turning Points
**Files**: `internal/ecosystem/turning_point.go` (NEW)

```go
type TurningPoint struct {
    ID        uuid.UUID
    WorldID   uuid.UUID
    Year      int64
    Trigger   TriggerType  // Interval, Event, PlayerRequest
    Options   []TurningPointOption
    Chosen    *TurningPointOption
}

type TriggerType string
const (
    TriggerInterval TriggerType = "interval"  // Every 1M years
    TriggerEvent    TriggerType = "event"     // After mass extinction
    TriggerPlayer   TriggerType = "player"    // Player requested
)
```

---

## Phase 7: Background Simulation & Player Visualization (Weeks 15-16)

### 7.1 Async Simulation Runner
**Files**: `internal/ecosystem/runner.go` (NEW)

```go
type SimulationRunner struct {
    WorldID       uuid.UUID
    Simulator     *PopulationSimulator
    Status        SimulationStatus
    TargetYear    int64
    CurrentYear   int64
    SnapshotFreq  int64  // Every 10 years for player visibility
    Logger        *SimulationLogger
    StopChan      chan struct{}
}

func (r *SimulationRunner) Run(ctx context.Context) {
    for r.CurrentYear < r.TargetYear {
        select {
        case <-ctx.Done():
            return
        case <-r.StopChan:
            return
        default:
            r.Simulator.SimulateYear()
            r.CurrentYear++
            
            if r.CurrentYear % r.SnapshotFreq == 0 {
                r.SaveSnapshot()
            }
            
            if r.CurrentYear % 1000000 == 0 {
                r.SaveCheckpoint()
            }
        }
    }
}
```

---

## Database Schema Changes

```sql
CREATE TABLE simulation_events (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    year BIGINT,
    event_type VARCHAR(50),
    severity FLOAT,
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE world_checkpoints (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    year BIGINT,
    checkpoint_type VARCHAR(20),
    state_data BYTEA,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Hex grid tectonic system
CREATE TABLE tectonic_plates (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    name VARCHAR(100),
    plate_type VARCHAR(20),  -- 'continental', 'oceanic'
    velocity_x FLOAT,
    velocity_y FLOAT,
    landmass_pct FLOAT
);

CREATE TABLE hex_cells (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    q INT,  -- Axial Q coordinate
    r INT,  -- Axial R coordinate
    plate_id UUID REFERENCES tectonic_plates(id),
    biome_id UUID,
    region_id UUID,
    elevation FLOAT,
    temperature FLOAT,
    moisture FLOAT,
    UNIQUE(world_id, q, r)
);
CREATE INDEX idx_hex_cells_coords ON hex_cells(world_id, q, r);

CREATE TABLE regions (
    id UUID PRIMARY KEY,
    biome_id UUID REFERENCES biomes(id),
    cells JSONB,  -- Array of {q, r} coordinates
    terrain_type VARCHAR(50),
    connections JSONB,
    last_updated BIGINT,
    isolation_years BIGINT  -- For gigantism/dwarfism calculation
);

CREATE TABLE pathogens (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    name VARCHAR(255),
    pathogen_type VARCHAR(50),
    traits JSONB,
    host_species_ids UUID[],
    origin_year BIGINT,
    status VARCHAR(20)
);

CREATE TABLE turning_points (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    year BIGINT,
    trigger_type VARCHAR(50),
    options JSONB,
    chosen_option_id UUID
);

CREATE TABLE sapient_species (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    species_id UUID,
    detection_year BIGINT,
    status VARCHAR(50),
    civilization_data JSONB
);

-- Extinct species with cause (for fossil loot flavor)
CREATE TABLE extinct_species (
    id UUID PRIMARY KEY,
    world_id UUID REFERENCES worlds(id),
    species_id UUID,
    name VARCHAR(255),
    traits JSONB,
    origin_year BIGINT,
    extinction_year BIGINT,
    extinction_cause VARCHAR(50),  -- 'starvation', 'disease', 'asteroid', etc.
    extinction_details TEXT,  -- 'Great Ash Winter of era 4B'
    peak_population BIGINT,
    region_ids UUID[]
);
```

---

## Testing Strategy

### Unit Tests (>80% coverage)
- [ ] Genetic distance calculations (with key gene weighting)
- [ ] Speciation threshold checks
- [ ] Pathogen spread mechanics
- [ ] Cascade effects
- [ ] Migration range calculations
- [ ] Inbreeding depression penalty
- [ ] Energy cost calculations
- [ ] Expression matrix phenotype derivation

### Integration Tests
- [ ] Full simulation run: 10M years
- [ ] Checkpoint save/restore
- [ ] Turning point triggering
- [ ] Hex grid tectonic plate movement

### Performance Benchmarks
- [ ] 1M years: < 1 minute
- [ ] 100M years: < 10 minutes
- [ ] 1B years: < 1 hour

### Scientific Validation
- [ ] Reproduction rate scaling: M^-0.25 (not M^-0.5)
- [ ] Megafauna don't go extinct too easily
- [ ] Carboniferous giant insects at 35% O2
- [ ] Punctuated equilibrium observable in history

---

## Technical Refinements Summary

These HIGH priority items must be addressed early:

| Item | Current | Fixed | Impact |
|------|---------|-------|--------|
| Genetic vectors | `float64` | `float32` | 50% memory savings |
| Reproduction rate | M^-0.5 | M^-0.25 | Megafauna survival |
| Continental config | Single float | Hex grid + plates | Spatial awareness |
| Trait changes | `++` operator | Proportional increase | Prevents scale breaking |
| Speciation distance | Raw Euclidean | Weighted by gene category | Realistic divergence |
| Small populations | No penalty | Inbreeding depression | Biological realism |
| Enhanced traits | No cost | Energy budget | Prevents power creep |

---

## Implementation Order

1. **Week 1**: Phase 0 (Logging & Checkpoints)
2. **Weeks 2-4**: Phase 1 (Unified Organism + Expression Matrix)
3. **Weeks 5-6**: Phase 2 (Hex Grid + Tectonic System + Regions)
4. **Weeks 7-9**: Phase 3 (Advanced Speciation with weighted distance)
5. **Weeks 10-11**: Phase 4 (Pathogen System)
6. **Week 12**: Phase 5 (Extinction Cascades + Fossil Causes)
7. **Weeks 13-14**: Phase 6 (Sapience + Magic Uplift + Turning Points)
8. **Weeks 15-16**: Phase 7 (Background Simulation + Player View)

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Performance bottleneck | High | High | Profile early, optimize hot paths, use float32 |
| Database growth | Medium | Medium | Aggressive pruning, compression |
| Feature creep | High | Medium | Stick to plan, defer nice-to-haves |
| Complex debugging | Medium | High | Comprehensive logging from start |
| Megafauna extinction | Medium | High | Fix reproduction rate to M^-0.25 |
| Hex grid complexity | Medium | Medium | Abstract behind Region interface |

---

## Out of Scope (Future Phases)

- Full civilization simulation
- NPC generation from species
- Player-controlled time manipulation
- Multiplayer timeline conflicts
- Advanced fossil discovery mechanics
- Full microbiome simulation

