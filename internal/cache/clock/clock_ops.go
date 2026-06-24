package clock

// Add inserts or replaces key, marking it recently used.
//
// It reports whether the entry was admitted;
// an entry heavier than the per-shard budget is rejected
// and leaves the cache unchanged.
func (clock *Clock[K, V]) Add(key K, value V) bool {
	return clock.shardFor(key).add(key, value, clock.weightFn(key, value))
}

// Get returns the value for key and marks it recently used.
func (clock *Clock[K, V]) Get(key K) (V, bool) {
	return clock.shardFor(key).get(key)
}

// Peek returns the value for key without changing its recency.
func (clock *Clock[K, V]) Peek(key K) (V, bool) {
	return clock.shardFor(key).peek(key)
}

// Len returns the number of cached entries.
func (clock *Clock[K, V]) Len() int {
	total := 0
	for _, shard := range clock.shards {
		total += shard.len()
	}

	return total
}

// Weight returns the current total weight across all shards.
func (clock *Clock[K, V]) Weight() uint64 {
	var total uint64
	for _, shard := range clock.shards {
		total += shard.loadWeight()
	}

	return total
}

// Clear removes all entries.
func (clock *Clock[K, V]) Clear() {
	for _, shard := range clock.shards {
		shard.clear()
	}
}
