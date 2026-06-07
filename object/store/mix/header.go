package mix

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadHeader reads object header data
// from the most-recently-used backend that has it.
func (mix *Mix) ReadHeader(id id.ObjectID) (typ.Type, uint64, error) {
	for _, backend := range mix.order.Keys() {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			mix.order.Touch(backend)

			return ty, size, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return typ.TypeUnknown, 0, fmt.Errorf("object/store/mix: read header: %w", err)
	}

	return typ.TypeUnknown, 0, store.ErrObjectNotFound
}

// ReadSize reads object content length
// from the most-recently-used backend that has it.
func (mix *Mix) ReadSize(id id.ObjectID) (uint64, error) {
	for _, backend := range mix.order.Keys() {
		size, err := backend.ReadSize(id)
		if err == nil {
			mix.order.Touch(backend)

			return size, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return 0, fmt.Errorf("object/store/mix: read size: %w", err)
	}

	return 0, store.ErrObjectNotFound
}
