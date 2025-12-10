package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	"tw-backend/internal/repository"
)

type DesireEngine struct {
	nc   *nats.Conn
	repo *repository.NPCMemoryRepository
}

func NewDesireEngine(nc *nats.Conn, repo *repository.NPCMemoryRepository) *DesireEngine {
	return &DesireEngine{
		nc:   nc,
		repo: repo,
	}
}

type DecideActionCommand struct {
	EntityID   string `json:"entityID"`
	WorldID    string `json:"worldID"`
	WorldState string `json:"worldState"`
}

type AIRequest struct {
	RequestID string `json:"requestID"`
	Prompt    string `json:"prompt"`
	EntityID  string `json:"entityID"`
}

type AIResponse struct {
	RequestID string `json:"requestID"`
	Response  string `json:"response"`
	EntityID  string `json:"entityID"`
}

// ListenForDecisions subscribes to npc.command.decide_action and ai.response.decision.*
func (e *DesireEngine) ListenForDecisions() error {
	// Listener 1: Decide Action
	_, err := e.nc.Subscribe("npc.command.decide_action", e.handleDecideAction)
	if err != nil {
		return fmt.Errorf("subscribe decide_action failed: %w", err)
	}

	// Listener 2: AI Response
	_, err = e.nc.Subscribe("ai.response.decision.*", e.handleAIResponse)
	if err != nil {
		return fmt.Errorf("subscribe ai_response failed: %w", err)
	}

	return nil
}

func (e *DesireEngine) handleDecideAction(msg *nats.Msg) {
	var cmd DecideActionCommand
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal decide action")
		return
	}

	// 1. Retrieve Context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	memories, err := e.repo.GetMemoriesByWorldID(ctx, cmd.WorldID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get memories")
		return
	}

	// Filter for this NPC and high importance
	var contextStr string
	for _, mem := range memories {
		if mem.NPCID == cmd.EntityID && mem.ImportanceScore > 0.5 {
			contextStr += fmt.Sprintf("- %s\n", mem.Content)
		}
	}

	// 2. Format Prompt
	prompt := fmt.Sprintf(`
You are an NPC in a MUD.
Personality: Grumpy, suspicious.
Memories:
%s
Current Situation:
%s
Decide your next action. Output ONLY the action command (e.g., "MOVE NORTH", "SAY Hello").
`, contextStr, cmd.WorldState)

	// 3. Call Gateway
	req := AIRequest{
		RequestID: fmt.Sprintf("req-%s-%d", cmd.EntityID, time.Now().UnixNano()),
		Prompt:    prompt,
		EntityID:  cmd.EntityID,
	}
	data, _ := json.Marshal(req)

	if err := e.nc.Publish("ai.request.decision", data); err != nil {
		log.Error().Err(err).Msg("Failed to publish AI request")
	}
}

func (e *DesireEngine) handleAIResponse(msg *nats.Msg) {
	var resp AIResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal AI response")
		return
	}

	log.Info().Str("entityID", resp.EntityID).Str("action", resp.Response).Msg("NPC Decided Action")

	// Publish final action to spatial service (assuming it's a move or generic action)
	// For now, we'll publish to a generic subject that the spatial service might listen to,
	// or just log it as the requirement says "publish the final action... for the spatial-service to process"
	// Let's assume spatial service listens to 'spatial.command.move' but that requires coordinates.
	// If the LLM outputs "MOVE NORTH", we'd need to translate that to coords.
	// Since we don't have a parser yet, we'll publish to 'npc.action.performed' which could be picked up.
	// OR we can assume the LLM outputs JSON with coords? No, prompt says "MOVE NORTH".
	// I'll publish to `spatial.command.action` for now.
	
	if err := e.nc.Publish("spatial.command.action", []byte(resp.Response)); err != nil {
		log.Error().Err(err).Msg("Failed to publish final action")
	}
}
