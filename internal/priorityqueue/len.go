package priorityqueue

// Len reports the number of queued items.
func (queue *Queue[T]) Len() int {
	return len(queue.items)
}
