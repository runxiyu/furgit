package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// PeelToBlobID peels tags until it reaches a blob object ID.
func (r *Fetcher) PeelToBlobID(id objectid.ObjectID) (objectid.ObjectID, error) {
	for {
		ty, _, err := r.Header(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}

		switch ty {
		case objecttype.TypeBlob:
			return id, nil
		case objecttype.TypeTag:
			tag, err := r.ExactTag(id)
			if err != nil {
				return objectid.ObjectID{}, err
			}

			id = tag.Object().TargetID
		case objecttype.TypeInvalid,
			objecttype.TypeCommit,
			objecttype.TypeTree,
			objecttype.TypeFuture,
			objecttype.TypeOfsDelta,
			objecttype.TypeRefDelta:
			return objectid.ObjectID{}, &giterrors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeBlob}
		default:
			return objectid.ObjectID{}, &giterrors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeBlob}
		}
	}
}
