package memory

import (
	"sync"

	objectid "codeberg.org/lindenii/furgit/object/id"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// Store reads and writes one in-memory Git reference namespace.
//
// Labels: Close-Caller.
type Store struct {
	mu   sync.RWMutex
	algo objectid.Algorithm
	refs map[string]storedRef
}

var (
	_ refstore.Reader        = (*Store)(nil)
	_ refstore.Transactioner = (*Store)(nil)
	_ refstore.Batcher       = (*Store)(nil)
)

// New builds one empty in-memory reference store for one object format.
func New(algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	return &Store{
		algo: algo,
		refs: make(map[string]storedRef),
	}, nil
}

// Close closes the in-memory reference store.
//
// Labels: MT-Unsafe.
func (store *Store) Close() error {
	return nil
}
