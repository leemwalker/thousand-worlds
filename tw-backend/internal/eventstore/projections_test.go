package eventstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadModel is a simple in-memory read model for testing.
type TestReadModel struct {
	WorldName string
	TickCount int64
}

// TestProjection updates TestReadModel.
type TestProjection struct {
	model *TestReadModel
}

func (p *TestProjection) Name() string {
	return "TestProjection"
}

func (p *TestProjection) HandleEvent(ctx context.Context, event Event) error {
	switch event.EventType {
	case "WorldCreated":
		var payload map[string]interface{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		if name, ok := payload["name"].(string); ok {
			p.model.WorldName = name
		}
	case "WorldTicked":
		p.model.TickCount++
	}
	return nil
}

func TestProjectionManager_ProjectEvent(t *testing.T) {
	model := &TestReadModel{}
	projection := &TestProjection{model: model}
	pm := NewProjectionManager()
	pm.RegisterProjection(projection)
	ctx := context.Background()

	t.Run("updates read model from events", func(t *testing.T) {
		// 1. WorldCreated
		evt1 := Event{
			ID:        "evt-1",
			EventType: "WorldCreated",
			Payload:   json.RawMessage(`{"name": "New World"}`),
			Timestamp: time.Now().UTC(),
		}
		err := pm.ProjectEvent(ctx, evt1)
		require.NoError(t, err)
		assert.Equal(t, "New World", model.WorldName)
		assert.Equal(t, int64(0), model.TickCount)

		// 2. WorldTicked
		evt2 := Event{
			ID:        "evt-2",
			EventType: "WorldTicked",
			Payload:   json.RawMessage(`{}`),
			Timestamp: time.Now().UTC(),
		}
		err = pm.ProjectEvent(ctx, evt2)
		require.NoError(t, err)
		assert.Equal(t, int64(1), model.TickCount)
	})
}
