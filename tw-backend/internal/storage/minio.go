// Package storage provides persistence abstractions for snapshot data.
// L3 storage uses MinIO (S3-compatible) for large binary snapshots.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"tw-backend/internal/circuitbreaker"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrSnapshotNotFound is returned when a requested snapshot doesn't exist.
var ErrSnapshotNotFound = errors.New("snapshot not found")

// SnapshotStore handles snapshot persistence to MinIO/S3.
type SnapshotStore struct {
	client     *minio.Client
	bucketName string
	cb         *circuitbreaker.CircuitBreaker
}

// NewSnapshotStore creates a new MinIO-backed snapshot store.
// endpoint should be "host:port" format (e.g., "minio:9000").
// useSSL should be true for HTTPS connections.
func NewSnapshotStore(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*SnapshotStore, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	// Configure circuit breaker for MinIO calls
	// ErrSnapshotNotFound is NOT a service failure, so we exclude it
	cbConfig := circuitbreaker.Config{
		Name:             "minio",
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		IsFailure: func(err error) bool {
			// "Not found" is expected - don't trip the circuit for it
			return err != nil && !errors.Is(err, ErrSnapshotNotFound)
		},
	}
	cb := circuitbreaker.New(cbConfig)

	// Log state transitions
	cb.OnStateChange(func(name string, from, to circuitbreaker.State) {
		fmt.Printf("[CircuitBreaker] %s: %s -> %s\n", name, from, to)
	})

	return &SnapshotStore{
		client:     client,
		bucketName: bucket,
		cb:         cb,
	}, nil
}

// CircuitBreakerStats returns the current circuit breaker statistics.
func (s *SnapshotStore) CircuitBreakerStats() circuitbreaker.Stats {
	return s.cb.Stats()
}

// EnsureBucket creates the bucket if it doesn't exist.
func (s *SnapshotStore) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}

	return nil
}

// Upload stores a snapshot and returns the object key.
// The data is expected to be already gzip-compressed.
// Uses circuit breaker to prevent cascading failures when MinIO is unavailable.
func (s *SnapshotStore) Upload(ctx context.Context, worldID string, year int64, data []byte) (string, error) {
	key := ObjectKey(worldID, year)

	err := s.cb.Execute(ctx, func(ctx context.Context) error {
		_, err := s.client.PutObject(ctx, s.bucketName, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
			ContentType:     "application/gzip",
			ContentEncoding: "gzip",
		})
		if err != nil {
			return fmt.Errorf("upload snapshot: %w", err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return key, nil
}

// Download retrieves a snapshot by its object key.
// Uses circuit breaker to prevent cascading failures when MinIO is unavailable.
// Note: ErrSnapshotNotFound is NOT treated as a circuit breaker failure.
func (s *SnapshotStore) Download(ctx context.Context, objectKey string) ([]byte, error) {
	var result []byte

	err := s.cb.Execute(ctx, func(ctx context.Context) error {
		obj, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("get object: %w", err)
		}
		defer obj.Close()

		data, err := io.ReadAll(obj)
		if err != nil {
			// Check if it's a not found error - this is NOT a service failure
			var minioErr minio.ErrorResponse
			if errors.As(err, &minioErr) {
				if minioErr.Code == "NoSuchKey" {
					result = nil
					return ErrSnapshotNotFound // This won't trip the circuit
				}
			}
			return fmt.Errorf("read snapshot: %w", err)
		}

		// Empty data means not found in some cases
		if len(data) == 0 {
			// Try to stat to confirm existence
			_, statErr := obj.Stat()
			if statErr != nil {
				result = nil
				return ErrSnapshotNotFound
			}
		}

		result = data
		return nil
	})

	// Handle "not found" specially - it's not a service error
	if errors.Is(err, ErrSnapshotNotFound) {
		return nil, ErrSnapshotNotFound
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}

// List returns all snapshots for a given world, ordered by year.
func (s *SnapshotStore) List(ctx context.Context, worldID string) ([]SnapshotInfo, error) {
	prefix := "snapshots/" + worldID + "/"

	var results []SnapshotInfo

	for obj := range s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("list objects: %w", obj.Err)
		}

		info, err := ParseObjectKey(obj.Key)
		if err != nil {
			continue // Skip malformed keys
		}

		info.Size = obj.Size
		info.CreatedAt = obj.LastModified
		results = append(results, *info)
	}

	return results, nil
}

// Delete removes a snapshot by its object key.
func (s *SnapshotStore) Delete(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete snapshot: %w", err)
	}
	return nil
}

// GetLatestSnapshot returns the most recent snapshot for a world.
func (s *SnapshotStore) GetLatestSnapshot(ctx context.Context, worldID string) (*SnapshotInfo, error) {
	snapshots, err := s.List(ctx, worldID)
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, ErrSnapshotNotFound
	}

	// Find the one with highest year
	var latest *SnapshotInfo
	for i := range snapshots {
		if latest == nil || snapshots[i].Year > latest.Year {
			latest = &snapshots[i]
		}
	}

	return latest, nil
}

// ObjectKey generates the storage key for a snapshot.
// Format: snapshots/{worldID}/year-{year}.bin.gz
func ObjectKey(worldID string, year int64) string {
	return fmt.Sprintf("snapshots/%s/year-%d.bin.gz", worldID, year)
}

// objectKeyPattern matches snapshot object keys
var objectKeyPattern = regexp.MustCompile(`^snapshots/([^/]+)/year-(\d+)\.bin\.gz$`)

// ParseObjectKey extracts world ID and year from an object key.
func ParseObjectKey(key string) (*SnapshotInfo, error) {
	matches := objectKeyPattern.FindStringSubmatch(key)
	if matches == nil {
		return nil, fmt.Errorf("invalid object key format: %s", key)
	}

	year, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse year: %w", err)
	}

	return &SnapshotInfo{
		WorldID: matches[1],
		Year:    year,
		Key:     key,
	}, nil
}

// SnapshotInfo represents metadata about a stored snapshot.
type SnapshotInfo struct {
	WorldID   string
	Year      int64
	Key       string
	Size      int64
	CreatedAt time.Time
}

// SnapshotStoreInterface defines the contract for snapshot storage.
// Useful for mocking in tests.
type SnapshotStoreInterface interface {
	Upload(ctx context.Context, worldID string, year int64, data []byte) (string, error)
	Download(ctx context.Context, objectKey string) ([]byte, error)
	List(ctx context.Context, worldID string) ([]SnapshotInfo, error)
	Delete(ctx context.Context, objectKey string) error
	GetLatestSnapshot(ctx context.Context, worldID string) (*SnapshotInfo, error)
}

// Verify SnapshotStore implements interface
var _ SnapshotStoreInterface = (*SnapshotStore)(nil)
