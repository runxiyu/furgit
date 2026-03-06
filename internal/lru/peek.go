package lru

// Peek returns value for key without changing recency.
//
//nolint:ireturn
func (cache *Cache[K, V]) Peek(key K) (V, bool) {
	elem, ok := cache.items[key]
	if !ok {
		var zero V

		return zero, false
	}
	//nolint:forcetypeassert
	return elem.Value.(*entry[K, V]).value, true
}
