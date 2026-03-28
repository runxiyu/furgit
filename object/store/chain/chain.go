// Package chain provides a wrapper object storage backend to query a chain of
// backends.
package chain

import objectstore "codeberg.org/lindenii/furgit/object/store"

// Chain queries multiple object databases in order.
//
// Labels: Close-Caller.
type Chain struct {
	backends []objectstore.ReadingStore
}
