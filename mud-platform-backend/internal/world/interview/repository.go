package interview

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository handles database operations for interviews and configurations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveInterview creates a new interview in the database
func (r *Repository) SaveInterview(session *InterviewSession) error {
	answersJSON, err := json.Marshal(session.State.Answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}

	historyJSON, err := json.Marshal(session.History)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	query := `
		INSERT INTO world_interviews (
			id, player_id, current_category, current_topic_index,
			answers, history, is_complete, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Exec(query,
		session.ID,
		session.PlayerID,
		session.State.CurrentCategory,
		session.State.CurrentTopicIndex,
		answersJSON,
		historyJSON,
		session.State.IsComplete,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert interview: %w", err)
	}

	return nil
}

// GetInterview retrieves an interview by ID
func (r *Repository) GetInterview(id uuid.UUID) (*InterviewSession, error) {
	query := `
		SELECT id, player_id, current_category, current_topic_index,
		       answers, history, is_complete, created_at, updated_at
		FROM world_interviews
		WHERE id = $1
	`

	var session InterviewSession
	var answersJSON, historyJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.PlayerID,
		&session.State.CurrentCategory,
		&session.State.CurrentTopicIndex,
		&answersJSON,
		&historyJSON,
		&session.State.IsComplete,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("interview not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query interview: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(answersJSON, &session.State.Answers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal answers: %w", err)
	}

	if err := json.Unmarshal(historyJSON, &session.History); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return &session, nil
}

// UpdateInterview updates an existing interview
func (r *Repository) UpdateInterview(session *InterviewSession) error {
	answersJSON, err := json.Marshal(session.State.Answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}

	historyJSON, err := json.Marshal(session.History)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	session.UpdatedAt = time.Now()

	query := `
		UPDATE world_interviews
		SET current_category = $1,
		    current_topic_index = $2,
		    answers = $3,
		    history = $4,
		    is_complete = $5,
		    updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(query,
		session.State.CurrentCategory,
		session.State.CurrentTopicIndex,
		answersJSON,
		historyJSON,
		session.State.IsComplete,
		session.UpdatedAt,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update interview: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("interview not found")
	}

	return nil
}

// GetActiveInterviewByPlayer retrieves an active (incomplete) interview for a player
func (r *Repository) GetActiveInterviewByPlayer(playerID uuid.UUID) (*InterviewSession, error) {
	query := `
		SELECT id, player_id, current_category, current_topic_index,
		       answers, history, is_complete, created_at, updated_at
		FROM world_interviews
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var session InterviewSession
	var answersJSON, historyJSON []byte

	err := r.db.QueryRow(query, playerID).Scan(
		&session.ID,
		&session.PlayerID,
		&session.State.CurrentCategory,
		&session.State.CurrentTopicIndex,
		&answersJSON,
		&historyJSON,
		&session.State.IsComplete,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No active interview, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query active interview: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(answersJSON, &session.State.Answers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal answers: %w", err)
	}

	if err := json.Unmarshal(historyJSON, &session.History); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return &session, nil
}

// SaveConfiguration saves a world configuration to the database
func (r *Repository) SaveConfiguration(config *WorldConfiguration) error {
	// Marshal JSON fields
	inspirationsJSON, _ := json.Marshal(config.Inspirations)
	conflictsJSON, _ := json.Marshal(config.MajorConflicts)
	featuresJSON, _ := json.Marshal(config.UniqueFeatures)
	extremeEnvJSON, _ := json.Marshal(config.ExtremeEnvironments)
	speciesJSON, _ := json.Marshal(config.SentientSpecies)
	valuesJSON, _ := json.Marshal(config.CulturalValues)
	religionsJSON, _ := json.Marshal(config.Religions)
	taboosJSON, _ := json.Marshal(config.Taboos)
	biomeWeightsJSON, _ := json.Marshal(config.BiomeWeights)
	resourceDistJSON, _ := json.Marshal(config.ResourceDistribution)
	speciesAttrsJSON, _ := json.Marshal(config.SpeciesStartAttributes)

	query := `
		INSERT INTO world_configurations (
			id, interview_id, world_id, created_by,
			world_name,
			theme, tone, inspirations, unique_aspect, major_conflicts,
			tech_level, magic_level, advanced_tech, magic_impact,
			planet_size, climate_range, land_water_ratio, unique_features, extreme_environments,
			sentient_species, political_structure, cultural_values, economic_system, religions, taboos,
			biome_weights, resource_distribution, species_start_attributes,
			created_at
		) VALUES (
			$1, $2, $3, $4,
			$5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25,
			$26, $27, $28,
			$29
		)
	`

	_, err := r.db.Exec(query,
		config.ID, config.InterviewID, config.WorldID, config.CreatedBy,
		config.WorldName,
		config.Theme, config.Tone, inspirationsJSON, config.UniqueAspect, conflictsJSON,
		config.TechLevel, config.MagicLevel, config.AdvancedTech, config.MagicImpact,
		config.PlanetSize, config.ClimateRange, config.LandWaterRatio, featuresJSON, extremeEnvJSON,
		speciesJSON, config.PoliticalStructure, valuesJSON, config.EconomicSystem, religionsJSON, taboosJSON,
		biomeWeightsJSON, resourceDistJSON, speciesAttrsJSON,
		config.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert configuration: %w", err)
	}

	return nil
}

// GetConfiguration retrieves a configuration by ID
func (r *Repository) GetConfiguration(id uuid.UUID) (*WorldConfiguration, error) {
	query := `
		SELECT id, interview_id, world_id, created_by,
		       world_name,
		       theme, tone, inspirations, unique_aspect, major_conflicts,
		       tech_level, magic_level, advanced_tech, magic_impact,
		       planet_size, climate_range, land_water_ratio, unique_features, extreme_environments,
		       sentient_species, political_structure, cultural_values, economic_system, religions, taboos,
		       biome_weights, resource_distribution, species_start_attributes,
		       created_at
		FROM world_configurations
		WHERE id = $1
	`

	return r.scanConfiguration(r.db.QueryRow(query, id))
}

// GetConfigurationByInterview retrieves a configuration by interview ID
func (r *Repository) GetConfigurationByInterview(interviewID uuid.UUID) (*WorldConfiguration, error) {
	query := `
		SELECT id, interview_id, world_id, created_by,
		       world_name,
		       theme, tone, inspirations, unique_aspect, major_conflicts,
		       tech_level, magic_level, advanced_tech, magic_impact,
		       planet_size, climate_range, land_water_ratio, unique_features, extreme_environments,
		       sentient_species, political_structure, cultural_values, economic_system, religions, taboos,
		       biome_weights, resource_distribution, species_start_attributes,
		       created_at
		FROM world_configurations
		WHERE interview_id = $1
	`

	return r.scanConfiguration(r.db.QueryRow(query, interviewID))
}

// GetConfigurationByWorldID retrieves a configuration by world ID
func (r *Repository) GetConfigurationByWorldID(worldID uuid.UUID) (*WorldConfiguration, error) {
	query := `
		SELECT id, interview_id, world_id, created_by,
		       world_name,
		       theme, tone, inspirations, unique_aspect, major_conflicts,
		       tech_level, magic_level, advanced_tech, magic_impact,
		       planet_size, climate_range, land_water_ratio, unique_features, extreme_environments,
		       sentient_species, political_structure, cultural_values, economic_system, religions, taboos,
		       biome_weights, resource_distribution, species_start_attributes,
		       created_at
		FROM world_configurations
		WHERE world_id = $1
	`

	return r.scanConfiguration(r.db.QueryRow(query, worldID))
}

// scanConfiguration is a helper to scan a configuration from a row
func (r *Repository) scanConfiguration(row *sql.Row) (*WorldConfiguration, error) {
	var config WorldConfiguration
	var worldID sql.NullString
	var inspirationsJSON, conflictsJSON, featuresJSON, extremeEnvJSON []byte
	var speciesJSON, valuesJSON, religionsJSON, taboosJSON []byte
	var biomeWeightsJSON, resourceDistJSON, speciesAttrsJSON []byte

	err := row.Scan(
		&config.ID, &config.InterviewID, &worldID, &config.CreatedBy,
		&config.WorldName,
		&config.Theme, &config.Tone, &inspirationsJSON, &config.UniqueAspect, &conflictsJSON,
		&config.TechLevel, &config.MagicLevel, &config.AdvancedTech, &config.MagicImpact,
		&config.PlanetSize, &config.ClimateRange, &config.LandWaterRatio, &featuresJSON, &extremeEnvJSON,
		&speciesJSON, &config.PoliticalStructure, &valuesJSON, &config.EconomicSystem, &religionsJSON, &taboosJSON,
		&biomeWeightsJSON, &resourceDistJSON, &speciesAttrsJSON,
		&config.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("configuration not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan configuration: %w", err)
	}

	// Handle nullable WorldID
	if worldID.Valid {
		parsed, _ := uuid.Parse(worldID.String)
		config.WorldID = &parsed
	}

	// Unmarshal JSON fields
	json.Unmarshal(inspirationsJSON, &config.Inspirations)
	json.Unmarshal(conflictsJSON, &config.MajorConflicts)
	json.Unmarshal(featuresJSON, &config.UniqueFeatures)
	json.Unmarshal(extremeEnvJSON, &config.ExtremeEnvironments)
	json.Unmarshal(speciesJSON, &config.SentientSpecies)
	json.Unmarshal(valuesJSON, &config.CulturalValues)
	json.Unmarshal(religionsJSON, &config.Religions)
	json.Unmarshal(taboosJSON, &config.Taboos)
	json.Unmarshal(biomeWeightsJSON, &config.BiomeWeights)
	json.Unmarshal(resourceDistJSON, &config.ResourceDistribution)
	json.Unmarshal(speciesAttrsJSON, &config.SpeciesStartAttributes)

	return &config, nil
}

// GetSessionByID retrieves a session by ID (interface method)
// Note: Interface uses string ID, but InterviewSession stores uuid.UUID
func (r *Repository) GetSessionByID(sessionID string) (*InterviewSession, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}
	return r.GetInterview(id)
}

// GetActiveSessionForUser retrieves active session for a user (interface method)
func (r *Repository) GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*InterviewSession, error) {
	return r.GetActiveInterviewByPlayer(userID)
}

// SaveSession saves a session (interface method)
func (r *Repository) SaveSession(session *InterviewSession) error {
	// Check if session exists - ID is already a UUID
	existing, _ := r.GetInterview(session.ID)
	if existing != nil {
		return r.UpdateInterview(session)
	}
	return r.SaveInterview(session)
}

// IsWorldNameTaken checks if a world name already exists (case-insensitive)
func (r *Repository) IsWorldNameTaken(name string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM world_configurations WHERE LOWER(world_name) = LOWER($1)`

	err := r.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check world name: %w", err)
	}

	return count > 0, nil
}
