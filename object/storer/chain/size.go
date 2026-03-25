package chain

import (
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstorer "codeberg.org/lindenii/furgit/object/storer"
)

// ReadSize reads object content length from the first backend that has it.
func (chain *Chain) ReadSize(id objectid.ObjectID) (int64, error) {
	for i, backend := range chain.backends {
		size, err := backend.ReadSize(id)
		if err == nil {
			return size, nil
		}

		if errors.Is(err, objectstorer.ErrObjectNotFound) {
			continue
		}

		return 0, fmt.Errorf("objectstorer: backend %d read size: %w", i, err)
	}

	return 0, objectstorer.ErrObjectNotFound
}
