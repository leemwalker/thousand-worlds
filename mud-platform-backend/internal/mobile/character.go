package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Character represents a game character
type Character struct {
	CharacterID string `json:"character_id"`
	UserID      string `json:"user_id"`
	WorldID     string `json:"world_id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Appearance  string `json:"appearance,omitempty"`
	Description string `json:"description,omitempty"`
	Occupation  string `json:"occupation,omitempty"`
}

// CreateCharacterRequest represents a request to create a character
type CreateCharacterRequest struct {
	WorldID     string `json:"world_id"`
	Name        string `json:"name"`
	Species     string `json:"species,omitempty"`
	Role        string `json:"role,omitempty"`
	Appearance  string `json:"appearance,omitempty"`
	Description string `json:"description,omitempty"`
	Occupation  string `json:"occupation,omitempty"`
}

// CreateCharacterResponse represents the response from creating a character
type CreateCharacterResponse struct {
	Character      *Character             `json:"character"`
	Attributes     map[string]interface{} `json:"attributes,omitempty"`
	SecondaryAttrs map[string]interface{} `json:"secondary_attributes,omitempty"`
}

// JoinGameResponse represents the response from joining a game
type JoinGameResponse struct {
	Character map[string]interface{} `json:"character"`
	WorldID   string                 `json:"world_id"`
	Message   string                 `json:"message"`
}

// CreateCharacter creates a new character in the specified world
func (c *Client) CreateCharacter(ctx context.Context, req *CreateCharacterRequest) (*Character, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/game/characters", req, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, handleErrorResponse(resp)
	}

	var createResp CreateCharacterResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to decode character response: %w", err)
	}

	return createResp.Character, nil
}

// GetCharacters retrieves all characters for the authenticated user
func (c *Client) GetCharacters(ctx context.Context) ([]*Character, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/game/characters", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var result struct {
		Characters []*Character `json:"characters"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode characters response: %w", err)
	}

	if result.Characters == nil {
		return []*Character{}, nil
	}

	return result.Characters, nil
}

// JoinGame joins a game world with the specified character
func (c *Client) JoinGame(ctx context.Context, characterID string) (*JoinGameResponse, error) {
	reqBody := map[string]string{
		"character_id": characterID,
	}

	resp, err := c.doRequest(ctx, "POST", "/api/game/join", reqBody, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var joinResp JoinGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&joinResp); err != nil {
		return nil, fmt.Errorf("failed to decode join game response: %w", err)
	}

	return &joinResp, nil
}
