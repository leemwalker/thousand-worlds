package storage

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapshotStore_Interface tests the public interface without MinIO.
// Integration tests with real MinIO are in minio_integration_test.go

func TestSnapshotStore_ObjectKey(t *testing.T) {
	tests := []struct {
		name    string
		worldID string
		year    int64
		want    string
	}{
		{
			name:    "basic key generation",
			worldID: "abc123",
			year:    1000000000,
			want:    "snapshots/abc123/year-1000000000.bin.gz",
		},
		{
			name:    "uuid worldID",
			worldID: "d8666327-0474-4b47-a461-9f9351052210",
			year:    2500000000,
			want:    "snapshots/d8666327-0474-4b47-a461-9f9351052210/year-2500000000.bin.gz",
		},
		{
			name:    "year zero",
			worldID: "world1",
			year:    0,
			want:    "snapshots/world1/year-0.bin.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := ObjectKey(tt.worldID, tt.year)
			assert.Equal(t, tt.want, key)
		})
	}
}

func TestSnapshotInfo_Parse(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantWorld string
		wantYear  int64
		wantErr   bool
	}{
		{
			name:      "valid key",
			key:       "snapshots/abc123/year-1000000000.bin.gz",
			wantWorld: "abc123",
			wantYear:  1000000000,
			wantErr:   false,
		},
		{
			name:      "uuid world",
			key:       "snapshots/d8666327-0474-4b47-a461-9f9351052210/year-2500000000.bin.gz",
			wantWorld: "d8666327-0474-4b47-a461-9f9351052210",
			wantYear:  2500000000,
			wantErr:   false,
		},
		{
			name:    "invalid format",
			key:     "invalid/key/format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseObjectKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantWorld, info.WorldID)
			assert.Equal(t, tt.wantYear, info.Year)
		})
	}
}

// MockSnapshotStore provides an in-memory implementation for unit testing
type MockSnapshotStore struct {
	objects map[string][]byte
}

func NewMockSnapshotStore() *MockSnapshotStore {
	return &MockSnapshotStore{
		objects: make(map[string][]byte),
	}
}

func (m *MockSnapshotStore) Upload(_ context.Context, worldID string, year int64, data []byte) (string, error) {
	key := ObjectKey(worldID, year)
	m.objects[key] = append([]byte{}, data...) // Copy
	return key, nil
}

func (m *MockSnapshotStore) Download(_ context.Context, objectKey string) ([]byte, error) {
	data, ok := m.objects[objectKey]
	if !ok {
		return nil, ErrSnapshotNotFound
	}
	return data, nil
}

func (m *MockSnapshotStore) List(_ context.Context, worldID string) ([]SnapshotInfo, error) {
	var results []SnapshotInfo
	prefix := "snapshots/" + worldID + "/"
	for key := range m.objects {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			info, _ := ParseObjectKey(key)
			if info != nil {
				results = append(results, *info)
			}
		}
	}
	return results, nil
}

func (m *MockSnapshotStore) Delete(_ context.Context, objectKey string) error {
	delete(m.objects, objectKey)
	return nil
}

func TestMockSnapshotStore_UploadDownload(t *testing.T) {
	store := NewMockSnapshotStore()
	ctx := context.Background()

	worldID := uuid.New().String()
	year := int64(1000000000)
	data := []byte("compressed heightmap data")

	// Upload
	key, err := store.Upload(ctx, worldID, year, data)
	require.NoError(t, err)
	assert.NotEmpty(t, key)

	// Download
	downloaded, err := store.Download(ctx, key)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(data, downloaded))
}

func TestMockSnapshotStore_List(t *testing.T) {
	store := NewMockSnapshotStore()
	ctx := context.Background()

	worldID := "test-world"

	// Upload multiple snapshots
	years := []int64{1000000000, 2000000000, 3000000000}
	for _, year := range years {
		_, err := store.Upload(ctx, worldID, year, []byte("data"))
		require.NoError(t, err)
	}

	// Upload to different world
	_, err := store.Upload(ctx, "other-world", 5000000000, []byte("other"))
	require.NoError(t, err)

	// List should only return our world's snapshots
	snapshots, err := store.List(ctx, worldID)
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)
}

func TestMockSnapshotStore_Delete(t *testing.T) {
	store := NewMockSnapshotStore()
	ctx := context.Background()

	key, err := store.Upload(ctx, "world", 1000000000, []byte("data"))
	require.NoError(t, err)

	// Delete
	err = store.Delete(ctx, key)
	require.NoError(t, err)

	// Should not be found
	_, err = store.Download(ctx, key)
	assert.ErrorIs(t, err, ErrSnapshotNotFound)
}

func TestMockSnapshotStore_NotFound(t *testing.T) {
	store := NewMockSnapshotStore()
	ctx := context.Background()

	_, err := store.Download(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrSnapshotNotFound)
}
