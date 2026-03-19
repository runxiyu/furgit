package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// PeelToBlob peels tags until it reaches a blob.
func (r *Resolver) PeelToBlob(id objectid.ObjectID) (*stored.Stored[*object.Blob], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *object.Blob:
			return stored.New(id, parsed), nil
		case *object.Tag:
			id = parsed.Target
		default:
			return nil, fmt.Errorf("object/resolve: expected blob-ish object %s, got %v", id, parsed.ObjectType())
		}
	}
}
