# Weather Package

Dynamic weather simulation based on geography.

## Architecture

```
weather/
├── service.go        # Main service with update loop
├── updates.go        # Weather state transitions
├── precipitation.go  # Rain/snow calculation
├── evaporation.go    # Evaporation from water bodies
├── wind.go           # Wind pattern simulation
├── extremes.go       # Extreme weather events
├── climate.go        # Climate zone classification
├── states.go         # Weather state machine
├── repository.go     # Persistence layer
└── types.go          # Core types (WeatherState, Season)
```

---

## Weather States

| State | Conditions |
|-------|------------|
| `Clear` | Low humidity, no precipitation |
| `Cloudy` | Medium humidity, no precipitation |
| `Rain` | High humidity, liquid precipitation |
| `Storm` | Very high humidity, high winds |
| `Snow` | Cold temperature, frozen precipitation |

---

## Key Functions

| Function | Description |
|----------|-------------|
| `UpdateWorldWeather()` | Updates all cells in a world |
| `GetCurrentWeather()` | Retrieves weather for a cell |
| `ForceWorldWeather()` | God-mode weather override |
| `CalculateEvaporation()` | Water → atmosphere |
| `SimulateWind()` | Wind patterns by latitude |

---

## Integration

Weather is integrated with `TickerManager` for real-time updates:

```go
// In TickerManager.tick() - every 30 game minutes
emotes, err := tm.weatherService.UpdateWorldWeather(ctx, worldID, calcTime, season)
```

---

## Usage

```go
svc := weather.NewService(repo)
svc.InitializeWorldWeather(ctx, worldID, states, cells)
emotes, _ := svc.UpdateWorldWeather(ctx, worldID, time.Now(), weather.SeasonSummer)
```

---

## Testing

```bash
go test -v ./internal/worldgen/weather/...
go test -cover ./internal/worldgen/weather/...
```
