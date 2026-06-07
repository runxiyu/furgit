package mru

// Touch moves key to the front, best-effort.
//
// When key is already the most-recently-used,
// Touch is lock-free and allocation-free;
// this is the common case.
// Otherwise Touch reorders under a non-blocking lock attempt,
// so a read never blocks merely to reorder.
// A contended attempt,
// or a key that is not a member,
// leaves the order unchanged.
func (order *Order[K]) Touch(key K) {
	keys := order.Keys()
	if len(keys) == 0 || keys[0] == key {
		return
	}

	if !order.mu.TryLock() {
		return
	}
	defer order.mu.Unlock()

	// The snapshot may have changed before we took the lock.
	keys = order.Keys()

	index := -1

	for i, candidate := range keys {
		if candidate == key {
			index = i

			break
		}
	}

	if index <= 0 {
		return
	}

	next := make([]K, 0, len(keys))
	next = append(next, key)
	next = append(next, keys[:index]...)
	next = append(next, keys[index+1:]...)

	order.snapshot.Store(&next)
}
