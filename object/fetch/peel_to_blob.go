package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/object/blob"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// PeelToBlob peels tags until it reaches a blob.
//
// Labels: Life-Parent.
func (r *Fetcher) PeelToBlob(id objectid.ObjectID) (*stored.Stored[*blob.Blob], error) {
	for {
		obj, err := r.ExactObject(id)
		if err != nil {
			return nil, err
		}

		switch parsed := obj.Object().(type) {
		case *blob.Blob:
			return stored.New(id, parsed), nil
		case *tag.Tag:
			id = parsed.Target
		default:
			return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeBlob}
		}
	}
}
