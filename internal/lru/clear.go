package lru

// Clear removes all entries from the cache.
func (cache *Cache[K, V]) Clear() {
	for elem := cache.lru.Front(); elem != nil; {
		next := elem.Next()
		cache.removeElem(elem)
		elem = next
	}
}
