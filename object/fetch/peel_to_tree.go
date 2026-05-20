package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
	"codeberg.org/lindenii/furgit/object/tree"
	objecttype "codeberg.org/lindenii/furgit/object/type"
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
