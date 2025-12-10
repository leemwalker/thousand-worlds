package interview_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"tw-backend/internal/world/interview"
)

type RepositoryIntegrationSuite struct {
	suite.Suite
	db        *sql.DB
	pool      *pgxpool.Pool
	repo      *interview.PostgresRepository
	container testcontainers.Container
}

func (s *RepositoryIntegrationSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgis/postgis:15-3.3",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
		}).WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		s.T().Skipf("Skipping integration test: %v", err)
		return
	}
	s.container = container

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dbURL := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	s.db, err = sql.Open("postgres", dbURL)
	s.Require().NoError(err)

	err = s.db.Ping()
	s.Require().NoError(err, "Failed to ping database")

	// Enable PostGIS
	_, err = s.db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	s.Require().NoError(err, "Failed to enable PostGIS")

	// Initialize pgxpool
	s.pool, err = pgxpool.New(ctx, dbURL)
	s.Require().NoError(err, "Failed to connect to database with pgxpool")

	s.runMigrations()
	s.repo = interview.NewRepository(s.pool)
}

func (s *RepositoryIntegrationSuite) runMigrations() {
	files := []string{
		"000001_create_worlds_table.up.sql",
		"000013_create_auth_tables.up.sql",
		"000014_create_interview_tables.up.sql",
		"000015_add_character_role_and_appearance.up.sql",
		"000016_add_character_description_occupation.up.sql",
		"000017_add_performance_indexes.up.sql",
		"000018_create_lobby_world.up.sql",
		"000019_add_username_to_users.up.sql",
		"000020_add_world_name_to_configurations.up.sql",
		"000021_add_last_world_id_to_users.up.sql",
		"000022_create_events_table.up.sql",
		"000023_refactor_interview_tables.up.sql",
		"000024_add_owner_id_to_worlds.up.sql",
		"000025_add_is_system_world.up.sql",
		"000026_add_spatial_navigation_tables.up.sql",
		"000027_create_skills_tables.up.sql",
		"000028_add_character_orientation.up.sql",
		"000030_add_character_last_world_visited.up.sql",
		"000031_unique_world_name.up.sql",
	}

	// Try to find the migrations directory
	basePaths := []string{
		"../../../migrations/postgres",    // for internal/group/package
		"../../../../migrations/postgres", // for internal/group/subgroup/package
		"../../migrations/postgres",       // for internal/package
		"../migrations/postgres",          // for package
		"migrations/postgres",             // for root
	}

	var migrationsDir string
	for _, path := range basePaths {
		if _, err := os.Stat(path); err == nil {
			migrationsDir = path
			break
		}
	}

	if migrationsDir == "" {
		s.T().Fatal("Could not find migrations directory")
	}

	for _, file := range files {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			s.T().Logf("Warning: Skipping migration %s: %v", file, err)
			continue
		}
		_, err = s.db.Exec(string(content))
		if err != nil {
			s.T().Logf("Warning: Migration %s failed (might be applied): %v", file, err)
		}
	}
}

func (s *RepositoryIntegrationSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		s.container.Terminate(context.Background())
	}
}

func (s *RepositoryIntegrationSuite) SetupTest() {
	if s.pool == nil {
		s.T().Skip("Database not initialized")
	}
	// Truncate tables
	tables := []string{"world_configurations", "world_interviews", "worlds", "users"}
	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
}

func (s *RepositoryIntegrationSuite) TestCreateAndGetInterview() {
	ctx := context.Background()
	playerID := uuid.New()
	// Create user first (FK constraint)
	_, err := s.db.Exec("INSERT INTO users (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		playerID, "test@example.com", "hash", time.Now())
	s.Require().NoError(err)

	// Create Interview
	created, err := s.repo.CreateInterview(ctx, playerID)
	s.NoError(err)
	s.NotNil(created)
	s.Equal(playerID, created.UserID)
	s.Equal(interview.StatusNotStarted, created.Status)

	// Get Interview
	retrieved, err := s.repo.GetInterview(ctx, playerID)
	s.NoError(err)
	s.NotNil(retrieved)
	s.Equal(created.ID, retrieved.ID)
	s.Equal(playerID, retrieved.UserID)
}

func (s *RepositoryIntegrationSuite) TestUpdateInterviewFlow() {
	ctx := context.Background()
	playerID := uuid.New()
	_, err := s.db.Exec("INSERT INTO users (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		playerID, "update@example.com", "hash", time.Now())
	s.Require().NoError(err)

	// Create
	created, err := s.repo.CreateInterview(ctx, playerID)
	s.NoError(err)

	// Update Status
	err = s.repo.UpdateInterviewStatus(ctx, created.ID, interview.StatusInProgress)
	s.NoError(err)

	// Update Question Index
	err = s.repo.UpdateQuestionIndex(ctx, created.ID, 1)
	s.NoError(err)

	// Save Answer
	err = s.repo.SaveAnswer(ctx, created.ID, 0, "My Answer")
	s.NoError(err)

	// Verify
	retrieved, err := s.repo.GetInterview(ctx, playerID)
	s.NoError(err)
	s.Equal(interview.StatusInProgress, retrieved.Status)
	s.Equal(1, retrieved.CurrentQuestionIndex)

	// Verify Answers
	answers, err := s.repo.GetAnswers(ctx, created.ID)
	s.NoError(err)
	s.Len(answers, 1)
	s.Equal("My Answer", answers[0].AnswerText)
	s.Equal(0, answers[0].QuestionIndex)
}

func (s *RepositoryIntegrationSuite) TestSaveAndGetConfiguration() {
	ctx := context.Background()
	playerID := uuid.New()
	_, err := s.db.Exec("INSERT INTO users (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		playerID, "config@example.com", "hash", time.Now())
	s.Require().NoError(err)

	// Create interview
	interviewObj, err := s.repo.CreateInterview(ctx, playerID)
	s.Require().NoError(err)

	worldID := uuid.New()
	// Create world record to satisfy FK constraint
	_, err = s.db.Exec(`
		INSERT INTO worlds (id, name, owner_id, shape, created_at) 
		VALUES ($1, $2, $3, $4, $5)
	`, worldID, "Test World", playerID, "sphere", time.Now())
	s.Require().NoError(err, "Failed to create world for FK constraint")

	config := &interview.WorldConfiguration{
		WorldID:         &worldID,
		InterviewID:     interviewObj.ID,
		CreatedBy:       playerID,
		WorldName:       "Test World",
		Theme:           "Fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "medium",
		SentientSpecies: []string{"Human", "Elf"},
		BiomeWeights:    map[string]float64{"forest": 0.5},
		CreatedAt:       time.Now(),
	}

	// Save
	err = s.repo.SaveConfiguration(ctx, config)
	s.NoError(err)

	// Get by WorldID
	retrieved, err := s.repo.GetConfigurationByWorldID(ctx, worldID)
	s.NoError(err)
	s.Equal(config.Theme, retrieved.Theme)
	s.Equal(config.TechLevel, retrieved.TechLevel)
	s.Len(retrieved.SentientSpecies, 2)
	s.Equal(0.5, retrieved.BiomeWeights["forest"])
	s.Equal("Test World", retrieved.WorldName)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RepositoryIntegrationSuite))
}
