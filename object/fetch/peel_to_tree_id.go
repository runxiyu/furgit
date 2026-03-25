package fetch

import (
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// PeelToTreeID peels tags until it reaches a tree object ID, or a commit whose
// root tree object ID is then returned.
func (r *Fetcher) PeelToTreeID(id objectid.ObjectID) (objectid.ObjectID, error) {
	for {
		ty, _, err := r.store.ReadHeader(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}

		switch ty {
		case objecttype.TypeTree:
			return id, nil
		case objecttype.TypeCommit:
			commit, err := r.ExactCommit(id)
			if err != nil {
				return objectid.ObjectID{}, err
			}

			return commit.Object().Tree, nil
		case objecttype.TypeTag:
			tag, err := r.ExactTag(id)
			if err != nil {
				return objectid.ObjectID{}, err
			}

			id = tag.Object().Target
		case objecttype.TypeInvalid,
			objecttype.TypeBlob,
			objecttype.TypeFuture,
			objecttype.TypeOfsDelta,
			objecttype.TypeRefDelta:
			return objectid.ObjectID{}, fmt.Errorf("object/fetch: expected tree-ish object %s, got %v", id, ty)
		default:
			return objectid.ObjectID{}, fmt.Errorf("object/fetch: expected tree-ish object %s, got %v", id, ty)
		}
	}
}
