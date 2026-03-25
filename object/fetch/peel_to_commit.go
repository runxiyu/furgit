package fetch

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
)

// PeelToCommit peels tags until it reaches a commit.
func (r *Fetcher) PeelToCommit(id objectid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *commit.Commit:
			return stored.New(id, parsed), nil
		case *tag.Tag:
			id = parsed.Target
		default:
			return nil, fmt.Errorf("object/fetch: expected commit-ish object %s, got %v", id, parsed.ObjectType())
		}
	}
}
