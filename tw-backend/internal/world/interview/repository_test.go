package interview

import (
	"context"
	"testing"

	"mud-platform-backend/internal/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *PostgresRepository {
	// We need a pgxpool.Pool, but testutil provides *sql.DB
	// So we need to create a pool manually using the same connection string

	// Note: In a real environment we would update testutil to support pgxpool
	// For now, we'll create a pool here.

	// Get connection string from env or default
	dbURL := "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	// Clean up
	t.Cleanup(func() {
		pool.Close()
	})

	// Run migrations using the sql.DB from testutil (it's easier)
	sqlDB := testutil.SetupTestDB(t)
	testutil.RunMigrations(t, sqlDB)
	testutil.TruncateTables(t, sqlDB)
	testutil.CloseDB(t, sqlDB)

	return NewRepository(pool)
}

func TestCreateInterview(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	// Create user first
	_, err := repo.db.Exec(ctx, "INSERT INTO users (user_id, email, password_hash, username) VALUES ($1, $2, $3, $4)", userID, "test@example.com", "hash", "testuser")
	require.NoError(t, err)

	interview, err := repo.CreateInterview(ctx, userID)
	require.NoError(t, err)
	assert.NotNil(t, interview)
	assert.Equal(t, userID, interview.UserID)
	assert.Equal(t, StatusNotStarted, interview.Status)
	assert.Equal(t, 0, interview.CurrentQuestionIndex)
}

func TestGetInterview(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	_, err := repo.db.Exec(ctx, "INSERT INTO users (user_id, email, password_hash, username) VALUES ($1, $2, $3, $4)", userID, "test@example.com", "hash", "testuser")
	require.NoError(t, err)

	// Should return nil if not found
	i, err := repo.GetInterview(ctx, userID)
	require.NoError(t, err)
	assert.Nil(t, i)

	// Create and get
	created, err := repo.CreateInterview(ctx, userID)
	require.NoError(t, err)

	fetched, err := repo.GetInterview(ctx, userID)
	require.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
}

func TestUpdateInterviewStatus(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	_, err := repo.db.Exec(ctx, "INSERT INTO users (user_id, email, password_hash, username) VALUES ($1, $2, $3, $4)", userID, "test@example.com", "hash", "testuser")
	require.NoError(t, err)

	interview, err := repo.CreateInterview(ctx, userID)
	require.NoError(t, err)

	err = repo.UpdateInterviewStatus(ctx, interview.ID, StatusInProgress)
	require.NoError(t, err)

	fetched, err := repo.GetInterview(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, StatusInProgress, fetched.Status)
}

func TestUpdateQuestionIndex(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	_, err := repo.db.Exec(ctx, "INSERT INTO users (user_id, email, password_hash, username) VALUES ($1, $2, $3, $4)", userID, "test@example.com", "hash", "testuser")
	require.NoError(t, err)

	interview, err := repo.CreateInterview(ctx, userID)
	require.NoError(t, err)

	err = repo.UpdateQuestionIndex(ctx, interview.ID, 5)
	require.NoError(t, err)

	fetched, err := repo.GetInterview(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 5, fetched.CurrentQuestionIndex)
}

func TestSaveAndGetAnswers(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	_, err := repo.db.Exec(ctx, "INSERT INTO users (user_id, email, password_hash, username) VALUES ($1, $2, $3, $4)", userID, "test@example.com", "hash", "testuser")
	require.NoError(t, err)

	interview, err := repo.CreateInterview(ctx, userID)
	require.NoError(t, err)

	// Save answers
	err = repo.SaveAnswer(ctx, interview.ID, 0, "Fantasy")
	require.NoError(t, err)
	err = repo.SaveAnswer(ctx, interview.ID, 1, "Forests")
	require.NoError(t, err)

	// Update an answer
	err = repo.SaveAnswer(ctx, interview.ID, 0, "Sci-Fi")
	require.NoError(t, err)

	// Get answers
	answers, err := repo.GetAnswers(ctx, interview.ID)
	require.NoError(t, err)
	assert.Len(t, answers, 2)

	assert.Equal(t, 0, answers[0].QuestionIndex)
	assert.Equal(t, "Sci-Fi", answers[0].AnswerText) // Should be updated
	assert.Equal(t, 1, answers[1].QuestionIndex)
	assert.Equal(t, "Forests", answers[1].AnswerText)
}
