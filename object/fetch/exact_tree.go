package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tree"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ExactTree reads, parses, and wraps the tree at id.
//
// Labels: Life-Parent.
func (r *Fetcher) ExactTree(id objectid.ObjectID) (*stored.Stored[*tree.Tree], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	tree, ok := parsed.(*tree.Tree)
	if !ok {
		return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeTree}
	}

	return stored.New(id, tree), nil
}
