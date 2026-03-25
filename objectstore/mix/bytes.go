package mix

import (
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/objectstore"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadBytesFull reads a full serialized object from one backend that has it.
func (mix *Mix) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			mix.touchBackend(backend)

			return full, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("objectstore: backend %d read bytes full: %w", i, err)
	}

	return nil, objectstore.ErrObjectNotFound
}

// ReadBytesContent reads an object's type and content bytes from one backend
// that has it.
func (mix *Mix) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			mix.touchBackend(backend)

			return ty, content, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, nil, fmt.Errorf("objectstore: backend %d read bytes content: %w", i, err)
	}

	return objecttype.TypeInvalid, nil, objectstore.ErrObjectNotFound
}
