package chain

import "codeberg.org/lindenii/furgit/objectstore"

// New creates an ordered object database chain.
func New(backends ...objectstore.Store) *Chain {
	return &Chain{
		backends: append([]objectstore.Store(nil), backends...),
	}
}
