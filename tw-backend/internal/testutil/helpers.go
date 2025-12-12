package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"tw-backend/internal/auth"
)

// CreateTestUser creates a minimal user for testing purposes
func CreateTestUser(t *testing.T, repo auth.Repository) *auth.User {
	t.Helper()

	user := &auth.User{
		UserID:    uuid.New(),
		Email:     GenerateTestEmail(),
		Username:  "TestUser" + uuid.New().String()[:8],
		CreatedAt: time.Now(),
	}

	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err, "Failed to create test user")

	return user
}

// CreateTestWorld creates a minimal world record for testing
func CreateTestWorld(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()

	// Create a dummy owner for the world
	ownerID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO users (user_id, email, password_hash, username, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, ownerID, "worldowner"+ownerID.String()[:8]+"@test.com", "hash", "Owner"+ownerID.String()[:8])
	require.NoError(t, err, "Failed to create world owner")

	worldID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO worlds (id, name, shape, created_at, owner_id)
		VALUES ($1, $2, $3, NOW(), $4)
	`, worldID, "Test World "+uuid.New().String()[:8], "sphere", ownerID)
	require.NoError(t, err, "Failed to create test world")

	return worldID
}

// CreateUserAndLogin creates a user, logs them in, and returns the auth token
func CreateUserAndLogin(t *testing.T, baseURL string) (email string, token string) {
	t.Helper()

	email = GenerateTestEmail()
	password := "TestPassword123"

	// Register
	registerReq := map[string]string{
		"email":    email,
		"username": "TestUser" + uuid.New().String()[:8],
		"password": password,
	}
	registerResp := PostJSON(t, http.DefaultClient, baseURL+"/api/auth/register", registerReq)
	require.Equal(t, 201, registerResp.StatusCode, "Failed to register user")
	registerResp.Body.Close()

	// Login
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	loginResp := PostJSON(t, http.DefaultClient, baseURL+"/api/auth/login", loginReq)
	require.Equal(t, 200, loginResp.StatusCode, "Failed to login")

	var loginData map[string]interface{}
	DecodeJSON(t, loginResp, &loginData)

	token, ok := loginData["token"].(string)
	require.True(t, ok, "Token should be a string")
	require.NotEmpty(t, token, "Token should not be empty")

	return email, token
}

// AssertErrorResponse asserts that response contains error with expected code
func AssertErrorResponse(t *testing.T, resp *http.Response, code string) {
	t.Helper()

	var errResp map[string]interface{}
	DecodeJSON(t, resp, &errResp)

	errorData, ok := errResp["error"].(map[string]interface{})
	require.True(t, ok, "Response should have 'error' field")

	actualCode, ok := errorData["code"].(string)
	require.True(t, ok, "Error should have 'code' field")

	require.Equal(t, code, actualCode, fmt.Sprintf("Expected error code %s, got %s", code, actualCode))
}
