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
func (r *Fetcher) ExactTree(id oid.ObjectID) (*stored.Stored[*tree.Tree], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	tree, ok := parsed.(*tree.Tree)
	if !ok {
		return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.TypeTree}
	}

	return stored.New(id, tree), nil
}

// PeelToTree peels tags until it reaches a tree or commit. If it reaches a
// commit, it returns the commit's root tree.
//
// Labels: Life-Parent.
func (r *Fetcher) PeelToTree(id oid.ObjectID) (*stored.Stored[*tree.Tree], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *tree.Tree:
			return stored.New(id, parsed), nil
		case *commit.Commit:
			return r.ExactTree(parsed.Tree)
		case *tag.Tag:
			id = parsed.TargetID
		default:
			return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.TypeTree}
		}
	}
}

// PeelToTreeID peels tags until it reaches a tree object ID, or a commit whose
// root tree object ID is then returned.
func (r *Fetcher) PeelToTreeID(id oid.ObjectID) (oid.ObjectID, error) {
	for {
		ty, _, err := r.Header(id)
		if err != nil {
			return oid.ObjectID{}, err
		}

		switch ty {
		case typ.TypeTree:
			return id, nil
		case typ.TypeCommit:
			commit, err := r.ExactCommit(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			return commit.Object().Tree, nil
		case typ.TypeTag:
			tag, err := r.ExactTag(id)
			if err != nil {
				return oid.ObjectID{}, err
			}

			id = tag.Object().TargetID
		case typ.TypeUnknown, typ.TypeBlob:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.TypeTree}
		default:
			return oid.ObjectID{}, &errs.ObjectTypeError{OID: id, Got: ty, Want: typ.TypeTree}
		}
	}
}
