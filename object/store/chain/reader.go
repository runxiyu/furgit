package chain

import (
	"errors"
	"fmt"
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadReaderFull reads a full serialized object stream
// from the first backend that has it.
func (chain *Chain) ReadReaderFull(id id.ObjectID) (io.ReadCloser, error) {
	for _, backend := range chain.backends {
		reader, err := backend.ReadReaderFull(id)
		if err == nil {
			return reader, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("object/store/chain: read reader full: %w", err)
	}

	return nil, store.ErrObjectNotFound
}

// ReadReaderContent reads an object's type, declared content length,
// and content stream from the first backend that has it.
func (chain *Chain) ReadReaderContent(id id.ObjectID) (typ.Type, uint64, io.ReadCloser, error) {
	for _, backend := range chain.backends {
		ty, size, reader, err := backend.ReadReaderContent(id)
		if err == nil {
			return ty, size, reader, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return typ.TypeUnknown, 0, nil, fmt.Errorf("object/store/chain: read reader content: %w", err)
	}

	return typ.TypeUnknown, 0, nil, store.ErrObjectNotFound
}
