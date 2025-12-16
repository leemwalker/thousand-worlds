package ecosystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestNewSimulationLogger(t *testing.T) {
	worldID := uuid.New()

	t.Run("creates logger without file output", func(t *testing.T) {
		cfg := SimulationLoggerConfig{
			WorldID:    worldID,
			Verbosity:  LogLevelInfo,
			FileOutput: false,
		}

		logger, err := NewSimulationLogger(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		if logger.worldID != worldID {
			t.Errorf("World ID mismatch: got %v, want %v", logger.worldID, worldID)
		}
	})

	t.Run("creates logger with file output", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := SimulationLoggerConfig{
			WorldID:    worldID,
			LogDir:     tmpDir,
			Verbosity:  LogLevelDebug,
			FileOutput: true,
		}

		logger, err := NewSimulationLogger(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Check log file was created
		logPath := filepath.Join(tmpDir, "world_simulation.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Log file was not created")
		}
	})
}

func TestSimulationLogger_Verbosity(t *testing.T) {
	worldID := uuid.New()

	t.Run("respects verbosity level", func(t *testing.T) {
		cfg := SimulationLoggerConfig{
			WorldID:    worldID,
			Verbosity:  LogLevelWarn,
			FileOutput: false,
		}

		logger, err := NewSimulationLogger(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		// Trace should not log
		if logger.shouldLog(LogLevelTrace) {
			t.Error("Trace should not log when verbosity is Warn")
		}

		// Warn should log
		if !logger.shouldLog(LogLevelWarn) {
			t.Error("Warn should log when verbosity is Warn")
		}

		// Error should log
		if !logger.shouldLog(LogLevelError) {
			t.Error("Error should log when verbosity is Warn")
		}
	})

	t.Run("SetVerbosity changes level", func(t *testing.T) {
		cfg := SimulationLoggerConfig{
			WorldID:    worldID,
			Verbosity:  LogLevelError,
			FileOutput: false,
		}

		logger, err := NewSimulationLogger(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		if logger.shouldLog(LogLevelInfo) {
			t.Error("Info should not log when verbosity is Error")
		}

		logger.SetVerbosity(LogLevelInfo)

		if !logger.shouldLog(LogLevelInfo) {
			t.Error("Info should log after SetVerbosity to Info")
		}
	})
}

func TestSimulationLogger_ConvenienceMethods(t *testing.T) {
	worldID := uuid.New()
	ctx := context.Background()

	cfg := SimulationLoggerConfig{
		WorldID:    worldID,
		Verbosity:  LogLevelTrace,
		FileOutput: false,
	}

	logger, err := NewSimulationLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test convenience methods don't panic
	t.Run("LogSpeciation", func(t *testing.T) {
		logger.LogSpeciation(ctx, 1000000, "Ancestor", "NewSpecies", "rainforest", 0.35)
	})

	t.Run("LogExtinction", func(t *testing.T) {
		logger.LogExtinction(ctx, 2000000, "Dinosaurus Rex", "asteroid", "Great Asteroid Winter", 1000000)
	})

	t.Run("LogMassExtinction", func(t *testing.T) {
		logger.LogMassExtinction(ctx, 3000000, "asteroid_impact", 0.95, 500)
	})

	t.Run("LogDiseaseOutbreak", func(t *testing.T) {
		logger.LogDiseaseOutbreak(ctx, 4000000, "Red Plague", "bacteria", "Primate", 0.3)
	})

	t.Run("LogTurningPoint", func(t *testing.T) {
		logger.LogTurningPoint(ctx, 5000000, "interval", "boost_intelligence")
	})

	t.Run("LogSapienceDetected", func(t *testing.T) {
		logger.LogSapienceDetected(ctx, 6000000, "Homo Sapiens", 0.8, 0.7, 0.5, 0.6)
	})

	t.Run("LogCheckpoint", func(t *testing.T) {
		logger.LogCheckpoint(ctx, 7000000, "full", 100, 50000000)
	})

	t.Run("LogYearSummary", func(t *testing.T) {
		logger.LogYearSummary(ctx, 8000000, 100, 50000000, 1000000, 500000)
	})
}
