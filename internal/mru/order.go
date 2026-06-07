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
