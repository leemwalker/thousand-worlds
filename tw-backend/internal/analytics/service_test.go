package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"tw-backend/internal/circuitbreaker"
	"tw-backend/internal/ecosystem"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewService_InvalidURL tests error handling for bad connection strings
func TestNewService_InvalidURL(t *testing.T) {
	// Invalid URL should fail
	_, err := NewService("postgres://invalid:invalid@localhost:9999/nonexistent?connect_timeout=1")
	require.Error(t, err)
}

// TestNewService_EmptyURL tests error handling for empty URL
func TestNewService_EmptyURL(t *testing.T) {
	_, err := NewService("")
	require.Error(t, err)
}

func TestNewServiceWithDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	// Expect initialization queries
	mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS timescaledb").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS world_metrics").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE INDEX IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))

	svc, err := NewServiceWithDB(db, cb)
	require.NoError(t, err)
	require.NotNil(t, svc)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewServiceWithDB_InitFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	// Fail on first query
	mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS timescaledb").WillReturnError(errors.New("db error"))

	svc, err := NewServiceWithDB(db, cb)
	require.Error(t, err)
	require.Nil(t, svc)
	assert.Contains(t, err.Error(), "init schema")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewServiceWithDB_InitFailure_TableCreation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS timescaledb").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS world_metrics").WillReturnError(errors.New("table init failed"))

	svc, err := NewServiceWithDB(db, cb)
	require.Error(t, err)
	require.Nil(t, svc)
	assert.Contains(t, err.Error(), "create metrics table")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewServiceWithDB_InitFailure_Hypertable(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	mock.ExpectExec("CREATE EXTENSION").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnError(errors.New("hypertable failed"))

	svc, err := NewServiceWithDB(db, cb)
	require.Error(t, err)
	require.Nil(t, svc)
	assert.Contains(t, err.Error(), "create hypertable")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewServiceWithDB_InitFailure_Index(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	mock.ExpectExec("CREATE EXTENSION").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE INDEX").WillReturnError(errors.New("index failed"))

	svc, err := NewServiceWithDB(db, cb)
	require.Error(t, err)
	require.Nil(t, svc)
	assert.Contains(t, err.Error(), "create index")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	// Init queries
	mock.ExpectExec("CREATE EXTENSION").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE INDEX").WillReturnResult(sqlmock.NewResult(0, 0))

	svc, err := NewServiceWithDB(db, cb)
	require.NoError(t, err)

	stats := ecosystem.GlobalStats{
		WorldID:          uuid.New(),
		Year:             2025,
		Population:       1000,
		AvgTemperature:   20.5,
		MaxTemperature:   30.0,
		MinTemperature:   10.0,
		AvgElevation:     500.0,
		SeaLevel:         0.0,
		AtmosphereCarbon: 420.0,
		AtmosphereOxygen: 20.9,
		RecordedAt:       time.Now(),
	}

	mock.ExpectExec("INSERT INTO world_metrics").
		WithArgs(stats.RecordedAt, stats.WorldID, stats.Year, stats.Population,
			stats.AvgTemperature, stats.MaxTemperature, stats.MinTemperature,
			stats.AvgElevation, stats.SeaLevel,
			stats.AtmosphereCarbon, stats.AtmosphereOxygen).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = svc.RecordStats(context.Background(), stats)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordStats_CircuitBreaker(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cfg := circuitbreaker.DefaultConfig("test")
	cfg.FailureThreshold = 1
	cfg.Timeout = 100 * time.Millisecond
	cb := circuitbreaker.New(cfg)

	// Init
	mock.ExpectExec("CREATE EXTENSION").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE INDEX").WillReturnResult(sqlmock.NewResult(0, 0))

	svc, err := NewServiceWithDB(db, cb)
	require.NoError(t, err)

	// 1. Failure triggers CB Open
	mock.ExpectExec("INSERT INTO world_metrics").WillReturnError(errors.New("db down"))
	err = svc.RecordStats(context.Background(), ecosystem.GlobalStats{})
	require.Error(t, err)

	// 2. Next call fails fast (Circuit Open)
	err = svc.RecordStats(context.Background(), ecosystem.GlobalStats{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")

	// Circuit breaker stats check
	stats := svc.CircuitBreakerStats()
	assert.Equal(t, "open", stats.State)
}

func TestClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig("test"))

	mock.ExpectExec("CREATE EXTENSION").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE INDEX").WillReturnResult(sqlmock.NewResult(0, 0))

	svc, err := NewServiceWithDB(db, cb)
	require.NoError(t, err)

	mock.ExpectClose()
	err = svc.Close()
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
