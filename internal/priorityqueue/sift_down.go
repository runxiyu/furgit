package priorityqueue

func (queue *Queue[T]) siftDown(idx int) {
	for {
		left := idx*2 + 1
		if left >= len(queue.items) {
			return
		}

		best := left
		right := left + 1
		if right < len(queue.items) && queue.less(queue.items[right], queue.items[left]) {
			best = right
		}

		if !queue.less(queue.items[best], queue.items[idx]) {
			return
		}

		queue.items[idx], queue.items[best] = queue.items[best], queue.items[idx]
		idx = best
	}
}
