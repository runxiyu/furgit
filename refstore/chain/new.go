package chain

import "codeberg.org/lindenii/furgit/refstore"

// New creates an ordered reference store chain.
func New(backends ...refstore.ReadingStore) *Chain {
	return &Chain{
		backends: append([]refstore.ReadingStore(nil), backends...),
	}
}
