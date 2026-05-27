package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

// PeelToCommitID peels tags until it reaches a commit object ID.
func (r *Fetcher) PeelToCommitID(id objectid.ObjectID) (objectid.ObjectID, error) {
	for {
		ty, _, err := r.Header(id)
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

			id = tag.Object().TargetID
		case objecttype.TypeInvalid,
			objecttype.TypeTree,
			objecttype.TypeBlob,
			objecttype.TypeFuture,
			objecttype.TypeOfsDelta,
			objecttype.TypeRefDelta:
			return objectid.ObjectID{}, &giterrors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
		default:
			return objectid.ObjectID{}, &giterrors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
		}
	}
}
