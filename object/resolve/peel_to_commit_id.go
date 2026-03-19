package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// PeelToCommitID peels tags until it reaches a commit object ID.
func (r *Resolver) PeelToCommitID(id objectid.ObjectID) (objectid.ObjectID, error) {
	for {
		ty, _, err := r.store.ReadHeader(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}

		switch ty {
		case objecttype.TypeCommit:
			return id, nil
		case objecttype.TypeTag:
			tag, err := r.ExactTag(id)
			if err != nil {
				return objectid.ObjectID{}, err
			}

			id = tag.Object().Target
		default:
			return objectid.ObjectID{}, fmt.Errorf("object/resolve: expected commit-ish object %s, got %v", id, ty)
		}
	}
}
