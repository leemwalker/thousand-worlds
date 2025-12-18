package player

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"

	"tw-backend/internal/auth"
	"tw-backend/internal/repository"
	"tw-backend/internal/spatial"
	worldspatial "tw-backend/internal/world/spatial"
	"tw-backend/internal/worldentity"

	"github.com/google/uuid"
)

// Collider represents an obstacle in the world defined in metadata
type Collider struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Radius  float64 `json:"radius"`
	Message string  `json:"message"`
}

// SpatialService handles character movement and spatial logic
type SpatialService struct {
	authRepo      auth.Repository
	worldRepo     repository.WorldRepository
	entityService *worldentity.Service
}

// NewSpatialService creates a new SpatialService
func NewSpatialService(authRepo auth.Repository, worldRepo repository.WorldRepository, entityService *worldentity.Service) *SpatialService {
	return &SpatialService{
		authRepo:      authRepo,
		worldRepo:     worldRepo,
		entityService: entityService,
	}
}

func (s *SpatialService) HandleMovementCommand(ctx context.Context, charID uuid.UUID, direction string) (string, error) {
	return s.HandleMovementCommandWithDistance(ctx, charID, direction, 1.0)
}

// HandleLongDistanceMovement processes a movement command with a specific distance
func (s *SpatialService) HandleMovementCommandWithDistance(ctx context.Context, charID uuid.UUID, direction string, distance float64) (string, error) {
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

	// Scale by distance
	dx *= distance
	dy *= distance

	// 4. Calculate New Position
	newX, newY, message, err := s.CalculateNewPosition(ctx, char, world, dx, dy)
	if err != nil {
		return message, nil // Return user-facing restriction message
	}

	// 5. Update Character
	char.PositionX = newX
	char.PositionY = newY
	// Update orientation to match movement direction unless strafing (not implemented)
	// For orientation, we use the normalized direction (original dx, dy before scaling usually, but sign matches)
	if distance > 0 {
		// Normalize orientation vector if possible, though simple clamp works for NESW
		// dx/dy are already scaled.
		// Orientation expects unit vector?
		// Check GetDirectionName... it normalizes.
		// But let's just use the signs or original parseDirection values?
		// We can re-parse direction for orientation to be clean
		normDx, normDy, _ := parseDirection(direction)
		char.OrientationX = normDx
		char.OrientationY = normDy
	}
	char.OrientationZ = 0 // Flat movement

	if err := s.authRepo.UpdateCharacter(ctx, char); err != nil {
		return "", fmt.Errorf("failed to update character position: %w", err)
	}

	if distance > 1 {
		return fmt.Sprintf("You travel %.0f units %s to (%.0f, %.0f). %s", distance, dirName, newX, newY, message), nil
	}
	return fmt.Sprintf("You move %s. %s", dirName, message), nil
}

// CalculateNewPosition calculates the new position based on a delta and world rules
func (s *SpatialService) CalculateNewPosition(ctx context.Context, char *auth.Character, world *repository.World, dx, dy float64) (float64, float64, string, error) {
	// Debug: Log world shape and bounds
	log.Printf("[SPATIAL] World %s shape=%s BoundsMin=%v BoundsMax=%v", world.ID, world.Shape, world.BoundsMin, world.BoundsMax)

	// Check if world is bounded (Cube or has bounds defined)
	if world.Shape == repository.WorldShapeCube || (world.BoundsMin != nil && world.BoundsMax != nil) {
		newX, newY, err := calculateBoundedPosition(char.PositionX, char.PositionY, dx, dy, world)
		if err != nil {
			return char.PositionX, char.PositionY, err.Error(), fmt.Errorf("blocked")
		}

		// Check WorldEntity collisions (database-backed entities)
		if s.entityService != nil {
			blocked, entity, checkErr := s.entityService.CheckCollision(ctx, world.ID, newX, newY)
			if checkErr == nil && blocked {
				msg := fmt.Sprintf("The way is blocked by the %s.", entity.Name)
				return char.PositionX, char.PositionY, msg, fmt.Errorf("blocked")
			}
		}

		return newX, newY, "", nil
	}

	// Default to Spherical Movement (Wrap around)
	circumference := 10000.0
	if world.Circumference != nil && *world.Circumference > 0 {
		circumference = *world.Circumference
	}
	dims := worldspatial.NewWorldDimensions(circumference)
	newX, newY, message := calculateSphericalPosition(char.PositionX, char.PositionY, dx, dy, "", dims) // Empty dirName as we formulate message caller-side or here

	// Check WorldEntity collisions for spherical worlds too
	if s.entityService != nil {
		blocked, entity, checkErr := s.entityService.CheckCollision(ctx, world.ID, newX, newY)
		if checkErr == nil && blocked {
			msg := fmt.Sprintf("The way is blocked by the %s.", entity.Name)
			return char.PositionX, char.PositionY, msg, fmt.Errorf("blocked")
		}
	}

	return newX, newY, message, nil
}

// GetPortalLocation returns a deterministic location on the world perimeter for a given target world ID
func (s *SpatialService) GetPortalLocation(world *repository.World, targetID uuid.UUID) (float64, float64) {
	// Defaults if bounds are missing
	minX, minY := 0.0, 0.0
	maxX, maxY := 10.0, 10.0

	if world.BoundsMin != nil {
		minX, minY = world.BoundsMin.X, world.BoundsMin.Y
	}
	if world.BoundsMax != nil {
		maxX, maxY = world.BoundsMax.X, world.BoundsMax.Y
	}

	width := maxX - minX
	length := maxY - minY

	// Use uuid hash to determine wall (0-3) and offset (0-10 relative to size)
	hash := targetID.ID()

	// Walls: 0: South, 1: North, 2: West, 3: East
	wallIdx := int(hash % 4)

	// Calculate offset along the wall (0.1 to 0.9 range to avoid corners)
	// (hash >> 2) % 10 gives 0-9. normalize to meters.
	// Actually let's map it to specific "slots" to be cleaner.
	// 5 slots per wall?
	// or just modulo width/length

	val := float64((hash >> 2) % 10)

	// Normalize val to be within the wall length
	// Use (val + 0.5) to center in 1m blocks if we assume 10 slots?
	// If width is 10, val (0-9) maps directly.

	ratio := (val + 0.5) / 10.0 // 0.05 to 0.95

	switch wallIdx {
	case 0: // South (y=minY, x varies)
		return minX + (width * ratio), minY
	case 1: // North (y=maxY, x varies)
		return minX + (width * ratio), maxY
	case 2: // West (x=minX, y varies)
		return minX, minY + (length * ratio)
	case 3: // East (x=maxX, y varies)
		return maxX, minY + (length * ratio)
	}
	return minX, minY
}

// CheckPortalProximity checks if a character is close enough to enter a portal
// Returns true if allowed, or false with a hint message if not.
// If isLobby is true, allows entry from anywhere (global entry).
func (s *SpatialService) CheckPortalProximity(charX, charY, portalX, portalY float64, isLobby bool) (bool, string) {
	if isLobby {
		return true, ""
	}

	// Euclidean distance <= 5
	dist := math.Sqrt(math.Pow(charX-portalX, 2) + math.Pow(charY-portalY, 2))
	if dist <= 5.0 {
		return true, ""
	}

	// Generate hint
	dx := portalX - charX
	dy := portalY - charY

	direction := ""
	if math.Abs(dy) > math.Abs(dx) {
		if dy > 0 {
			direction = "North"
		} else {
			direction = "South"
		}
	} else {
		if dx > 0 {
			direction = "East"
		} else {
			direction = "West"
		}
	}

	return false, fmt.Sprintf("You are too far from the portal. Try moving %s.", direction)
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
		return 1, 1, "northeast"
	case "nw", "northwest":
		return -1, 1, "northwest"
	case "se", "southeast":
		return 1, -1, "southeast"
	case "sw", "southwest":
		return -1, -1, "southwest"
	default:
		return 0, 0, ""
	}
}

func calculateBoundedPosition(x, y, dx, dy float64, world *repository.World) (float64, float64, error) {
	newX := x + dx
	newY := y + dy

	// Get Bounds (default to 0-10 if not set, though caller handles this check usually)
	minX, minY := 0.0, 0.0
	maxX, maxY := 10.0, 10.0
	if world.BoundsMin != nil {
		minX, minY = world.BoundsMin.X, world.BoundsMin.Y
	}
	if world.BoundsMax != nil {
		maxX, maxY = world.BoundsMax.X, world.BoundsMax.Y
	}

	// Check Walls
	if newX < minX || newX > maxX {
		wall := "western"
		if newX > maxX {
			wall = "eastern"
		}
		return x, y, fmt.Errorf("You cannot go further %s. The %s wall blocks your way.", getDirectionName(dx, 0), wall)
	}

	if newY < minY || newY > maxY {
		end := "southern"
		if newY > maxY {
			end = "northern"
		}
		return x, y, fmt.Errorf("You cannot go further %s. The %s wall blocks your way.", getDirectionName(0, dy), end)
	}

	// Check Colliders from Metadata
	// Expecting metadata["colliders"] to be []Collider or equivalent JSON array if not unmarshaled to struct yet
	// Since world repository reads JSONB into map[string]interface{}, nested structs are usually []interface{} of map[string]interface{}

	if val, ok := world.Metadata["colliders"]; ok {
		// Helper to parse generic interface to colliders
		colliders := parseColliders(val)
		for _, c := range colliders {
			dist := math.Sqrt(math.Pow(newX-c.X, 2) + math.Pow(newY-c.Y, 2))
			if dist < c.Radius {
				msg := c.Message
				if msg == "" {
					msg = "Something blocks your path."
				}
				return x, y, fmt.Errorf("%s", msg)
			}
		}
	}

	return newX, newY, nil
}

func parseColliders(data interface{}) []Collider {
	var colliders []Collider

	// If it's already a slice of interfaces
	if list, ok := data.([]interface{}); ok {
		for _, item := range list {
			if m, ok := item.(map[string]interface{}); ok {
				// Safely extract fields
				x, _ := getFloat(m["x"])
				y, _ := getFloat(m["y"])
				r, _ := getFloat(m["radius"])
				msg, _ := m["message"].(string)

				colliders = append(colliders, Collider{
					X:       x,
					Y:       y,
					Radius:  r,
					Message: msg,
				})
			}
		}
	} else {
		// Try re-marshaling if it's some other structure (less likely but robust)
		// Or if explicitly passed as []Collider (in tests)
		if cList, ok := data.([]Collider); ok {
			return cList
		}

		// Fallback: try json roundtrip
		b, err := json.Marshal(data)
		if err == nil {
			_ = json.Unmarshal(b, &colliders)
		}
	}

	return colliders
}

func getFloat(v interface{}) (float64, bool) {
	switch i := v.(type) {
	case float64:
		return i, true
	case int:
		return float64(i), true
	case int64:
		return float64(i), true
	}
	return 0, false
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
	message := ""

	// --- Meter-Based Movement Logic ---
	// We treat the world as a rectangle in meters where:
	// X: [0, Circumference] (Longitude)
	// Y: [-QuarterCircumference, QuarterCircumference] (Latitude -90 to 90 degrees converted to meters)
	// QuarterCircumference = Circumference / 4

	circumference := dims.CircumferenceM
	quarterCircumference := circumference / 4.0

	// 1. Apply movement in meters
	newX := lon + dx
	newY := lat + dy

	// 2. Handle Pole Crossing (Y axis)
	// If Y goes beyond the poles (quarterCircumference), we flip longitude and adjust Y
	if newY > quarterCircumference {
		// Crossed North Pole
		overshoot := newY - quarterCircumference
		newY = quarterCircumference - overshoot
		newX += circumference / 2.0 // Flip to opposite side of globe
		message += " You cross the North Pole and the world spins beneath you."
	} else if newY < -quarterCircumference {
		// Crossed South Pole
		overshoot := -quarterCircumference - newY
		newY = -quarterCircumference + overshoot
		newX += circumference / 2.0 // Flip to opposite side of globe
		message += " You cross the South Pole and the world spins beneath you."
	}

	// 3. Handle Circumnavigation (X axis)
	// Wrap X within [0, Circumference)
	// We use a loop for robustness against large jumps, or modulo
	if newX < 0 {
		newX = math.Mod(newX, circumference)
		if newX < 0 {
			newX += circumference
		}
		message += " You've circled back around the world."
	} else if newX >= circumference {
		newX = math.Mod(newX, circumference)
		message += " You've circled back around the world."
	}

	return newX, newY, message
}
