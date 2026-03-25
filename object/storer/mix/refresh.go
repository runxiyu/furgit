package mix

import (
	"errors"

	"codeberg.org/lindenii/furgit/object/storer"
)

// Refresh forwards refresh calls to refresh-capable backends.
func (mix *Mix) Refresh() error {
	mix.mu.RLock()

	backends := make([]objectstorer.Store, 0, len(mix.backendNodeByStore))
	for node := mix.backendHead; node != nil; node = node.next {
		backends = append(backends, node.backend)
	}

	mix.mu.RUnlock()

	var errs []error

	for _, backend := range backends {
		err := backend.Refresh()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
