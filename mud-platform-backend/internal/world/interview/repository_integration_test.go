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
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"mud-platform-backend/internal/world/interview"
)

type RepositoryIntegrationSuite struct {
	suite.Suite
	db        *sql.DB
	repo      *interview.Repository
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

	s.runMigrations()
	s.repo = interview.NewRepository(s.db)
}

func (s *RepositoryIntegrationSuite) runMigrations() {
	// Adjust path to find migrations from internal/world/interview
	migrationsDir := "../../../migrations/postgres"

	files := []string{
		"000001_create_worlds_table.up.sql",
		"000013_create_auth_tables.up.sql",
		"000014_create_interview_tables.up.sql",
		"000015_add_world_name_to_configurations.up.sql",
	}

	for _, file := range files {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		s.Require().NoError(err, "Failed to read migration %s", file)
		_, err = s.db.Exec(string(content))
		s.Require().NoError(err, "Failed to execute migration %s", file)
	}
}

func (s *RepositoryIntegrationSuite) TearDownSuite() {
	if s.container != nil {
		s.container.Terminate(context.Background())
	}
}

func (s *RepositoryIntegrationSuite) SetupTest() {
	if s.db == nil {
		s.T().Skip("Database not initialized")
	}
	// Truncate tables
	tables := []string{"world_configurations", "world_interviews", "worlds", "users"}
	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
}

func (s *RepositoryIntegrationSuite) TestSaveAndGetInterview() {
	playerID := uuid.New()
	// Create user first (FK constraint)
	_, err := s.db.Exec("INSERT INTO users (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		playerID, "test@example.com", "hash", time.Now())
	s.Require().NoError(err)

	session := &interview.InterviewSession{
		ID:       uuid.New(),
		PlayerID: playerID,
		State: interview.InterviewState{
			CurrentCategory:   "Theme",
			CurrentTopicIndex: 0,
			Answers:           map[string]string{"Theme": "Fantasy"},
			IsComplete:        false,
		},
		History: []interview.ConversationTurn{
			{Question: "Q1", Answer: "A1"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save
	err = s.repo.SaveInterview(session)
	s.NoError(err)

	// Get
	retrieved, err := s.repo.GetInterview(session.ID)
	s.NoError(err)
	s.Equal(session.ID, retrieved.ID)
	s.Equal(session.PlayerID, retrieved.PlayerID)
	s.Equal("Fantasy", retrieved.State.Answers["Theme"])
	s.Len(retrieved.History, 1)
}

func (s *RepositoryIntegrationSuite) TestUpdateInterview() {
	playerID := uuid.New()
	_, err := s.db.Exec("INSERT INTO users (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		playerID, "update@example.com", "hash", time.Now())
	s.Require().NoError(err)

	session := &interview.InterviewSession{
		ID:       uuid.New(),
		PlayerID: playerID,
		State: interview.InterviewState{
			CurrentCategory:   "Theme",
			CurrentTopicIndex: 0,
			Answers:           make(map[string]string),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.SaveInterview(session)
	s.NoError(err)

	// Update
	session.State.Answers["Theme"] = "Sci-Fi"
	session.State.CurrentTopicIndex = 1
	err = s.repo.UpdateInterview(session)
	s.NoError(err)

	// Verify
	retrieved, err := s.repo.GetInterview(session.ID)
	s.NoError(err)
	s.Equal("Sci-Fi", retrieved.State.Answers["Theme"])
	s.Equal(1, retrieved.State.CurrentTopicIndex)
}

func (s *RepositoryIntegrationSuite) TestSaveAndGetConfiguration() {
	playerID := uuid.New()
	_, err := s.db.Exec("INSERT INTO users (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		playerID, "config@example.com", "hash", time.Now())
	s.Require().NoError(err)

	interviewID := uuid.New()
	// Create interview (FK constraint)
	_, err = s.db.Exec(`INSERT INTO world_interviews (id, player_id, current_category, current_topic_index, answers, history, is_complete, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		interviewID, playerID, "Done", 10, "{}", "[]", true, time.Now(), time.Now())
	s.Require().NoError(err)

	config := &interview.WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     interviewID,
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
	err = s.repo.SaveConfiguration(config)
	s.NoError(err)

	// Get
	retrieved, err := s.repo.GetConfiguration(config.ID)
	s.NoError(err)
	s.Equal(config.Theme, retrieved.Theme)
	s.Equal(config.TechLevel, retrieved.TechLevel)
	s.Len(retrieved.SentientSpecies, 2)
	s.Equal(0.5, retrieved.BiomeWeights["forest"])
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RepositoryIntegrationSuite))
}
