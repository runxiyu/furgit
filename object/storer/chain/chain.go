// Package chain provides a wrapper object storage backend to query a chain of
// backends.
package chain

import objectstorer "codeberg.org/lindenii/furgit/object/storer"

// Chain queries multiple object databases in order.
//
// Chain borrows its backend stores.
type Chain struct {
	backends []objectstorer.Store
}
