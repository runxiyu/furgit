package object

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
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
		return ParseBlob(body)
	case objecttype.TypeTree:
		return ParseTree(body, algo)
	case objecttype.TypeCommit:
		return ParseCommit(body, algo)
	case objecttype.TypeTag:
		return ParseTag(body, algo)
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, fmt.Errorf("object: unsupported object type %d", ty)
	default:
		return nil, fmt.Errorf("object: unsupported object type %d", ty)
	}
}
