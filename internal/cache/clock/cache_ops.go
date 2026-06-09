package clock

// Add inserts or replaces key, marking it recently used.
//
// It reports whether the entry was admitted;
// an entry heavier than the per-shard budget is rejected
// and leaves the cache unchanged.
func (cache *Cache[K, V]) Add(key K, value V) bool {
	return cache.shardFor(key).add(key, value, cache.weightFn(key, value))
}

// Get returns the value for key and marks it recently used.
//
//nolint:ireturn
func (cache *Cache[K, V]) Get(key K) (V, bool) {
	return cache.shardFor(key).get(key)
}

// Peek returns the value for key without changing its recency.
//
//nolint:ireturn
func (cache *Cache[K, V]) Peek(key K) (V, bool) {
	return cache.shardFor(key).peek(key)
}

// Len returns the number of cached entries.
func (cache *Cache[K, V]) Len() int {
	total := 0
	for _, shard := range cache.shards {
		total += shard.len()
	}

	return total
}

// Weight returns the current total weight across all shards.
func (cache *Cache[K, V]) Weight() uint64 {
	var total uint64
	for _, shard := range cache.shards {
		total += shard.loadWeight()
	}

	return total
}

// Clear removes all entries.
func (cache *Cache[K, V]) Clear() {
	for _, shard := range cache.shards {
		shard.clear()
	}
}
