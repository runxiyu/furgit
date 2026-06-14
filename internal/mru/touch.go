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
//
// When the order has a reorder interval above 1,
// an eligible (non-front) Touch records its recency
// but applies the reorder only once per interval such calls;
// the recording itself is lock-free and allocation-free.
func (order *Order[K]) Touch(key K) {
	keys := order.Keys()
	if len(keys) == 0 || keys[0] == key {
		return
	}

	if order.interval > 1 && order.pending.Add(1)%order.interval != 0 {
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
