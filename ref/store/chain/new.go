package chain

import refstore "codeberg.org/lindenii/furgit/ref/store"

// New creates an ordered reference store chain.
//
// The provided backends must be non-nil and distinct.
// Chain borrows the provided backends and does not close them in Close.
func New(backends ...refstore.ReadingStore) *Chain {
	return &Chain{
		backends: append([]refstore.ReadingStore(nil), backends...),
	}
}
