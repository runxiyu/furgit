package mix

import (
	"errors"

	"lindenii.org/go/furgit/internal/mru"
	"lindenii.org/go/furgit/object/store"
)

// Mix queries multiple object databases
// with a most-recently-used backend preference.
//
// Labels: Close-Caller.
type Mix struct {
	order *mru.Order[store.ObjectReader]
}

var _ store.ObjectReader = (*Mix)(nil)

// New creates a Mix from backends.
//
// The provided backends must be non-nil and distinct.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(backends ...store.ObjectReader) *Mix {
	present := make(map[store.ObjectReader]struct{}, len(backends))
	for _, backend := range backends {
		present[backend] = struct{}{}
	}

	order := mru.New[store.ObjectReader]()
	order.Sync(present)

	return &Mix{
		order: order,
	}
}

// Refresh forwards refresh calls to all backends.
func (mix *Mix) Refresh() error {
	var errs []error

	for _, backend := range mix.order.Keys() {
		err := backend.Refresh()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
