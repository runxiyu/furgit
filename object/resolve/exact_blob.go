package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/blob"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
)

// ExactBlob reads, parses, and wraps the blob at id.
func (r *Resolver) ExactBlob(id objectid.ObjectID) (*stored.Stored[*blob.Blob], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	blob, ok := parsed.(*blob.Blob)
	if !ok {
		return nil, fmt.Errorf("object/resolve: expected blob object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, blob), nil
}
