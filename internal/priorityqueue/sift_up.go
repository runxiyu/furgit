package priorityqueue

func (queue *Queue[T]) siftUp(idx int) {
	for idx > 0 {
		parent := (idx - 1) / 2
		if !queue.less(queue.items[idx], queue.items[parent]) {
			return
		}

		queue.items[idx], queue.items[parent] = queue.items[parent], queue.items[idx]
		idx = parent
	}
}
