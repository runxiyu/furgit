package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// ExactTag reads, parses, and wraps the tag at id.
func (r *Resolver) ExactTag(id objectid.ObjectID) (*stored.Stored[*object.Tag], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	tag, ok := parsed.(*object.Tag)
	if !ok {
		return nil, fmt.Errorf("object/resolve: expected tag object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, tag), nil
}
