// Package packed provides packfile reading and associated indexes.
package packed

import (
	"os"
	"sync"
	"sync/atomic"

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
	// refreshPolicy controls automatic candidate refresh on lookup misses.
	refreshPolicy RefreshPolicy

	// candidates stores the latest immutable candidate snapshot.
	candidates atomic.Pointer[candidateSnapshot]
	// refreshMu serializes candidate refresh.
	refreshMu sync.Mutex
	// mruMu guards candidate MRU linked-list state.
	mruMu sync.RWMutex
	// mruHead is the first pack in MRU order.
	mruHead *packCandidateNode
	// mruTail is the last pack in MRU order.
	mruTail *packCandidateNode
	// mruNodeByPack maps pack basename to MRU node.
	mruNodeByPack map[string]*packCandidateNode
	// idxByPack caches opened and parsed indexes by pack basename.
	idxByPack map[string]*idxFile

	// stateMu guards pack cache and close state.
	stateMu sync.RWMutex
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

var _ objectstore.Store = (*Store)(nil)
