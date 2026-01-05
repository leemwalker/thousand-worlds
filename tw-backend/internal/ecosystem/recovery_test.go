package ecosystem

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pb "tw-backend/api/proto"
	"tw-backend/internal/events"
	"tw-backend/internal/repository"
	"tw-backend/internal/spatial"
	"tw-backend/internal/storage"
	"tw-backend/internal/worldgen/geography"
)

// --- Mocks ---

type MockSnapshotStore struct {
	mock.Mock
}

func (m *MockSnapshotStore) EnsureBucket(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSnapshotStore) Upload(ctx context.Context, worldID string, year int64, data []byte) (string, error) {
	args := m.Called(ctx, worldID, year, data)
	return args.String(0), args.Error(1)
}

func (m *MockSnapshotStore) Download(ctx context.Context, objectKey string) ([]byte, error) {
	args := m.Called(ctx, objectKey)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSnapshotStore) List(ctx context.Context, worldID string) ([]storage.SnapshotInfo, error) {
	args := m.Called(ctx, worldID)
	return args.Get(0).([]storage.SnapshotInfo), args.Error(1)
}

func (m *MockSnapshotStore) Delete(ctx context.Context, objectKey string) error {
	args := m.Called(ctx, objectKey)
	return args.Error(0)
}

func (m *MockSnapshotStore) GetLatestSnapshot(ctx context.Context, worldID string) (*storage.SnapshotInfo, error) {
	args := m.Called(ctx, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.SnapshotInfo), args.Error(1)
}

type MockEventConsumer struct {
	mock.Mock
}

func (m *MockEventConsumer) GetEventsFromSequence(ctx context.Context, worldID string, fromSeq uint64) ([]*pb.SimulationEvent, error) {
	args := m.Called(ctx, worldID, fromSeq)
	return args.Get(0).([]*pb.SimulationEvent), args.Error(1)
}

func (m *MockEventConsumer) GetLatestSequence(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEventConsumer) GetStreamInfo(ctx context.Context) (*jetstream.StreamInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(*jetstream.StreamInfo), args.Error(1)
}

func (m *MockEventConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockSaveRepository struct {
	mock.Mock
}

func (m *MockSaveRepository) CreateSave(ctx context.Context, save *repository.WorldSave) error {
	args := m.Called(ctx, save)
	return args.Error(0)
}

func (m *MockSaveRepository) GetLatestSave(ctx context.Context, worldID uuid.UUID) (*repository.WorldSave, error) {
	args := m.Called(ctx, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorldSave), args.Error(1)
}

func (m *MockSaveRepository) GetSavesForWorld(ctx context.Context, worldID uuid.UUID, limit int) ([]*repository.WorldSave, error) {
	args := m.Called(ctx, worldID, limit)
	return args.Get(0).([]*repository.WorldSave), args.Error(1)
}

func (m *MockSaveRepository) GetSavesForPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]*repository.WorldSave, error) {
	args := m.Called(ctx, playerID, limit)
	return args.Get(0).([]*repository.WorldSave), args.Error(1)
}

func (m *MockSaveRepository) DeleteSave(ctx context.Context, saveID uuid.UUID) error {
	args := m.Called(ctx, saveID)
	return args.Error(0)
}

func (m *MockSaveRepository) DeleteOldSaves(ctx context.Context, worldID uuid.UUID, keepCount int) (int64, error) {
	args := m.Called(ctx, worldID, keepCount)
	return args.Get(0).(int64), args.Error(1)
}

// --- Tests ---

func TestRecoveryService_RecoverWorld(t *testing.T) {
	// Setup generic test data
	worldID := uuid.New()
	snapshotKey := "world-snapshots/test-snapshot.bin.gz"
	var resolution int32 = 10

	// Create a dummy heightmap to serialize
	topo := spatial.NewCubeSphereTopology(int(resolution))
	hm := geography.NewSphereHeightmap(topo)
	// Add some data
	hm.GetFace(0).Elevations[0] = 100.0
	hm.UpdateMinMax()

	serializedHM, err := events.SerializeHeightmap(hm)
	assert.NoError(t, err)

	t.Run("Success with events", func(t *testing.T) {
		mockSnapshotStore := new(MockSnapshotStore)
		mockEventConsumer := new(MockEventConsumer)
		mockSaveRepo := new(MockSaveRepository)

		service := NewRecoveryService(mockSnapshotStore, mockEventConsumer, mockSaveRepo)

		// Expectations
		save := &repository.WorldSave{
			WorldID:       worldID,
			SnapshotKey:   snapshotKey,
			EventSequence: 100,
			Year:          500,
		}
		mockSaveRepo.On("GetLatestSave", mock.Anything, worldID).Return(save, nil)
		mockSnapshotStore.On("Download", mock.Anything, snapshotKey).Return(serializedHM, nil)

		// Events to replay
		replayEvents := []*pb.SimulationEvent{
			{
				Event: &pb.SimulationEvent_Tectonic{
					Tectonic: &pb.TectonicUpdate{Year: 510},
				},
			},
			{
				Event: &pb.SimulationEvent_Erosion{
					Erosion: &pb.ErosionUpdate{Year: 520},
				},
			},
		}
		mockEventConsumer.On("GetEventsFromSequence", mock.Anything, worldID.String(), uint64(100)).
			Return(replayEvents, nil)

		// Execute
		result, err := service.RecoverWorld(context.Background(), worldID)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(500), result.SnapshotYear)
		assert.Equal(t, int64(520), result.CurrentYear)
		assert.Equal(t, 2, result.EventsReplayed)
		assert.NotNil(t, result.Heightmap)

		// Verify heightmap restored
		val := result.Heightmap.GetFace(0).Elevations[0]
		assert.Equal(t, 100.0, val)

		mockSaveRepo.AssertExpectations(t)
		mockSnapshotStore.AssertExpectations(t)
		mockEventConsumer.AssertExpectations(t)
	})

	t.Run("No save found", func(t *testing.T) {
		mockSaveRepo := new(MockSaveRepository)
		service := NewRecoveryService(nil, nil, mockSaveRepo)

		mockSaveRepo.On("GetLatestSave", mock.Anything, worldID).Return(nil, repository.ErrSaveNotFound)

		result, err := service.RecoverWorld(context.Background(), worldID)

		assert.ErrorIs(t, err, repository.ErrSaveNotFound)
		assert.Nil(t, result)
	})

	t.Run("Snapshot download failed", func(t *testing.T) {
		mockSaveRepo := new(MockSaveRepository)
		mockSnapshotStore := new(MockSnapshotStore)
		service := NewRecoveryService(mockSnapshotStore, nil, mockSaveRepo)

		save := &repository.WorldSave{SnapshotKey: snapshotKey}
		mockSaveRepo.On("GetLatestSave", mock.Anything, worldID).Return(save, nil)
		mockSnapshotStore.On("Download", mock.Anything, snapshotKey).Return([]byte{}, errors.New("S3 error"))

		result, err := service.RecoverWorld(context.Background(), worldID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "download snapshot")
		assert.Nil(t, result)
	})

	t.Run("Event replay failed (should not fail recovery)", func(t *testing.T) {
		mockSnapshotStore := new(MockSnapshotStore)
		mockEventConsumer := new(MockEventConsumer)
		mockSaveRepo := new(MockSaveRepository)

		service := NewRecoveryService(mockSnapshotStore, mockEventConsumer, mockSaveRepo)

		save := &repository.WorldSave{
			WorldID:       worldID,
			SnapshotKey:   snapshotKey,
			EventSequence: 100,
			Year:          500,
		}
		mockSaveRepo.On("GetLatestSave", mock.Anything, worldID).Return(save, nil)
		mockSnapshotStore.On("Download", mock.Anything, snapshotKey).Return(serializedHM, nil)

		// Event replay failure
		mockEventConsumer.On("GetEventsFromSequence", mock.Anything, worldID.String(), uint64(100)).
			Return([]*pb.SimulationEvent{}, errors.New("NATS error"))

		// Execute
		result, err := service.RecoverWorld(context.Background(), worldID)

		// Verify - should succeed but with 0 events
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(500), result.SnapshotYear)
		assert.Equal(t, int64(500), result.CurrentYear) // Still at snapshot year
		assert.Equal(t, 0, result.EventsReplayed)
	})
}

func TestRecoveryService_CreateRecoverySave(t *testing.T) {
	mockSaveRepo := new(MockSaveRepository)
	service := NewRecoveryService(nil, nil, mockSaveRepo)

	worldID := uuid.New()
	year := int64(1000)
	key := "key123"
	seq := uint64(50)

	mockSaveRepo.On("CreateSave", mock.Anything, mock.MatchedBy(func(s *repository.WorldSave) bool {
		return s.WorldID == worldID && s.Year == year && s.SnapshotKey == key && s.EventSequence == seq
	})).Return(nil)

	mockSaveRepo.On("DeleteOldSaves", mock.Anything, worldID, 10).Return(int64(2), nil)

	err := service.CreateRecoverySave(context.Background(), worldID, year, key, seq)

	assert.NoError(t, err)
	mockSaveRepo.AssertExpectations(t)
}
