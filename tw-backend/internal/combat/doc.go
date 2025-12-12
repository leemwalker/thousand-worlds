// Package combat implements the turn-based combat system for Thousand Worlds.
//
// # Subsystems
//
// The combat package is organized into three subsystems:
//
//   - action: Combat queue, turn management, action validation
//   - damage: Damage calculation, critical hits, durability
//   - effects: Status effects (buffs, debuffs, DoTs)
//
// # Combat Flow
//
//  1. Actions are queued via action.NewCombatQueue()
//  2. Actions are validated against stamina, cooldowns, stun status
//  3. Reaction times are calculated based on agility and action type
//  4. Damage is calculated with type resistances and critical multipliers
//  5. Effects are applied (poison, stun, buffs)
//  6. Equipment durability is reduced
//
// # Usage
//
//	queue := action.NewCombatQueue()
//	queue.Push(action.NewCombatAction(attacker, target, AttackLight, reactionTime))
//	resolver := action.NewCombatResolver()
//	resolver.Resolve(queue)
package combat
