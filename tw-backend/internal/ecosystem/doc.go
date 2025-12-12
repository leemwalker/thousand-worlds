// Package ecosystem implements wildlife simulation including creature spawning,
// evolution, geological events, and pathfinding for the Thousand Worlds game.
//
// # Core Components
//
//   - Service: Main ecosystem manager that coordinates all subsystems
//   - Spawner: Creates creatures appropriate for each biome
//   - EvolutionManager: Handles creature reproduction and genetic inheritance
//   - WorldGeology: Simulates geological events over time (volcanoes, earthquakes)
//   - FindPath: A* pathfinding for entity movement
//
// # Usage
//
//	sim := ecosystem.NewService(seed)
//	sim.SpawnBiomes(biomes)
//	sim.Tick() // Called every game tick
//
// # Subsystems
//
// The ecosystem package works closely with:
//   - internal/ecosystem/state: Entity state management
//   - internal/npc/genetics: DNA and trait inheritance
//   - internal/worldgen/geography: Biome definitions
package ecosystem
