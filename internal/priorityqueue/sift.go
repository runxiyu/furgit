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
