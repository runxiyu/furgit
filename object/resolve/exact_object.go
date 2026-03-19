package resolve

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// ExactObject reads, parses, and wraps the object at id without constraining
// its concrete object kind.
func (r *Resolver) ExactObject(id objectid.ObjectID) (*stored.Stored[object.Object], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	return stored.New(id, parsed), nil
}
