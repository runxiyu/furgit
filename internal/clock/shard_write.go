package clock

// add inserts or replaces key, then evicts down to budget.
//
// It reports whether the entry was admitted;
// an entry heavier than the shard budget is rejected
// and leaves the shard unchanged.
func (shard *shard[K, V]) add(key K, value V, weight uint64) bool {
	if weight > shard.maxWeight {
		return false
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if old, ok := shard.items.Load(key); ok {
		shard.unlink(old)
		shard.items.Delete(key)
		shard.weight -= old.weight
		shard.count--
	}

	e := &entry[K, V]{ //nolint:exhaustruct
		key:    key,
		value:  value,
		weight: weight,
	}
	e.referenced.Store(true)

	shard.linkBeforeHand(e)
	shard.items.Store(key, e)
	shard.weight += weight
	shard.count++

	shard.evict()

	return true
}

// evict advances the clock hand until the shard is within budget.
//
// A referenced entry is spared once, its bit cleared;
// after a full rotation of spared entries the next is evicted regardless,
// so a flood of concurrent reads cannot stall progress.
func (shard *shard[K, V]) evict() {
	skips := 0

	for shard.weight > shard.maxWeight && shard.hand != nil {
		victim := shard.hand

		if victim.referenced.Load() && skips < shard.count {
			victim.referenced.Store(false)
			shard.hand = victim.next
			skips++

			continue
		}

		shard.unlink(victim)
		shard.items.Delete(victim.key)
		shard.weight -= victim.weight
		shard.count--
		skips = 0
	}
}

// linkBeforeHand inserts e just behind the hand,
// so a full rotation passes before the sweep examines it.
func (shard *shard[K, V]) linkBeforeHand(e *entry[K, V]) {
	if shard.hand == nil {
		e.prev = e
		e.next = e
		shard.hand = e

		return
	}

	tail := shard.hand.prev
	tail.next = e
	e.prev = tail
	e.next = shard.hand
	shard.hand.prev = e
}

// unlink removes e from the ring,
// moving the hand off e when it points at it.
func (shard *shard[K, V]) unlink(e *entry[K, V]) {
	if e.next == e {
		shard.hand = nil
		e.prev = nil
		e.next = nil

		return
	}

	e.prev.next = e.next
	e.next.prev = e.prev

	if shard.hand == e {
		shard.hand = e.next
	}

	e.prev = nil
	e.next = nil
}
