package commitquery

import internalheap "codeberg.org/lindenii/furgit/internal/heap"

// priorityQueue orders internal nodes using one query context's comparator.
type priorityQueue struct {
	items *internalheap.Heap[nodeIndex]
}

// newPriorityQueue builds one empty priority queue over one query context.
func newPriorityQueue(query *query) *priorityQueue {
	return &priorityQueue{
		items: internalheap.New(func(left, right nodeIndex) bool {
			return query.compare(left, right) > 0
		}),
	}
}

// Len reports the number of queued items.
func (queue *priorityQueue) Len() int {
	return queue.items.Len()
}

// PushNode inserts one internal node.
func (queue *priorityQueue) PushNode(idx nodeIndex) {
	queue.items.Push(idx)
}

// PopNode removes the highest-priority internal node.
func (queue *priorityQueue) PopNode() (nodeIndex, bool) {
	return queue.items.Pop()
}
