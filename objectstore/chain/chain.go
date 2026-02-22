// Package chain provides an adaptive wrapper over multiple object storage
// backends.
package chain

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
)

// Chain queries multiple object databases with an MRU backend preference.
type Chain struct {
	mu sync.RWMutex

	backendHead        *backendNode
	backendTail        *backendNode
	backendNodeByStore map[objectstore.Store]*backendNode
}

type backendNode struct {
	backend objectstore.Store
	prev    *backendNode
	next    *backendNode
}

// New creates a Chain from backends.
func New(backends ...objectstore.Store) *Chain {
	nodeByStore := make(map[objectstore.Store]*backendNode, len(backends))
	var head *backendNode
	var tail *backendNode

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

	return &Chain{
		backendHead:        head,
		backendTail:        tail,
		backendNodeByStore: nodeByStore,
	}
}

func (chain *Chain) firstBackend() objectstore.Store {
	chain.mu.RLock()
	defer chain.mu.RUnlock()
	if chain.backendHead == nil {
		return nil
	}
	return chain.backendHead.backend
}

func (chain *Chain) nextBackend(current objectstore.Store) objectstore.Store {
	chain.mu.RLock()
	defer chain.mu.RUnlock()
	node := chain.backendNodeByStore[current]
	if node == nil || node.next == nil {
		return nil
	}
	return node.next.backend
}

func (chain *Chain) touchBackend(backend objectstore.Store) {
	if backend == nil {
		return
	}
	if !chain.mu.TryLock() {
		return
	}
	defer chain.mu.Unlock()

	node := chain.backendNodeByStore[backend]
	if node == nil || node == chain.backendHead {
		return
	}
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if chain.backendTail == node {
		chain.backendTail = node.prev
	}

	node.prev = nil
	node.next = chain.backendHead
	if chain.backendHead != nil {
		chain.backendHead.prev = node
	}
	chain.backendHead = node
	if chain.backendTail == nil {
		chain.backendTail = node
	}
}

// ReadBytesFull reads a full serialized object from one backend that has it.
func (chain *Chain) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	for i, backend := 0, chain.firstBackend(); backend != nil; i, backend = i+1, chain.nextBackend(backend) {
		full, err := backend.ReadBytesFull(id)
		if err == nil {
			chain.touchBackend(backend)
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
func (chain *Chain) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	for i, backend := 0, chain.firstBackend(); backend != nil; i, backend = i+1, chain.nextBackend(backend) {
		ty, content, err := backend.ReadBytesContent(id)
		if err == nil {
			chain.touchBackend(backend)
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
func (chain *Chain) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	for i, backend := 0, chain.firstBackend(); backend != nil; i, backend = i+1, chain.nextBackend(backend) {
		reader, err := backend.ReadReaderFull(id)
		if err == nil {
			chain.touchBackend(backend)
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
func (chain *Chain) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	for i, backend := 0, chain.firstBackend(); backend != nil; i, backend = i+1, chain.nextBackend(backend) {
		ty, size, reader, err := backend.ReadReaderContent(id)
		if err == nil {
			chain.touchBackend(backend)
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
func (chain *Chain) ReadSize(id objectid.ObjectID) (int64, error) {
	for i, backend := 0, chain.firstBackend(); backend != nil; i, backend = i+1, chain.nextBackend(backend) {
		size, err := backend.ReadSize(id)
		if err == nil {
			chain.touchBackend(backend)
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
func (chain *Chain) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	for i, backend := 0, chain.firstBackend(); backend != nil; i, backend = i+1, chain.nextBackend(backend) {
		ty, size, err := backend.ReadHeader(id)
		if err == nil {
			chain.touchBackend(backend)
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
	chain.mu.RLock()
	backends := make([]objectstore.Store, 0, len(chain.backendNodeByStore))
	for node := chain.backendHead; node != nil; node = node.next {
		backends = append(backends, node.backend)
	}
	chain.mu.RUnlock()

	var errs []error
	for _, backend := range backends {
		if err := backend.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
