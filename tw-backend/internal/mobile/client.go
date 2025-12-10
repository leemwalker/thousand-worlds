package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the mobile SDK client for interacting with the game backend
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string
}

// NewClient creates a new mobile client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: "",
	}
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken retrieves the current authentication token
func (c *Client) GetToken() string {
	return c.token
}

// ClearToken clears the authentication token
func (c *Client) ClearToken() {
	c.token = ""
}

// Logout clears the authentication token (alias for ClearToken)
func (c *Client) Logout() {
	c.ClearToken()
}

// doRequest performs an HTTP request with optional authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, requireAuth bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if requireAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleErrorResponse parses error responses from the API
func handleErrorResponse(resp *http.Response) error {
	defer resp.Body.Close()

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("HTTP %d: failed to parse error response", resp.StatusCode)
	}

	return &errResp
}

// Register creates a new user account
func (c *Client) Register(ctx context.Context, email, username, password string) (*User, error) {
	reqBody := map[string]string{
		"email":    email,
		"username": username,
		"password": password,
	}

	resp, err := c.doRequest(ctx, "POST", "/api/auth/register", reqBody, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, handleErrorResponse(resp)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// Login authenticates a user and returns a JWT token
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, err := c.doRequest(ctx, "POST", "/api/auth/login", reqBody, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	// Automatically set the token
	c.SetToken(loginResp.Token)

	return &loginResp, nil
}

// GetMe retrieves the current authenticated user's information
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/auth/me", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}
