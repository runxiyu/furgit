package priorityqueue

// New builds one empty priority queue ordered by less.
func New[T any](less func(left, right T) bool) *Queue[T] {
	return &Queue[T]{less: less}
}
