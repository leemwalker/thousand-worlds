package world

// TickBroadcast represents the payload broadcast on each tick
type TickBroadcast struct {
	WorldID        string  `json:"worldId"`
	TickNumber     int64   `json:"tickNumber"`
	GameTimeMs     int64   `json:"gameTimeMs"`
	RealTimeMs     int64   `json:"realTimeMs"`
	DilationFactor float64 `json:"dilationFactor"`
}
