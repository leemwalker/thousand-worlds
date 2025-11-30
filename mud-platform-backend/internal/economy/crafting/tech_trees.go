package crafting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LoadTechTreeFromFile loads a tech tree definition from a JSON file
func LoadTechTreeFromFile(path string) ([]*TechNode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var nodes []*TechNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}

	// Ensure IDs are generated if missing (for static data)
	// In a real scenario, we might want stable IDs
	for _, node := range nodes {
		if node.NodeID == uuid.Nil {
			node.NodeID = uuid.New()
		}
	}

	return nodes, nil
}

// TechTreeManager handles tech tree operations
type TechTreeManager struct {
	repo Repository
}

func NewTechTreeManager(repo Repository) *TechTreeManager {
	return &TechTreeManager{repo: repo}
}

// UnlockTechNode attempts to unlock a tech node for an entity
func (m *TechTreeManager) UnlockTechNode(entityID uuid.UUID, nodeID uuid.UUID) error {
	// 1. Get the node
	node, err := m.repo.GetTechNode(nodeID)
	if err != nil {
		return err
	}

	// 2. Check if already unlocked
	unlocked, err := m.repo.IsTechUnlocked(entityID, nodeID)
	if err != nil {
		return err
	}
	if unlocked {
		return nil // Already unlocked
	}

	// 3. Check prerequisites
	for _, prereqID := range node.Prerequisites {
		prereqUnlocked, err := m.repo.IsTechUnlocked(entityID, prereqID)
		if err != nil {
			return err
		}
		if !prereqUnlocked {
			return fmt.Errorf("prerequisite not met: %s", prereqID)
		}
	}

	// 4. Unlock the node
	if err := m.repo.UnlockTech(entityID, nodeID); err != nil {
		return err
	}

	// 5. Auto-discover recipes unlocked by this node
	for _, recipeID := range node.UnlocksRecipes {
		knowledge := &RecipeKnowledge{
			EntityID:    entityID,
			RecipeID:    recipeID,
			Proficiency: 0.0,
			TimesUsed:   0,
			Source:      "research",
		}
		// Ignore error if already known
		_ = m.repo.DiscoverRecipe(knowledge)
	}

	return nil
}

// LoadAllTechTrees loads all JSON files from the data directory
func LoadAllTechTrees(dataDir string) (map[TechLevel][]*TechNode, error) {
	result := make(map[TechLevel][]*TechNode)

	levels := []TechLevel{TechPrimitive, TechMedieval, TechIndustrial, TechModern, TechFuturistic}

	for _, level := range levels {
		filename := fmt.Sprintf("%s.json", level)
		path := filepath.Join(dataDir, filename)

		// Skip if file doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		nodes, err := LoadTechTreeFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", level, err)
		}

		result[level] = nodes
	}

	return result, nil
}
