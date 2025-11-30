package resources

import (
	"github.com/google/uuid"
)

// Repository defines the interface for resource persistence
type Repository interface {
	// Resource node operations
	CreateResourceNode(node *ResourceNode) error
	GetResourceNode(nodeID uuid.UUID) (*ResourceNode, error)
	GetAllResourceNodes() ([]*ResourceNode, error)
	GetResourceNodesByType(resourceType ResourceType) ([]*ResourceNode, error)
	GetResourceNodesByBiome(biomeType string) ([]*ResourceNode, error)
	GetResourceNodesInRadius(x, y, radius float64) ([]*ResourceNode, error)
	UpdateResourceNode(node *ResourceNode) error
	DeleteResourceNode(nodeID uuid.UUID) error

	// Mineral deposit operations (Phase 8.2b queries - read-only)
	GetMineralDeposits() ([]*MineralDeposit, error)
	GetMineralDepositByID(depositID uuid.UUID) (*MineralDeposit, error)
}
