package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// World represents a game world
type World struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Shape     string    `json:"shape"`
	CreatedAt time.Time `json:"created_at"`
}

// ListWorlds retrieves all available game worlds
func (c *Client) ListWorlds(ctx context.Context) ([]*World, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/game/worlds", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var worlds []*World
	if err := json.NewDecoder(resp.Body).Decode(&worlds); err != nil {
		return nil, fmt.Errorf("failed to decode worlds response: %w", err)
	}

	if worlds == nil {
		return []*World{}, nil
	}

	return worlds, nil
}
