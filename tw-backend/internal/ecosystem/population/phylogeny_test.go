package population

import (
	"testing"

	"github.com/google/uuid"
)

func TestPhylogeneticTree_AddRoot(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	species := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Founder Species",
		Diet:      DietHerbivore,
	}

	node := tree.AddRoot(species, 0)

	t.Run("node added correctly", func(t *testing.T) {
		if node.SpeciesID != species.SpeciesID {
			t.Error("Species ID mismatch")
		}
		if node.Depth != 0 {
			t.Error("Root should have depth 0")
		}
		if !node.IsRoot() {
			t.Error("Root should return true for IsRoot()")
		}
	})

	t.Run("tree updated correctly", func(t *testing.T) {
		if len(tree.Roots) != 1 {
			t.Errorf("Should have 1 root, got %d", len(tree.Roots))
		}
		if tree.ExtantCount != 1 {
			t.Errorf("Should have 1 extant, got %d", tree.ExtantCount)
		}
	})
}

func TestPhylogeneticTree_AddSpeciation(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	parent := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Parent Species",
		Diet:      DietHerbivore,
	}
	tree.AddRoot(parent, 0)

	child := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Child Species",
		Diet:      DietHerbivore,
	}

	node := tree.AddSpeciation(parent, child, SpeciationAllopatric, 100000)

	t.Run("child linked to parent", func(t *testing.T) {
		if node.ParentID == nil || *node.ParentID != parent.SpeciesID {
			t.Error("Child should be linked to parent")
		}
		if node.Depth != 1 {
			t.Errorf("Child depth should be 1, got %d", node.Depth)
		}
	})

	t.Run("parent has child", func(t *testing.T) {
		parentNode := tree.GetNode(parent.SpeciesID)
		if len(parentNode.ChildIDs) != 1 {
			t.Errorf("Parent should have 1 child, got %d", len(parentNode.ChildIDs))
		}
	})

	t.Run("counts updated", func(t *testing.T) {
		if tree.ExtantCount != 2 {
			t.Errorf("Should have 2 extant, got %d", tree.ExtantCount)
		}
		if tree.MaxDepth != 1 {
			t.Errorf("Max depth should be 1, got %d", tree.MaxDepth)
		}
	})
}

func TestPhylogeneticTree_MarkExtinct(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	species := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Doomed Species",
	}
	tree.AddRoot(species, 0)

	tree.MarkExtinct(species.SpeciesID, 50000)

	t.Run("species marked extinct", func(t *testing.T) {
		node := tree.GetNode(species.SpeciesID)
		if node.IsExtant() {
			t.Error("Species should be extinct")
		}
		if node.ExtinctionYear != 50000 {
			t.Errorf("Extinction year should be 50000, got %d", node.ExtinctionYear)
		}
	})

	t.Run("counts updated", func(t *testing.T) {
		if tree.ExtantCount != 0 {
			t.Errorf("Should have 0 extant, got %d", tree.ExtantCount)
		}
		if tree.ExtinctCount != 1 {
			t.Errorf("Should have 1 extinct, got %d", tree.ExtinctCount)
		}
	})
}

func TestPhylogeneticTree_Ancestors(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	// Create a lineage: grandparent -> parent -> child
	grandparent := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Grandparent"}
	parent := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Parent"}
	child := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Child"}

	tree.AddRoot(grandparent, 0)
	tree.AddSpeciation(grandparent, parent, SpeciationAllopatric, 100000)
	tree.AddSpeciation(parent, child, SpeciationAllopatric, 200000)

	ancestors := tree.GetAncestors(child.SpeciesID)

	if len(ancestors) != 2 {
		t.Errorf("Child should have 2 ancestors, got %d", len(ancestors))
	}
	if ancestors[0].SpeciesID != parent.SpeciesID {
		t.Error("First ancestor should be parent")
	}
	if ancestors[1].SpeciesID != grandparent.SpeciesID {
		t.Error("Second ancestor should be grandparent")
	}
}

func TestPhylogeneticTree_Descendants(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	// Create a branching tree
	root := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Root"}
	child1 := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Child1"}
	child2 := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Child2"}
	grandchild := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Grandchild"}

	tree.AddRoot(root, 0)
	tree.AddSpeciation(root, child1, SpeciationAllopatric, 100000)
	tree.AddSpeciation(root, child2, SpeciationSympatric, 100000)
	tree.AddSpeciation(child1, grandchild, SpeciationAllopatric, 200000)

	descendants := tree.GetDescendants(root.SpeciesID)

	if len(descendants) != 3 {
		t.Errorf("Root should have 3 descendants, got %d", len(descendants))
	}
}

func TestPhylogeneticTree_CommonAncestor(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	root := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Root"}
	child1 := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Child1"}
	child2 := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Child2"}
	grandchild1 := &SpeciesPopulation{SpeciesID: uuid.New(), Name: "Grandchild1"}

	tree.AddRoot(root, 0)
	tree.AddSpeciation(root, child1, SpeciationAllopatric, 100000)
	tree.AddSpeciation(root, child2, SpeciationAllopatric, 100000)
	tree.AddSpeciation(child1, grandchild1, SpeciationAllopatric, 200000)

	t.Run("siblings share parent", func(t *testing.T) {
		ancestor := tree.GetCommonAncestor(child1.SpeciesID, child2.SpeciesID)
		if ancestor == nil {
			t.Fatal("Should find common ancestor")
		}
		if ancestor.SpeciesID != root.SpeciesID {
			t.Error("Common ancestor of siblings should be root")
		}
	})

	t.Run("parent-child share parent", func(t *testing.T) {
		ancestor := tree.GetCommonAncestor(child1.SpeciesID, grandchild1.SpeciesID)
		if ancestor == nil {
			t.Fatal("Should find common ancestor")
		}
		if ancestor.SpeciesID != child1.SpeciesID {
			t.Error("Common ancestor should be child1")
		}
	})

	t.Run("cousins share grandparent", func(t *testing.T) {
		ancestor := tree.GetCommonAncestor(grandchild1.SpeciesID, child2.SpeciesID)
		if ancestor == nil {
			t.Fatal("Should find common ancestor")
		}
		if ancestor.SpeciesID != root.SpeciesID {
			t.Error("Common ancestor of cousins should be root")
		}
	})
}

func TestPhylogeneticTree_PhylogeneticDistance(t *testing.T) {
	tree := NewPhylogeneticTree(uuid.New())

	root := &SpeciesPopulation{SpeciesID: uuid.New()}
	child1 := &SpeciesPopulation{SpeciesID: uuid.New()}
	child2 := &SpeciesPopulation{SpeciesID: uuid.New()}
	grandchild := &SpeciesPopulation{SpeciesID: uuid.New()}

	tree.AddRoot(root, 0)
	tree.AddSpeciation(root, child1, SpeciationAllopatric, 100000)
	tree.AddSpeciation(root, child2, SpeciationAllopatric, 100000)
	tree.AddSpeciation(child1, grandchild, SpeciationAllopatric, 200000)

	t.Run("self distance is 0", func(t *testing.T) {
		dist := tree.GetPhylogeneticDistance(root.SpeciesID, root.SpeciesID)
		if dist != 0 {
			t.Errorf("Self distance should be 0, got %d", dist)
		}
	})

	t.Run("parent-child distance is 1", func(t *testing.T) {
		dist := tree.GetPhylogeneticDistance(root.SpeciesID, child1.SpeciesID)
		if dist != 1 {
			t.Errorf("Parent-child distance should be 1, got %d", dist)
		}
	})

	t.Run("siblings distance is 2", func(t *testing.T) {
		dist := tree.GetPhylogeneticDistance(child1.SpeciesID, child2.SpeciesID)
		if dist != 2 {
			t.Errorf("Sibling distance should be 2, got %d", dist)
		}
	})

	t.Run("cousins distance is 3", func(t *testing.T) {
		dist := tree.GetPhylogeneticDistance(grandchild.SpeciesID, child2.SpeciesID)
		if dist != 3 {
			t.Errorf("Cousin distance should be 3, got %d", dist)
		}
	})
}
