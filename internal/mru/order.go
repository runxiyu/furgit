package mru

import (
	"sync"
	"sync/atomic"
)

// Order is a concurrent most-recently-used ordering over keys.
//
// The front is the most-recently-used key.
// Order expresses recency only,
// never priority;
// a caller that needs a fixed priority order
// must arrange its keys in that order itself.
// Order never evicts.
//
// Reads are lock-free over an immutable snapshot,
// so a concurrent Touch or Sync
// never perturbs an in-progress walk.
//
// Labels: MT-Safe.
type Order[K comparable] struct {
	snapshot atomic.Pointer[[]K] //exhaustruct:optional
	mu       sync.Mutex          //exhaustruct:optional

	interval uint64
	pending  atomic.Uint64 //exhaustruct:optional
}

// Options configures a new Order.
type Options struct {
	// Interval applies a reorder at most once per Interval
	// eligible (non-front, member) Touch calls.
	//
	// A larger Interval decreases recency precision
	// but uses fewer allocations.
	// Each applied reorder allocates one snapshot,
	// so throttling decreases the snapshot-allocation rate
	// by roughly Interval.
	//
	// An Interval of 1 reorders on every eligible Touch.
	Interval uint64
}

// New returns a new, empty order configured by opts.
func New[K comparable](opts Options) *Order[K] {
	if opts.Interval == 0 {
		panic("internal/mru: Options.Interval must be at least 1")
	}

	return &Order[K]{interval: opts.Interval}
}

// Len returns the number of keys in the order.
func (order *Order[K]) Len() int {
	return len(order.Keys())
}

// Keys returns the keys in most-recently-used order,
// front first.
//
// The result is the immutable snapshot current at the call:
// a concurrent Touch or Sync does not affect it.
//
// Labels: Mut-No.
func (order *Order[K]) Keys() []K {
	keys := order.snapshot.Load()
	if keys == nil {
		return nil
	}

	return *keys
}
