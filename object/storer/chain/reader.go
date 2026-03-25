package chain

import (
	"errors"
	"fmt"
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/storer"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadReaderFull reads a full serialized object stream from the first backend that has it.
func (chain *Chain) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	for i, backend := range chain.backends {
		reader, err := backend.ReadReaderFull(id)
		if err == nil {
			return reader, nil
		}

		if errors.Is(err, objectstorer.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("objectstorer: backend %d read reader full: %w", i, err)
	}

	return nil, objectstorer.ErrObjectNotFound
}

// ReadReaderContent reads an object's type, declared content length, and content stream from the first backend that has it.
func (chain *Chain) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	for i, backend := range chain.backends {
		ty, size, reader, err := backend.ReadReaderContent(id)
		if err == nil {
			return ty, size, reader, nil
		}

		if errors.Is(err, objectstorer.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, 0, nil, fmt.Errorf("objectstorer: backend %d read reader content: %w", i, err)
	}

	return objecttype.TypeInvalid, 0, nil, objectstorer.ErrObjectNotFound
}
