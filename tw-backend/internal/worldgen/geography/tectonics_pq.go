package geography

import (
	"tw-backend/internal/spatial"
)

// borderItem represents a cell in the Dijkstra priority queue
type borderItem struct {
	coord    spatial.Coordinate
	plateIdx int
	cost     float64
	index    int
}

// borderPQ implements heap.Interface for borderItem
type borderPQ []*borderItem

func (pq borderPQ) Len() int           { return len(pq) }
func (pq borderPQ) Less(i, j int) bool { return pq[i].cost < pq[j].cost }
func (pq borderPQ) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }

func (pq *borderPQ) Push(x interface{}) {
	n := len(*pq)
	item := x.(*borderItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *borderPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
