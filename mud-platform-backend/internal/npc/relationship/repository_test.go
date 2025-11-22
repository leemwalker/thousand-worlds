package relationship

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRepository_CRUD(t *testing.T) {
	repo := NewMockRepository()
	npcID := uuid.New()
	targetID := uuid.New()

	// Create
	rel, err := repo.CreateRelationship(npcID, targetID)
	assert.NoError(t, err)
	assert.Equal(t, npcID, rel.NPCID)
	assert.Equal(t, 0, rel.CurrentAffinity.Affection)

	// Get
	fetched, err := repo.GetRelationship(npcID, targetID)
	assert.NoError(t, err)
	assert.Equal(t, rel.ID, fetched.ID)

	// Update Affinity
	newAffinity := Affinity{Affection: 10, Trust: 5, Fear: 0}
	err = repo.UpdateAffinity(npcID, targetID, newAffinity)
	assert.NoError(t, err)

	fetched, _ = repo.GetRelationship(npcID, targetID)
	assert.Equal(t, 10, fetched.CurrentAffinity.Affection)

	// Record Interaction
	interaction := Interaction{
		Timestamp:  time.Now(),
		ActionType: "gift",
	}
	err = repo.RecordInteraction(npcID, targetID, interaction)
	assert.NoError(t, err)

	fetched, _ = repo.GetRelationship(npcID, targetID)
	assert.Len(t, fetched.RecentInteractions, 1)
	assert.Equal(t, "gift", fetched.RecentInteractions[0].ActionType)
}
