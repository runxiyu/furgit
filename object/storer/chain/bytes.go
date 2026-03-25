package chain

import (
	"errors"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/storer"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadBytesFull reads a full serialized object from the first backend that has it.
func (chain *Chain) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	for i, backend := range chain.backends {
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			return full, nil
		}

		if errors.Is(err, objectstorer.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("objectstorer: backend %d read bytes full: %w", i, err)
	}

	return nil, objectstorer.ErrObjectNotFound
}

// ReadBytesContent reads an object's type and content bytes from the first backend that has it.
func (chain *Chain) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	for i, backend := range chain.backends {
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			return ty, content, nil
		}

		if errors.Is(err, objectstorer.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, nil, fmt.Errorf("objectstorer: backend %d read bytes content: %w", i, err)
	}

	return objecttype.TypeInvalid, nil, objectstorer.ErrObjectNotFound
}
