package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
)

// ExactCommit reads, parses, and wraps the commit at id.
func (r *Resolver) ExactCommit(id objectid.ObjectID) (*stored.Stored[*object.Commit], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	commit, ok := parsed.(*object.Commit)
	if !ok {
		return nil, fmt.Errorf("object/resolve: expected commit object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, commit), nil
}
