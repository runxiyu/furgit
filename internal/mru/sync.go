package mru

// Sync reconciles the order's membership to present.
//
// Surviving keys keep their relative recency order,
// absent keys are dropped,
// and newly present keys are appended after the survivors.
func (order *Order[K]) Sync(present map[K]struct{}) {
	order.mu.Lock()
	defer order.mu.Unlock()

	keys := order.Keys()

	next := make([]K, 0, len(present))
	seen := make(map[K]struct{}, len(present))

	for _, key := range keys {
		if _, ok := present[key]; ok {
			next = append(next, key)
			seen[key] = struct{}{}
		}
	}

	for key := range present {
		if _, ok := seen[key]; !ok {
			next = append(next, key)
		}
	}

	order.snapshot.Store(&next)
}
