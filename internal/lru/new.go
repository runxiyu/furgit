package lru

import "container/list"

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
