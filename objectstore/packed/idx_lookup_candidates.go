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

// packCandidateNode is one node in the candidate MRU order list.
type packCandidateNode struct {
	candidate packCandidate
	prev      *packCandidateNode
	next      *packCandidateNode
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

// touchCandidate moves one candidate to the front of the lookup order.
// This is done on a best-effort basis.
func (store *Store) touchCandidate(packName string) {
	if !store.candidatesMu.TryLock() {
		return
	}
	defer store.candidatesMu.Unlock()

	node := store.candidateNodeByPack[packName]
	if node == nil || node == store.candidateHead {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	}

	if store.candidateTail == node {
		store.candidateTail = node.prev
	}

	node.prev = nil

	node.next = store.candidateHead
	if store.candidateHead != nil {
		store.candidateHead.prev = node
	}

	store.candidateHead = node
	if store.candidateTail == nil {
		store.candidateTail = node
	}
}

// firstCandidatePackName returns the current head pack name, or "" when none
// are available.
func (store *Store) firstCandidatePackName() string {
	store.candidatesMu.RLock()
	defer store.candidatesMu.RUnlock()

	if store.candidateHead == nil {
		return ""
	}

	return store.candidateHead.candidate.packName
}

// nextCandidatePackName returns the pack name after currentPack in current MRU
// order, or "" at end / when currentPack is not present.
func (store *Store) nextCandidatePackName(currentPack string) string {
	store.candidatesMu.RLock()
	defer store.candidatesMu.RUnlock()

	node := store.candidateNodeByPack[currentPack]
	if node == nil || node.next == nil {
		return ""
	}

	return node.next.candidate.packName
}
