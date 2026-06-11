package chain

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadBytesFull reads a full serialized object
// from the first backend that has it.
func (chain *Chain) ReadBytesFull(id id.ObjectID) ([]byte, error) {
	for _, backend := range chain.backends {
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			return full, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("object/store/chain: read bytes full: %w", err)
	}

	return nil, store.ErrObjectNotFound
}

// ReadBytesContent reads an object's type and content bytes
// from the first backend that has it.
func (chain *Chain) ReadBytesContent(id id.ObjectID) (typ.Type, []byte, error) {
	for _, backend := range chain.backends {
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			return ty, content, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return typ.Unknown, nil, fmt.Errorf("object/store/chain: read bytes content: %w", err)
	}

	return typ.Unknown, nil, store.ErrObjectNotFound
}
