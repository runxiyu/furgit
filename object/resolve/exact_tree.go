package resolve

import (
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tree"
)

// ExactTree reads, parses, and wraps the tree at id.
func (r *Resolver) ExactTree(id objectid.ObjectID) (*stored.Stored[*tree.Tree], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	tree, ok := parsed.(*tree.Tree)
	if !ok {
		return nil, fmt.Errorf("object/resolve: expected tree object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, tree), nil
}
