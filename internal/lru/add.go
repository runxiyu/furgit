package lru

// Add inserts or replaces key and marks it most-recently-used.
//
// Add returns false when the entry's weight exceeds MaxWeight even for an empty
// cache. In that case the cache is unchanged.
//
// Add panics if weightFn returns a negative weight.
func (cache *Cache[K, V]) Add(key K, value V) bool {
	w := cache.weightFn(key, value)
	if w < 0 {
		panic("lru: negative entry weight")
	}

	if w > cache.maxWeight {
		return false
	}

	if elem, ok := cache.items[key]; ok {
		cache.removeElem(elem)
	}

	ent := &entry[K, V]{
		key:    key,
		value:  value,
		weight: w,
	}
	elem := cache.lru.PushBack(ent)
	cache.items[key] = elem
	cache.weight += w

	cache.evictOverBudget()

	return true
}
