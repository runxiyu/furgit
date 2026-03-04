package packed

// packCandidateNode is one node in the candidate MRU order list.
type packCandidateNode struct {
	candidate packCandidate
	prev      *packCandidateNode
	next      *packCandidateNode
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
