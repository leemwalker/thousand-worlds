package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	"tw-backend/internal/service"
)

type EventListener struct {
	nc      *nats.Conn
	service *service.SpatialService
}

func NewEventListener(nc *nats.Conn, svc *service.SpatialService) *EventListener {
	return &EventListener{
		nc:      nc,
		service: svc,
	}
}

type MoveCommand struct {
	EntityID  string    `json:"entityID"` // String UUID from JSON
	NewCoords NewCoords `json:"newCoords"`
}

type NewCoords struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (l *EventListener) ListenForMove() error {
	_, err := l.nc.Subscribe("spatial.command.move", func(msg *nats.Msg) {
		var cmd MoveCommand
		if err := json.Unmarshal(msg.Data, &cmd); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal move command")
			return
		}

		entityID, err := uuid.Parse(cmd.EntityID)
		if err != nil {
			log.Error().Err(err).Str("entityID", cmd.EntityID).Msg("Invalid entity ID")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.service.UpdateLocation(ctx, entityID, cmd.NewCoords.X, cmd.NewCoords.Y, cmd.NewCoords.Z); err != nil {
			log.Error().Err(err).Msg("Failed to update location")
			return
		}

		log.Info().
			Str("entityID", cmd.EntityID).
			Float64("x", cmd.NewCoords.X).
			Float64("y", cmd.NewCoords.Y).
			Float64("z", cmd.NewCoords.Z).
			Msg("Entity moved")
	})

	if err != nil {
		return fmt.Errorf("eventListener.ListenForMove: subscribe failed: %w", err)
	}

	return nil
}
