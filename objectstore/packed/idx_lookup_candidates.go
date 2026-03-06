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

// ensureCandidates discovers pack/index pairs once.
func (store *Store) ensureCandidates() error {
	store.discoverOnce.Do(func() {
		candidates, err := store.discoverCandidates()
		candidateByPack := make(map[string]packCandidate, len(candidates))
		nodeByPack := make(map[string]*packCandidateNode, len(candidates))

		var (
			head *packCandidateNode
			tail *packCandidateNode
		)

		for _, candidate := range candidates {
			node := &packCandidateNode{
				candidate: candidate,
				prev:      tail,
			}
			if tail != nil {
				tail.next = node
			}

			if head == nil {
				head = node
			}

			tail = node
			candidateByPack[candidate.packName] = candidate
			nodeByPack[candidate.packName] = node
		}

		store.candidatesMu.Lock()
		store.candidateHead = head
		store.candidateTail = tail
		store.candidateByPack = candidateByPack
		store.candidateNodeByPack = nodeByPack
		store.discoverErr = err
		store.candidatesMu.Unlock()
	})

	store.candidatesMu.RLock()
	err := store.discoverErr
	store.candidatesMu.RUnlock()

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
