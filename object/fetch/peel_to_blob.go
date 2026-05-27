package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	"lindenii.org/go/furgit/object/blob"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	objecttype "lindenii.org/go/furgit/object/type"
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
			id = parsed.TargetID
		default:
			return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeBlob}
		}
	}
}
