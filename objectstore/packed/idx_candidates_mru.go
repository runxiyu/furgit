package packed

// packCandidateNode is one node in the candidate MRU order list.
type packCandidateNode struct {
	packName string
	prev     *packCandidateNode
	next     *packCandidateNode
}

func (store *Store) reconcileMRU(candidates []packCandidate) {
	store.mruMu.Lock()
	defer store.mruMu.Unlock()

	if store.mruNodeByPack == nil {
		store.mruNodeByPack = make(map[string]*packCandidateNode, len(candidates))
	}

	present := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		present[candidate.packName] = struct{}{}
	}

	ordered := make([]string, 0, len(candidates))

	for node := store.mruHead; node != nil; node = node.next {
		if _, ok := present[node.packName]; !ok {
			continue
		}

		ordered = append(ordered, node.packName)
		delete(present, node.packName)
	}

	for _, candidate := range candidates {
		if _, ok := present[candidate.packName]; !ok {
			continue
		}

		ordered = append(ordered, candidate.packName)
		delete(present, candidate.packName)
	}

	store.mruHead = nil
	store.mruTail = nil
	store.mruNodeByPack = make(map[string]*packCandidateNode, len(ordered))

	for _, packName := range ordered {
		node := &packCandidateNode{
			packName: packName,
			prev:     store.mruTail,
		}
		if store.mruTail != nil {
			store.mruTail.next = node
		}

		if store.mruHead == nil {
			store.mruHead = node
		}

		store.mruTail = node
		store.mruNodeByPack[packName] = node
	}
}

// touchCandidate moves one candidate to the front of the lookup order.
// This is done on a best-effort basis.
func (store *Store) touchCandidate(packName string) {
	if !store.mruMu.TryLock() {
		return
	}
	defer store.mruMu.Unlock()

	node := store.mruNodeByPack[packName]
	if node == nil || node == store.mruHead {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	}

	if store.mruTail == node {
		store.mruTail = node.prev
	}

	node.prev = nil
	node.next = store.mruHead
	if store.mruHead != nil {
		store.mruHead.prev = node
	}

	store.mruHead = node
	if store.mruTail == nil {
		store.mruTail = node
	}
}

// firstCandidatePackName returns the current head pack name, or "" when none
// are available.
func (store *Store) firstCandidatePackName(snapshot *candidateSnapshot) string {
	store.mruMu.RLock()
	defer store.mruMu.RUnlock()

	for node := store.mruHead; node != nil; node = node.next {
		if _, ok := snapshot.candidateByPack[node.packName]; ok {
			return node.packName
		}
	}

	return ""
}

// nextCandidatePackName returns the pack name after currentPack in current MRU
// order, or "" at end / when currentPack is not present.
func (store *Store) nextCandidatePackName(currentPack string, snapshot *candidateSnapshot) string {
	store.mruMu.RLock()
	defer store.mruMu.RUnlock()

	node := store.mruNodeByPack[currentPack]
	if node == nil {
		return ""
	}

	for node = node.next; node != nil; node = node.next {
		if _, ok := snapshot.candidateByPack[node.packName]; ok {
			return node.packName
		}
	}

	return ""
}
