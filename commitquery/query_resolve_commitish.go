package commitquery

import (
	stderrors "errors"

	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// resolveCommitish peels one commit-ish object ID and resolves the commit.
func (query *query) resolveCommitish(id objectid.ObjectID) (nodeIndex, error) {
	for {
		obj, err := query.fetcher.ExactObject(id)
		if err != nil {
			if stderrors.Is(err, objectstore.ErrObjectNotFound) {
				return 0, &giterrors.ObjectMissingError{OID: id}
			}

			return 0, err
		}

		switch parsed := obj.Object().(type) {
		case *commit.Commit:
			return query.resolveOID(id)
		case *tag.Tag:
			id = parsed.Target
		default:
			return 0, &giterrors.ObjectTypeError{
				OID:  id,
				Got:  parsed.ObjectType(),
				Want: objecttype.TypeCommit,
			}
		}
	}
}
