package packed

import "fmt"

// verifyPackMatchesIndexes checks that one opened pack's trailer hash matches
// every loaded index that references the same pack name.
func (store *Store) verifyPackMatchesIndexes(pack *packFile) error {
	snapshot, err := store.ensureCandidates()
	if err != nil {
		return err
	}

	candidate, ok := snapshot.candidateByPack[pack.name]
	if !ok {
		return fmt.Errorf("objectstore/packed: missing index for pack %q", pack.name)
	}

	index, err := store.openIndex(candidate)
	if err != nil {
		return err
	}

	err = verifyMappedPackMatchesMappedIdx(pack.data, index.data, store.algo)
	if err != nil {
		return fmt.Errorf("objectstore/packed: pack %q does not match idx %q: %w", pack.name, index.idxName, err)
	}

	return nil
}
