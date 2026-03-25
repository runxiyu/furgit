package fetch

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
	"codeberg.org/lindenii/furgit/object/tree"
)

// PeelToTree peels tags until it reaches a tree or commit. If it reaches a
// commit, it returns the commit's root tree.
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
			id = parsed.Target
		default:
			return nil, fmt.Errorf("object/fetch: expected tree-ish object %s, got %v", id, parsed.ObjectType())
		}
	}
}
