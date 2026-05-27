// Package chain provides a wrapper object storage backend to query a chain of
// backends.
package chain

import objectstore "lindenii.org/go/furgit/object/store"

// Chain queries multiple object databases in order.
//
// Labels: Close-Caller.
type Chain struct {
	backends []objectstore.Reader
}
