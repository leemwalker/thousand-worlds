package world

// TickBroadcast represents the payload broadcast on each tick
type TickBroadcast struct {
	WorldID        string  `json:"worldId"`
	TickNumber     int64   `json:"tickNumber"`
	GameTimeMs     int64   `json:"gameTimeMs"`
	RealTimeMs     int64   `json:"realTimeMs"`
	DilationFactor float64 `json:"dilationFactor"`
	TimeOfDay      string  `json:"timeOfDay"`
	SunPosition    float64 `json:"sunPosition"`
	CurrentSeason  string  `json:"currentSeason"`
	SeasonProgress float64 `json:"seasonProgress"`
}
