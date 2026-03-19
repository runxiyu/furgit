package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// PeelToCommit peels tags until it reaches a commit.
func (r *Resolver) PeelToCommit(id objectid.ObjectID) (*stored.Stored[*object.Commit], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *object.Commit:
			return stored.New(id, parsed), nil
		case *object.Tag:
			id = parsed.Target
		default:
			return nil, fmt.Errorf("object/resolve: expected commit-ish object %s, got %v", id, parsed.ObjectType())
		}
	}
}
