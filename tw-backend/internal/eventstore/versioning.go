package eventstore

import "fmt"

// Upcaster transforms an event from an older version to a newer version.
type Upcaster interface {
	Upcast(event Event) (Event, error)
	SourceType() EventType
	TargetType() EventType
}

// VersioningManager manages event upcasting.
type VersioningManager struct {
	upcasters map[EventType]Upcaster
}

func NewVersioningManager() *VersioningManager {
	return &VersioningManager{
		upcasters: make(map[EventType]Upcaster),
	}
}

func (vm *VersioningManager) RegisterUpcaster(u Upcaster) {
	vm.upcasters[u.SourceType()] = u
}

func (vm *VersioningManager) UpcastEvent(event Event) (Event, error) {
	// Check if there is an upcaster for this event type
	// This is a simple 1-step upcast. For multi-step, we'd need a chain.
	// For now, let's assume simple V1 -> V2 migration where type changes or payload changes.
	// If the event type stays the same but schema changes, we might need a version field in metadata or payload.
	// The prompt says "Support event schema migrations (V1 -> V2 transformations)".

	// Let's assume the upcaster changes the EventType or modifies the Payload.
	// If we find an upcaster for the current EventType, we apply it.
	// We should loop until no more upcasters apply (chaining).

	currentEvent := event
	for {
		upcaster, ok := vm.upcasters[currentEvent.EventType]
		if !ok {
			break
		}

		upcasted, err := upcaster.Upcast(currentEvent)
		if err != nil {
			return event, fmt.Errorf("failed to upcast event %s from %s: %w", event.ID, currentEvent.EventType, err)
		}
		currentEvent = upcasted
	}

	return currentEvent, nil
}
