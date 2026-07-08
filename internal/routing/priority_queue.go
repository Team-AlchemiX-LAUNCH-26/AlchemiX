package routing

import "container/heap"

type routeState struct {
	previous string
	current  string
	cost     float64
	path     []string
	index    int
}

type priorityQueue []*routeState

func (p priorityQueue) Len() int { return len(p) }
func (p priorityQueue) Less(i, j int) bool {
	return stateLess(p[i], p[j])
}
func (p priorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}
func (p *priorityQueue) Push(x any) {
	item := x.(*routeState)
	item.index = len(*p)
	*p = append(*p, item)
}
func (p *priorityQueue) Pop() any {
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*p = old[:n-1]
	return item
}

func pushState(pq *priorityQueue, state *routeState) { heap.Push(pq, state) }
func popState(pq *priorityQueue) *routeState         { return heap.Pop(pq).(*routeState) }
