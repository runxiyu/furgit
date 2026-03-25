package object

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/blob"
	"codeberg.org/lindenii/furgit/object/commit"
	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tag"
	"codeberg.org/lindenii/furgit/object/tree"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ParseObjectWithHeader parses a loose object in "type size\x00body" format.
//
//nolint:ireturn
func ParseObjectWithHeader(raw []byte, algo objectid.Algorithm) (Object, error) {
	ty, size, headerLen, ok := objectheader.Parse(raw)
	if !ok {
		return nil, fmt.Errorf("object: malformed object header")
	}

	body := raw[headerLen:]
	if int64(len(body)) != size {
		return nil, fmt.Errorf("object: size mismatch: header says %d bytes, body has %d", size, len(body))
	}

	return ParseObjectWithoutHeader(ty, body, algo)
}

// ParseObjectWithoutHeader parses a typed object body.
//
//nolint:ireturn
func ParseObjectWithoutHeader(ty objecttype.Type, body []byte, algo objectid.Algorithm) (Object, error) {
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
