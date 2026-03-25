package mix

import (
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store"
)

// ReadSize reads object content length from one backend that has it.
func (mix *Mix) ReadSize(id objectid.ObjectID) (int64, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		size, err := backend.ReadSize(id)
		if err == nil {
			mix.touchBackend(backend)

			return size, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return 0, fmt.Errorf("objectstore: backend %d read size: %w", i, err)
	}

	return 0, objectstore.ErrObjectNotFound
}
