package constants

import "github.com/google/uuid"

// LobbyWorldID is the reserved UUID for the Lobby world
var LobbyWorldID = uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000000"))

// IsLobby checks if the given world ID is the Lobby
func IsLobby(worldID uuid.UUID) bool {
	return worldID == LobbyWorldID
}
