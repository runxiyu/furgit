package fetch

import (
	giterrors "lindenii.org/go/furgit/errors"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ExactTag reads, parses, and wraps the tag at id.
//
// Labels: Life-Parent.
func (r *Fetcher) ExactTag(id objectid.ObjectID) (*stored.Stored[*tag.Tag], error) {
	parsed, err := r.parseObject(id)
	if err != nil {
		return nil, err
	}

	tag, ok := parsed.(*tag.Tag)
	if !ok {
		return nil, &giterrors.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: objecttype.TypeTag}
	}

	return stored.New(id, tag), nil
}
