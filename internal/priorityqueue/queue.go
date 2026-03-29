package priorityqueue

// Queue is one slice-backed priority queue.
type Queue[T any] struct {
	items []T
	less  func(left, right T) bool
}
