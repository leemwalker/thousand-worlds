// Package population provides phylogenetic tree tracking for species lineages.
// This enables visualization of evolutionary history and lineage-based queries.
package population

import (
	"github.com/google/uuid"
)

// PhylogeneticNode represents a node in the tree of life
type PhylogeneticNode struct {
	SpeciesID      uuid.UUID      `json:"species_id"`
	Name           string         `json:"name"`
	ParentID       *uuid.UUID     `json:"parent_id,omitempty"`
	ChildIDs       []uuid.UUID    `json:"child_ids"`
	OriginYear     int64          `json:"origin_year"`
	ExtinctionYear int64          `json:"extinction_year,omitempty"` // 0 if extant
	SpeciationType SpeciationType `json:"speciation_type"`
	Diet           DietType       `json:"diet"`
	GeneticCode    *GeneticCode   `json:"genetic_code,omitempty"`
	Depth          int            `json:"depth"` // Distance from root
}

// IsExtant returns true if the species is not extinct
func (n *PhylogeneticNode) IsExtant() bool {
	return n.ExtinctionYear == 0
}

// IsRoot returns true if this is a root node (no parent)
func (n *PhylogeneticNode) IsRoot() bool {
	return n.ParentID == nil
}

// Duration returns how long the species existed (or has existed)
func (n *PhylogeneticNode) Duration(currentYear int64) int64 {
	endYear := n.ExtinctionYear
	if endYear == 0 {
		endYear = currentYear
	}
	return endYear - n.OriginYear
}

// PhylogeneticTree represents the complete tree of life for a world
type PhylogeneticTree struct {
	WorldID      uuid.UUID                       `json:"world_id"`
	Nodes        map[uuid.UUID]*PhylogeneticNode `json:"nodes"`
	Roots        []uuid.UUID                     `json:"roots"` // Original founder species
	CurrentYear  int64                           `json:"current_year"`
	ExtantCount  int                             `json:"extant_count"`
	ExtinctCount int                             `json:"extinct_count"`
	MaxDepth     int                             `json:"max_depth"`
}

// NewPhylogeneticTree creates a new empty phylogenetic tree
func NewPhylogeneticTree(worldID uuid.UUID) *PhylogeneticTree {
	return &PhylogeneticTree{
		WorldID: worldID,
		Nodes:   make(map[uuid.UUID]*PhylogeneticNode),
		Roots:   make([]uuid.UUID, 0),
	}
}

// AddRoot adds a founder species (root node) to the tree
func (pt *PhylogeneticTree) AddRoot(species *SpeciesPopulation, originYear int64) *PhylogeneticNode {
	node := &PhylogeneticNode{
		SpeciesID:      species.SpeciesID,
		Name:           species.Name,
		ParentID:       nil,
		ChildIDs:       make([]uuid.UUID, 0),
		OriginYear:     originYear,
		SpeciationType: "", // Root has no speciation type
		Diet:           species.Diet,
		GeneticCode:    species.GeneticCode,
		Depth:          0,
	}

	pt.Nodes[species.SpeciesID] = node
	pt.Roots = append(pt.Roots, species.SpeciesID)
	pt.ExtantCount++

	return node
}

// AddSpeciation adds a new species that descended from a parent
func (pt *PhylogeneticTree) AddSpeciation(
	parent *SpeciesPopulation,
	child *SpeciesPopulation,
	speciationType SpeciationType,
	year int64,
) *PhylogeneticNode {
	parentNode := pt.Nodes[parent.SpeciesID]
	if parentNode == nil {
		// Parent not in tree - add it as root first
		parentNode = pt.AddRoot(parent, year-1)
	}

	childNode := &PhylogeneticNode{
		SpeciesID:      child.SpeciesID,
		Name:           child.Name,
		ParentID:       &parent.SpeciesID,
		ChildIDs:       make([]uuid.UUID, 0),
		OriginYear:     year,
		SpeciationType: speciationType,
		Diet:           child.Diet,
		GeneticCode:    child.GeneticCode,
		Depth:          parentNode.Depth + 1,
	}

	pt.Nodes[child.SpeciesID] = childNode
	parentNode.ChildIDs = append(parentNode.ChildIDs, child.SpeciesID)
	pt.ExtantCount++

	if childNode.Depth > pt.MaxDepth {
		pt.MaxDepth = childNode.Depth
	}

	return childNode
}

// MarkExtinct marks a species as extinct
func (pt *PhylogeneticTree) MarkExtinct(speciesID uuid.UUID, year int64) {
	node := pt.Nodes[speciesID]
	if node != nil && node.ExtinctionYear == 0 {
		node.ExtinctionYear = year
		pt.ExtantCount--
		pt.ExtinctCount++
	}
}

// GetNode returns a node by species ID
func (pt *PhylogeneticTree) GetNode(speciesID uuid.UUID) *PhylogeneticNode {
	return pt.Nodes[speciesID]
}

// GetAncestors returns all ancestors of a species (parent, grandparent, etc.)
func (pt *PhylogeneticTree) GetAncestors(speciesID uuid.UUID) []*PhylogeneticNode {
	ancestors := make([]*PhylogeneticNode, 0)

	node := pt.Nodes[speciesID]
	for node != nil && node.ParentID != nil {
		parent := pt.Nodes[*node.ParentID]
		if parent != nil {
			ancestors = append(ancestors, parent)
			node = parent
		} else {
			break
		}
	}

	return ancestors
}

// GetDescendants returns all descendants of a species (children, grandchildren, etc.)
func (pt *PhylogeneticTree) GetDescendants(speciesID uuid.UUID) []*PhylogeneticNode {
	descendants := make([]*PhylogeneticNode, 0)

	node := pt.Nodes[speciesID]
	if node == nil {
		return descendants
	}

	// BFS through children
	queue := make([]uuid.UUID, len(node.ChildIDs))
	copy(queue, node.ChildIDs)

	for len(queue) > 0 {
		childID := queue[0]
		queue = queue[1:]

		child := pt.Nodes[childID]
		if child != nil {
			descendants = append(descendants, child)
			queue = append(queue, child.ChildIDs...)
		}
	}

	return descendants
}

// GetCommonAncestor finds the most recent common ancestor of two species
func (pt *PhylogeneticTree) GetCommonAncestor(species1ID, species2ID uuid.UUID) *PhylogeneticNode {
	// Get all ancestors of species1
	ancestors1 := make(map[uuid.UUID]bool)
	ancestors1[species1ID] = true
	for _, a := range pt.GetAncestors(species1ID) {
		ancestors1[a.SpeciesID] = true
	}

	// Walk up species2's lineage until we find a match
	node := pt.Nodes[species2ID]
	for node != nil {
		if ancestors1[node.SpeciesID] {
			return node
		}
		if node.ParentID != nil {
			node = pt.Nodes[*node.ParentID]
		} else {
			break
		}
	}

	return nil // No common ancestor (different roots)
}

// GetPhylogeneticDistance returns the number of speciation events between two species
func (pt *PhylogeneticTree) GetPhylogeneticDistance(species1ID, species2ID uuid.UUID) int {
	ancestor := pt.GetCommonAncestor(species1ID, species2ID)
	if ancestor == nil {
		return -1 // Different trees
	}

	// Count steps from each species to ancestor
	dist1 := pt.distanceToAncestor(species1ID, ancestor.SpeciesID)
	dist2 := pt.distanceToAncestor(species2ID, ancestor.SpeciesID)

	return dist1 + dist2
}

// distanceToAncestor counts steps from a node to an ancestor
func (pt *PhylogeneticTree) distanceToAncestor(speciesID, ancestorID uuid.UUID) int {
	if speciesID == ancestorID {
		return 0
	}

	distance := 0
	node := pt.Nodes[speciesID]
	for node != nil && node.SpeciesID != ancestorID {
		distance++
		if node.ParentID != nil {
			node = pt.Nodes[*node.ParentID]
		} else {
			break
		}
	}

	return distance
}

// GetExtantSpecies returns all currently living species
func (pt *PhylogeneticTree) GetExtantSpecies() []*PhylogeneticNode {
	extant := make([]*PhylogeneticNode, 0, pt.ExtantCount)
	for _, node := range pt.Nodes {
		if node.IsExtant() {
			extant = append(extant, node)
		}
	}
	return extant
}

// GetExtinctSpecies returns all extinct species
func (pt *PhylogeneticTree) GetExtinctSpecies() []*PhylogeneticNode {
	extinct := make([]*PhylogeneticNode, 0, pt.ExtinctCount)
	for _, node := range pt.Nodes {
		if !node.IsExtant() {
			extinct = append(extinct, node)
		}
	}
	return extinct
}

// GetSpeciesAtYear returns species that existed at a specific year
func (pt *PhylogeneticTree) GetSpeciesAtYear(year int64) []*PhylogeneticNode {
	species := make([]*PhylogeneticNode, 0)
	for _, node := range pt.Nodes {
		if node.OriginYear <= year {
			if node.ExtinctionYear == 0 || node.ExtinctionYear > year {
				species = append(species, node)
			}
		}
	}
	return species
}

// GetDiversityOverTime returns species count at each interval
func (pt *PhylogeneticTree) GetDiversityOverTime(interval, endYear int64) map[int64]int {
	diversity := make(map[int64]int)

	for year := int64(0); year <= endYear; year += interval {
		diversity[year] = len(pt.GetSpeciesAtYear(year))
	}

	return diversity
}

// GetLineageSurvivors returns extant species that descend from a given ancestor
func (pt *PhylogeneticTree) GetLineageSurvivors(ancestorID uuid.UUID) []*PhylogeneticNode {
	survivors := make([]*PhylogeneticNode, 0)

	// Check if ancestor itself is extant
	node := pt.Nodes[ancestorID]
	if node != nil && node.IsExtant() {
		survivors = append(survivors, node)
	}

	// Check all descendants
	for _, desc := range pt.GetDescendants(ancestorID) {
		if desc.IsExtant() {
			survivors = append(survivors, desc)
		}
	}

	return survivors
}

// Clone creates a deep copy of the tree
func (pt *PhylogeneticTree) Clone() *PhylogeneticTree {
	clone := NewPhylogeneticTree(pt.WorldID)
	clone.CurrentYear = pt.CurrentYear
	clone.ExtantCount = pt.ExtantCount
	clone.ExtinctCount = pt.ExtinctCount
	clone.MaxDepth = pt.MaxDepth

	clone.Roots = make([]uuid.UUID, len(pt.Roots))
	copy(clone.Roots, pt.Roots)

	for id, node := range pt.Nodes {
		nodeCopy := *node
		nodeCopy.ChildIDs = make([]uuid.UUID, len(node.ChildIDs))
		copy(nodeCopy.ChildIDs, node.ChildIDs)
		if node.GeneticCode != nil {
			nodeCopy.GeneticCode = node.GeneticCode.Clone()
		}
		clone.Nodes[id] = &nodeCopy
	}

	return clone
}
