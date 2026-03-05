package lru

import "container/list"

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
