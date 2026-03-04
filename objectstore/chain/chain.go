// Package chain provides a wrapper object storage backend to query a chain of backends.
package chain

import (
	"codeberg.org/lindenii/furgit/objectstore"
)

// Chain queries multiple object databases in order.
type Chain struct {
	backends []objectstore.Store
}
