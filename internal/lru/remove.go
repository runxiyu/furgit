package lru

import "container/list"

// Remove deletes key from the cache.
//
//nolint:ireturn
func (cache *Cache[K, V]) Remove(key K) (V, bool) {
	elem, ok := cache.items[key]
	if !ok {
		var zero V

		return zero, false
	}

	ent := cache.removeElem(elem)

	return ent.value, true
}

func (cache *Cache[K, V]) removeElem(elem *list.Element) *entry[K, V] {
	//nolint:forcetypeassert
	ent := elem.Value.(*entry[K, V])
	cache.lru.Remove(elem)
	delete(cache.items, ent.key)

	cache.weight -= ent.weight
	if cache.onEvict != nil {
		cache.onEvict(ent.key, ent.value)
	}

	return ent
}
