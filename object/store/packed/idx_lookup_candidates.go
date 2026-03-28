package packed

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// packCandidate describes one discovered pack/index pair.
type packCandidate struct {
	// packName is the .pack basename.
	packName string
	// idxName is the .idx basename.
	idxName string
	// mtime is the pack file modification time for initial ordering.
	mtime int64
}

type candidateSnapshot struct {
	candidates      []packCandidate
	candidateByPack map[string]packCandidate
}

// Refresh rescans objects/pack and atomically installs a fresh candidate list
// for future lookups.
//
// Refresh does not invalidate existing readers. Cached pack/index mappings,
// including ones for previously visible candidates, may be retained until
// Close.
func (store *Store) Refresh() error {
	store.refreshMu.Lock()
	defer store.refreshMu.Unlock()

	candidates, err := store.discoverCandidates()
	if err != nil {
		return err
	}

	candidateByPack := make(map[string]packCandidate, len(candidates))
	for _, candidate := range candidates {
		candidateByPack[candidate.packName] = candidate
	}

	store.reconcileMRU(candidates)

	store.candidates.Store(&candidateSnapshot{
		candidates:      candidates,
		candidateByPack: candidateByPack,
	})

	return nil
}

func (store *Store) ensureCandidates() (*candidateSnapshot, error) {
	snapshot := store.candidates.Load()
	if snapshot != nil {
		return snapshot, nil
	}

	err := store.Refresh()
	if err != nil {
		return nil, err
	}

	return store.candidates.Load(), nil
}

// discoverCandidates scans the objects/pack root and returns sorted pack/index
// pairs.
func (store *Store) discoverCandidates() ([]packCandidate, error) {
	dir, err := store.root.Open(".")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	defer func() { _ = dir.Close() }()

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	candidates := make([]packCandidate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".idx") {
			continue
		}

		idxName := entry.Name()
		packName := strings.TrimSuffix(idxName, ".idx") + ".pack"

		packInfo, err := store.root.Stat(packName)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("objectstore/packed: missing pack file for index %q", idxName)
			}

			return nil, err
		}

		candidates = append(candidates, packCandidate{
			packName: packName,
			idxName:  idxName,
			mtime:    packInfo.ModTime().UnixNano(),
		})
	}

	slices.SortFunc(candidates, func(a, b packCandidate) int {
		if a.mtime != b.mtime {
			if a.mtime > b.mtime {
				return -1
			}

			return 1
		}

		return strings.Compare(a.packName, b.packName)
	})

	return candidates, nil
}
