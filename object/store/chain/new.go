package chain

import objectstore "codeberg.org/lindenii/furgit/object/store"

// New creates an ordered object database chain.
//
// The provided backends must be non-nil and distinct.
// Chain borrows the provided backends and does not close them in Close.
func New(backends ...objectstore.ReadingStore) *Chain {
	return &Chain{
		backends: append([]objectstore.ReadingStore(nil), backends...),
	}
}
