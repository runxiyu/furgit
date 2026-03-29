package priorityqueue

// Pop removes one highest-priority item.
func (queue *Queue[T]) Pop() (T, bool) {
	if len(queue.items) == 0 {
		var zero T

		return zero, false
	}

	last := len(queue.items) - 1
	top := queue.items[0]
	queue.items[0] = queue.items[last]
	queue.items = queue.items[:last]

	if len(queue.items) > 0 {
		queue.siftDown(0)
	}

	return top, true
}
