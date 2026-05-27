package fetch

import (
	"lindenii.org/go/furgit/object"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
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
