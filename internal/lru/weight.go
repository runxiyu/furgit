package lru

// WeightFunc reports one entry's weight used for eviction budgeting.
//
// Returned weights MUST be non-negative.
type WeightFunc[K comparable, V any] func(key K, value V) int64

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
