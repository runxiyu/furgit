package mru

// Sync reconciles the order's membership to present.
//
// Surviving keys keep their relative recency order,
// absent keys are dropped,
// and newly present keys are placed before the survivors,
// since new members are presumed recently used.
func (order *Order[K]) Sync(present map[K]struct{}) {
	order.mu.Lock()
	defer order.mu.Unlock()

	keys := order.Keys()

	survivors := make(map[K]struct{}, len(present))

	for _, key := range keys {
		if _, ok := present[key]; ok {
			survivors[key] = struct{}{}
		}
	}

	next := make([]K, 0, len(present))

	for key := range present {
		if _, ok := survivors[key]; !ok {
			next = append(next, key)
		}
	}

	for _, key := range keys {
		if _, ok := survivors[key]; ok {
			next = append(next, key)
		}
	}

	order.snapshot.Store(&next)
}
