package chain

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadHeader reads object header data
// from the first backend that has it.
func (chain *Chain) ReadHeader(id id.ObjectID) (typ.Type, uint64, error) {
	for _, backend := range chain.backends {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			return ty, size, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return typ.TypeUnknown, 0, fmt.Errorf("object/store/chain: read header: %w", err)
	}

	return typ.TypeUnknown, 0, store.ErrObjectNotFound
}

// ReadSize reads object content length
// from the first backend that has it.
func (chain *Chain) ReadSize(id id.ObjectID) (uint64, error) {
	for _, backend := range chain.backends {
		size, err := backend.ReadSize(id)
		if err == nil {
			return size, nil
		}

		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		return 0, fmt.Errorf("object/store/chain: read size: %w", err)
	}

	return 0, store.ErrObjectNotFound
}
