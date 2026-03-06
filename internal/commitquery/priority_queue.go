package commitquery

import "container/heap"

// PriorityQueue orders internal nodes using one query context's comparator.
type PriorityQueue struct {
	ctx   *Context
	items []NodeIndex
}

// NewPriorityQueue builds one empty priority queue over one query context.
func NewPriorityQueue(ctx *Context) *PriorityQueue {
	queue := &PriorityQueue{ctx: ctx}
	heap.Init(queue)

	return queue
}

// Len reports the number of queued items.
func (queue *PriorityQueue) Len() int {
	return len(queue.items)
}

// Less reports whether one heap slot sorts ahead of another.
func (queue *PriorityQueue) Less(left, right int) bool {
	return queue.ctx.Compare(queue.items[left], queue.items[right]) > 0
}

// Swap exchanges two heap slots.
func (queue *PriorityQueue) Swap(left, right int) {
	queue.items[left], queue.items[right] = queue.items[right], queue.items[left]
}

// Push appends one heap element.
func (queue *PriorityQueue) Push(item any) {
	idx, ok := item.(NodeIndex)
	if !ok {
		panic("commitquery: heap push item is not a NodeIndex")
	}

	queue.items = append(queue.items, idx)
}

// Pop removes one heap element.
func (queue *PriorityQueue) Pop() any {
	last := len(queue.items) - 1
	item := queue.items[last]
	queue.items = queue.items[:last]

	return item
}

// PushNode inserts one internal node.
func (queue *PriorityQueue) PushNode(idx NodeIndex) {
	heap.Push(queue, idx)
}

// PopNode removes the highest-priority internal node.
func (queue *PriorityQueue) PopNode() NodeIndex {
	item := heap.Pop(queue)

	idx, ok := item.(NodeIndex)
	if !ok {
		panic("commitquery: heap pop item is not a NodeIndex")
	}

	return idx
}
