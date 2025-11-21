package eventstore

import (
	"context"
	"fmt"
)

// Projection updates a read model based on events.
type Projection interface {
	Name() string
	HandleEvent(ctx context.Context, event Event) error
}

// ProjectionManager manages projections and dispatches events to them.
type ProjectionManager struct {
	projections map[string]Projection
}

func NewProjectionManager() *ProjectionManager {
	return &ProjectionManager{
		projections: make(map[string]Projection),
	}
}

func (pm *ProjectionManager) RegisterProjection(p Projection) {
	pm.projections[p.Name()] = p
}

func (pm *ProjectionManager) ProjectEvent(ctx context.Context, event Event) error {
	// Dispatch event to all registered projections
	// In a real system, this might be async, but for now we do it synchronously
	// to ensure consistency for the test requirement "CQRS read model stays consistent with event stream".

	for name, p := range pm.projections {
		if err := p.HandleEvent(ctx, event); err != nil {
			return fmt.Errorf("projection %s failed to handle event %s: %w", name, event.ID, err)
		}
	}
	return nil
}
