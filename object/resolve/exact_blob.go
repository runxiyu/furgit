package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// ExactBlob reads, parses, and wraps the blob at id.
func (r *Resolver) ExactBlob(id objectid.ObjectID) (*stored.Stored[*object.Blob], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	blob, ok := parsed.(*object.Blob)
	if !ok {
		return nil, fmt.Errorf("object/resolve: expected blob object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, blob), nil
}
