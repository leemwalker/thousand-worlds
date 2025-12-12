# NPC Package

The `internal/npc` package implements the NPC (Non-Player Character) system for Thousand Worlds. NPCs are autonomous agents with genetic traits, memories, relationships, and personality-driven behavior.

## Architecture

```
npc/
├── appearance/      # Physical appearance generation from DNA
├── desire/          # Need-based decision engine (survival > social > achievement > pleasure)
├── emotion/         # Emotional state and mood system
├── genetics/        # Mendelian inheritance and trait mutations
├── interaction/     # NPC-to-NPC and NPC-to-player interactions
├── memory/          # MongoDB-backed memory with decay/rehearsal
├── personality/     # Personality traits derived from genetics + experience
└── relationship/    # Affection, trust, fear tracking between entities
```

---

## Subsystems

### genetics/
Implements Mendelian inheritance with dominant/recessive alleles.

**Key Files:**
| File | Description |
|------|-------------|
| `types.go` | DNA struct, Gene struct with alleles |
| `traits.go` | Maps genes to attribute bonuses (SS: +10, Ss: +5, ss: +0) |
| `inheritance.go` | Parent → child gene inheritance (Punnett square) |
| `mutation.go` | 5% mutation chance per gene |
| `diversity.go` | Initial population genetic variation |
| `appearance.go` | Physical traits from DNA (height, build, features) |

**Example:**
```go
child := genetics.Inherit(parent1.DNA, parent2.DNA)
mods := genetics.CalculateAttributeBonuses(child)
// mods.Attributes.Might = sum of Strength + Muscle gene bonuses
```

---

### memory/
MongoDB-backed NPC memory system with realistic decay and recall.

**Key Files:**
| File | Description |
|------|-------------|
| `types.go` | Memory struct (observation, conversation, event, relationship) |
| `repository.go` | Memory storage interface |
| `mongo_repository.go` | MongoDB implementation |
| `decay.go` | Linear decay over time (clarity degrades) |
| `rehearsal.go` | Accessed memories decay slower |
| `corruption.go` | Fuzzy/altered details on low-clarity recall |
| `emotional_scoring.go` | High-emotion memories more persistent (0.0-1.0) |
| `relevance.go` | Context-aware memory retrieval |
| `retention.go` | `retention = baseRetention * (1 - decayRate * time) * (1 + rehearsalBonus * accessCount)` |
| `tagging.go` | Entity, location, emotion, keyword tagging |
| `jobs.go` | Background decay/cleanup jobs |

---

### desire/
Rule-based AI for NPC decision-making following a need hierarchy.

**Priority (highest to lowest):**
1. **Survival** - hunger, thirst, sleep, safety (0-100 scales)
2. **Social** - companionship, conversation, affection
3. **Achievement** - tasks, objectives, long-term plans
4. **Pleasure/Exploration** - curiosity, hedonism

```go
priority = needUrgency * personalityWeight
```

---

### relationship/
Tracks NPC-to-NPC and NPC-to-player relationships.

**Tracked Metrics:**
- `affection` (-100 to +100) - Updated by gifts, help, kindness
- `trust` (-100 to +100) - Updated by honesty, reliability
- `fear` (-100 to +100) - Updated by threats, violence

---

### personality/
Derives personality from genetics + life experiences using Big Five model:
- Extraversion
- Conscientiousness
- Openness
- Agreeableness
- Neuroticism

---

### emotion/
Transient emotional states affecting behavior and dialogue:
- joy, anger, fear, sadness, surprise, disgust
- Mood system (temporary emotional state)

---

### interaction/
Handles NPC-to-NPC and NPC-to-player interactions:
- Conversation initiation (proximity, relationships, needs)
- Dialogue flow (greeting, topic selection, response)
- Relationship updates from conversations
- Memory formation during interactions

---

### appearance/
Generates physical descriptions from genetic traits:
- Height (tall/short)
- Build (muscular/lean)
- Features (hair color, eye color)
- Age-based changes (child → adult → elder)

---

## Integration Points

- **World Simulation** - NPCs act on world tick events
- **Combat System** - NPC behavior in combat from `internal/combat`
- **LLM Dialogue** - Personality/mood passed to `internal/ai` for natural dialogue
- **Event Store** - NPC actions logged via `internal/eventstore`

## Testing

```bash
# Run all NPC tests
go test ./internal/npc/...

# With coverage
go test -cover ./internal/npc/...
```
