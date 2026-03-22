package chain

import "codeberg.org/lindenii/furgit/objectstore"

// New creates an ordered object database chain.
//
// The provided backends must be non-nil and distinct.
// Chain borrows the provided backends and does not close them in Close.
func New(backends ...objectstore.Store) *Chain {
	return &Chain{
		backends: append([]objectstore.Store(nil), backends...),
	}
}
