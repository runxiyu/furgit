package fetch

import (
	"lindenii.org/go/furgit/errs"
	"lindenii.org/go/furgit/object/commit"
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

// ExactCommit reads, parses, and wraps the commit at id.
//
// Labels: Life-Parent.
func (r *Fetcher) ExactCommit(id oid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	commit, ok := parsed.(*commit.Commit)
	if !ok {
		return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.TypeCommit}
	}

	return stored.New(id, commit), nil
}

// PeelToCommit peels tags until it reaches a commit.
//
// Labels: Life-Parent.
func (r *Fetcher) PeelToCommit(id oid.ObjectID) (*stored.Stored[*commit.Commit], error) {
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
			return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.TypeCommit}
		}
	}
}

// PeelToCommitID peels tags until it reaches a commit object ID.
func (r *Fetcher) PeelToCommitID(id oid.ObjectID) (oid.ObjectID, error) {
	for {
		ty, _, err := r.Header(id)
		if err != nil {
			return oid.ObjectID{}, err
		}

		switch ty {
		case typ.TypeCommit:
			return id, nil
		case typ.TypeTag:
			tag, err := r.ExactTag(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			id = tag.Object().TargetID
		case typ.TypeUnknown, typ.TypeTree, typ.TypeBlob:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.TypeCommit}
		default:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.TypeCommit}
		}
	}
}
