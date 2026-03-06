package lru

// Get returns value for key and marks it most-recently-used.
//
//nolint:ireturn
func (cache *Cache[K, V]) Get(key K) (V, bool) {
	elem, ok := cache.items[key]
	if !ok {
		var zero V

		return zero, false
	}

	cache.lru.MoveToBack(elem)
	//nolint:forcetypeassert
	return elem.Value.(*entry[K, V]).value, true
}
