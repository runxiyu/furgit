package chain

import "lindenii.org/go/furgit/ref/store"

// Chain queries an ordered list of reference stores,
// first match wins.
//
// Labels: MT-Safe, Deps-Borrowed, Life-Parent.
type Chain struct {
	backends []store.Reader
}

var _ store.Reader = (*Chain)(nil)

// New builds one reference store chain over the given backends.
//
// The backends are queried in the order given.
// They must each be non-nil and distinct.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(backends ...store.Reader) *Chain {
	return &Chain{
		backends: append([]store.Reader(nil), backends...),
	}
}
