package player

import (
	"context"
	"fmt"
	"math"
	"strings"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/spatial"
	worldspatial "mud-platform-backend/internal/world/spatial"

	"github.com/google/uuid"
)

const (
	LobbyWorldName        = "Lobby"
	LobbyLengthNorthSouth = 1000.0
	LobbyWidthEastWest    = 10.0
	LobbyCenterX          = 5.0
	LobbyCenterY          = 500.0
)

// SpatialService handles character movement and spatial logic
type SpatialService struct {
	authRepo  auth.Repository
	worldRepo repository.WorldRepository
}

// NewSpatialService creates a new SpatialService
func NewSpatialService(authRepo auth.Repository, worldRepo repository.WorldRepository) *SpatialService {
	return &SpatialService{
		authRepo:  authRepo,
		worldRepo: worldRepo,
	}
}

// HandleMovementCommand processes a movement command (n, s, e, w, etc.)
func (s *SpatialService) HandleMovementCommand(ctx context.Context, charID uuid.UUID, direction string) (string, error) {
	// 1. Get Character
	char, err := s.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return "", fmt.Errorf("failed to get character: %w", err)
	}

	// 2. Get World
	world, err := s.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return "", fmt.Errorf("failed to get world: %w", err)
	}

	// 3. Parse Direction
	dx, dy, dirName := parseDirection(direction)
	if dirName == "" {
		return "Invalid direction. Use north, south, east, or west.", nil
	}

	// 4. Calculate New Position based on World Type
	newX, newY := char.PositionX, char.PositionY
	message := fmt.Sprintf("You move %s.", dirName)

	// Check if Lobby
	// Assuming Lobby has a specific ID or Name or Shape?
	// Prompt says "Lobby has special world_id" in example struct, but in DB likely just a world.
	// We can check name or a hardcoded ID if we had one.
	// Or check world.Name == "Lobby"
	isLobby := world.Name == "Lobby" || strings.Contains(strings.ToLower(world.Name), "lobby")

	if isLobby {
		// Lobby Movement (Cartesian with walls)
		newX, newY, err = calculateLobbyPosition(char.PositionX, char.PositionY, dx, dy)
		if err != nil {
			return err.Error(), nil // Return user-facing error message
		}
	} else {
		// Spherical Movement (Wrap around)
		// Default circumference if not set
		circumference := 10000.0
		if world.Circumference != nil && *world.Circumference > 0 {
			circumference = *world.Circumference
		}
		dims := worldspatial.NewWorldDimensions(circumference)
		newX, newY, message = calculateSphericalPosition(char.PositionX, char.PositionY, dx, dy, dirName, dims)
	}

	// 5. Update Character (TODO: Consume Stamina via Movement.Move if needed, assuming simple move for now)
	char.PositionX = newX
	char.PositionY = newY
	// Keep Z same for now
	// char.PositionZ = char.PositionZ

	if err := s.authRepo.UpdateCharacter(ctx, char); err != nil {
		return "", fmt.Errorf("failed to update character position: %w", err)
	}

	// 6. Return Message
	return message, nil
}

// CalculateDistance calculates the distance between two points on a sphere
func (s *SpatialService) CalculateDistance(lat1, lon1, lat2, lon2, radius float64) float64 {
	return spatial.GreatCircleDistance(lat1, lon1, lat2, lon2, radius)
}

func parseDirection(input string) (dx, dy float64, name string) {
	input = strings.ToLower(strings.TrimSpace(input))
	switch input {
	case "n", "north":
		return 0, 1, "north"
	case "s", "south":
		return 0, -1, "south"
	case "e", "east":
		return 1, 0, "east"
	case "w", "west":
		return -1, 0, "west"
	default:
		return 0, 0, ""
	}
}

func calculateLobbyPosition(x, y, dx, dy float64) (float64, float64, error) {
	newX := x + dx
	newY := y + dy

	// Lobby Boundaries: 0-10 (x), 0-1000 (y)
	// Check Walls
	if newX < 0 || newX > LobbyWidthEastWest {
		wall := "western"
		if newX > LobbyWidthEastWest {
			wall = "eastern"
		}
		return x, y, fmt.Errorf("You cannot go further %s. The %s wall blocks your way.", getDirectionName(dx, 0), wall)
	}

	if newY < 0 || newY > LobbyLengthNorthSouth {
		end := "southern"
		if newY > LobbyLengthNorthSouth {
			end = "northern"
		}
		return x, y, fmt.Errorf("You cannot go further %s. You have reached the %s end of the hallway.", getDirectionName(0, dy), end)
	}

	return newX, newY, nil
}

func getDirectionName(dx, dy float64) string {
	if dy > 0 {
		return "north"
	}
	if dy < 0 {
		return "south"
	}
	if dx > 0 {
		return "east"
	}
	if dx < 0 {
		return "west"
	}
	return ""
}

func calculateSphericalPosition(lon, lat, dx, dy float64, dirName string, dims worldspatial.WorldDimensions) (float64, float64, string) {
	message := fmt.Sprintf("You move %s.", dirName)

	// Convert 1 meter to degrees
	// dx, dy are in meters (1 unit = 1 meter per command)

	// Latitude change (y)
	// 1 degree lat = Circumference / 360
	deltaLat := dy / dims.MetersPerDegreeY
	rawLat := lat + deltaLat

	// Longitude change (x) depends on latitude
	// radius at lat = radius_equator * cos(lat)
	// circumference at lat = circumference_equator * cos(lat)
	// meters per degree lon = circumference_at_lat / 360

	// Use destination latitude for conservation of angular momentum-ish behavior,
	// or start latitude. Let's use start latitude to avoid infinite recursion if we were to solve it perfectly.
	// However, if we cross the pole, longitude flips.
	// Let's calculate raw New Lon first based on current lat.

	cosLat := math.Cos(lat * math.Pi / 180.0)
	if math.Abs(cosLat) < 0.0001 {
		// At pole, can't move east/west effectively.
		// Allow movement but effectively 0 change or very small.
		cosLat = 0.0001
	}

	metersPerDegreeLon := dims.MetersPerDegreeX * cosLat
	deltaLon := dx / metersPerDegreeLon
	rawLon := lon + deltaLon

	// Normalize coordinates using shared logic
	newLat, newLon := spatial.NormalizeCoordinates(rawLat, rawLon)

	// Detect events
	// 1. Pole Crossing: Latitude passed 90/-90 (rawLat vs newLat check is tricky because of the flip)
	// Easier check: if we were not at pole, and now longitude shifted by ~180 without us moving East/West?
	// Or check rawLat.
	if rawLat > 90 || rawLat < -90 {
		message += " You cross the pole and the world spins beneath you."
	}

	// 2. Circumnavigation (Date Line)
	// If longitude wrapped.
	// Check difference between newLon and rawLon (normalized to same phase)
	// Or just check if rawLon was outside -180, 180
	if rawLon > 180 || rawLon <= -180 {
		message += " The landscape seems familiar - you've circled back around the world."
	}

	return newLon, newLat, message
}
