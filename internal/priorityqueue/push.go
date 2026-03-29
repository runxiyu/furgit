package priorityqueue

// Push inserts one item.
func (queue *Queue[T]) Push(item T) {
	queue.items = append(queue.items, item)
	queue.siftUp(len(queue.items) - 1)
}
