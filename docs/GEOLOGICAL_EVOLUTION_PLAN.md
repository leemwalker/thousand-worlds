# Implementation Plan - Geological Terrain Evolution

## Goal
Integrate existing worldgen geology code into `world simulate` to evolve terrain over time, including tectonic plates, volcanism, erosion, and sea level changes.

## Existing Code to Reuse
| File | Functions | Purpose |
|------|-----------|---------|
| `tectonics.go` | `GeneratePlates`, `SimulateTectonics` | Create plates, calculate boundaries |
| `heightmap.go` | `GenerateHeightmap` | Full terrain generation |
| `erosion.go` | `ApplyThermalErosion`, `ApplyHydraulicErosion` | Sculpt terrain |
| `volcanism.go` | `ApplyHotspots`, `applyVolcano` | Add volcanic features |

## Additional Geological Events (Balanced)
Current events + 2 new ones for realism without granularity:
1. **Volcanic Winter** ✓ (temp/sunlight) - Add: volcanic mountains
2. **Asteroid Impact** ✓ (temp/sunlight) - Add: crater formation
3. **Ice Age** ✓ (temp) - Add: sea level drop, glacier erosion
4. **Ocean Anoxia** ✓ (temp/oxygen) - Keep as is (no terrain effect)
5. **[NEW] Continental Drift** - Move plates over time → new mountains
6. **[NEW] Flood Basalt** - Large volcanic provinces

## Proposed Changes

### 1. Create `WorldGeology` service (`internal/ecosystem/geology.go`)
- [ ] Store heightmap, plates, sea level per world
- [ ] `InitializeGeology(worldID, seed, circumference)` - create baseline
- [ ] `SimulateGeology(years)` - advance plates, apply erosion
- [ ] `ApplyEvent(event)` - terrain changes from events

### 2. Wire to `handleWorldSimulate`
- [ ] Initialize geology if not exists
- [ ] Call `SimulateGeology` periodically (every 10k years simulated)
- [ ] Apply terrain effects from geological events

### 3. Scale Considerations
- Plate movement: ~2cm/year → 20km per 1M years
- Mountain building: thousands of years
- Erosion: tens of thousands of years
- Sea level: ±100m over ice age cycles

## Verification
- `world info` shows terrain stats (avg elevation, sea level)
- `ecosystem status` shows geographic distribution of life
