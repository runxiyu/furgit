package packed

import (
	"errors"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// lookup resolves one object ID to its pack location.
func (store *Store) lookup(id objectid.ObjectID) (location, error) {
	var zero location
	if id.Algorithm() != store.algo {
		return zero, errors.New("objectstore/packed: object id algorithm mismatch")
	}

	err := store.ensureCandidates()
	if err != nil {
		return zero, err
	}

	nextPackName := store.firstCandidatePackName()
	for nextPackName != "" {
		candidate, ok := store.candidateForPack(nextPackName)
		if !ok {
			nextPackName = store.firstCandidatePackName()

			continue
		}

		nextPackName = store.nextCandidatePackName(candidate.packName)

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
