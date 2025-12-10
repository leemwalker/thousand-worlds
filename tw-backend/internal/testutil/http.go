package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// PostJSON makes a POST request with JSON body
func PostJSON(t *testing.T, client *http.Client, url string, body interface{}) *http.Response {
	t.Helper()

	jsonData, err := json.Marshal(body)
	require.NoError(t, err, "Failed to marshal JSON")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute request")

	return resp
}

// PostJSONWithAuth makes an authenticated POST request with JSON body
func PostJSONWithAuth(t *testing.T, client *http.Client, url string, body interface{}, token string) *http.Response {
	t.Helper()

	jsonData, err := json.Marshal(body)
	require.NoError(t, err, "Failed to marshal JSON")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute request")

	return resp
}

// GetWithAuth makes an authenticated GET request
func GetWithAuth(t *testing.T, client *http.Client, url string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute request")

	return resp
}

// Get makes a GET request
func Get(t *testing.T, client *http.Client, url string) *http.Response {
	t.Helper()

	resp, err := client.Get(url)
	require.NoError(t, err, "Failed to execute GET request")

	return resp
}

// DecodeJSON decodes JSON response body
func DecodeJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()

	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err, "Failed to decode JSON response")
}

// ReadBody reads the entire response body as string
func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	return string(body)
}

// AssertStatus asserts HTTP status code
func AssertStatus(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()

	if resp.StatusCode != expectedStatus {
		body := ReadBody(t, resp)
		t.Fatalf("Expected status %d, got %d. Body: %s", expectedStatus, resp.StatusCode, body)
	}
}
