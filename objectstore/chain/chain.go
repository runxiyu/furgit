// Package chain provides an ordered object database chain implementation.
package chain

import (
	"errors"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
)

// Chain queries multiple object databases in order.
type Chain struct {
	backends []objectstore.Store
}

// New creates an ordered object database chain.
func New(backends ...objectstore.Store) *Chain {
	return &Chain{
		backends: append([]objectstore.Store(nil), backends...),
	}
}

// ReadBytesFull reads a full serialized object from the first backend that has it.
func (chain *Chain) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			return full, nil
		}
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}
		return nil, fmt.Errorf("objectstore: backend %d read bytes full: %w", i, err)
	}
	return nil, objectstore.ErrObjectNotFound
}

// ReadBytesContent reads an object's type and content bytes from the first backend that has it.
func (chain *Chain) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			return ty, content, nil
		}
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}
		return objecttype.TypeInvalid, nil, fmt.Errorf("objectstore: backend %d read bytes content: %w", i, err)
	}
	return objecttype.TypeInvalid, nil, objectstore.ErrObjectNotFound
}

// ReadReaderFull reads a full serialized object stream from the first backend that has it.
func (chain *Chain) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
		reader, err := backend.ReadReaderFull(id)
		if err == nil {
			return reader, nil
		}
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}
		return nil, fmt.Errorf("objectstore: backend %d read reader full: %w", i, err)
	}
	return nil, objectstore.ErrObjectNotFound
}

// ReadReaderContent reads an object's type and content stream from the first backend that has it.
func (chain *Chain) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, io.ReadCloser, error) {
	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
		ty, reader, err := backend.ReadReaderContent(id)
		if err == nil {
			return ty, reader, nil
		}
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}
		return objecttype.TypeInvalid, nil, fmt.Errorf("objectstore: backend %d read reader content: %w", i, err)
	}
	return objecttype.TypeInvalid, nil, objectstore.ErrObjectNotFound
}

// ReadHeader reads object header data from the first backend that has it.
func (chain *Chain) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			return ty, size, nil
		}
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}
		return objecttype.TypeInvalid, 0, fmt.Errorf("objectstore: backend %d read header: %w", i, err)
	}
	return objecttype.TypeInvalid, 0, objectstore.ErrObjectNotFound
}

// Close closes all backends and joins close errors.
func (chain *Chain) Close() error {
	var errs []error
	for _, backend := range chain.backends {
		if backend == nil {
			continue
		}
		if err := backend.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
