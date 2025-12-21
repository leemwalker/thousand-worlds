# Sapience Package

Detection and tracking of sapient species emergence.

## Architecture

```
sapience/
├── detection.go  # SapienceDetector - monitors for emergence
└── types.go      # SapienceLevel, SapienceCandidate
```

---

## Sapience Levels

| Level | Description |
|-------|-------------|
| `None` | No significant cognition |
| `ProtoSapient` | Early tool use, basic communication |
| `Sapient` | Full sapience - language, culture |
| `Advanced` | Advanced technology/magic |

---

## Detection Thresholds

| Path | Intelligence | Social | Other |
|------|--------------|--------|-------|
| Standard | 80+ | 70+ | Tool use 60+ |
| Magic-Assisted | 50+ | 50+ | Magic affinity 80+ |

---

## Key Functions

| Function | Description |
|----------|-------------|
| `NewSapienceDetector()` | Creates detector |
| `Evaluate()` | Check species for sapience |
| `GetCandidates()` | All proto/sapient species |
| `HasAnySapience()` | Check if any sapient exists |
| `CalculateSapienceProgress()` | World progress (0-1) |
| `PredictSapienceYear()` | Estimate emergence year |

---

## Usage

```go
detector := sapience.NewSapienceDetector(worldID, magicEnabled)
candidate := detector.Evaluate(speciesID, name, traits, year)
progress := detector.CalculateSapienceProgress()
```

---

## Testing

```bash
go test -v ./internal/ecosystem/sapience/...
```
