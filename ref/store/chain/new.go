package chain

import refstore "codeberg.org/lindenii/furgit/ref/store"

// New creates an ordered reference store chain.
//
// The provided backends must be non-nil and distinct.
//
// Labels: Deps-Borrowed.
func New(backends ...refstore.ReadingStore) *Chain {
	return &Chain{
		backends: append([]refstore.ReadingStore(nil), backends...),
	}
}
