package fetch

import (
	"lindenii.org/go/furgit/errs"
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

// ExactTag reads, parses, and wraps the tag at id.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) ExactTag(id oid.ObjectID) (*stored.Stored[*tag.Tag], error) {
	parsed, err := fetcher.parseObject(id)
	if err != nil {
		return nil, err
	}

	tag, ok := parsed.(*tag.Tag)
	if !ok {
		return nil, &errs.ObjectTypeError{OID: id, Got: parsed.ObjectType(), Want: typ.Tag}
	}

	return stored.New(id, tag), nil
}
