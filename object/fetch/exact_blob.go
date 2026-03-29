package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/object/blob"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	objecttype "codeberg.org/lindenii/furgit/object/type"
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
