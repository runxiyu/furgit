package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	"lindenii.org/go/furgit/object/blob"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ExactBlob reads, parses, and wraps the blob at id.
//
// Labels: Life-Parent.
func (r *Fetcher) ExactBlob(id objectid.ObjectID) (*stored.Stored[*blob.Blob], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	blob, ok := parsed.(*blob.Blob)
	if !ok {
		return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeBlob}
	}

	return stored.New(id, blob), nil
}
