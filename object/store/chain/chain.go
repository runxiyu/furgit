package chain

import (
	"errors"

	"lindenii.org/go/furgit/object/store"
)

// Chain queries multiple object databases in order.
//
// Labels: Close-Caller.
type Chain struct {
	backends []store.ObjectReader
}

var _ store.ObjectReader = (*Chain)(nil)

// New creates an ordered object database chain.
//
// The provided backends must be non-nil and distinct.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(backends ...store.ObjectReader) *Chain {
	return &Chain{
		backends: append([]store.ObjectReader(nil), backends...),
	}
}

// Refresh forwards refresh calls to all backends.
func (chain *Chain) Refresh() error {
	var errs []error

	for _, backend := range chain.backends {
		err := backend.Refresh()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
