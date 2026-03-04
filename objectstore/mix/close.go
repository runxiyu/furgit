package mix

import (
	"errors"

	"codeberg.org/lindenii/furgit/objectstore"
)

// Close closes all backends and joins close errors.
func (mix *Mix) Close() error {
	mix.mu.RLock()

	backends := make([]objectstore.Store, 0, len(mix.backendNodeByStore))
	for node := mix.backendHead; node != nil; node = node.next {
		backends = append(backends, node.backend)
	}

	mix.mu.RUnlock()

	var errs []error

	for _, backend := range backends {
		err := backend.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
