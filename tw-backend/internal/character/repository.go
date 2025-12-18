package character

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tw-backend/internal/eventstore"

	"github.com/google/uuid"

	apperrors "tw-backend/internal/errors"
)

// CharacterRepository handles persistence of characters via event sourcing
type CharacterRepository interface {
	Save(ctx context.Context, char *Character, newEvents []interface{}) error
	Load(ctx context.Context, id uuid.UUID) (*Character, error)
}

type eventSourcedRepository struct {
	store eventstore.EventStore
}

// NewCharacterRepository creates a new event-sourced repository
func NewCharacterRepository(store eventstore.EventStore) CharacterRepository {
	return &eventSourcedRepository{store: store}
}

func (r *eventSourcedRepository) Save(ctx context.Context, char *Character, newEvents []interface{}) error {
	for _, e := range newEvents {
		// Determine event type and marshal payload
		var eventType eventstore.EventType
		var payload []byte
		var err error

		switch v := e.(type) {
		case CharacterCreatedViaGenerationEvent:
			eventType = eventstore.EventType(EventTypeCharacterCreatedViaGeneration)
			payload, err = json.Marshal(v)
		case CharacterCreatedViaInhabitanceEvent:
			eventType = eventstore.EventType(EventTypeCharacterCreatedViaInhabitance)
			payload, err = json.Marshal(v)
		case AttributeModifiedEvent:
			eventType = eventstore.EventType(EventTypeAttributeModified)
			payload, err = json.Marshal(v)
		default:
			return fmt.Errorf("unknown event type: %T", e)
		}

		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		event := eventstore.Event{
			ID:            uuid.New().String(),
			EventType:     eventType,
			AggregateID:   char.ID.String(),
			AggregateType: "Character",
			Version:       0, // TODO: Handle optimistic concurrency
			Timestamp:     time.Now(),
			Payload:       json.RawMessage(payload),
		}

		if err := r.store.AppendEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (r *eventSourcedRepository) Load(ctx context.Context, id uuid.UUID) (*Character, error) {
	events, err := r.store.GetEventsByAggregate(ctx, id.String(), 0)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, apperrors.NewNotFound("character not found: %s", id)
	}

	char := &Character{ID: id}
	for _, event := range events {
		if err := r.applyEvent(char, event); err != nil {
			return nil, err
		}
	}

	return char, nil
}

func (r *eventSourcedRepository) applyEvent(char *Character, event eventstore.Event) error {
	// Unmarshal payload if it's raw JSON, or type switch if it's already a struct
	// Assuming EventStore returns typed events or we need to unmarshal based on Type

	// var payload interface{} // This line is no longer needed
	// If event.Data is json.RawMessage or []byte, unmarshal it
	// But we need to know the type.
	// Let's assume event.Type tells us.

	switch string(event.EventType) {
	case EventTypeCharacterCreatedViaGeneration:
		var e CharacterCreatedViaGenerationEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			return err
		}
		char.PlayerID = e.PlayerID
		char.Name = e.Name
		char.Species = e.Species
		char.BaseAttrs = e.FinalAttributes // Use final attributes as base for simplicity in current model
		char.SecAttrs = CalculateSecondaryAttributes(char.BaseAttrs)
		char.CreatedAt = e.Timestamp
		char.UpdatedAt = e.Timestamp

	case EventTypeCharacterCreatedViaInhabitance:
		var e CharacterCreatedViaInhabitanceEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			return err
		}
		char.PlayerID = e.PlayerID
		// char.Name would come from NPC data, but here we might need to fetch it or it's in the event?
		// The event definition didn't have Name, but maybe it should?
		// For now, we'll leave it empty or set a placeholder
		char.Name = "Inhabited NPC"
		char.CreatedAt = e.Timestamp
		char.UpdatedAt = e.Timestamp

	case EventTypeAttributeModified:
		var e AttributeModifiedEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			return err
		}
		// Apply modification logic
		// This requires mapping string attribute names to struct fields
		// We can use a helper or reflection, or a big switch
		r.applyAttributeModification(char, e.Attribute, e.NewValue)
		char.UpdatedAt = e.Timestamp
	}

	return nil
}

func (r *eventSourcedRepository) applyAttributeModification(char *Character, attr string, value int) {
	switch attr {
	case AttrMight:
		char.BaseAttrs.Might = value
	case AttrAgility:
		char.BaseAttrs.Agility = value
	case AttrEndurance:
		char.BaseAttrs.Endurance = value
	case AttrReflexes:
		char.BaseAttrs.Reflexes = value
	case AttrVitality:
		char.BaseAttrs.Vitality = value
	case AttrIntellect:
		char.BaseAttrs.Intellect = value
	case AttrCunning:
		char.BaseAttrs.Cunning = value
	case AttrWillpower:
		char.BaseAttrs.Willpower = value
	case AttrPresence:
		char.BaseAttrs.Presence = value
	case AttrIntuition:
		char.BaseAttrs.Intuition = value
	case AttrSight:
		char.BaseAttrs.Sight = value
	case AttrHearing:
		char.BaseAttrs.Hearing = value
	case AttrSmell:
		char.BaseAttrs.Smell = value
	case AttrTaste:
		char.BaseAttrs.Taste = value
	case AttrTouch:
		char.BaseAttrs.Touch = value
	}
	// Recalculate secondary attributes
	char.SecAttrs = CalculateSecondaryAttributes(char.BaseAttrs)
}
