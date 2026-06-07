package mru

// Len returns the number of keys in the order.
func (order *Order[K]) Len() int {
	return len(order.Keys())
}
