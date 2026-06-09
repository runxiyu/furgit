package clock

import (
	"sync"
	"sync/atomic"

	lsync "lindenii.org/go/lgo/sync"
)

// entry is one cached key/value with CLOCK.
type entry[K comparable, V any] struct {
	key        K
	value      V
	weight     uint64
	prev, next *entry[K, V]

	// referenced is set on access and cleared by the eviction sweep;
	// prev and next link the entry into its shard's ring.
	referenced atomic.Bool
}

// shard is an independently locked CLOCK cache.
type shard[K comparable, V any] struct {
	items lsync.Map[K, *entry[K, V]]

	hand      *entry[K, V]
	weight    uint64
	count     int
	maxWeight uint64

	// mu protects the ring, hand, totals, and writes.
	mu sync.Mutex
}

// newShard returns an empty shard with the given weight budget.
func newShard[K comparable, V any](maxWeight uint64) *shard[K, V] {
	return &shard[K, V]{ //nolint:exhaustruct
		maxWeight: maxWeight,
	}
}

// len returns the shard's entry count.
func (shard *shard[K, V]) len() int {
	shard.mu.Lock()
	defer shard.mu.Unlock()

	return shard.count
}

// loadWeight returns the shard's current total weight.
func (shard *shard[K, V]) loadWeight() uint64 {
	shard.mu.Lock()
	defer shard.mu.Unlock()

	return shard.weight
}

// clear removes every entry.
func (shard *shard[K, V]) clear() {
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.items.Clear()
	shard.hand = nil
	shard.weight = 0
	shard.count = 0
}
