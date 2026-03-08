package chain

import "errors"

// Refresh forwards refresh calls to all backends.
func (chain *Chain) Refresh() error {
	var errs []error

	for _, backend := range chain.backends {
		if backend == nil {
			continue
		}

		err := backend.Refresh()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
