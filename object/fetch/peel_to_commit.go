package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// PeelToCommit peels tags until it reaches a commit.
//
// Labels: Life-Parent.
func (r *Fetcher) PeelToCommit(id objectid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *commit.Commit:
			return stored.New(id, parsed), nil
		case *tag.Tag:
			id = parsed.TargetID
		default:
			return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeCommit}
		}
	}
}
