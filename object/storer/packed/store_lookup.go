package packed

import (
	"errors"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/storer"
)

// lookup resolves one object ID to its pack location.
func (store *Store) lookup(id objectid.ObjectID) (location, error) {
	var zero location
	if id.Algorithm() != store.algo {
		return zero, errors.New("objectstorer/packed: object id algorithm mismatch")
	}

	snapshot, err := store.ensureCandidates()
	if err != nil {
		return zero, err
	}

	loc, ok, err := store.lookupInCandidates(id, snapshot)
	if err != nil {
		return zero, err
	}

	if ok {
		return loc, nil
	}

	if store.refreshPolicy == RefreshPolicyOnMissing { //nolint:nestif
		err = store.Refresh()
		if err != nil {
			return zero, err
		}

		refreshed := store.candidates.Load()
		if refreshed != nil && refreshed != snapshot {
			loc, ok, err = store.lookupInCandidates(id, refreshed)
			if err != nil {
				return zero, err
			}

			if ok {
				return loc, nil
			}
		}
	}

	return zero, objectstorer.ErrObjectNotFound
}

func (store *Store) lookupInCandidates(
	id objectid.ObjectID,
	snapshot *candidateSnapshot,
) (location, bool, error) {
	var zero location

	nextPackName := store.firstCandidatePackName(snapshot)
	for nextPackName != "" {
		candidate, ok := snapshot.candidateByPack[nextPackName]
		if !ok {
			nextPackName = store.firstCandidatePackName(snapshot)

			continue
		}

		nextPackName = store.nextCandidatePackName(candidate.packName, snapshot)

		index, err := store.openIndex(candidate)
		if err != nil {
			return zero, false, err
		}

		offset, ok, err := index.lookup(id)
		if err != nil {
			return zero, false, err
		}

		if ok {
			store.touchCandidate(candidate.packName)

			return location{packName: index.packName, offset: offset}, true, nil
		}
	}

	for _, candidate := range snapshot.candidates {
		index, err := store.openIndex(candidate)
		if err != nil {
			return zero, false, err
		}

		offset, ok, err := index.lookup(id)
		if err != nil {
			return zero, false, err
		}

		if ok {
			store.touchCandidate(candidate.packName)

			return location{packName: index.packName, offset: offset}, true, nil
		}
	}

	return zero, false, nil
}
