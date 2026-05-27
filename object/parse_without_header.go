package object

import (
	"fmt"

	"lindenii.org/go/furgit/object/blob"
	"lindenii.org/go/furgit/object/commit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/tree"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ParseWithoutHeader parses a typed object body.
//
//nolint:ireturn
func ParseWithoutHeader(ty objecttype.Type, body []byte, algo objectid.Algorithm) (Object, error) {
	switch ty {
	case objecttype.TypeBlob:
		return blob.Parse(body)
	case objecttype.TypeTree:
		return tree.Parse(body, algo)
	case objecttype.TypeCommit:
		return commit.Parse(body, algo)
	case objecttype.TypeTag:
		return tag.Parse(body, algo)
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, fmt.Errorf("object: unsupported object type %d", ty)
	default:
		return nil, fmt.Errorf("object: unsupported object type %d", ty)
	}
}
