package mix

import (
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/objectstore"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadHeader reads object header data from one backend that has it.
func (mix *Mix) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			mix.touchBackend(backend)

			return ty, size, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore: backend %d read header: %w", i, err)
	}

	return objecttype.TypeInvalid, 0, objectstore.ErrObjectNotFound
}
