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

	// 4. Calculate New Position
	newX, newY, message, err := s.CalculateNewPosition(char, world, dx, dy)
	if err != nil {
		return message, nil // Return user-facing restriction message
	}

	// 5. Update Character
	char.PositionX = newX
	char.PositionY = newY
	// Update orientation to match movement direction unless strafing (not implemented)
	char.OrientationX = dx
	char.OrientationY = dy
	char.OrientationZ = 0 // Flat movement

	if err := s.authRepo.UpdateCharacter(ctx, char); err != nil {
		return "", fmt.Errorf("failed to update character position: %w", err)
	}

	return fmt.Sprintf("You move %s. %s", dirName, message), nil
}

// CalculateNewPosition calculates the new position based on a delta and world rules
func (s *SpatialService) CalculateNewPosition(char *auth.Character, world *repository.World, dx, dy float64) (float64, float64, string, error) {
	isLobby := world.Name == "Lobby" || strings.Contains(strings.ToLower(world.Name), "lobby")

	if isLobby {
		// Lobby Movement (Cartesian with walls)
		newX, newY, err := calculateLobbyPosition(char.PositionX, char.PositionY, dx, dy)
		if err != nil {
			return char.PositionX, char.PositionY, err.Error(), fmt.Errorf("blocked")
		}
		return newX, newY, "", nil
	}

	// Spherical Movement (Wrap around)
	circumference := 10000.0
	if world.Circumference != nil && *world.Circumference > 0 {
		circumference = *world.Circumference
	}
	dims := worldspatial.NewWorldDimensions(circumference)
	newX, newY, message := calculateSphericalPosition(char.PositionX, char.PositionY, dx, dy, "", dims) // Empty dirName as we formulate message caller-side or here

	// Strip "You move..." from the message if it exists, as we are reusing logic
	// The original calculateSphericalPosition returned "You move [dir]. [Extra]".
	// We should refactor strictly, but for now let's just use the logic.

	return newX, newY, message, nil
}

// GetOrientationVector returns the x, y, z vector for a named direction
func (s *SpatialService) GetOrientationVector(direction string) (float64, float64, float64, string) {
	direction = strings.ToLower(strings.TrimSpace(direction))
	switch direction {
	case "n", "north":
		return 0, 1, 0, "North"
	case "s", "south":
		return 0, -1, 0, "South"
	case "e", "east":
		return 1, 0, 0, "East"
	case "w", "west":
		return -1, 0, 0, "West"
	case "ne", "northeast":
		return 0.707, 0.707, 0, "Northeast"
	case "nw", "northwest":
		return -0.707, 0.707, 0, "Northwest"
	case "se", "southeast":
		return 0.707, -0.707, 0, "Southeast"
	case "sw", "southwest":
		return -0.707, -0.707, 0, "Southwest"
	case "u", "up":
		return 0, 0, 1, "Up"
	case "d", "down":
		return 0, 0, -1, "Down"
	default:
		return 0, 0, 0, ""
	}
}

// GetDirectionName from vector (approximate)
func (s *SpatialService) GetDirectionName(x, y, z float64) string {
	if z > 0.5 {
		return "Up"
	}
	if z < -0.5 {
		return "Down"
	}

	// Normalize 2D
	mag := math.Sqrt(x*x + y*y)
	if mag < 0.1 {
		return "Unknown"
	}

	nx, ny := x/mag, y/mag

	// Dot products with cardinals
	if ny > 0.9 {
		return "North"
	}
	if ny < -0.9 {
		return "South"
	}
	if nx > 0.9 {
		return "East"
	}
	if nx < -0.9 {
		return "West"
	}

	if nx > 0 && ny > 0 {
		return "Northeast"
	}
	if nx < 0 && ny > 0 {
		return "Northwest"
	}
	if nx > 0 && ny < 0 {
		return "Southeast"
	}
	if nx < 0 && ny < 0 {
		return "Southwest"
	}

	return "Unknown"
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
	case "ne", "northeast":
		return 0.707, 0.707, "northeast"
	case "nw", "northwest":
		return -0.707, 0.707, "northwest"
	case "se", "southeast":
		return 0.707, -0.707, "southeast"
	case "sw", "southwest":
		return -0.707, -0.707, "southwest"
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
