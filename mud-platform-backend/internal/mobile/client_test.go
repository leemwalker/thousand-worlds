package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_NewClient tests client initialization
func TestClient_NewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.BaseURL)
	assert.NotNil(t, client.HTTPClient)
	assert.Empty(t, client.GetToken())
}

// TestClient_TokenManagement tests token get/set/clear
func TestClient_TokenManagement(t *testing.T) {
	client := NewClient("http://localhost:8080")

	// Initially empty
	assert.Empty(t, client.GetToken())

	// Set token
	client.SetToken("test-token-123")
	assert.Equal(t, "test-token-123", client.GetToken())

	// Clear token
	client.ClearToken()
	assert.Empty(t, client.GetToken())
}

// TestClient_Register_Success tests successful user registration
func TestClient_Register_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/register", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test@example.com", req["email"])
		assert.Equal(t, "Password123", req["password"])

		// Send response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":    "user-123",
			"email":      "test@example.com",
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	user, err := client.Register(context.Background(), "test@example.com", "Password123")

	require.NoError(t, err)
	assert.Equal(t, "user-123", user.UserID)
	assert.Equal(t, "test@example.com", user.Email)
}

// TestClient_Register_DuplicateEmail tests duplicate email error
func TestClient_Register_DuplicateEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "CONFLICT",
			"message": "User already exists",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	user, err := client.Register(context.Background(), "duplicate@example.com", "Password123")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")
}

// TestClient_Register_InvalidEmail tests email validation
func TestClient_Register_InvalidEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "INVALID_INPUT",
			"message": "Invalid email format",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	user, err := client.Register(context.Background(), "not-an-email", "Password123")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "email")
}

// TestClient_Register_WeakPassword tests password validation
func TestClient_Register_WeakPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "INVALID_INPUT",
			"message": "Password must be at least 8 characters",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	user, err := client.Register(context.Background(), "test@example.com", "short")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "Password")
}

// TestClient_Login_Success tests successful login
func TestClient_Login_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/login", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": "jwt-token-here",
			"user": map[string]interface{}{
				"user_id":    "user-123",
				"email":      "test@example.com",
				"created_at": "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	loginResp, err := client.Login(context.Background(), "test@example.com", "Password123")

	require.NoError(t, err)
	assert.Equal(t, "jwt-token-here", loginResp.Token)
	assert.Equal(t, "user-123", loginResp.User.UserID)
	// Token should be automatically set
	assert.Equal(t, "jwt-token-here", client.GetToken())
}

// TestClient_Login_InvalidCredentials tests invalid credentials
func TestClient_Login_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "UNAUTHORIZED",
			"message": "Invalid credentials",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	loginResp, err := client.Login(context.Background(), "wrong@example.com", "WrongPassword")

	require.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Contains(t, err.Error(), "credentials")
}

// TestClient_GetMe_Success tests getting current user
func TestClient_GetMe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/me", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": "user-123",
			"email":   "test@example.com",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	user, err := client.GetMe(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "user-123", user.UserID)
	assert.Equal(t, "test@example.com", user.Email)
}

// TestClient_GetMe_Unauthorized tests unauthorized access
func TestClient_GetMe_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "UNAUTHORIZED",
			"message": "Invalid or expired token",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	// No token set

	user, err := client.GetMe(context.Background())

	require.Error(t, err)
	assert.Nil(t, user)
}

// TestClient_Logout tests logout functionality
func TestClient_Logout(t *testing.T) {
	client := NewClient("http://localhost:8080")
	client.SetToken("test-token")

	assert.NotEmpty(t, client.GetToken())

	client.Logout()

	assert.Empty(t, client.GetToken())
}

// Helper function to create a test HTTP response
func createTestResponse(statusCode int, body interface{}) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bodyReader),
		Header:     make(http.Header),
	}
}
