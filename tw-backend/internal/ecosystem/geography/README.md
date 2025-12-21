# Ecosystem Geography Package

Tectonic plate and region simulation for population dynamics.

## Architecture

```
ecosystem/geography/
├── tectonics.go   # TectonicSystem - plate movement and boundaries
├── hex_grid.go    # HexGrid - hexagonal world grid
└── region.go      # RegionSystem - continent/island tracking
```

---

## Key Types

| Type | Description |
|------|-------------|
| `TectonicSystem` | Manages plates, boundaries, fragmentation |
| `TectonicPlate` | Individual plate with movement vector |
| `PlateBoundary` | Divergent/convergent/transform boundaries |
| `HexGrid` | Hexagonal grid for spatial queries |
| `RegionSystem` | Tracks landmasses and isolation |

---

## Key Functions

| Function | Description |
|----------|-------------|
| `NewTectonicSystem()` | Creates system with initial plates |
| `TectonicSystem.Update()` | Advances plates by years |
| `CalculateFragmentation()` | Continental fragmentation (0-1) |
| `IsBoundaryCell()` | Check if cell on plate boundary |
| `GetBoundaryActivity()` | Tectonic activity level (0-1) |

---

## Usage

```go
ts := geography.NewTectonicSystem(worldID, seed)
ts.Update(10000) // Advance 10,000 years
frag := ts.CalculateFragmentation()
```

---

## Testing

```bash
go test -v ./internal/ecosystem/geography/...
```
