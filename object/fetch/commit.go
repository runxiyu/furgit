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
func (fetcher *Fetcher) ExactCommit(id oid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	parsed, err := fetcher.parseObject(id)
	if err != nil {
		return nil, err
	}

	commit, ok := parsed.(*commit.Commit)
	if !ok {
		return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Commit}
	}

	return stored.New(id, commit), nil
}

// PeelToCommit peels tags until it reaches a commit.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) PeelToCommit(id oid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	for {
		obj, err := fetcher.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *commit.Commit:
			return stored.New(id, parsed), nil
		case *tag.Tag:
			id = parsed.TargetID
		default:
			return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Commit}
		}
	}
}

// PeelToCommitID peels tags until it reaches a commit object ID.
func (fetcher *Fetcher) PeelToCommitID(id oid.ObjectID) (oid.ObjectID, error) {
	for {
		ty, _, err := fetcher.Header(id)
		if err != nil {
			return oid.ObjectID{}, err
		}

		switch ty {
		case typ.Commit:
			return id, nil
		case typ.Tag:
			tag, err := fetcher.ExactTag(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			id = tag.Object().TargetID
		case typ.Unknown, typ.Tree, typ.Blob:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.Commit}
		default:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.Commit}
		}
	}
}
