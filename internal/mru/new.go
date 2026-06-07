package mru

// New returns a new, empty order.
func New[K comparable]() *Order[K] {
	return &Order[K]{} //nolint:exhaustruct
}
