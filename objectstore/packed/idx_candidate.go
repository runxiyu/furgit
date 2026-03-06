package packed

// candidateForPack returns one discovered candidate for a pack basename.
func (store *Store) candidateForPack(packName string) (packCandidate, bool) {
	store.candidatesMu.RLock()
	candidate, ok := store.candidateByPack[packName]
	store.candidatesMu.RUnlock()

	return candidate, ok
}
