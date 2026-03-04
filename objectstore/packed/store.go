// Package packed provides packfile reading and associated indexes.
package packed

import (
	"os"
	"sync"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Store reads Git objects from pack/index files under an objects/pack root.
//
// Store owns root and closes it in Close.
type Store struct {
	// root is the objects/pack capability used for all file access.
	root *os.Root
	// algo is the expected object ID algorithm for lookups.
	algo objectid.Algorithm

	// discoverOnce guards one-time pack candidate discovery.
	discoverOnce sync.Once
	// discoverErr stores candidate discovery failures.
	discoverErr error
	// candidateHead is the first candidate in lookup priority order.
	candidateHead *packCandidateNode
	// candidateTail is the last candidate in lookup priority order.
	candidateTail *packCandidateNode
	// candidateByPack maps pack basename to discovered candidate.
	candidateByPack map[string]packCandidate
	// candidateNodeByPack maps pack basename to linked-list node.
	candidateNodeByPack map[string]*packCandidateNode
	// idxByPack caches opened and parsed indexes by pack basename.
	idxByPack map[string]*idxFile

	// stateMu guards pack cache and close state.
	stateMu sync.RWMutex
	// candidatesMu guards discovered candidates and MRU order.
	candidatesMu sync.RWMutex
	// idxMu guards parsed index cache.
	idxMu sync.RWMutex
	// cacheMu guards delta cache operations.
	cacheMu sync.RWMutex
	// packs caches opened .pack handles by basename.
	packs map[string]*packFile
	// deltaCache caches resolved base objects by pack location.
	deltaCache *deltaCache
	// closed reports whether Close has been called.
	closed bool
}

const defaultDeltaCacheMaxBytes = 32 << 20

var _ objectstore.Store = (*Store)(nil)

// New creates a packed-object store rooted at an objects/pack directory.
func New(root *os.Root, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	return &Store{
		root:                root,
		algo:                algo,
		candidateByPack:     make(map[string]packCandidate),
		candidateNodeByPack: make(map[string]*packCandidateNode),
		idxByPack:           make(map[string]*idxFile),
		packs:               make(map[string]*packFile),
		deltaCache:          newDeltaCache(defaultDeltaCacheMaxBytes),
	}, nil
}

// Close releases mapped pack/index resources associated with the store.
func (store *Store) Close() error {
	store.stateMu.Lock()

	if store.closed {
		store.stateMu.Unlock()

		return nil
	}

	store.closed = true
	root := store.root
	packs := store.packs
	store.stateMu.Unlock()
	store.idxMu.RLock()
	indexes := store.idxByPack
	store.idxMu.RUnlock()

	var closeErr error

	for _, pack := range packs {
		err := pack.close()
		if err != nil && closeErr == nil {
			closeErr = err
		}
	}

	for _, index := range indexes {
		err := index.close()
		if err != nil && closeErr == nil {
			closeErr = err
		}
	}

	store.cacheMu.Lock()
	store.deltaCache.clear()
	store.cacheMu.Unlock()

	err := root.Close()
	if err != nil && closeErr == nil {
		closeErr = err
	}

	return closeErr
}
