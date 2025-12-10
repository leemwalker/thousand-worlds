package processor

import (
	"context"
	"fmt"
	"log"
	"strings"

	"tw-backend/cmd/game-server/websocket"
)

// handleFace handles the face command to change orientation and look
func (p *GameProcessor) handleFace(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// 1. Get Character
	charID := client.GetCharacterID()
	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// 2. Handle simple "face" (status check)
	if cmd.Text == "face" || (cmd.Target == nil && (len(strings.Fields(cmd.Text)) <= 1)) {
		// Just report current facing
		direction := p.spatialService.GetDirectionName(char.OrientationX, char.OrientationY, char.OrientationZ)
		client.SendGameMessage("info", fmt.Sprintf("You are facing %s.", direction), nil)
		return nil
	}

	// 3. Parse target direction
	// The command parser might put the direction in Target if it was "face north"
	targetDir := ""
	if cmd.Target != nil {
		targetDir = *cmd.Target
	} else if len(strings.Split(cmd.Text, " ")) > 1 {
		// Fallback manual parse if parser didn't catch it
		parts := strings.Split(cmd.Text, " ")
		targetDir = parts[1]
	}

	if targetDir == "" {
		client.SendGameMessage("error", "Face which direction?", nil)
		return nil
	}

	// 4. Validate Direction
	dx, dy, dz, dirName := p.spatialService.GetOrientationVector(targetDir)
	if dirName == "" {
		client.SendGameMessage("error", "Invalid direction.", nil)
		return nil
	}

	// 5. Update Orientation
	char.OrientationX = dx
	char.OrientationY = dy
	char.OrientationZ = dz
	if err := p.authRepo.UpdateCharacter(ctx, char); err != nil {
		return fmt.Errorf("failed to update orientation: %w", err)
	}

	// 6. Look in that direction
	// Calculate "view" position (e.g. 50 meters ahead? 1 tile ahead?)
	// For "what do you see", let's look at the next tile/step.
	world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return err
	}

	// Look 1 unit ahead
	viewX, viewY, _, err := p.spatialService.CalculateNewPosition(char, world, dx, dy)
	if err != nil {
		// Even if movement is blocked, we might see the wall/obstacle.
		// CalculateNewPosition returns the blocked pos or current pos with error?
		// Current impl returns current pos + error message if blocked.
		// If blocked by wall, we look AT the wall.
	}

	// Get description
	viewDesc, err := p.lookService.DescribeView(ctx, char.WorldID, char, viewX, viewY)
	if err != nil {
		log.Printf("[FACE] Error describing view: %v", err)
		viewDesc = "You see nothing special."
	}

	msg := fmt.Sprintf("You face %s. You see: %s", dirName, viewDesc)
	client.SendGameMessage("info", msg, nil)

	return nil
}
