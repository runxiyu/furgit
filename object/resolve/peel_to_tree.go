package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// PeelToTree peels tags until it reaches a tree or commit. If it reaches a
// commit, it returns the commit's root tree.
func (r *Resolver) PeelToTree(id objectid.ObjectID) (*stored.Stored[*object.Tree], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *object.Tree:
			return stored.New(id, parsed), nil
		case *object.Commit:
			return r.ExactTree(parsed.Tree)
		case *object.Tag:
			id = parsed.Target
		default:
			return nil, fmt.Errorf("object/resolve: expected tree-ish object %s, got %v", id, parsed.ObjectType())
		}
	}
}
