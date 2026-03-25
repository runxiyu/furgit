// Package chain provides a wrapper reference storage backend to query a chain
// of backends.
package chain

import refstore "codeberg.org/lindenii/furgit/ref/store"

// Chain queries multiple reference stores in order.
//
// Chain borrows its backend stores.
type Chain struct {
	backends []refstore.ReadingStore
}
