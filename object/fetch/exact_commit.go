package fetch

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
)

// ExactCommit reads, parses, and wraps the commit at id.
func (r *Fetcher) ExactCommit(id objectid.ObjectID) (*stored.Stored[*commit.Commit], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	commit, ok := parsed.(*commit.Commit)
	if !ok {
		return nil, fmt.Errorf("object/fetch: expected commit object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, commit), nil
}
