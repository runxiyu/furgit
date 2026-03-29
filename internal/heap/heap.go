package heap

// Heap is one slice-backed binary heap.
type Heap[T any] struct {
	items []T
	less  func(left, right T) bool
}

// New builds one empty heap ordered by less.
func New[T any](less func(left, right T) bool) *Heap[T] {
	return &Heap[T]{less: less}
}

// Len reports the number of queued items.
func (heap *Heap[T]) Len() int {
	return len(heap.items)
}

// Empty reports whether the heap is empty.
func (heap *Heap[T]) Empty() bool {
	return len(heap.items) == 0
}

// Peek returns the top item without removing it.
func (heap *Heap[T]) Peek() (T, bool) {
	if len(heap.items) == 0 {
		var zero T

		return zero, false
	}

	return heap.items[0], true
}

// Push inserts one item.
func (heap *Heap[T]) Push(item T) {
	heap.items = append(heap.items, item)
	heap.siftUp(len(heap.items) - 1)
}

// Pop removes one top item.
func (heap *Heap[T]) Pop() (T, bool) {
	if len(heap.items) == 0 {
		var zero T

		return zero, false
	}

	last := len(heap.items) - 1
	top := heap.items[0]
	heap.items[0] = heap.items[last]
	heap.items = heap.items[:last]

	if len(heap.items) > 0 {
		heap.siftDown(0)
	}

	return top, true
}

func (heap *Heap[T]) siftUp(idx int) {
	for idx > 0 {
		parent := (idx - 1) / 2
		if !heap.less(heap.items[idx], heap.items[parent]) {
			return
		}

		heap.items[idx], heap.items[parent] = heap.items[parent], heap.items[idx]
		idx = parent
	}
}

func (heap *Heap[T]) siftDown(idx int) {
	for {
		left := idx*2 + 1
		if left >= len(heap.items) {
			return
		}

		best := left
		right := left + 1
		if right < len(heap.items) && heap.less(heap.items[right], heap.items[left]) {
			best = right
		}

		if !heap.less(heap.items[best], heap.items[idx]) {
			return
		}

		heap.items[idx], heap.items[best] = heap.items[best], heap.items[idx]
		idx = best
	}
}
