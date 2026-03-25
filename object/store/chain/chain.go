// Package chain provides a wrapper object storage backend to query a chain of
// backends.
package chain

import (
	"codeberg.org/lindenii/furgit/object/store"
)

// Chain queries multiple object databases in order.
//
// Chain borrows its backend stores.
type Chain struct {
	backends []objectstore.Store
}
