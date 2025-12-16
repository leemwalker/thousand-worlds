// Package ecosystem provides the simulation logger for world ecosystem events.
// This logger provides dual output (file + database) specifically for simulation events
// that may need to be queried later for debugging, rewinding, or player-facing history.
package ecosystem

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// LogLevel represents the verbosity level for simulation logging
type LogLevel int

const (
	// LogLevelTrace logs every year's calculations (development only)
	LogLevelTrace LogLevel = iota
	// LogLevelDebug logs speciation, extinction, disease outbreaks (development only)
	LogLevelDebug
	// LogLevelInfo logs major events (mass extinctions, turning points) - production
	LogLevelInfo
	// LogLevelWarn logs simulation anomalies - production
	LogLevelWarn
	// LogLevelError logs failures - production
	LogLevelError
)

// SimulationEventType categorizes simulation events for querying
type SimulationEventType string

const (
	EventTypeSpeciation     SimulationEventType = "speciation"
	EventTypeExtinction     SimulationEventType = "extinction"
	EventTypeMassExtinction SimulationEventType = "mass_extinction"
	EventTypeDisease        SimulationEventType = "disease_outbreak"
	EventTypeMigration      SimulationEventType = "migration"
	EventTypeTurningPoint   SimulationEventType = "turning_point"
	EventTypeBiomeShift     SimulationEventType = "biome_shift"
	EventTypeTectonic       SimulationEventType = "tectonic"
	EventTypeClimate        SimulationEventType = "climate"
	EventTypeSapience       SimulationEventType = "sapience_detected"
	EventTypeCheckpoint     SimulationEventType = "checkpoint"
	EventTypeYearTick       SimulationEventType = "year_tick"

	// V2 Event Types
	EventTypeCascade           SimulationEventType = "extinction_cascade"
	EventTypePhylogeny         SimulationEventType = "phylogeny_update"
	EventTypeKeystoneEffect    SimulationEventType = "keystone_effect"
	EventTypeRegionIsolation   SimulationEventType = "region_isolation"
	EventTypeEndemicEvolution  SimulationEventType = "endemic_evolution"
	EventTypeZoonoticJump      SimulationEventType = "zoonotic_jump"
	EventTypeProtoSapience     SimulationEventType = "proto_sapience"
	EventTypeSapienceAchieved  SimulationEventType = "sapience_achieved"
	EventTypeIntervention      SimulationEventType = "intervention"
	EventTypeAdaptiveRadiation SimulationEventType = "adaptive_radiation"
)

// SimulationEvent represents a logged simulation event
type SimulationEvent struct {
	ID        uuid.UUID           `json:"id"`
	WorldID   uuid.UUID           `json:"world_id"`
	Year      int64               `json:"year"`
	EventType SimulationEventType `json:"event_type"`
	Severity  float64             `json:"severity"` // 0.0-1.0 for importance
	Details   json.RawMessage     `json:"details"`
	Timestamp time.Time           `json:"timestamp"`
}

// DBEventLogger interface for database logging (to be implemented by repository)
type DBEventLogger interface {
	LogEvent(ctx context.Context, event *SimulationEvent) error
	GetEvents(ctx context.Context, worldID uuid.UUID, fromYear, toYear int64) ([]SimulationEvent, error)
	GetEventsByType(ctx context.Context, worldID uuid.UUID, eventType SimulationEventType) ([]SimulationEvent, error)
}

// SimulationLogger provides dual-output logging for simulation events
type SimulationLogger struct {
	worldID    uuid.UUID
	fileLogger zerolog.Logger
	dbLogger   DBEventLogger // nil if no DB configured
	verbosity  LogLevel
	mu         sync.Mutex
	file       *os.File
}

// SimulationLoggerConfig holds configuration for creating a simulation logger
type SimulationLoggerConfig struct {
	WorldID    uuid.UUID
	LogDir     string   // Directory for log files, defaults to "logs"
	Verbosity  LogLevel // Minimum level to log
	DBLogger   DBEventLogger
	FileOutput bool // Whether to write to file
}

// NewSimulationLogger creates a new simulation logger with dual output
func NewSimulationLogger(cfg SimulationLoggerConfig) (*SimulationLogger, error) {
	sl := &SimulationLogger{
		worldID:   cfg.WorldID,
		dbLogger:  cfg.DBLogger,
		verbosity: cfg.Verbosity,
	}

	if cfg.FileOutput {
		// Ensure log directory exists
		logDir := cfg.LogDir
		if logDir == "" {
			logDir = "logs"
		}
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		// Create or open log file
		logPath := filepath.Join(logDir, "world_simulation.log")
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		sl.file = file

		// Create zerolog writer with both console and file output
		multi := io.MultiWriter(os.Stderr, file)
		sl.fileLogger = zerolog.New(multi).With().
			Timestamp().
			Str("world_id", cfg.WorldID.String()).
			Logger()
	} else {
		// Console only
		sl.fileLogger = zerolog.New(os.Stderr).With().
			Timestamp().
			Str("world_id", cfg.WorldID.String()).
			Logger()
	}

	return sl, nil
}

// Close closes the log file if open
func (sl *SimulationLogger) Close() error {
	if sl.file != nil {
		return sl.file.Close()
	}
	return nil
}

// SetVerbosity changes the logging verbosity level
func (sl *SimulationLogger) SetVerbosity(level LogLevel) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.verbosity = level
}

// shouldLog returns true if the given level should be logged
func (sl *SimulationLogger) shouldLog(level LogLevel) bool {
	return level >= sl.verbosity
}

// logEvent logs to both file and database
func (sl *SimulationLogger) logEvent(ctx context.Context, level LogLevel, year int64, eventType SimulationEventType, severity float64, message string, details map[string]interface{}) {
	if !sl.shouldLog(level) {
		return
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Create event for DB
	detailsJSON, _ := json.Marshal(details)
	event := &SimulationEvent{
		ID:        uuid.New(),
		WorldID:   sl.worldID,
		Year:      year,
		EventType: eventType,
		Severity:  severity,
		Details:   detailsJSON,
		Timestamp: time.Now(),
	}

	// Log to file
	var logEvent *zerolog.Event
	switch level {
	case LogLevelTrace:
		logEvent = sl.fileLogger.Trace()
	case LogLevelDebug:
		logEvent = sl.fileLogger.Debug()
	case LogLevelInfo:
		logEvent = sl.fileLogger.Info()
	case LogLevelWarn:
		logEvent = sl.fileLogger.Warn()
	case LogLevelError:
		logEvent = sl.fileLogger.Error()
	default:
		logEvent = sl.fileLogger.Info()
	}

	logEvent.
		Int64("year", year).
		Str("event_type", string(eventType)).
		Float64("severity", severity).
		Interface("details", details).
		Msg(message)

	// Log to database if available
	if sl.dbLogger != nil {
		// Fire and forget - don't block simulation for DB writes
		go func(e *SimulationEvent) {
			if err := sl.dbLogger.LogEvent(ctx, e); err != nil {
				sl.fileLogger.Error().Err(err).Msg("Failed to log event to database")
			}
		}(event)
	}
}

// Trace logs trace-level events (every year's calculations)
func (sl *SimulationLogger) Trace(ctx context.Context, year int64, eventType SimulationEventType, message string, details map[string]interface{}) {
	sl.logEvent(ctx, LogLevelTrace, year, eventType, 0.1, message, details)
}

// Debug logs debug-level events (speciation, extinction, disease outbreaks)
func (sl *SimulationLogger) Debug(ctx context.Context, year int64, eventType SimulationEventType, message string, details map[string]interface{}) {
	sl.logEvent(ctx, LogLevelDebug, year, eventType, 0.3, message, details)
}

// Info logs info-level events (major events like mass extinctions, turning points)
func (sl *SimulationLogger) Info(ctx context.Context, year int64, eventType SimulationEventType, message string, details map[string]interface{}) {
	sl.logEvent(ctx, LogLevelInfo, year, eventType, 0.5, message, details)
}

// Warn logs warning-level events (simulation anomalies)
func (sl *SimulationLogger) Warn(ctx context.Context, year int64, eventType SimulationEventType, message string, details map[string]interface{}) {
	sl.logEvent(ctx, LogLevelWarn, year, eventType, 0.7, message, details)
}

// Error logs error-level events (failures)
func (sl *SimulationLogger) Error(ctx context.Context, year int64, eventType SimulationEventType, err error, message string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()
	sl.logEvent(ctx, LogLevelError, year, eventType, 1.0, message, details)
}

// --- Convenience methods for common events ---

// LogSpeciation logs a speciation event
func (sl *SimulationLogger) LogSpeciation(ctx context.Context, year int64, parentSpecies, newSpecies string, biome string, geneticDistance float64) {
	sl.Debug(ctx, year, EventTypeSpeciation, "New species emerged", map[string]interface{}{
		"parent_species":   parentSpecies,
		"new_species":      newSpecies,
		"biome":            biome,
		"genetic_distance": geneticDistance,
	})
}

// LogExtinction logs an extinction event
func (sl *SimulationLogger) LogExtinction(ctx context.Context, year int64, species string, cause string, details string, peakPopulation int64) {
	sl.Debug(ctx, year, EventTypeExtinction, "Species went extinct", map[string]interface{}{
		"species":         species,
		"cause":           cause,
		"details":         details,
		"peak_population": peakPopulation,
	})
}

// LogMassExtinction logs a mass extinction event
func (sl *SimulationLogger) LogMassExtinction(ctx context.Context, year int64, extinctionType string, severity float64, speciesLost int) {
	sl.Info(ctx, year, EventTypeMassExtinction, "Mass extinction event", map[string]interface{}{
		"extinction_type": extinctionType,
		"severity":        severity,
		"species_lost":    speciesLost,
	})
}

// LogDiseaseOutbreak logs a disease outbreak
func (sl *SimulationLogger) LogDiseaseOutbreak(ctx context.Context, year int64, pathogenName, pathogenType string, hostSpecies string, mortality float64) {
	sl.Debug(ctx, year, EventTypeDisease, "Disease outbreak", map[string]interface{}{
		"pathogen_name": pathogenName,
		"pathogen_type": pathogenType,
		"host_species":  hostSpecies,
		"mortality":     mortality,
	})
}

// LogTurningPoint logs a turning point event
func (sl *SimulationLogger) LogTurningPoint(ctx context.Context, year int64, trigger string, chosenOption string) {
	sl.Info(ctx, year, EventTypeTurningPoint, "Turning point reached", map[string]interface{}{
		"trigger":       trigger,
		"chosen_option": chosenOption,
	})
}

// LogSapienceDetected logs when a species reaches sapience threshold
func (sl *SimulationLogger) LogSapienceDetected(ctx context.Context, year int64, species string, intelligence, social, toolUse, communication float64) {
	sl.Info(ctx, year, EventTypeSapience, "Proto-sapient species detected", map[string]interface{}{
		"species":       species,
		"intelligence":  intelligence,
		"social":        social,
		"tool_use":      toolUse,
		"communication": communication,
	})
}

// LogCheckpoint logs a checkpoint save
func (sl *SimulationLogger) LogCheckpoint(ctx context.Context, year int64, checkpointType string, speciesCount int, populationCount int64) {
	sl.Info(ctx, year, EventTypeCheckpoint, "Checkpoint saved", map[string]interface{}{
		"checkpoint_type":  checkpointType,
		"species_count":    speciesCount,
		"population_count": populationCount,
	})
}

// LogYearSummary logs a summary of a simulated year (trace level)
func (sl *SimulationLogger) LogYearSummary(ctx context.Context, year int64, speciesCount int, totalPopulation int64, births, deaths int64) {
	sl.Trace(ctx, year, EventTypeYearTick, "Year simulated", map[string]interface{}{
		"species_count":    speciesCount,
		"total_population": totalPopulation,
		"births":           births,
		"deaths":           deaths,
	})
}

// --- V2 Feature Logging Methods ---

// LogSpeciationV2 logs a speciation event with V2 details (speciation type, genetic distance)
func (sl *SimulationLogger) LogSpeciationV2(ctx context.Context, year int64, parentSpecies, newSpecies, speciationType string, geneticDistance float64, regionID string, isMagicAssisted bool) {
	sl.Debug(ctx, year, EventTypeSpeciation, "Species speciation event", map[string]interface{}{
		"parent_species":    parentSpecies,
		"new_species":       newSpecies,
		"speciation_type":   speciationType, // "allopatric", "sympatric", "peripatric"
		"genetic_distance":  geneticDistance,
		"region_id":         regionID,
		"is_magic_assisted": isMagicAssisted,
	})
}

// LogExtinctionCascade logs an extinction cascade event
func (sl *SimulationLogger) LogExtinctionCascade(ctx context.Context, year int64, triggerSpecies string, cascadeType string, affectedCount int, secondaryExtinctions int) {
	sl.Info(ctx, year, EventTypeCascade, "Extinction cascade triggered", map[string]interface{}{
		"trigger_species":       triggerSpecies,
		"cascade_type":          cascadeType, // "co_extinction", "food_loss", "keystone"
		"affected_species":      affectedCount,
		"secondary_extinctions": secondaryExtinctions,
	})
}

// LogKeystoneExtinction logs when a keystone species goes extinct
func (sl *SimulationLogger) LogKeystoneExtinction(ctx context.Context, year int64, species string, importance float64, ecosystemImpact string) {
	sl.Info(ctx, year, EventTypeKeystoneEffect, "Keystone species extinction", map[string]interface{}{
		"species":             species,
		"keystone_importance": importance,
		"ecosystem_impact":    ecosystemImpact,
	})
}

// LogRegionIsolation logs when a region becomes isolated
func (sl *SimulationLogger) LogRegionIsolation(ctx context.Context, year int64, regionName string, isolationType string, speciesCount int, isIsland bool) {
	sl.Debug(ctx, year, EventTypeRegionIsolation, "Region isolation detected", map[string]interface{}{
		"region_name":    regionName,
		"isolation_type": isolationType, // "mountain", "ocean", "desert"
		"species_count":  speciesCount,
		"is_island":      isIsland,
	})
}

// LogPathogenOutbreakV2 logs a disease outbreak with V2 details
func (sl *SimulationLogger) LogPathogenOutbreakV2(ctx context.Context, year int64, pathogenName, pathogenType, transmission string, hostSpecies string, r0 float32, virulence float32, peakInfected int64) {
	sl.Debug(ctx, year, EventTypeDisease, "Pathogen outbreak", map[string]interface{}{
		"pathogen_name": pathogenName,
		"pathogen_type": pathogenType, // "virus", "bacteria", "fungus", "prion", "parasite"
		"transmission":  transmission, // "airborne", "contact", "vector", "water"
		"host_species":  hostSpecies,
		"r0":            r0,
		"virulence":     virulence,
		"peak_infected": peakInfected,
	})
}

// LogEndemicEvolution logs when a pathogen becomes endemic
func (sl *SimulationLogger) LogEndemicEvolution(ctx context.Context, year int64, pathogenName string, initialVirulence, finalVirulence float32, mutationCount int) {
	sl.Info(ctx, year, EventTypeEndemicEvolution, "Pathogen became endemic", map[string]interface{}{
		"pathogen_name":     pathogenName,
		"initial_virulence": initialVirulence,
		"final_virulence":   finalVirulence,
		"mutation_count":    mutationCount,
		"virulence_change":  initialVirulence - finalVirulence,
	})
}

// LogZoonoticJump logs when a pathogen jumps to a new host species
func (sl *SimulationLogger) LogZoonoticJump(ctx context.Context, year int64, pathogenName string, sourceSpecies, targetSpecies string, newHostSpecificity float32) {
	sl.Info(ctx, year, EventTypeZoonoticJump, "Zoonotic disease jump", map[string]interface{}{
		"pathogen_name":        pathogenName,
		"source_species":       sourceSpecies,
		"target_species":       targetSpecies,
		"new_host_specificity": newHostSpecificity,
	})
}

// LogProtoSapience logs when a species shows proto-sapient traits
func (sl *SimulationLogger) LogProtoSapience(ctx context.Context, year int64, species string, intelligence, toolUse, communication float64, sapienceScore float64) {
	sl.Info(ctx, year, EventTypeProtoSapience, "Proto-sapience detected", map[string]interface{}{
		"species":        species,
		"intelligence":   intelligence,
		"tool_use":       toolUse,
		"communication":  communication,
		"sapience_score": sapienceScore,
	})
}

// LogSapienceAchieved logs when a species achieves full sapience
func (sl *SimulationLogger) LogSapienceAchieved(ctx context.Context, year int64, species string, isMagicAssisted bool, intelligence, social, toolUse, communication, magicAffinity float64) {
	sl.Info(ctx, year, EventTypeSapienceAchieved, "SAPIENCE ACHIEVED", map[string]interface{}{
		"species":           species,
		"is_magic_assisted": isMagicAssisted,
		"intelligence":      intelligence,
		"social":            social,
		"tool_use":          toolUse,
		"communication":     communication,
		"magic_affinity":    magicAffinity,
	})
}

// LogIntervention logs a player intervention at a turning point
func (sl *SimulationLogger) LogIntervention(ctx context.Context, year int64, interventionName, interventionType string, targetType, targetName string, cost int) {
	sl.Info(ctx, year, EventTypeIntervention, "Player intervention", map[string]interface{}{
		"intervention_name": interventionName,
		"intervention_type": interventionType, // "nudge", "protection", "cataclysm", "magic"
		"target_type":       targetType,       // "species", "biome", "region"
		"target_name":       targetName,
		"cost":              cost,
	})
}

// LogAdaptiveRadiation logs the start of an adaptive radiation period
func (sl *SimulationLogger) LogAdaptiveRadiation(ctx context.Context, year int64, triggerEvent string, radiationBonus float64, affectedRegions []string) {
	sl.Info(ctx, year, EventTypeAdaptiveRadiation, "Adaptive radiation begun", map[string]interface{}{
		"trigger_event":    triggerEvent, // "post_extinction", "new_niche", "boundary_change"
		"radiation_bonus":  radiationBonus,
		"affected_regions": affectedRegions,
	})
}

// LogPhylogenyMilestone logs phylogenetic tree milestones
func (sl *SimulationLogger) LogPhylogenyMilestone(ctx context.Context, year int64, milestone string, totalSpecies, maxDepth int, longestLineageYears int64) {
	sl.Debug(ctx, year, EventTypePhylogeny, "Phylogenetic milestone", map[string]interface{}{
		"milestone":             milestone, // "100_species", "first_branch", "max_depth_10"
		"total_species":         totalSpecies,
		"max_tree_depth":        maxDepth,
		"longest_lineage_years": longestLineageYears,
	})
}

// LogPandemic logs when an outbreak reaches pandemic severity
func (sl *SimulationLogger) LogPandemic(ctx context.Context, year int64, pathogenName string, hostSpecies string, totalInfected, totalDeaths int64, affectedBiomes int) {
	sl.Info(ctx, year, EventTypeDisease, "PANDEMIC", map[string]interface{}{
		"pathogen_name":   pathogenName,
		"host_species":    hostSpecies,
		"total_infected":  totalInfected,
		"total_deaths":    totalDeaths,
		"affected_biomes": affectedBiomes,
		"severity":        "pandemic",
	})
}
