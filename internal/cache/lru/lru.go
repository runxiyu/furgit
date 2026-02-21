// Package lru provides a weighted least-recently-used cache.
package lru

import "container/list"

// WeightFunc reports one entry's weight used for eviction budgeting.
//
// Returned weights MUST be non-negative.
type WeightFunc[K comparable, V any] func(key K, value V) int64

// OnEvictFunc runs when an entry leaves the cache.
//
// It is called for evictions, explicit removals, Clear, and replacement by Add.
type OnEvictFunc[K comparable, V any] func(key K, value V)

// Cache is a non-concurrent weighted LRU cache.
//
// Methods on Cache are not safe for concurrent use.
type Cache[K comparable, V any] struct {
	maxWeight int64
	weightFn  WeightFunc[K, V]
	onEvict   OnEvictFunc[K, V]

	weight int64
	items  map[K]*list.Element
	lru    list.List
}

type entry[K comparable, V any] struct {
	key    K
	value  V
	weight int64
}

// New creates a cache with a maximum total weight.
//
// New panics if maxWeight is negative or weightFn is nil.
func New[K comparable, V any](maxWeight int64, weightFn WeightFunc[K, V], onEvict OnEvictFunc[K, V]) *Cache[K, V] {
	if maxWeight < 0 {
		panic("lru: negative max weight")
	}
	if weightFn == nil {
		panic("lru: nil weight function")
	}
	return &Cache[K, V]{
		maxWeight: maxWeight,
		weightFn:  weightFn,
		onEvict:   onEvict,
		items:     make(map[K]*list.Element),
	}
}

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

// Get returns value for key and marks it most-recently-used.
func (cache *Cache[K, V]) Get(key K) (V, bool) {
	elem, ok := cache.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	cache.lru.MoveToBack(elem)
	return elem.Value.(*entry[K, V]).value, true
}

// Peek returns value for key without changing recency.
func (cache *Cache[K, V]) Peek(key K) (V, bool) {
	elem, ok := cache.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	return elem.Value.(*entry[K, V]).value, true
}

// Remove deletes key from the cache.
func (cache *Cache[K, V]) Remove(key K) (V, bool) {
	elem, ok := cache.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	ent := cache.removeElem(elem)
	return ent.value, true
}

// Clear removes all entries from the cache.
func (cache *Cache[K, V]) Clear() {
	for elem := cache.lru.Front(); elem != nil; {
		next := elem.Next()
		cache.removeElem(elem)
		elem = next
	}
}

// Len returns the number of cached entries.
func (cache *Cache[K, V]) Len() int {
	return len(cache.items)
}

// Weight returns the current total weight.
func (cache *Cache[K, V]) Weight() int64 {
	return cache.weight
}

// MaxWeight returns the configured total weight budget.
func (cache *Cache[K, V]) MaxWeight() int64 {
	return cache.maxWeight
}

// SetMaxWeight updates the total weight budget and evicts LRU entries as
// needed.
//
// SetMaxWeight panics if maxWeight is negative.
func (cache *Cache[K, V]) SetMaxWeight(maxWeight int64) {
	if maxWeight < 0 {
		panic("lru: negative max weight")
	}
	cache.maxWeight = maxWeight
	cache.evictOverBudget()
}

func (cache *Cache[K, V]) evictOverBudget() {
	for cache.weight > cache.maxWeight {
		elem := cache.lru.Front()
		if elem == nil {
			return
		}
		cache.removeElem(elem)
	}
}

func (cache *Cache[K, V]) removeElem(elem *list.Element) *entry[K, V] {
	ent := elem.Value.(*entry[K, V])
	cache.lru.Remove(elem)
	delete(cache.items, ent.key)
	cache.weight -= ent.weight
	if cache.onEvict != nil {
		cache.onEvict(ent.key, ent.value)
	}
	return ent
}
