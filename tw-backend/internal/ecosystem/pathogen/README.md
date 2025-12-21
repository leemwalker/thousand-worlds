# Pathogen Package

Disease simulation and outbreak management.

## Architecture

```
pathogen/
├── simulation.go  # DiseaseSystem - outbreak management
└── types.go       # Pathogen, Outbreak, PathogenType
```

---

## Key Types

| Type | Description |
|------|-------------|
| `DiseaseSystem` | Manages all pathogens and outbreaks |
| `Pathogen` | Disease characteristics (R0, mortality) |
| `Outbreak` | Active disease spread |

---

## Pathogen Types

| Type | Transmission | Example |
|------|--------------|---------|
| `Bacterial` | Contact/Airborne | Plague |
| `Viral` | High spread | Influenza |
| `Parasitic` | Vector-borne | Malaria |
| `Fungal` | Spores | Cordyceps |

---

## Key Functions

| Function | Description |
|----------|-------------|
| `NewDiseaseSystem()` | Creates disease manager |
| `CheckSpontaneousOutbreak()` | Random outbreak based on density |
| `CheckZoonoticTransfer()` | Cross-species transmission |
| `Update()` | Advance all outbreaks by year |
| `GetImpact()` | Deaths/infections for species |

---

## Usage

```go
ds := pathogen.NewDiseaseSystem(worldID, seed)
pathogen, outbreak := ds.CheckSpontaneousOutbreak(speciesID, name, pop, density)
ds.Update(year, speciesData)
```

---

## Testing

```bash
go test -v ./internal/ecosystem/pathogen/...
```
