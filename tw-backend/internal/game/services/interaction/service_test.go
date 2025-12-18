package interaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessDialogue(t *testing.T) {
	svc := NewService(nil) // Mock dependencies if needed later

	ctx := context.Background()
	charID := uuid.New()
	target := "Guard"
	msg := "Hello there"

	resp, err := svc.ProcessDialogue(ctx, charID, target, msg)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Text, "I heard you say")
	assert.Equal(t, target, resp.NPCName)
}
