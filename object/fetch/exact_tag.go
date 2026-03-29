package fetch

import (
	giterrors "codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
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
