package ecosystem

import (
	"container/heap"
	"math"
)

// Point represents a location in the world
type Point struct {
	X, Y float64
}

// Node represents a point in the pathfinding grid
type Node struct {
	Point     Point
	Cost      float64
	Heuristic float64
	Parent    *Node
	Index     int // For heap interface
}

// Pathdinder interface allows different world representations to be navigated
type WorldMap interface {
	GetNeighbors(p Point) []Point
	Cost(from, to Point) float64
	IsBlocked(p Point) bool
}

// PriorityQueue implements heap.Interface for Nodes
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return (pq[i].Cost + pq[i].Heuristic) < (pq[j].Cost + pq[j].Heuristic)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n, _ := x.(*Node) // Type assertion guaranteed by heap.Push caller
	n.Index = len(*pq)
	*pq = append(*pq, n)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// FindPath implements A* algorithm
func FindPath(world WorldMap, start, end Point) []Point {
	startNode := &Node{Point: start, Cost: 0, Heuristic: distance(start, end)}
	openSet := &PriorityQueue{startNode}
	heap.Init(openSet)

	visited := make(map[Point]bool)
	nodeCache := make(map[Point]*Node)
	nodeCache[start] = startNode

	for openSet.Len() > 0 {
		current, _ := heap.Pop(openSet).(*Node) // Type assertion guaranteed by heap implementation

		if distance(current.Point, end) < 1.0 { // Reached destination (within tolerance)
			return reconstructPath(current)
		}

		visited[current.Point] = true

		for _, neighborPos := range world.GetNeighbors(current.Point) {
			if visited[neighborPos] || world.IsBlocked(neighborPos) {
				continue
			}

			newCost := current.Cost + world.Cost(current.Point, neighborPos)

			neighborNode, seen := nodeCache[neighborPos]
			if !seen {
				neighborNode = &Node{
					Point:     neighborPos,
					Cost:      math.Inf(1),
					Heuristic: distance(neighborPos, end),
				}
				nodeCache[neighborPos] = neighborNode
			}

			if newCost < neighborNode.Cost {
				neighborNode.Cost = newCost
				neighborNode.Parent = current
				if !seen {
					heap.Push(openSet, neighborNode)
				} else {
					heap.Fix(openSet, neighborNode.Index)
				}
			}
		}
	}

	return nil // No path found
}

func distance(a, b Point) float64 {
	return math.Sqrt(math.Pow(a.X-b.X, 2) + math.Pow(a.Y-b.Y, 2))
}

func reconstructPath(node *Node) []Point {
	var path []Point
	for node != nil {
		path = append([]Point{node.Point}, path...)
		node = node.Parent
	}
	return path
}
