package packed

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// location identifies one object entry in a specific pack file.
type location struct {
	packName string
	offset   uint64
}

// packCandidate describes one discovered pack/index pair.
type packCandidate struct {
	// packName is the .pack basename.
	packName string
	// idxName is the .idx basename.
	idxName string
	// mtime is the pack file modification time for initial ordering.
	mtime int64
}

// ensureCandidates discovers pack/index pairs once.
func (store *Store) ensureCandidates() error {
	store.discoverOnce.Do(func() {
		candidates, err := store.discoverCandidates()
		candidateByPack := make(map[string]packCandidate, len(candidates))
		for _, candidate := range candidates {
			candidateByPack[candidate.packName] = candidate
		}
		store.stateMu.Lock()
		store.candidates = candidates
		store.candidateByPack = candidateByPack
		store.discoverErr = err
		store.stateMu.Unlock()
	})

	store.stateMu.RLock()
	err := store.discoverErr
	store.stateMu.RUnlock()
	return err
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

// touchCandidate moves one candidate to the front of the lookup order.
func (store *Store) touchCandidate(packName string) {
	store.stateMu.Lock()
	defer store.stateMu.Unlock()
	for i := range store.candidates {
		if store.candidates[i].packName != packName {
			continue
		}
		if i == 0 {
			return
		}
		candidate := store.candidates[i]
		copy(store.candidates[1:i+1], store.candidates[0:i])
		store.candidates[0] = candidate
		return
	}
}
