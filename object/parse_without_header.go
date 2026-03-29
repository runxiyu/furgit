package object

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/blob"
	"codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tag"
	"codeberg.org/lindenii/furgit/object/tree"
	objecttype "codeberg.org/lindenii/furgit/object/type"
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
