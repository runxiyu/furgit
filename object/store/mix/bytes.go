package mix

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadBytesFull reads a full serialized object
// from the most-recently-used backend that has it.
func (mix *Mix) ReadBytesFull(id id.ObjectID) ([]byte, error) {
	for _, backend := range mix.order.Keys() {
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			mix.order.Touch(backend)

			return full, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("object/store/mix: read bytes full: %w", err)
	}

	return nil, store.ErrObjectNotFound
}

// ReadBytesContent reads an object's type and content bytes
// from the most-recently-used backend that has it.
func (mix *Mix) ReadBytesContent(id id.ObjectID) (typ.Type, []byte, error) {
	for _, backend := range mix.order.Keys() {
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			mix.order.Touch(backend)

			return ty, content, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return typ.Unknown, nil, fmt.Errorf("object/store/mix: read bytes content: %w", err)
	}

	return typ.Unknown, nil, store.ErrObjectNotFound
}
