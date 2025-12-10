package character

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"tw-backend/internal/eventstore"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventStore is a mock implementation of eventstore.EventStore
type MockEventStore struct {
	mock.Mock
}

func (m *MockEventStore) AppendEvent(ctx context.Context, event eventstore.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventStore) GetEventsByAggregate(ctx context.Context, aggregateID string, fromVersion int64) ([]eventstore.Event, error) {
	args := m.Called(ctx, aggregateID, fromVersion)
	return args.Get(0).([]eventstore.Event), args.Error(1)
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, eventType eventstore.EventType, fromTimestamp, toTimestamp time.Time) ([]eventstore.Event, error) {
	args := m.Called(ctx, eventType, fromTimestamp, toTimestamp)
	return args.Get(0).([]eventstore.Event), args.Error(1)
}

func (m *MockEventStore) GetAllEvents(ctx context.Context, fromTimestamp time.Time, limit int) ([]eventstore.Event, error) {
	args := m.Called(ctx, fromTimestamp, limit)
	return args.Get(0).([]eventstore.Event), args.Error(1)
}

func TestRepository_Save(t *testing.T) {
	mockStore := new(MockEventStore)
	repo := NewCharacterRepository(mockStore)
	ctx := context.Background()

	charID := uuid.New()
	char := &Character{ID: charID}
	events := []interface{}{
		CharacterCreatedViaGenerationEvent{CharacterID: charID},
	}

	// Expect AppendEvent to be called
	mockStore.On("AppendEvent", ctx, mock.MatchedBy(func(e eventstore.Event) bool {
		return e.AggregateID == charID.String() && string(e.EventType) == EventTypeCharacterCreatedViaGeneration
	})).Return(nil)

	err := repo.Save(ctx, char, events)
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestRepository_Load(t *testing.T) {
	mockStore := new(MockEventStore)
	repo := NewCharacterRepository(mockStore)
	ctx := context.Background()

	charID := uuid.New()
	playerID := uuid.New()
	timestamp := time.Now()

	// Create a sample event
	createdEvent := CharacterCreatedViaGenerationEvent{
		CharacterID: charID,
		PlayerID:    playerID,
		Name:        "Test Hero",
		Species:     SpeciesHuman,
		FinalAttributes: Attributes{
			Might: 60,
		},
		Timestamp: timestamp,
	}
	data, _ := json.Marshal(createdEvent)

	storedEvents := []eventstore.Event{
		{
			EventType: eventstore.EventType(EventTypeCharacterCreatedViaGeneration),
			Payload:   json.RawMessage(data),
		},
	}

	mockStore.On("GetEventsByAggregate", ctx, charID.String(), int64(0)).Return(storedEvents, nil)

	char, err := repo.Load(ctx, charID)
	assert.NoError(t, err)
	assert.NotNil(t, char)
	assert.Equal(t, charID, char.ID)
	assert.Equal(t, "Test Hero", char.Name)
	assert.Equal(t, 60, char.BaseAttrs.Might)
	mockStore.AssertExpectations(t)
}

func TestRepository_Save_AllTypes(t *testing.T) {
	mockStore := new(MockEventStore)
	repo := NewCharacterRepository(mockStore)
	ctx := context.Background()

	charID := uuid.New()
	char := &Character{ID: charID}

	// Test saving all event types
	events := []interface{}{
		CharacterCreatedViaGenerationEvent{CharacterID: charID},
		CharacterCreatedViaInhabitanceEvent{CharacterID: charID},
		AttributeModifiedEvent{CharacterID: charID, Attribute: AttrMight, NewValue: 10},
	}

	mockStore.On("AppendEvent", ctx, mock.Anything).Return(nil).Times(3)

	err := repo.Save(ctx, char, events)
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestRepository_Load_AttributeModified(t *testing.T) {
	mockStore := new(MockEventStore)
	repo := NewCharacterRepository(mockStore)
	ctx := context.Background()

	charID := uuid.New()

	// Sequence: Created -> Modified
	createdEvent := CharacterCreatedViaGenerationEvent{
		CharacterID:     charID,
		FinalAttributes: Attributes{Might: 50},
	}
	createdData, _ := json.Marshal(createdEvent)

	modEvent := AttributeModifiedEvent{
		CharacterID: charID,
		Attribute:   AttrMight,
		NewValue:    60,
		Timestamp:   time.Now(),
	}
	modData, _ := json.Marshal(modEvent)

	storedEvents := []eventstore.Event{
		{EventType: eventstore.EventType(EventTypeCharacterCreatedViaGeneration), Payload: json.RawMessage(createdData)},
		{EventType: eventstore.EventType(EventTypeAttributeModified), Payload: json.RawMessage(modData)},
	}

	mockStore.On("GetEventsByAggregate", ctx, charID.String(), int64(0)).Return(storedEvents, nil)

	char, err := repo.Load(ctx, charID)
	assert.NoError(t, err)
	assert.Equal(t, 60, char.BaseAttrs.Might)
	mockStore.AssertExpectations(t)
}

func TestRepository_Load_Inhabitance(t *testing.T) {
	mockStore := new(MockEventStore)
	repo := NewCharacterRepository(mockStore)
	ctx := context.Background()

	charID := uuid.New()

	event := CharacterCreatedViaInhabitanceEvent{
		CharacterID: charID,
		NPCID:       uuid.New(),
		Timestamp:   time.Now(),
	}
	data, _ := json.Marshal(event)

	storedEvents := []eventstore.Event{
		{EventType: eventstore.EventType(EventTypeCharacterCreatedViaInhabitance), Payload: json.RawMessage(data)},
	}

	mockStore.On("GetEventsByAggregate", ctx, charID.String(), int64(0)).Return(storedEvents, nil)

	char, err := repo.Load(ctx, charID)
	assert.NoError(t, err)
	assert.Equal(t, "Inhabited NPC", char.Name)
	mockStore.AssertExpectations(t)
}
