package chain

import "errors"

// Close closes all backends and joins close errors.
func (chain *Chain) Close() error {
	var errs []error

	for _, backend := range chain.backends {
		if backend == nil {
			continue
		}

		err := backend.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
