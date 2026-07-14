package priorityqueue

// Queue is one slice-backed priority queue.
//
// Labels: MT-Unsafe.
type Queue[T any] struct {
	items []T //exhaustruct:optional
	less  func(left, right T) bool
}

// New builds one empty priority queue ordered by less.
func New[T any](less func(left, right T) bool) *Queue[T] {
	return &Queue[T]{less: less}
}

// Len reports the number of queued items.
func (queue *Queue[T]) Len() int {
	return len(queue.items)
}

// Pop removes one highest-priority item.
func (queue *Queue[T]) Pop() (T, bool) { //nolint:ireturn
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

// Push inserts one item.
func (queue *Queue[T]) Push(item T) {
	queue.items = append(queue.items, item)
	queue.siftUp(len(queue.items) - 1)
}
