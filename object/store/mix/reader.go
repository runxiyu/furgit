package mix

import (
	"errors"
	"fmt"
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadReaderFull reads a full serialized object stream
// from the most-recently-used backend that has it.
func (mix *Mix) ReadReaderFull(id id.ObjectID) (io.ReadCloser, error) {
	for _, backend := range mix.order.Keys() {
		reader, err := backend.ReadReaderFull(id)
		if err == nil {
			mix.order.Touch(backend)

			return reader, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("object/store/mix: read reader full: %w", err)
	}

	return nil, store.ErrObjectNotFound
}

// ReadReaderContent reads an object's type, declared content length,
// and content stream from the most-recently-used backend that has it.
func (mix *Mix) ReadReaderContent(id id.ObjectID) (typ.Type, int, io.ReadCloser, error) {
	for _, backend := range mix.order.Keys() {
		ty, size, reader, err := backend.ReadReaderContent(id)
		if err == nil {
			mix.order.Touch(backend)

			return ty, size, reader, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return typ.Unknown, 0, nil, fmt.Errorf("object/store/mix: read reader content: %w", err)
	}

	return typ.Unknown, 0, nil, store.ErrObjectNotFound
}
