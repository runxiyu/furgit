package lru

// Len returns the number of cached entries.
func (cache *Cache[K, V]) Len() int {
	return len(cache.items)
}
