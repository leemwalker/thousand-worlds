package processor

import (
	"context"
	"fmt"
	"strings"

	"tw-backend/cmd/game-server/websocket"
)

// handleReply sends a reply to the last person who sent a tell
func (p *GameProcessor) handleReply(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	// Validate message is not empty
	if cmd.Message == nil || strings.TrimSpace(*cmd.Message) == "" {
		client.SendGameMessage("error", "What do you want to say?", nil)
		return nil
	}

	message := strings.TrimSpace(*cmd.Message)

	// Get last tell sender directly from interface
	lastSender := client.GetLastTellSender()
	if lastSender == "" {
		client.SendGameMessage("error", "You haven't received any messages to reply to.", nil)
		return nil
	}

	senderUsername := client.GetUsername()
	senderCharID := client.GetCharacterID()
	lastSenderLower := strings.ToLower(lastSender)

	// Special case: If replying to statue, route to interview service
	if lastSenderLower == "statue" {
		userID := client.GetUserID()

		// Try to get existing interview
		_, err := p.interviewService.GetActiveInterview(ctx, userID)
		if err != nil {
			client.SendGameMessage("error", "You don't have an active interview with the statue.", nil)
			return nil
		}

		// Send thinking emote first
		client.SendGameMessage("emote", "The statue listens intently...", map[string]interface{}{
			"source": "Statue",
		})

		// Process the response
		nextQuestion, isComplete, err := p.interviewService.ProcessResponse(ctx, userID, message)
		if err != nil {
			client.SendGameMessage("error", fmt.Sprintf("The statue seems confused. %v", err), nil)
			return err
		}

		// Send emote
		client.SendGameMessage("emote", "The statue's eyes glow warmly.", map[string]interface{}{
			"source": "Statue",
		})

		// Send the actual response from the interview service
		var response string
		if isComplete {
			response = fmt.Sprintf("A voice resonates in your mind:\n\n%s\n\nYour world is being forged. You may now enter it using 'enter <world_name>'.", nextQuestion)
		} else {
			response = fmt.Sprintf("A voice resonates in your mind:\n\n%s", nextQuestion)
		}

		client.SendGameMessage("tell", response, map[string]interface{}{
			"sender_name": "Statue",
			"is_complete": isComplete,
		})

		// Keep statue as last tell sender so user can continue replying
		client.SetLastTellSender("statue")
		return nil
	}

	// Find target client (last sender)
	allClients := p.Hub.GetAllClients()
	var targetClient websocket.GameClient
	for _, c := range allClients {
		if strings.ToLower(c.GetUsername()) == lastSenderLower {
			targetClient = c
			break
		}
	}

	// Check if target was found
	if targetClient == nil {
		client.SendGameMessage("error", fmt.Sprintf("%s is no longer online.", lastSender), nil)
		return nil
	}

	// Check if replying to self (edge case)
	if targetClient.GetCharacterID() == senderCharID {
		client.SendGameMessage("error", "You cannot send a message to yourself.", nil)
		return nil
	}

	// Send to sender (show as tell_self)
	client.SendGameMessage("tell_self", fmt.Sprintf("You tell %s, '%s'", targetClient.GetUsername(), message), map[string]interface{}{
		"sender_id":    senderCharID.String(),
		"sender_name":  senderUsername,
		"recipient_id": targetClient.GetCharacterID().String(),
		"recipient":    targetClient.GetUsername(),
		"message":      message,
	})

	// Send to recipient
	targetClient.SendGameMessage("tell", fmt.Sprintf("%s tells you, '%s'", senderUsername, message), map[string]interface{}{
		"sender_id":   senderCharID.String(),
		"sender_name": senderUsername,
		"message":     message,
	})

	// Update recipient's last tell sender (so they can reply back)
	targetClient.SetLastTellSender(senderUsername)

	return nil
}
