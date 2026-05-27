package chain

import (
	"errors"
	"fmt"

	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ReadHeader reads object header data from the first backend that has it.
func (chain *Chain) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	for i, backend := range chain.backends {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			return ty, size, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore: backend %d read header: %w", i, err)
	}

	return objecttype.TypeInvalid, 0, objectstore.ErrObjectNotFound
}
