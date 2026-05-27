package chain

import objectstore "lindenii.org/go/furgit/object/store"

// New creates an ordered object database chain.
//
// The provided backends must be non-nil and distinct.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(backends ...objectstore.Reader) *Chain {
	return &Chain{
		backends: append([]objectstore.Reader(nil), backends...),
	}
}
