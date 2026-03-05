package lru

// OnEvictFunc runs when an entry leaves the cache.
//
// It is called for evictions, explicit removals, Clear, and replacement by Add.
type OnEvictFunc[K comparable, V any] func(key K, value V)

func (cache *Cache[K, V]) evictOverBudget() {
	for cache.weight > cache.maxWeight {
		elem := cache.lru.Front()
		if elem == nil {
			return
		}

		cache.removeElem(elem)
	}
}
