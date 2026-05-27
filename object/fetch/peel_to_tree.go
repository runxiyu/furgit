package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	"lindenii.org/go/furgit/object/commit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/tree"
	objecttype "lindenii.org/go/furgit/object/type"
)

// PeelToTree peels tags until it reaches a tree or commit. If it reaches a
// commit, it returns the commit's root tree.
//
// Labels: Life-Parent.
func (r *Fetcher) PeelToTree(id objectid.ObjectID) (*stored.Stored[*tree.Tree], error) {
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
			return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeTree}
		}
	}
}
