package eventstore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpcaster implements Upcaster for testing.
type TestUpcaster struct{}

func (u *TestUpcaster) SourceType() EventType {
	return "TestEventV1"
}

func (u *TestUpcaster) TargetType() EventType {
	return "TestEventV2"
}

func (u *TestUpcaster) Upcast(event Event) (Event, error) {
	// Transform payload: {"foo": "bar"} -> {"foo": "bar", "version": 2}
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return event, err
	}
	payload["version"] = 2

	newPayload, err := json.Marshal(payload)
	if err != nil {
		return event, err
	}

	event.EventType = "TestEventV2"
	event.Payload = newPayload
	return event, nil
}

func TestVersioningManager_UpcastEvent(t *testing.T) {
	vm := NewVersioningManager()
	vm.RegisterUpcaster(&TestUpcaster{})

	t.Run("upcasts event correctly", func(t *testing.T) {
		event := Event{
			ID:        "evt-v1",
			EventType: "TestEventV1",
			Payload:   json.RawMessage(`{"foo": "bar"}`),
			Timestamp: time.Now().UTC(),
		}

		upcasted, err := vm.UpcastEvent(event)
		require.NoError(t, err)
		assert.Equal(t, EventType("TestEventV2"), upcasted.EventType)

		var payload map[string]interface{}
		err = json.Unmarshal(upcasted.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, float64(2), payload["version"])
	})

	t.Run("returns original event if no upcaster", func(t *testing.T) {
		event := Event{
			ID:        "evt-other",
			EventType: "OtherEvent",
			Payload:   json.RawMessage(`{}`),
		}

		upcasted, err := vm.UpcastEvent(event)
		require.NoError(t, err)
		assert.Equal(t, event, upcasted)
	})
}
