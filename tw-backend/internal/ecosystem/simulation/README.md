# Ecosystem Simulation Package

**Package**: `internal/ecosystem/simulation/`

## Overview

The simulation engine is the heart of the "live" world, responsible for time-advanced progression of geological and biological systems.

## Core Components

### Step Orchestration (`step.go`)
The `Step()` function is the canonical entry point for advancing a world by a given amount of time. It handles:
- **Sub-stepping**: Automagically breaking down large time jumps into stable simulation steps.
- **Dependency Ordering**: Ensuring geology evolves before biological populations react.
- **Parallel Execution**: Processing independent cells/regions in parallel.

### Auto-Resolver (`auto_resolver.go`)
Handles conflict resolution and state transitions between simulation ticks:
- Resource competition outcome.
- Population migration patterns.
- Disaster impact application.

## Constants (`constants.go`)
Defines simulation boundaries, tick durations, and stability thresholds.

## Usage

```go
import "tw-backend/internal/ecosystem/simulation"

// Advance world by 1,000 years
result, err := simulation.Step(ctx, worldID, 1000)
```

## Testing

```bash
go test ./internal/ecosystem/simulation/...
```
