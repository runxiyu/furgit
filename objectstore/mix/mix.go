// Package mix provides an adaptive wrapper over multiple object storage
// backends.
package mix

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
)

// Mix queries multiple object databases with an MRU backend preference.
type Mix struct {
	mu sync.RWMutex

	backendHead        *backendNode
	backendTail        *backendNode
	backendNodeByStore map[objectstore.Store]*backendNode
}

// New creates a Mix from backends.
func New(backends ...objectstore.Store) *Mix {
	nodeByStore := make(map[objectstore.Store]*backendNode, len(backends))

	var (
		head *backendNode
		tail *backendNode
	)

	for _, backend := range backends {
		if backend == nil {
			continue
		}

		node := &backendNode{
			backend: backend,
			prev:    tail,
		}
		if tail != nil {
			tail.next = node
		}

		if head == nil {
			head = node
		}

		tail = node
		nodeByStore[backend] = node
	}

	return &Mix{
		backendHead:        head,
		backendTail:        tail,
		backendNodeByStore: nodeByStore,
	}
}

// ReadBytesFull reads a full serialized object from one backend that has it.
func (mix *Mix) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			mix.touchBackend(backend)

			return full, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return nil, fmt.Errorf("objectstore: backend %d read bytes full: %w", i, err)
	}

	return nil, objectstore.ErrObjectNotFound
}

// ReadBytesContent reads an object's type and content bytes from one backend
// that has it.
func (mix *Mix) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			mix.touchBackend(backend)

			return ty, content, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return objecttype.TypeInvalid, nil, fmt.Errorf("objectstore: backend %d read bytes content: %w", i, err)
	}

	return objecttype.TypeInvalid, nil, objectstore.ErrObjectNotFound
}

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

// ReadSize reads object content length from one backend that has it.
func (mix *Mix) ReadSize(id objectid.ObjectID) (int64, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		size, err := backend.ReadSize(id)
		if err == nil {
			mix.touchBackend(backend)

			return size, nil
		}

		if errors.Is(err, objectstore.ErrObjectNotFound) {
			continue
		}

		return 0, fmt.Errorf("objectstore: backend %d read size: %w", i, err)
	}

	return 0, objectstore.ErrObjectNotFound
}

// ReadHeader reads object header data from one backend that has it.
func (mix *Mix) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	for i, backend := 0, mix.firstBackend(); backend != nil; i, backend = i+1, mix.nextBackend(backend) {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			mix.touchBackend(backend)

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
func (mix *Mix) Close() error {
	mix.mu.RLock()

	backends := make([]objectstore.Store, 0, len(mix.backendNodeByStore))
	for node := mix.backendHead; node != nil; node = node.next {
		backends = append(backends, node.backend)
	}

	mix.mu.RUnlock()

	var errs []error

	for _, backend := range backends {
		err := backend.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type backendNode struct {
	backend objectstore.Store
	prev    *backendNode
	next    *backendNode
}

func (mix *Mix) firstBackend() objectstore.Store {
	mix.mu.RLock()
	defer mix.mu.RUnlock()

	if mix.backendHead == nil {
		return nil
	}

	return mix.backendHead.backend
}

func (mix *Mix) nextBackend(current objectstore.Store) objectstore.Store {
	mix.mu.RLock()
	defer mix.mu.RUnlock()

	node := mix.backendNodeByStore[current]
	if node == nil || node.next == nil {
		return nil
	}

	return node.next.backend
}

func (mix *Mix) touchBackend(backend objectstore.Store) {
	if backend == nil {
		return
	}

	if !mix.mu.TryLock() {
		return
	}
	defer mix.mu.Unlock()

	node := mix.backendNodeByStore[backend]
	if node == nil || node == mix.backendHead {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	}

	if mix.backendTail == node {
		mix.backendTail = node.prev
	}

	node.prev = nil

	node.next = mix.backendHead
	if mix.backendHead != nil {
		mix.backendHead.prev = node
	}

	mix.backendHead = node
	if mix.backendTail == nil {
		mix.backendTail = node
	}
}
