package fetch

import (
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
)

// ExactObject reads, parses, and wraps the object at id without constraining
// its concrete object kind.
//
// Labels: Life-Parent.
func (r *Fetcher) ExactObject(id objectid.ObjectID) (*stored.Stored[object.Object], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	return stored.New(id, parsed), nil
}
