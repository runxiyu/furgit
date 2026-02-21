// Package packed provides packfile reading and associated indexes.
package packed

import (
	"errors"
	"fmt"
	"os"
	"sync"

	packchecksum "codeberg.org/lindenii/furgit/format/pack/checksum"
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
	// candidates stores known pack/index pairs in lookup priority order.
	candidates []packCandidate
	// candidateByPack maps pack basename to discovered candidate.
	candidateByPack map[string]packCandidate
	// idxByPack caches opened and parsed indexes by pack basename.
	idxByPack map[string]*idxFile

	// stateMu guards index publication, pack cache, and close state.
	stateMu sync.RWMutex
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
		root:            root,
		algo:            algo,
		candidateByPack: make(map[string]packCandidate),
		idxByPack:       make(map[string]*idxFile),
		packs:           make(map[string]*packFile),
		deltaCache:      newDeltaCache(defaultDeltaCacheMaxBytes),
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
	indexes := store.idxByPack
	store.stateMu.Unlock()

	var closeErr error
	for _, pack := range packs {
		if err := pack.close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	for _, index := range indexes {
		if err := index.close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	store.cacheMu.Lock()
	store.deltaCache.clear()
	store.cacheMu.Unlock()

	if err := root.Close(); err != nil && closeErr == nil {
		closeErr = err
	}
	return closeErr
}

// lookup resolves one object ID to its pack location.
func (store *Store) lookup(id objectid.ObjectID) (location, error) {
	var zero location
	if id.Algorithm() != store.algo {
		return zero, errors.New("objectstore/packed: object id algorithm mismatch")
	}
	if err := store.ensureCandidates(); err != nil {
		return zero, err
	}

	store.stateMu.RLock()
	candidates := append([]packCandidate(nil), store.candidates...)
	store.stateMu.RUnlock()

	for _, candidate := range candidates {
		index, err := store.openIndex(candidate)
		if err != nil {
			return zero, err
		}
		offset, ok, err := index.lookup(id)
		if err != nil {
			return zero, err
		}
		if ok {
			store.touchCandidate(candidate.packName)
			return location{packName: index.packName, offset: offset}, nil
		}
	}
	return zero, objectstore.ErrObjectNotFound
}

// openPack returns one opened and validated pack handle.
func (store *Store) openPack(name string) (*packFile, error) {
	store.stateMu.RLock()
	if pack, ok := store.packs[name]; ok {
		store.stateMu.RUnlock()
		return pack, nil
	}
	store.stateMu.RUnlock()

	file, err := store.root.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	pack, err := openPackFile(name, file, info.Size())
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if err := store.verifyPackMatchesIndexes(pack); err != nil {
		_ = pack.close()
		return nil, err
	}

	store.stateMu.Lock()
	if existing, ok := store.packs[name]; ok {
		store.stateMu.Unlock()
		_ = pack.close()
		return existing, nil
	}
	store.packs[name] = pack
	store.stateMu.Unlock()
	return pack, nil
}

// verifyPackMatchesIndexes checks that one opened pack's trailer hash matches
// every loaded index that references the same pack name.
func (store *Store) verifyPackMatchesIndexes(pack *packFile) error {
	if err := store.ensureCandidates(); err != nil {
		return err
	}
	candidate, ok := store.candidateForPack(pack.name)
	if !ok {
		return fmt.Errorf("objectstore/packed: missing index for pack %q", pack.name)
	}
	index, err := store.openIndex(candidate)
	if err != nil {
		return err
	}
	if err := packchecksum.VerifyPackMatchesIdx(pack.data, index.data, store.algo); err != nil {
		return fmt.Errorf("objectstore/packed: pack %q does not match idx %q: %w", pack.name, index.idxName, err)
	}
	return nil
}

// entryMetaAt parses one pack entry header at location.
func (store *Store) entryMetaAt(loc location) (*packFile, entryMeta, error) {
	pack, err := store.openPack(loc.packName)
	if err != nil {
		return nil, entryMeta{}, err
	}
	meta, err := parseEntryMeta(pack, store.algo, loc.offset)
	if err != nil {
		return nil, entryMeta{}, err
	}
	return pack, meta, nil
}
