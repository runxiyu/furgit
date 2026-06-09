package clock

import (
	"hash/maphash"
	"runtime"

	"lindenii.org/go/lgo/intconv"
)

// maxShards bounds the shard count.
//
// Keep it relatively modest
// so the per-shard budget
// stays large enough to admit sizable values.
const maxShards = 16

// WeightFunc reports one entry's weight, used for eviction budgeting.
type WeightFunc[K comparable, V any] func(key K, value V) uint64

// Cache is a concurrent, weight-bounded cache
// with CLOCK eviction.
//
// Reads are lock-free;
// writes lock only the shard that owns the key.
//
// Labels: MT-Safe.
type Cache[K comparable, V any] struct {
	shards   []*shard[K, V]
	seed     maphash.Seed
	mask     uint64
	weightFn WeightFunc[K, V]
}

// New returns a cache bounded to maxWeight total weight,
// weighing entries with weightFn.
//
// New panics if weightFn is nil.
func New[K comparable, V any](maxWeight uint64, weightFn WeightFunc[K, V]) *Cache[K, V] {
	if weightFn == nil {
		panic("internal/clock: nil weight function")
	}

	count, mask := shardLayout(maxWeight)
	perShard := maxWeight / (mask + 1)

	shards := make([]*shard[K, V], count)
	for i := range shards {
		shards[i] = newShard[K, V](perShard)
	}

	return &Cache[K, V]{
		shards:   shards,
		seed:     maphash.MakeSeed(),
		mask:     mask,
		weightFn: weightFn,
	}
}

// shardLayout picks a power-of-two shard count and its address mask.
//
// Tracks GOMAXPROCS, capped at maxShards,
// and is shrunk so the per-shard budget
// stays at least one while maxWeight is nonzero.
func shardLayout(maxWeight uint64) (int, uint64) {
	count := 1
	for count < runtime.GOMAXPROCS(0) && count < maxShards {
		count *= 2
	}

	countU, err := intconv.IntToUint64(count)
	if err != nil {
		return 1, 0
	}

	for countU > maxWeight && countU > 1 {
		count /= 2
		countU /= 2
	}

	return count, countU - 1
}

// shardFor returns the shard that owns key.
func (cache *Cache[K, V]) shardFor(key K) *shard[K, V] {
	return cache.shards[maphash.Comparable(cache.seed, key)&cache.mask]
}
