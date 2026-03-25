package chain

import (
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstorer "codeberg.org/lindenii/furgit/object/storer"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadHeader reads object header data from the first backend that has it.
func (chain *Chain) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	for i, backend := range chain.backends {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			return ty, size, nil
		}

		if errors.Is(err, objectstorer.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, 0, fmt.Errorf("objectstorer: backend %d read header: %w", i, err)
	}

	return objecttype.TypeInvalid, 0, objectstorer.ErrObjectNotFound
}
