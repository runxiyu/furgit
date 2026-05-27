package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	"lindenii.org/go/furgit/object/commit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ExactCommit reads, parses, and wraps the commit at id.
//
// Labels: Life-Parent.
func (r *Fetcher) ExactCommit(id objectid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	commit, ok := parsed.(*commit.Commit)
	if !ok {
		return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeCommit}
	}

	return stored.New(id, commit), nil
}
