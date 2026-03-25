package mix

import (
	"errors"
	"fmt"
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadReaderFull reads a full serialized object stream from one backend that
// has it.
func (mix *Mix) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		reader, err := backend.ReadReaderFull(id)
		if err == nil {
			mix.touchBackend(backend)

			return reader, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("objectstore: backend %d read reader full: %w", i, err)
	}

	return nil, objectstore.ErrObjectNotFound
}

// ReadReaderContent reads an object's type, declared content length, and
// content stream from one backend that has it.
func (mix *Mix) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		ty, size, reader, err := backend.ReadReaderContent(id)
		if err == nil {
			mix.touchBackend(backend)

			return ty, size, reader, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, 0, nil, fmt.Errorf("objectstore: backend %d read reader content: %w", i, err)
	}

	return objecttype.TypeInvalid, 0, nil, objectstore.ErrObjectNotFound
}
