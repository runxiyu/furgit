package chain

import (
	"errors"
	"fmt"

	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
)

// ReadSize reads object content length from the first backend that has it.
func (chain *Chain) ReadSize(id objectid.ObjectID) (int64, error) {
	for i, backend := range chain.backends {
		size, err := backend.ReadSize(id)
		if err == nil {
			return size, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return 0, fmt.Errorf("objectstore: backend %d read size: %w", i, err)
	}

	return 0, objectstore.ErrObjectNotFound
}
