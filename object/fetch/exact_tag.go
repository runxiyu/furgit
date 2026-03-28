package fetch

import (
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
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
		return nil, fmt.Errorf("object/fetch: expected tag object %s, got %v", id, parsed.ObjectType())
	}

	return stored.New(id, tag), nil
}
