package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
)

// ExactTree reads, parses, and wraps the tree at id.
func (r *Resolver) ExactTree(id objectid.ObjectID) (*stored.Stored[*object.Tree], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	tree, ok := parsed.(*object.Tree)
	if !ok {
		return nil, fmt.Errorf("object/resolve: expected tree object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, tree), nil
}
