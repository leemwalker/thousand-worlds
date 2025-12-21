# Geography Package

Procedural terrain generation through geological simulation.

## Architecture

```
geography/
├── tectonics.go   # Tectonic plate simulation
├── volcanism.go   # Volcanic activity (hotspots, eruptions)
├── heightmap.go   # Elevation grid generation
├── erosion.go     # Hydraulic and thermal erosion
├── rivers.go      # River generation via A* pathfinding
├── biomes.go      # Biome classification (Whittaker)
├── ocean.go       # Sea level and ocean placement
├── shapes.go      # World shape handling
├── noise.go       # Perlin noise utilities
├── seismology.go  # Earthquake simulation
├── crust.go       # Crustal composition
└── types.go       # Core types (TectonicPlate, Biome, Heightmap)
```

---

## Key Functions

### Tectonics (`tectonics.go`)

| Function | Description |
|----------|-------------|
| `GeneratePlates()` | Creates tectonic plates with random centroids |
| `SimulateTectonics()` | Calculates elevation from plate interactions |
| `SimulateWilsonCycle()` | Determines phase (Rifting/Spreading/Subduction/Orogeny) |
| `SimulateContinentalRift()` | Rift formation and volcanic activity |

### Volcanism (`volcanism.go`)

| Function | Description |
|----------|-------------|
| `ApplyHotspots()` | Creates volcanic chains over moving plates |
| `ApplyVolcano()` | Adds volcanic cone to heightmap |
| `GetEruptionStyle()` | Determines behavior by magma type |
| `SimulateFloodBasalt()` | Large Igneous Province impacts |

### Erosion (`erosion.go`)

| Function | Description |
|----------|-------------|
| `ApplyHydraulicErosion()` | Rain/water flow to carve valleys |
| `ApplyThermalErosion()` | Slope stability and material transfer |

### Biomes (`biomes.go`)

| Function | Description |
|----------|-------------|
| `AssignBiomes()` | Whittaker classification by temp/moisture |
| `resolveBiome()` | Maps temperature + moisture to biome type |

---

## Biome Types

| Type | Conditions |
|------|------------|
| `Tundra` | Cold (<-5°C), low moisture |
| `Taiga` | Cold, high moisture |
| `Grassland` | Temperate, medium moisture |
| `DeciduousForest` | Temperate, high moisture |
| `Rainforest` | Hot (>20°C), very high moisture |
| `Desert` | Hot, low moisture |
| `Alpine` | High elevation or frozen mountain |
| `Ocean` | Below sea level |

---

## Usage

```go
// 1. Generate plates
plates := geography.GeneratePlates(10, 100, 100, seed)

// 2. Simulate tectonics
hm := geography.SimulateTectonics(plates, 100, 100)

// 3. Apply erosion
geography.ApplyHydraulicErosion(hm, 5000, seed)

// 4. Assign biomes
biomes := geography.AssignBiomes(hm, seaLevel, seed, 0.0)
```

---

## Testing

```bash
go test -v ./internal/worldgen/geography/...
go test -cover ./internal/worldgen/geography/...
```
