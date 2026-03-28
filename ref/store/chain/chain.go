// Package chain provides a wrapper reference storage backend to query a chain
// of backends.
package chain

import refstore "codeberg.org/lindenii/furgit/ref/store"

// Chain queries multiple reference stores in order.
//
// Labels: Close-Caller.
type Chain struct {
	backends []refstore.ReadingStore
}
