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
	snapshot atomic.Pointer[[]K]
	mu       sync.Mutex
}

// New returns a new, empty order.
func New[K comparable]() *Order[K] {
	return &Order[K]{} //nolint:exhaustruct
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
