package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tree"
	objecttype "lindenii.org/go/furgit/object/type"
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
