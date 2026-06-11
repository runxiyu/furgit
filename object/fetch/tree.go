package fetch

import (
	"lindenii.org/go/furgit/errs"
	"lindenii.org/go/furgit/object/commit"
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/typ"
)

// ExactTree reads, parses, and wraps the tree at id.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) ExactTree(id oid.ObjectID) (*stored.Stored[*tree.Tree], error) {
	parsed, err := fetcher.parseObject(id)
	if err != nil {
		return nil, err
	}

	tree, ok := parsed.(*tree.Tree)
	if !ok {
		return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Tree}
	}

	return stored.New(id, tree), nil
}

// PeelToTree peels tags until it reaches a tree or commit. If it reaches a
// commit, it returns the commit's root tree.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) PeelToTree(id oid.ObjectID) (*stored.Stored[*tree.Tree], error) {
	for {
		obj, err := fetcher.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *tree.Tree:
			return stored.New(id, parsed), nil
		case *commit.Commit:
			return fetcher.ExactTree(parsed.Tree)
		case *tag.Tag:
			id = parsed.TargetID
		default:
			return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Tree}
		}
	}
}

// PeelToTreeID peels tags until it reaches a tree object ID, or a commit whose
// root tree object ID is then returned.
func (fetcher *Fetcher) PeelToTreeID(id oid.ObjectID) (oid.ObjectID, error) {
	for {
		ty, _, err := fetcher.Header(id)
		if err != nil {
			return oid.ObjectID{}, err
		}

		switch ty {
		case typ.Tree:
			return id, nil
		case typ.Commit:
			commit, err := fetcher.ExactCommit(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			return commit.Object().Tree, nil
		case typ.Tag:
			tag, err := fetcher.ExactTag(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			id = tag.Object().TargetID
		case typ.Unknown, typ.Blob:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.Tree}
		default:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.Tree}
		}
	}
}
