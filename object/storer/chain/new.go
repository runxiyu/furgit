package chain

import objectstorer "codeberg.org/lindenii/furgit/object/storer"

// New creates an ordered object database chain.
//
// The provided backends must be non-nil and distinct.
// Chain borrows the provided backends and does not close them in Close.
func New(backends ...objectstorer.Store) *Chain {
	return &Chain{
		backends: append([]objectstorer.Store(nil), backends...),
	}
}
