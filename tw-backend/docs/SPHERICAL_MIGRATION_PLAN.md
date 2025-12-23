# Spherical World Migration Plan

**Status:** Prototype Complete | **Target:** Full Spherical Planet Simulation

---

## Executive Summary

The `CubeSphereTopology` prototype in `internal/spatial/` proves the cube-sphere approach is viable. This document assesses migration effort for the world generation engine.

---

## Current State Audit

### Tectonics (`worldgen/geography/tectonics.go`)

| Component | Current | Migration Effort |
|-----------|---------|------------------|
| `TectonicPlate.Centroid` | `Point{X, Y float64}` | **Medium** - Convert to `Coordinate{Face, X, Y}` |
| `GeneratePlates()` | Random X/Y in `[0, Width)` | **Medium** - Generate across 6 faces |
| `SimulateTectonics()` | Voronoi distance on flat grid | **High** - Use `Topology.Distance()` for spherical Voronoi |
| Boundary detection | 4-neighbor grid check | **Medium** - Use `Topology.GetNeighbor()` |

**Key Challenge:** Plate movement vectors need 3D representation via `ToSphere`/`FromVector`.

---

### Weather (`worldgen/weather/`)

| Component | Current | Migration Effort |
|-----------|---------|------------------|
| `CalculateWind()` | Uses latitude (float64) | **Low** - Derive latitude from `ToSphere().Y` |
| `SimulateAdvection()` | `x + 1` with modulo wrap | **Medium** - Use `Topology.GetNeighbor(East)` |
| `GeographyCell` | Indexed by UUID | **Low** - Add `Coordinate` field |

**Key Benefit:** Wind can now cross poles naturally via face transitions.

---

### Heightmap (`worldgen/geography/types.go`)

| Component | Current | Migration Effort |
|-----------|---------|------------------|
| `Heightmap` struct | `[]float64` size `Width*Height` | **High** - 6 faces × `Resolution²` |
| `Get(x, y)` | Direct array index | **Medium** - `Get(Coordinate)` |
| `Set(x, y, val)` | Direct array index | **Medium** - `Set(Coordinate, val)` |

**Recommendation:** Create `SphereHeightmap` wrapper using `Topology`.

---

## Approach Comparison

| Criterion | Cube Sphere | Voronoi/Icosphere |
|-----------|-------------|-------------------|
| Code reuse | **90%** - Arrays preserved | 20% - Full graph rewrite |
| Distortion | Low (normalized mapping) | None (geodesic) |
| Implementation | **Done** (prototype) | 2-3 weeks |
| Weather simulation | Simple (grid-based) | Complex (graph traversal) |
| Rendering | Standard UV mapping | Specialized shaders |

**Recommendation:** Proceed with **Cube Sphere**. Voronoi only if distortion becomes problematic.

---

## Migration Strategy

### Phase 1: Parallel Data Structures (Low Risk)
1. Add `Coordinate` field to `GeographyCell`
2. Create `SphereHeightmap` as adapter over existing `Heightmap`
3. Keep flat grid as fallback

### Phase 2: Weather Migration (Medium Risk)
1. Inject `Topology` into weather `Service`
2. Replace `SimulateAdvection` to use `GetNeighbor`
3. Test polar storm crossing

### Phase 3: Tectonics Migration (High Risk)
1. Convert `TectonicPlate` to use 3D vectors
2. Update boundary detection for spherical Voronoi
3. Validate mountain formation at poles

---

## Estimated Effort

| Phase | Effort | Risk |
|-------|--------|------|
| Phase 1 | 2-3 days | Low |
| Phase 2 | 3-5 days | Medium |
| Phase 3 | 1-2 weeks | High |
| **Total** | **2-3 weeks** | |

---

## Files Created

- [topology.go](file:///Users/walker/git/thousand-worlds/tw-backend/internal/spatial/topology.go) - Interface + types
- [cube_sphere.go](file:///Users/walker/git/thousand-worlds/tw-backend/internal/spatial/cube_sphere.go) - Implementation
- [topology_test.go](file:///Users/walker/git/thousand-worlds/tw-backend/internal/spatial/topology_test.go) - Contract tests
- [cube_sphere_test.go](file:///Users/walker/git/thousand-worlds/tw-backend/internal/spatial/cube_sphere_test.go) - 88.8% coverage
